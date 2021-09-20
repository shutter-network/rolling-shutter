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

	host                host.Host
	pubSub              *pubsub.PubSub
	TopicGossips        map[string]*TopicGossip
	TopicGossipMessages chan *Message
}

type Config struct {
	ListenAddr     multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	PrivKey        crypto.PrivKey
}

func NewP2P(config Config) *P2P {
	p := P2P{
		Config:              config,
		host:                nil,
		pubSub:              nil,
		TopicGossips:        make(map[string]*TopicGossip),
		TopicGossipMessages: make(chan *Message, 1),
	}
	return &p
}

func (p *P2P) Run(ctx context.Context, topicNames []string) error {
	if err := p.createHost(ctx); err != nil {
		return err
	}
	if err := p.joinTopics(topicNames); err != nil {
		return err
	}

	// listen to gossip on all topics
	errorgroup, errorgroupctx := errgroup.WithContext(ctx)
	errorgroup.Go(func() error {
		return p.listenTopicGossip(errorgroupctx)
	})
	errorgroup.Go(func() error {
		return p.managePeers(errorgroupctx)
	})

	return errorgroup.Wait()
}

func (p *P2P) listenTopicGossip(ctx context.Context) error {
	errorgroup, errorgroupctx := errgroup.WithContext(ctx)
	for _, topicGossip := range p.TopicGossips {
		// we can't use the loop variable directly in the goroutines below, but need to create a
		// new variable
		tg := topicGossip

		errorgroup.Go(func() error {
			return tg.readLoop(errorgroupctx)
		})

		// gather messages from all topics in TopicGossipMessages channel
		errorgroup.Go(func() error {
			for {
				select {
				case msg := <-tg.Messages:
					select {
					case p.TopicGossipMessages <- msg:
					case <-ctx.Done():
						return ctx.Err()
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}
	return errorgroup.Wait()
}

func (p *P2P) Publish(ctx context.Context, topic string, message string) error {
	topicGossip, ok := p.TopicGossips[topic]
	if !ok {
		log.Printf("dropping message to not (yet) subscribed topic %s", topic)
		return nil
	}
	return topicGossip.Publish(ctx, message)
}

func (p *P2P) createHost(ctx context.Context) error {
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

func (p *P2P) joinTopics(topicNames []string) error {
	for _, topicName := range topicNames {
		if err := p.joinTopic(topicName); err != nil {
			return err
		}
	}
	return nil
}

// JoinTopic tries to subscribe to the PubSub topic.
func (p *P2P) joinTopic(topicName string) error {
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
	p.TopicGossips[topicName] = topicGossip
	return nil
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
