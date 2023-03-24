// Package service provides methods to run background services in an error group
package service

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Runner interface {
	Go(f func() error)
	StartService(s ...Service) error
}

type Service interface {
	Start(context.Context, Runner) error
}

type runner struct {
	group *errgroup.Group
	ctx   context.Context
}

func (r *runner) Go(f func() error) {
	r.group.Go(f)
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

func Run(ctx context.Context, services ...Service) error {
	group, ctx := errgroup.WithContext(ctx)
	r := runner{group: group, ctx: ctx}
	group.Go(func() error {
		return r.StartService(services...)
	})
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
