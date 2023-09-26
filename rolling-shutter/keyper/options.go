package keyper

import (
	"reflect"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
)

type Option func(*options) error

type options struct {
	events       []*eventsyncer.EventType
	eventHandler map[reflect.Type]chainobserver.EventHandlerFunc
	trigger      chan broker.Event[*epochkghandler.DecryptionTrigger]
}

func newDefaultOptions() *options {
	ops := &options{
		events:       []*eventsyncer.EventType{},
		eventHandler: map[reflect.Type]chainobserver.EventHandlerFunc{},
		// FIXME what channel to use as default?
		trigger: make(chan broker.Event[*epochkghandler.DecryptionTrigger]),
	}
	// TODO set the defaults
	return ops
}

func HandleEvent[T any](event *eventsyncer.EventType, handler chainobserver.EventHandlerFuncGeneric[T]) Option {
	wrappedHandler := chainobserver.MakeHandler(handler)
	return func(o *options) error {
		o.events = append(o.events, event)
		o.eventHandler[event.Type] = wrappedHandler
		// TODO raise error when same type multiple times
		return nil
	}
}

func DecryptionTrigger(trigger chan broker.Event[*epochkghandler.DecryptionTrigger]) Option {
	return func(o *options) error {
		return nil
	}
}
