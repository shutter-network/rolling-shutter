// Package service provides methods to run background services in an error group
package service

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Service interface {
	Start(ctx context.Context, group *errgroup.Group) error
}

func Run(ctx context.Context, services []Service) error {
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		for _, s := range services {
			err := s.Start(ctx, group)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return group.Wait()
}

type ServiceFn struct {
	Fn func(ctx context.Context) error
}

func (sf ServiceFn) Start(ctx context.Context, group *errgroup.Group) error { //nolint:unparam
	group.Go(func() error {
		return sf.Fn(ctx)
	})
	return nil
}
