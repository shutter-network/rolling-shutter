package p2p

import (
	"context"
	"encoding/json"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

// MessagesBufSize is the number of incoming messages to buffer for each topic.
const MessagesBufSize = 128

// TopicGossip represents a subscription to a single PubSub topic. Messages
// can be published to the topic with TopicGossip.Publish, and received
// messages are pushed to the Messages channel.
type TopicGossip struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *Message

	ps    *pubsub.PubSub
	Topic *pubsub.Topic
	sub   *pubsub.Subscription

	topicName string
	Self      peer.ID
}

// Message gets converted to/from JSON and sent in the body of pubsub messages.
type Message struct {
	Message  string
	SenderID string
}

type P2P struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr

	TopicGossips map[string]*TopicGossip
	host         host.Host
	ps           *pubsub.PubSub
}

func NewP2p() *P2P {
	p := P2P{}
	p.TopicGossips = make(map[string]*TopicGossip)
	p.PeerMultiaddrs = []multiaddr.Multiaddr{}
	return &p
}

func (p *P2P) CreateHost(ctx context.Context, listenAddress multiaddr.Multiaddr) error {
	if p.host != nil {
		return errors.New("Cannot create host on p2p with existing host")
	}
	p.ListenAddress = listenAddress
	// create a new libp2p Host
	h, err := libp2p.New(ctx, libp2p.ListenAddrs(listenAddress))
	if err != nil {
		return err
	}
	p.host = h

	// print the node's PeerInfo in multiaddr format
	peerInfo := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return err
	}
	log.Println("libp2p node address:", addrs[0])

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return err
	}
	p.ps = ps

	return nil
}

func (p *P2P) JoinTopics(ctx context.Context, topicNames []string) error {
	for _, topicName := range topicNames {
		if err := p.JoinTopic(ctx, topicName); err != nil {
			return err
		}
	}
	return nil
}

func (p *P2P) JoinTopic(ctx context.Context, topicName string) error {
	if _, ok := p.TopicGossips[topicName]; ok {
		return errors.New("Cannot join new topic if already joined")
	}
	topicGossip, err := JoinTopic(ctx, p.ps, p.host.ID(), topicName)
	if err != nil {
		return err
	}
	p.TopicGossips[topicName] = topicGossip
	return nil
}

func (p *P2P) ConnectToPeers(ctx context.Context, peerMultiaddrs []multiaddr.Multiaddr) error {
	for _, address := range peerMultiaddrs {
		peerAddr, err := peer.AddrInfoFromP2pAddr(address)
		if err != nil {
			return err
		}
		if err := p.host.Connect(ctx, *peerAddr); err != nil {
			return err
		}
		p.PeerMultiaddrs = append(p.PeerMultiaddrs, address)
	}
	return nil
}

func (p *P2P) ConnectToPeer(ctx context.Context, peerMultiaddr multiaddr.Multiaddr) error {
	return p.ConnectToPeers(ctx, []multiaddr.Multiaddr{peerMultiaddr})
}

func (p *P2P) ConnectedPeers() []peer.ID {
	return p.host.Network().Peers()
}

func (p *P2P) GetMultiaddr() (multiaddr.Multiaddr, error) {
	peerInfo := peer.AddrInfo{
		ID:    p.host.ID(),
		Addrs: p.host.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if len(addrs) != 0 {
		return addrs[0], err
	}
	return nil, err
}

func (p *P2P) HostID() string {
	return p.host.ID().Pretty()
}

// JoinTopic tries to subscribe to the PubSub topic returning a TopicGossip on success.
func JoinTopic(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, topicName string) (*TopicGossip, error) {
	// join the pubsub topic
	topic, err := ps.Join(topicName)
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	topicGossip := &TopicGossip{
		ps:        ps,
		Topic:     topic,
		sub:       sub,
		Self:      selfID,
		topicName: topicName,
		Messages:  make(chan *Message, MessagesBufSize),
	}

	// start reading messages from the subscription in a loop
	go topicGossip.readLoop(ctx)
	return topicGossip, nil
}

// Publish sends a message to the pubsub topic.
func (topicGossip *TopicGossip) Publish(ctx context.Context, message string) error {
	m := Message{
		Message:  message,
		SenderID: topicGossip.Self.Pretty(),
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return topicGossip.Topic.Publish(ctx, msgBytes)
}

func (topicGossip *TopicGossip) ListPeers() []peer.ID {
	return topicGossip.ps.ListPeers(topicGossip.topicName)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (topicGossip *TopicGossip) readLoop(ctx context.Context) {
	for {
		msg, err := topicGossip.sub.Next(ctx)
		if err != nil {
			close(topicGossip.Messages)
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == topicGossip.Self {
			continue
		}
		m := new(Message)
		err = json.Unmarshal(msg.Data, m)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel
		topicGossip.Messages <- m
	}
}
