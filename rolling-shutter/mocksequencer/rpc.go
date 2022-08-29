package mocksequencer

import (
	"context"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

func injectHTTPLogger(handler http.Handler) http.Handler {
	logger := log.With().
		Timestamp().
		Str("role", "my-service").
		Logger()

	c := alice.New()
	c = c.Append(hlog.NewHandler(logger))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("finished request")
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))
	return c.Then(handler)
}

func (proc *SequencerProcessor) ListenAndServe(ctx context.Context, rpcServices ...RPCService) error {
	rpcServer := rpc.NewServer()
	for _, service := range rpcServices {
		service.injectProcessor(proc)
		err := rpcServer.RegisterName(service.name(), service)
		if err != nil {
			return errors.Wrap(err, "error while trying to register RPCService")
		}
	}

	mux := http.NewServeMux()
	handler := injectHTTPLogger(rpcServer)
	mux.Handle("/", handler)

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

func stringToAddress(addr string) (common.Address, error) {
	if !common.IsHexAddress(addr) {
		var a common.Address
		return a, errors.New("not a valid ethereum address")
	}
	return common.HexToAddress(addr), nil
}
