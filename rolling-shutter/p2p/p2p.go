package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type P2P struct {
	Config Config

	TopicGossips map[string]*TopicGossip
	host         host.Host
	pubSub       *pubsub.PubSub
	errgroup     *errgroup.Group
	errgroupctx  context.Context
	cancel       context.CancelFunc
}

type Config struct {
	ListenAddr     multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	PrivKey        crypto.PrivKey
}

func NewP2P(config Config) *P2P {
	ctx, cancel := context.WithCancel(context.Background())
	errgroup, errgroupctx := errgroup.WithContext(ctx)

	p := P2P{
		Config:       config,
		TopicGossips: make(map[string]*TopicGossip),
		host:         nil,
		pubSub:       nil,
		errgroup:     errgroup,
		errgroupctx:  errgroupctx,
		cancel:       cancel,
	}
	return &p
}

func (p *P2P) Close() error {
	if p.host == nil {
		return nil
	}
	p.cancel()
	_ = p.errgroup.Wait()

	return p.host.Close()
}

func (p *P2P) CreateHost(ctx context.Context) error {
	var err error
	if p.host != nil {
		return errors.New("Cannot create host on p2p with existing host")
	}
	privKey := p.Config.PrivKey
	if privKey == nil {
		privKey, _, err = crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return err
		}
	}

	// create a new libp2p Host
	p.host, err = libp2p.New(ctx, libp2p.ListenAddrs(p.Config.ListenAddr), libp2p.Identity(privKey))
	if err != nil {
		return err
	}
	// print the node's PeerInfo in multiaddr format
	log.Println("libp2p node address:", p.P2PAddress())

	// create a new PubSub service using the GossipSub router
	pubSub, err := pubsub.NewGossipSub(ctx, p.host)
	if err != nil {
		return err
	}
	p.pubSub = pubSub

	return nil
}

// P2PAddress returns the node's PeerInfo in multiaddr format.
func (p *P2P) P2PAddress() string {
	if p.host == nil {
		return "<not connected yet>"
	}
	peerInfo := peer.AddrInfo{
		ID:    p.host.ID(),
		Addrs: p.host.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return fmt.Sprintf("<error: %s>", err)
	}
	return addrs[0].String()
}

func (p *P2P) JoinTopics(ctx context.Context, topicNames []string) error {
	for _, topicName := range topicNames {
		if err := p.JoinTopic(ctx, topicName); err != nil {
			return err
		}
	}
	return nil
}

// JoinTopic tries to subscribe to the PubSub topic.
func (p *P2P) JoinTopic(ctx context.Context, topicName string) error {
	if _, ok := p.TopicGossips[topicName]; ok {
		return errors.New("Cannot join new topic if already joined")
	}
	// join the pubsub topic
	topic, err := p.pubSub.Join(topicName)
	if err != nil {
		return err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return err
	}

	topicGossip := &TopicGossip{
		pubSub:       p.pubSub,
		Topic:        topic,
		subscription: sub,
		Self:         p.host.ID(),
		topicName:    topicName,
		Messages:     make(chan *Message, MessagesBufSize),
	}

	// start reading messages from the subscription in a loop
	p.errgroup.Go(func() error {
		return topicGossip.readLoop(p.errgroupctx)
	})
	p.TopicGossips[topicName] = topicGossip
	return nil
}

func (p *P2P) ConnectToPeers(ctx context.Context) error {
	for _, address := range p.Config.PeerMultiaddrs {
		err := p.ConnectToPeer(ctx, address)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *P2P) ConnectToPeer(ctx context.Context, address multiaddr.Multiaddr) error {
	peerAddr, err := peer.AddrInfoFromP2pAddr(address)
	if err != nil {
		return errors.Wrapf(err, "ConnectToPeer %s", address)
	}
	err = p.host.Connect(ctx, *peerAddr)
	if err != nil {
		return err
	}
	return nil
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
