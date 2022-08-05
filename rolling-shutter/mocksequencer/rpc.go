package mocksequencer

import (
	"context"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
)

func (proc *SequencerProcessor) ListenAndServe(ctx context.Context, rpcServices ...RPCService) error {
	rpcServer := rpc.NewServer()
	for _, service := range rpcServices {
		service.injectProcessor(proc)
		rpcServer.RegisterName(service.name(), service)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", http.HandlerFunc(rpcServer.ServeHTTP))

	server := &http.Server{Addr: ":8545", Handler: mux}

	failed := make(chan error)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			failed <- err
		}
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("shutting down server failed")
		}
		cancel()
	}()

	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

type RPCService interface {
	name() string
	injectProcessor(*SequencerProcessor)
}

func stringToAddress(addr string) common.Address {
	return common.HexToAddress(addr)
}
