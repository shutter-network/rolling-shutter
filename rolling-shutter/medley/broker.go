package medley

import (
	"github.com/rs/zerolog/log"
)

// Broker allows to distribute a value to multiple
// receiving channels, so that all channels receive the same value.
// This is different to the Fan-Out pattern where multiple channels
// are listening for a send on one channel but only the first channel
// to select the receive operation receives this value.
//
// The Broker allows to set whether the internal send operation is
// blocking or non blocking. In the blocking case, all receiving
// (subscribed) channels have to successively select the receive in order
// for the broker to continue operation. So if one receiver blocks,
// everything else will block as well.
// In the non-blocking case, a subscribed channel that is blocking
// while the published value is sent will miss the value (depending
// on the channels buffer).
//
// One has to be careful in when to select which case.
// The blocking behavior can be partially subverted by increasing
// the subscribers buffer size - but if a subscriber is blocked forever,
// this will only work until this subscribers buffer is full.
//
// The Publish channel field is the means to send a value to the subscribers.
// If the Publish channel is closed, the broker will stop it's background loop.
type Broker[T any] struct {
	Publish         chan T
	stop            chan struct{}
	subscribe       chan chan T
	unsubscribe     chan chan T
	nonBlockingSend bool
}

func StartNewBroker[T any](nonBlocking bool) *Broker[T] {
	b := &Broker[T]{
		Publish:         make(chan T, 1),
		stop:            make(chan struct{}),
		subscribe:       make(chan chan T),
		unsubscribe:     make(chan chan T),
		nonBlockingSend: nonBlocking,
	}

	subscriptions := map[chan T]struct{}{}
	// this is started in the background,
	// without being handled by an errorgroup.
	// this goroutine could be hanging in the backround
	// indefinitely.
	// This is why the sender to the Publish channel has to
	// handle an eventual context done.
	go func() {
		publish := b.Publish
		for {
			select {
			case <-b.stop:
				for subscribeChan := range subscriptions {
					b.removeSubscriber(subscriptions, subscribeChan)
				}
				return
			case subscribeChan := <-b.subscribe:
				subscriptions[subscribeChan] = struct{}{}
			case subscribeChan := <-b.unsubscribe:
				b.removeSubscriber(subscriptions, subscribeChan)
			case val, ok := <-publish:
				if !ok {
					// the run loop will be closed when the publish channel is closed
					// that way we still send out all buffered values put in Publish
					// before we close
					close(b.stop)
					publish = nil
					continue
				}
				b.sendToSubscribers(subscriptions, val)
			}
		}
	}()
	return b
}

func (b *Broker[T]) removeSubscriber(subscriptions map[chan T]struct{}, subscribeChan chan T) {
	delete(subscriptions, subscribeChan)
	close(subscribeChan)
	log.Debug().Int("num-subscribers", len(subscriptions)).Msg("Unsubscribed")
}

func (b *Broker[T]) sendToSubscribers(subscriptions map[chan T]struct{}, value T) {
	if b.nonBlockingSend {
		// use non-blocking send to protect the broker
		// and calls to Publish()
		// receivers could miss published values when they are currently busy
		// -> higher non-zero 'bufferSizeSubscriber' values could circumvent this somewhat
		for msgCh := range subscriptions {
			select {
			case msgCh <- value:
			default:
			}
		}
	} else {
		// use blocking send to make sure all subscribers
		// got the value,
		// could block the whole broker (+ successive Publish() calls) if a receiver
		// is currently busy
		for msgCh := range subscriptions {
			msgCh <- value
		}
	}
}

// Subscribe creates and registers a new channel to
// receive published values.
//
// The returned channel will receive the published value
// as soon as the previous registered subscribers in line
// received the value (in the blockingSend case),
// or as soon as the value is published AND the subscriber
// is currently waiting on the receive operation
// (in the non blockingSend case).
//
// The channel will get closed by the Broker,
// either when the Broker's loop is stopped,
// or when the listener asks the Broker to
// unsubscribe the channel via the Unsubscribe() method.
func (b *Broker[T]) Subscribe(chanSize int) chan T {
	channel := make(chan T, chanSize)
	select {
	case <-b.stop:
		// Don't block the Subscribe() method when the broker
		// is stopped.
		// To comply with the interface, return a closed channel.
		close(channel)
	case b.subscribe <- channel:
	}
	return channel
}

// Unsubscribe deregisters the subscribed channel
// from the internal send operation and close the
// subscribed channel.
// This is a noop when the loop is not running.
func (b *Broker[T]) Unsubscribe(channel chan T) {
	select {
	case <-b.stop:
		// Don't block the Unsubscribe() method when the broker
		// is stopped

		// Contrary to the Subscribe, this is a noop and the channel
		// is not closed here.
		// This is because there could be a race (close of closed channel):
		// The observer doesn't know when the broker is stopped
		// so he could call unsubscribe just after the broker
		// being stopped and the channel being closed.

		// If the passed in channel IS a subscription channel instantiated
		// by the broker, it would be closed during the shutdown
		// of the broker anyways.

		// Calling unsubscribe with a channel not instantiated by the
		// is discouraged!
		return
	case b.unsubscribe <- channel:
		return
	}
}
