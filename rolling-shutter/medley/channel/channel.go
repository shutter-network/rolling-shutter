package channel

import "context"

func Forward[T any](ctx context.Context,
	receive <-chan T,
	send chan<- T,
) error {
	for {
		select {
		case val, ok := <-receive:
			if !ok {
				return nil
			}
			send <- val
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
