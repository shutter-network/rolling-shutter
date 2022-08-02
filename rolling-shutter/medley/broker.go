package medley

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// Broker allows to distribute a value to multiple
// receiving channels, so that all channels receive the same value.
// This is different to the Fan-Out pattern where multiple channels
// are listening for a send on one channel, but only the first channel
// to select the receive operation receives this value.
//
// The Broker allows to set whether the internal send operation is
// blocking or non blocking. In the blocking case, all receiving
// (subscribed) channels have to successively select the receive in order
// for the broker to continue operation. So if one receiver blocks,
// everything else will block as well.
// In the non-blocking case, a subscribed channel that is blocking
// while the published value is sent will miss the value.
//
// One has to be careful in when to select which case.
// The blocking behavior can be partially subverted by increasing
// the subscribers buffer size - but if a subscriber is blocked forever,
// this will only work until this subscribers buffer is full.
type Broker[T any] struct {
	mux                  sync.RWMutex
	stop                 chan struct{}
	publish              chan T
	subscribe            chan chan T
	unsubscribe          chan chan T
	running              bool
	bufferSizeSubscriber int
	nonBlockingSend      bool
}

func NewBroker[T any](buffer int, nonBlocking bool) *Broker[T] {
	return &Broker[T]{
		stop:            make(chan struct{}),
		publish:         make(chan T, 1),
		subscribe:       make(chan chan T),
		unsubscribe:     make(chan chan T),
		nonBlockingSend: nonBlocking,
		running:         false,
	}
}

// Start starts the Broker's internal loop that processes
// the subscription and publish operations.
// initialSubscriptions will be registered
// before the loop is started.
// This is useful when the channels must not miss a published value
// after the start or when the subscribers buffer size
// should diverge from the default value `bufferSizeSubscriber`
// set in the constructor.
func (b *Broker[T]) Start(initialSubscriptions ...chan T) {
	subscriptions := map[chan T]struct{}{}
	for _, sub := range initialSubscriptions {
		log.Debug().Int("num-subscribers", len(subscriptions)).Msg("subscribed initial subscription")
		subscriptions[sub] = struct{}{}
	}
	started := make(chan struct{})
	go func() {
		b.mux.Lock()
		b.running = true
		b.mux.Unlock()
		close(started)
		for {
			select {
			case <-b.stop:
				for msgCh := range subscriptions {
					close(msgCh)
					delete(subscriptions, msgCh)
				}
				b.mux.Lock()
				b.running = false
				b.mux.Unlock()
				return
			case msgCh := <-b.subscribe:
				subscriptions[msgCh] = struct{}{}
			case msgCh := <-b.unsubscribe:
				delete(subscriptions, msgCh)
				close(msgCh)
				log.Debug().Int("num-subscribers", len(subscriptions)).Msg("Unsubscribed")
			case msg := <-b.publish:
				if b.nonBlockingSend {
					// use blocking send to make sure all subscribers
					// got the value,
					// could block the whole broker (+ successive Publish() calls) if a receiver
					// is currently busy
					for msgCh := range subscriptions {
						msgCh <- msg
					}
				} else {
					// use non-blocking send to protect the broker
					// and calls to Publish()
					// receivers could miss published values when they are currently busy
					// -> higher non-zero 'bufferSizeSubscriber' values could circumvent this somewhat
					for msgCh := range subscriptions {
						select {
						case msgCh <- msg:
						default:
						}
					}
				}
			}
		}
	}()
	<-started
}

// Stop will close the internal stop channel
// and will stop the Brokers internal loop
// as soon as the closed stop channel is selected.
// A call to Stop() will also cause the subscribed
// channels to be closed.
func (b *Broker[_]) Stop() {
	close(b.stop)
}

// Subscribe registers the channel to the internal
// send operation. It will receive a published value
// as soon as the other registered subscribers in line
// received the value (in the blockingSend case),
// or as soon as the value is published AND the subscriber
// is currently waiting on the receive operation
// (in the non blockingSend case).
// This is blocking, even if the loop is not running.
func (b *Broker[T]) Subscribe() chan T {
	channel := make(chan T, b.bufferSizeSubscriber)
	b.subscribe <- channel
	return channel
}

// Unsubscribe deregisters the subscribed channel
// from the internal send operation and close the
// subscribed channel.
// This is a noop when the loop is not running.
func (b *Broker[T]) Unsubscribe(channel chan T) {
	b.mux.RLock()
	if !b.running {
		// return directly in case the loop is not running,
		// otherwise this would block forever
		b.mux.RUnlock()
		return
	}
	b.mux.RUnlock()
	for {
		select {
		case <-b.stop:
			return
		case b.unsubscribe <- channel:
			return
		case <-channel:
			// Allow to flush the message channel
		}
	}
}

// Publish distributes the `value` to
// all subscribers as soon as the
// internal loop is not busy anymore.
// Note that this could potentially block
// if the loop is busy, although the internal
// publish channel has a buffer of 1.
func (b *Broker[T]) Publish(value T) {
	// noop when not running
	b.mux.RLock()
	if !b.running {
		// return directly in case the loop is not running,
		// otherwise this would block forever
		b.mux.RUnlock()
		return
	}
	b.mux.RUnlock()
	select {
	case <-b.stop:
		return
	case b.publish <- value:
		return
	}
}
