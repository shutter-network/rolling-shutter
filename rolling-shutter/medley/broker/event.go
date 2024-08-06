package broker

import (
	"errors"
	"sync"
	"time"
)

var ErrResultAlreadySet = errors.New("result has already been set")

func NewEvent[T any](value T) *Event[T] {
	return &Event[T]{
		Value:   value,
		Time:    time.Now(),
		resultC: make(chan *Result, 1),
		resMux:  &sync.Mutex{},
	}
}

type Result struct {
	OK    bool
	Error error
	Time  time.Time
	Where string
}

func (r *Result) Set(err error) {
	r.OK = bool(err == nil)
	r.Error = err
	r.Time = time.Now()
}

type Event[T any] struct {
	Value T
	Time  time.Time

	resMux  *sync.Mutex
	result  *Result
	resultC chan *Result
}

func (e *Event[T]) Result() <-chan *Result {
	return e.resultC
}

func (e *Event[T]) setResult(err error) error {
	e.resMux.Lock()
	defer e.resMux.Unlock()
	if e.result != nil {
		return ErrResultAlreadySet
	}
	e.result = &Result{}
	e.result.Set(err)
	return nil
}

func (e *Event[T]) SetResult(err error) error {
	if err := e.setResult(err); err != nil {
		return err
	}
	// capacity 1 and empty channel,
	// so this can never block
	e.resultC <- e.result
	return nil
}

func (e *Event[T]) IsRecent(delta time.Duration) bool {
	return e.Time.After(time.Now().Add(-delta))
}
