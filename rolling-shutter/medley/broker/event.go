package broker

import "time"

func NewEvent[T any](value T) *Event[T] {
	return &Event[T]{
		Value: value,
		Time:  time.Now(),
	}
}

type Event[T any] struct {
	Value T
	Time  time.Time
}

func (e *Event[T]) IsRecent(delta time.Duration) bool {
	return e.Time.After(time.Now().Add(-delta))
}
