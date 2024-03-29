package mocksequencer

import (
	"context"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

type RPCService interface {
	Name() string
	InjectProcessor(*Sequencer)
}

func injectHTTPLogger(handler http.Handler) http.Handler {
	logger := log.With().
		Timestamp().
		Str("role", "my-service").
		Logger()

	c := alice.New()
	c = c.Append(hlog.NewHandler(logger))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Trace().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("finished request")
	}))
	//nolint:godox //this is not worth an issue at the moment
	// TODO(ezdac) It would be good to decode the request body already
	// and deduct some domain-specific information about the request,
	// mainly the JSON RPC method.
	// In go this means copying the request body though, because
	// the stream-buffer can only be read once ...
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))
	return c.Then(handler)
}

func (proc *Sequencer) ListenAndServe(ctx context.Context, rpcServices ...RPCService) error {
	rpcServer := rpc.NewServer()
	backgroundError := proc.RunBackgroundTasks(ctx)

	for _, service := range rpcServices {
		service.InjectProcessor(proc)
		err := rpcServer.RegisterName(service.Name(), service)
		if err != nil {
			return errors.Wrap(err, "error while trying to register RPCService")
		}
	}
	return RPCListenAndServe(ctx, rpcServer, proc.URL, backgroundError)
}

func RPCListenAndServe(
	ctx context.Context,
	rpcServer *rpc.Server,
	url string,
	backgroundError <-chan error,
) error {
	mux := http.NewServeMux()
	handler := injectHTTPLogger(rpcServer)
	mux.Handle("/", handler)

	server := &http.Server{Addr: url, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

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
	case err := <-backgroundError:
		// For now, fail the whole server when the background task
		// of the processor fails.
		return err
	case err := <-failed:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
