package mocksequencer

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/rpc"
)

var (
	l1RPCURL     string
	sequencerURL string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock-sequencer",
		Short: "Run a node that pretends to be a sequencer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mockSequencerMain()
		},
	}
	cmd.PersistentFlags().StringVarP(&l1RPCURL, "l1", "l", "", "layer-1 node JSON RPC endpoint")
	cmd.PersistentFlags().StringVarP(&sequencerURL, "rpc", "r", ":8545", "url of the sequencer's JSON RPC endpoint")
	cmd.MarkPersistentFlagRequired("l1-url")
	cmd.MarkPersistentFlagRequired("rpc")
	return cmd
}

func mockSequencerMain() error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

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

	sequencer := mocksequencer.New(big.NewInt(1), sequencerURL, l1RPCURL)

	err := sequencer.ListenAndServe(
		ctx,
		&rpc.AdminService{},
		&rpc.EthService{},
		&rpc.ShutterService{},
	)
	if err == context.Canceled {
		log.Info().Msg("Bye.")
		return nil
	}
	return err
}
