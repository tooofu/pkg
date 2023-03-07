package tracer

import (
	"fmt"
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

// HandleTracer handle tracer
func HandleTracer(tr opentracing.Tracer, logger *logrus.Logger, handler http.Handler) http.Handler {
	mw := nethttp.Middleware(
		tr,
		handler,
		nethttp.OperationNameFunc(func(r *http.Request) string {
			// return "HTTP " + r.Method + " " + pattern
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),

		// Jaeger SDK was able to accept `jaeger-baggage` header even for requests without am active trace.
		// OTEL Bridge does not support that, so we use Baggage propagator to manually extract the baggage
		// into Context (in otelBaggageExtractor handler below), and once the Bridge creates a Span,
		// we use this SpanObserver to copy OTEL baggage from Context into the Span.

		// nethttp.MWSpanObserver(func(span opentracing.Span, r *http.Request) {
		//     if !tm.copyBaggage {
		//         return
		//     }
		//     bag := baggage.FromContext(r.Context())
		//     for _, m := range bag.Members() {
		//         if b := span.BaggageItem(m.Key()); b == "" {
		//             span.SetBaggageItem(m.Key(), m.Value())
		//         }
		//     }
		// }),
	)
	return mw
}
