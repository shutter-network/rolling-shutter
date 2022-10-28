package mocksequencer

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/rpc"
)

var (
	l1RPCURL     string
	sequencerURL string
	chainID      uint64
	debugPtr     *bool
	adminPtr     *bool
)

type Config struct{}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock-sequencer",
		Short: "Run a node that pretends to be a sequencer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mockSequencerMain()
		},
	}
	debugPtr = cmd.PersistentFlags().Bool("debug", false, "debug mode (higher log verbosity)")
	adminPtr = cmd.PersistentFlags().Bool("admin", true, "expose the 'admin_' RPC namespace methods")
	cmd.PersistentFlags().StringVarP(&l1RPCURL, "l1", "l", "http://localhost:8545", "layer-1 node JSON RPC endpoint")
	cmd.PersistentFlags().StringVarP(&sequencerURL, "rpc", "r", ":8545", "url of the sequencer's JSON RPC endpoint")
	cmd.PersistentFlags().Uint64VarP(&chainID, "chain-id", "c", 4242, "the chain-id")
	return cmd
}

func mockSequencerMain() error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugPtr != nil {
		if *debugPtr {
			zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
	}

	log.Info().Msgf("Starting mock sequencer version %s", shversion.Version())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termChan
		log.Info().Str("signal", sig.String()).Msg("Received signal, shutting down")
		cancel()
	}()

	sequencer := mocksequencer.New(new(big.Int).SetUint64(chainID), sequencerURL, l1RPCURL)

	services := []mocksequencer.RPCService{
		&rpc.EthService{},
		&rpc.ShutterService{},
	}
	if adminPtr != nil {
		if *adminPtr {
			services = append(services, &rpc.AdminService{})
		}
	}
	err := sequencer.ListenAndServe(
		ctx,
		services...,
	)
	if err == context.Canceled {
		log.Info().Msg("Bye.")
		return nil
	}
	return err
}
