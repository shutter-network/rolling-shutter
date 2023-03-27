// Package service provides methods to run background services in an error group
package service

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
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
	group, ctx := errgroup.WithContext(ctx)
	r := runner{group: group, ctx: ctx}
	group.Go(func() error {
		return r.StartService(services...)
	})
	defer r.runShutdownFuncs()
	return group.Wait()
}

type ServiceFn struct {
	Fn func(ctx context.Context) error
}

func (sf ServiceFn) Start(ctx context.Context, group Runner) error { //nolint:unparam
	group.Go(func() error {
		return sf.Fn(ctx)
	})
	return nil
}
