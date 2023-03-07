package tracer

import (
	"context"
	"fmt"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otbridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	// "github.com/jaegertracing/jaeger/examples/hotrod/pkg/log"
	// "github.com/jaegertracing/jaeger/examples/hotrod/pkg/tracing/rpcmetrics"
	// "github.com/jaegertracing/jaeger/pkg/metrics"
)

var once sync.Once

// Init initializes OpenTelemetry SDK and uses OTel-OpenTracing Bridge
// to return an OpenTracing-compatible tracer.
// func Init(serviceName string, exporterType string, metricsFactory metrics.Factory, logger log.Factory) opentracing.Tracer {
func Init(serviceName, exporterType, endpoint string, logger *logrus.Logger) opentracing.Tracer {
	// func Init(serviceName, exporterType, endpoint string, fraction float64, logger *logrus.Logger) opentracing.Tracer {
	once.Do(func() {
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			))
	})

	exp, err := createOtelExporter(exporterType, endpoint)
	if err != nil {
		logger.Fatalf("cannot create exporter, exportType=%v,err=%v", exporterType, err)
	}
	logger.Debug("using " + exporterType + " trace exporter")

	// rpcmetricsObserver := rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		// sdktrace.WithSpanProcessor(rpcmetricsObserver),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
		sdktrace.WithSampler(
			// sdktrace.AlwaysSample(),
			// sdktrace.TraceIDRatioBased(fraction),
			sdktrace.TraceIDRatioBased(1),
		),
	)
	otTracer, _ := otbridge.NewTracerPair(tp.Tracer(""))
	logger.Debugf("created OTEL->OT brige, service-name=%v", serviceName)
	return otTracer
}

func createOtelExporter(exporterType, collectorEndpoint string) (sdktrace.SpanExporter, error) {
	var exporter sdktrace.SpanExporter
	var err error
	switch exporterType {
	case "jaeger":
		exporter, err = jaeger.New(
			jaeger.WithCollectorEndpoint(
				jaeger.WithEndpoint(collectorEndpoint),
				// jaeger.WithUsername(),
				// jaeger.WithPassword(),
			),
			// jaeger.WithAgentEndpoint(
			//     jaeger.WithAgentHost(agentHost),
			//     jaeger.WithAgentPort(agentPort),
			// ),
		)
	case "otlp":
		client := otlptracehttp.NewClient(
			otlptracehttp.WithInsecure(),
		)
		exporter, err = otlptrace.New(context.Background(), client)
	case "stdout":
		exporter, err = stdouttrace.New()
	default:
		return nil, fmt.Errorf("unrecognized exporter type %s", exporterType)
	}
	return exporter, err
}
