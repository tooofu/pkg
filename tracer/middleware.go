package tracer

// import (
//     "net/http"

//     "github.com/opentracing-contrib/go-stdlib/nethttp"
//     "github.com/opentracing/opentracing-go"
//     "github.com/sirupsen/logrus"
//     "go.opentelemetry.io/otel/propagation"
// )

// // RequestTracer request tracer middleware
// func RequestTracer(tr opentracing.Tracer, logger *logrus.Logger) http.Handler {
//     mw := nethttp.Middleware()
//     return http.HandlerFunc(requestTracer(mw))
// }

// func requestTracer(next http.Handler) http.Handler {
//     propagator := propagation.Baggage{}
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         carrier := propagation.HeaderCarrier(r.Header)
//         ctx := propagator.Extract(r.Context(), carrier)
//         r = r.WithContext(ctx)
//         next.ServeHTTP(w, r)
//     })
// }
