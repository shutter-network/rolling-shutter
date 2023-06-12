package p2p

import (
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/env"
)

var DefaultOptions []dht.Option

const (
	// resulting protocol string will be: '/<prefix>{/<extension>}/kad/1.0.0'.
	dhtProtocolPrefix           protocol.ID = "/shutter"
	dhtProtocolExtensionStaging protocol.ID = "/staging"
	dhtProtocolExtensionLocal   protocol.ID = "/local"
)

func dhtRoutingOptions(
	environment env.Environment,
	bootstrapPeers ...peer.AddrInfo,
) []dht.Option {
	// options with higher index in the array will overwrite existing ones
	opts := []dht.Option{
		dht.ProtocolPrefix(dhtProtocolPrefix),
	}

	switch environment { //nolint: exhaustive
	case env.EnvironmentStaging:
		opts = append(opts,
			dht.ProtocolExtension(dhtProtocolExtensionStaging),
		)
	case env.EnvironmentLocal:
		opts = append(opts,
			dht.ProtocolExtension(dhtProtocolExtensionLocal),
			// auto mode will not work when the AutoNAT sets the
			// reachability to "private" when we are not reachable
			// over a public IP.
			dht.Mode(dht.ModeServer),
		)
	default:
	}

	if len(bootstrapPeers) > 0 {
		// this overwrites the option set before
		opts = append(opts, dht.BootstrapPeers(bootstrapPeers...))
	}

	return opts
}
