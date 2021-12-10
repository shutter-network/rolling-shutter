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
func Retry(ctx context.Context, f func() (interface{}, error)) (interface{}, error) {
	res, err := f()
	if err == nil {
		return res, nil
	}

	log.Printf("retrying request after error: %s", err)
	for i := 0; i < numRetries-1; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
		res, err = f()
		if err == nil {
			return res, nil
		}
	}
	return res, err
}
