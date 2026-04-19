package tracing

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"go-init/config"

	"gitlab.com/go-init/go-init-common/default/logger"
)

const instrumentationName = "go-init-manager"

type noopSpanExporter struct{}

func (noopSpanExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }
func (noopSpanExporter) Shutdown(context.Context) error                            { return nil }

// Init всегда регистрирует SDK TracerProvider, чтобы в context были валидные trace/span id
// (их подхватывает slog ContextHandler в go-init-common → поля trace_id / traceId в JSON).
// При включённом OTLP спаны дополнительно экспортируются в Tempo.
func Init(ctx context.Context, log *logger.Logger, cfg config.TracingConfig, serviceName, version string) (func(context.Context) error, error) {
	ep := strings.TrimSpace(cfg.OTLPEndpoint)
	if ep == "" {
		ep = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}
	exportOTLP := cfg.Enabled
	if !exportOTLP && ep != "" {
		exportOTLP = true
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("tracing: resource: %w", err)
	}

	var processor sdktrace.SpanProcessor
	switch {
	case exportOTLP && ep == "":
		return nil, fmt.Errorf("tracing: OTLP enabled but otlp_endpoint is empty (set tracing.otlp_endpoint or OTEL_EXPORTER_OTLP_ENDPOINT)")
	case exportOTLP && ep != "":
		u, err := url.Parse(ep)
		if err != nil {
			return nil, fmt.Errorf("tracing: parse otlp endpoint: %w", err)
		}
		if u.Host == "" {
			return nil, fmt.Errorf("tracing: invalid otlp endpoint %q (need host:port or URL)", ep)
		}
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(u.Host)}
		if u.Scheme == "http" || u.Scheme == "" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exp, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("tracing: otlp exporter: %w", err)
		}
		processor = sdktrace.NewBatchSpanProcessor(exp)
		log.Info("OpenTelemetry OTLP export enabled",
			logger.String("endpoint", ep))
	default:
		processor = sdktrace.NewSimpleSpanProcessor(noopSpanExporter{})
		log.Info("OpenTelemetry: trace ids in logs only (OTLP export off or no endpoint)")
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(processor),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// HTTPHandler оборачивает обработчик входящих HTTP-запросов (server span).
func HTTPHandler(path string, next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + path
		}),
	)
}

// Tracer возвращает трассировщик для ручных спанов.
func Tracer() trace.Tracer {
	return otel.Tracer(instrumentationName)
}
