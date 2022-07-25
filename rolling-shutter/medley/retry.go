package medley

import (
	"context"
	"log"
	"time"
)

const (
	numRetries    = 3
	retryInterval = 2 * time.Second
)

// Retry calls the given function multiple times until it doesn't return an error.
func Retry[T any](ctx context.Context, f func() (T, error)) (T, error) {
	var null T
	res, err := f()
	if err == nil {
		return res, nil
	}

	for i := 0; i < numRetries-1; i++ {
		select {
		case <-ctx.Done():
			return null, ctx.Err()
		case <-time.After(retryInterval):
		}
		log.Printf("retrying request after error: %s", err)
		res, err = f()
		if err == nil {
			return res, nil
		}
	}
	return res, err
}
