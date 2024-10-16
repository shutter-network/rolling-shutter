package p2p

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/trace"
)

const resourceName = "p2p"

// Inject set cross-cutting concerns from the Context into the carrier.
func InjectTraceContext(ctx context.Context, carrier *p2pmsg.TraceContext) {
	if carrier == nil {
		return
	}
	sc := oteltrace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		log.Debug().Msg("serialized span context is not valid")
		return
	}
	tid := sc.TraceID()
	sid := sc.SpanID()
	flags := sc.TraceFlags()
	carrier.TraceId = tid[:]
	carrier.SpanId = sid[:]
	carrier.TraceFlags = []byte{byte(flags)}
	carrier.TraceState = sc.TraceState().String()
}

// Extract reads cross-cutting concerns from the carrier into a Context.
func ExtractTraceContext(ctx context.Context, carrier *p2pmsg.TraceContext) (context.Context, error) {
	if carrier == nil {
		return ctx, errors.New("carrier is nil")
	}
	cTraceID := carrier.GetTraceId()
	cSpanID := carrier.GetSpanId()
	cFlags := carrier.TraceFlags
	if len(cTraceID) != 16 || len(cSpanID) != 8 || len(cFlags) != 1 {
		return ctx, errors.New("invalid context")
	}

	traceID := (*oteltrace.TraceID)(cTraceID[:16])
	spanID := (*oteltrace.SpanID)(cSpanID[:8])
	flags := oteltrace.TraceFlags((cFlags[0]))
	traceState, err := oteltrace.ParseTraceState(carrier.GetTraceState())
	if err != nil {
		return ctx, errors.Wrap(err, "could not parse trace-state")
	}
	scc := oteltrace.SpanContextConfig{
		TraceID:    *traceID,
		SpanID:     *spanID,
		TraceFlags: flags,
		TraceState: traceState,
	}
	sc := oteltrace.NewSpanContext(scc)
	if !sc.IsValid() {
		return ctx, errors.Wrap(err, "deserialized span context invalid")
	}
	return oteltrace.ContextWithRemoteSpanContext(ctx, sc), nil
}

func newSpanForReceive(
	ctx context.Context,
	p2pnode *P2PNode,
	traceContext *p2pmsg.TraceContext,
	msg *pubsub.Message,
	p2pMsg p2pmsg.Message,
) (context.Context, oteltrace.Span, trace.ErrorWrapper) {
	opName := "receive"

	attrs := []attribute.KeyValue{}
	var netPeer, consumer, producer peer.ID
	var spanKind oteltrace.SpanKind

	netPeer = msg.ReceivedFrom
	producer = msg.GetFrom()
	spanKind = oteltrace.SpanKindConsumer

	h := p2pnode.host
	if h != nil {
		consumer = h.ID()
		peerProtocols, err := h.Peerstore().GetProtocols(netPeer)
		if peerProtocols != nil && err != nil {
			attrs = append(attrs,
				attribute.StringSlice("net.peer.protocols", protocol.ConvertToStrings(peerProtocols)),
			)
		}

		attrs = append(attrs,
			attribute.String("net.peer.connectedness", h.Network().Connectedness(netPeer).String()),
			attribute.String("messaging.producer.connectedness", h.Network().Connectedness(producer).String()),
			attribute.String("messaging.consumer.id", consumer.String()),
		)
	}

	attrs = append(attrs,
		attribute.String("messaging.source.kind", "topic"),
		attribute.String("messaging.source.name", msg.GetTopic()),
		attribute.String("net.peer.name", msg.GetFrom().String()),
	)
	msgName := string(proto.MessageName(p2pMsg).Name())
	attrs = append(attrs,
		attribute.String("messaging.system", "libp2p"),
		attribute.String("messaging.message.type", msgName),
		attribute.String("messaging.operation", opName),
		attribute.String("messaging.producer.id", producer.String()),
		attribute.Int64("shutter.instance.id", int64(p2pMsg.GetInstanceId())),
	)
	log.Debug().Interface("attrs", attrs).Msg("span with attrs")
	spanName := opName + " " + msg.GetTopic()

	if traceContext != nil && trace.IsEnabled() {
		var err error
		ctx, err = ExtractTraceContext(ctx, traceContext)
		if err != nil {
			err = errors.Wrap(err, "unable to extract trace-context, skipped injection in context")
			log.Error().Err(err).Msg("unable to extract trace-context")
		}
	}
	ctx, span := otel.Tracer(resourceName).
		Start(ctx, spanName,
			oteltrace.WithSpanKind(spanKind),
			oteltrace.WithAttributes(attrs...))
	reportError := func(err error) error {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}
	return ctx, span, reportError
}

func newSpanForPublish(
	ctx context.Context,
	p2pnode *P2PNode,
	traceContext *p2pmsg.TraceContext,
	p2pMsg p2pmsg.Message,
) (context.Context, oteltrace.Span, trace.ErrorWrapper) {
	opName := "publish"

	attrs := []attribute.KeyValue{}
	var spanKind oteltrace.SpanKind

	if traceContext != nil && trace.IsEnabled() {
		InjectTraceContext(ctx, traceContext)
	}

	h := p2pnode.host
	if h != nil {
		producer := h.ID()
		attrs = append(attrs, attribute.String("messaging.producer.id", producer.String()))
	}
	spanKind = oteltrace.SpanKindProducer

	attrs = append(attrs,
		attribute.String("messaging.destination.kind", "topic"),
		attribute.String("messaging.destination.name", p2pMsg.Topic()),
	)

	msgName := string(proto.MessageName(p2pMsg).Name())
	attrs = append(attrs,
		attribute.String("messaging.system", "libp2p"),
		attribute.String("messaging.message.type", msgName),
		attribute.String("messaging.operation", opName),
		attribute.Int64("shutter.instance.id", int64(p2pMsg.GetInstanceId())),
	)
	spanName := opName + " " + p2pMsg.Topic()
	ctx, span := otel.Tracer(resourceName).
		Start(ctx, spanName,
			oteltrace.WithSpanKind(spanKind),
			oteltrace.WithAttributes(attrs...))
	reportError := func(err error) error {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}
	return ctx, span, reportError
}
