// Package service provides methods to run background services in an error group
package service

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
)

type Runner interface {
	// Go starts a background go routine in an error group.
	Go(f func() error)
	// StartService starts the given (sub)services.
	StartService(s ...Service) error
	// Defer registers the given function to be called when shutting down the service.
	Defer(f func())
}

type Service interface {
	Start(context.Context, Runner) error
}

type runner struct {
	group         *errgroup.Group
	ctx           context.Context
	mux           sync.Mutex
	shutdownFuncs []func()
}

func (r *runner) Go(f func() error) {
	r.group.Go(f)
}

func (r *runner) Defer(f func()) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.shutdownFuncs = append(r.shutdownFuncs, f)
}

func (r *runner) StartService(services ...Service) error {
	for _, s := range services {
		err := s.Start(r.ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) runShutdownFuncs() {
	r.mux.Lock()
	defer r.mux.Unlock()

	for i := len(r.shutdownFuncs) - 1; i >= 0; i-- {
		r.shutdownFuncs[i]()
	}
}

func Run(ctx context.Context, services ...Service) error {
	group, deferFn := RunBackground(ctx, services...)
	defer deferFn()
	return group.Wait()
}

// RunBackground runs the services within the context of an errgroup.Group.
// It returns the errgroup.Group as well as the cleanup function to be deferred.
// This is a low-level executor of services, the defer function is expected
// to be called!
func RunBackground(ctx context.Context, services ...Service) (*errgroup.Group, func()) {
	group, ctx := errgroup.WithContext(ctx)
	r := runner{group: group, ctx: ctx}
	group.Go(func() error {
		return r.StartService(services...)
	})
	return group, r.runShutdownFuncs
}

// notifyTermination creates a context that is canceled, when the process receives SIGINT or
// SIGTERM. Similar to signal.NotifyContext, but with additional log output.
func notifyTermination(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	if ctx.Err() == nil {
		go func() {
			select {
			case sig := <-termChan:
				log.Info().Str("signal", sig.String()).Msg("received OS signal, shutting down")
				cancel()
			case <-ctx.Done():
			}
		}()
	}
	return ctx, cancel
}

// RunWithSighandler runs the given services until they fail or the process receives a
// SIGINT/SIGTERM signal.
func RunWithSighandler(ctx context.Context, services ...Service) error {
	ctx, cancel := notifyTermination(ctx)
	defer cancel()
	err := Run(ctx, services...)
	if err == context.Canceled {
		log.Info().Msg("bye")
		return nil
	}
	if errors.Is(err, medley.ErrShutdownRequested) {
		log.Info().Msg("user shut down service")
		log.Info().Msg("bye")
		return nil
	}

	return err
}

type Function struct {
	Func func(ctx context.Context, group Runner) error
}

func (sf Function) Start(ctx context.Context, group Runner) error { //nolint:unparam
	group.Go(func() error {
		return sf.Func(ctx, group)
	})
	return nil
}
