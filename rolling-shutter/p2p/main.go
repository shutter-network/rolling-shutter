package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	multiaddr "github.com/multiformats/go-multiaddr"
)

const (
	topicName = "testTopic"
	message   = "test message"
)

func main() {
	// parse some flags to set our nickname and the room to join
	addressFlag := flag.String("address", "", "address of other node to connect to")
	flag.Parse()
	otherAddress := *addressFlag

	ctx := context.Background()

	// create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New(ctx, libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		panic(err)
	}

	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Println("libp2p node address:", addrs[0])

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	// join the chat room
	topicGossip, err := JoinTopic(ctx, ps, h.ID(), topicName)
	if err != nil {
		panic(err)
	}

	// if a remote peer has been passed on the command line, connect to it
	// and send messages
	if otherAddress != "" {
		addr, err := multiaddr.NewMultiaddr(otherAddress)
		if err != nil {
			panic(err)
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			panic(err)
		}
		if err := h.Connect(ctx, *peer); err != nil {
			panic(err)
		}
		fmt.Println("sending message to", addr)
		fmt.Println(topicGossip.ListPeers())

		for {
			if err := topicGossip.Publish(ctx, message); err != nil {
				panic(err)
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		msg := <-topicGossip.Messages
		fmt.Println(*msg)
	}

	// shut the node down
	if err := h.Close(); err != nil {
		panic(err)
	}
}
