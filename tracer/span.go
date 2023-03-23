package tracer

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func Span(c context.Context, tr opentracing.Tracer, name string) (ctx context.Context, span opentracing.Span) {
	span = opentracing.SpanFromContext(c)
	if span == nil {
		return c, nil
	}

	span = tr.StartSpan(name, opentracing.ChildOf(span.Context()))
	ctx = opentracing.ContextWithSpan(c, span)

	return
}

func RPCSpan(c context.Context, tr opentracing.Tracer, name string) (ctx context.Context, span opentracing.Span) {
	span = opentracing.SpanFromContext(c)
	if span == nil {
		return c, nil
	}

	span = tr.StartSpan(name, opentracing.ChildOf(span.Context()))
	ext.SpanKindRPCClient.Set(span)
	ctx = opentracing.ContextWithSpan(c, span)

	return
}
