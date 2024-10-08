package tracing

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/khulnasoft/kengine/v2"
	"github.com/khulnasoft/kengine/v2/modules/kenginehttp"
)

const (
	webEngineName                  = "Kengine"
	defaultSpanName                = "handler"
	nextCallCtxKey  kengine.CtxKey = "nextCall"
)

// nextCall store the next handler, and the error value return on calling it (if any)
type nextCall struct {
	next kenginehttp.Handler
	err  error
}

// openTelemetryWrapper is responsible for the tracing injection, extraction and propagation.
type openTelemetryWrapper struct {
	propagators propagation.TextMapPropagator

	handler http.Handler

	spanName string
}

// newOpenTelemetryWrapper is responsible for the openTelemetryWrapper initialization using provided configuration.
func newOpenTelemetryWrapper(
	ctx context.Context,
	spanName string,
) (openTelemetryWrapper, error) {
	if spanName == "" {
		spanName = defaultSpanName
	}

	ot := openTelemetryWrapper{
		spanName: spanName,
	}

	version, _ := kengine.Version()
	res, err := ot.newResource(webEngineName, version)
	if err != nil {
		return ot, fmt.Errorf("creating resource error: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return ot, fmt.Errorf("creating trace exporter error: %w", err)
	}

	ot.propagators = autoprop.NewTextMapPropagator()

	tracerProvider := globalTracerProvider.getTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	ot.handler = otelhttp.NewHandler(http.HandlerFunc(ot.serveHTTP),
		ot.spanName,
		otelhttp.WithTracerProvider(tracerProvider),
		otelhttp.WithPropagators(ot.propagators),
		otelhttp.WithSpanNameFormatter(ot.spanNameFormatter),
	)

	return ot, nil
}

// serveHTTP injects a tracing context and call the next handler.
func (ot *openTelemetryWrapper) serveHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ot.propagators.Inject(ctx, propagation.HeaderCarrier(r.Header))
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		traceID := spanCtx.TraceID().String()
		// Add a trace_id placeholder, accessible via `{http.vars.trace_id}`.
		kenginehttp.SetVar(ctx, "trace_id", traceID)
		// Add the trace id to the log fields for the request.
		if extra, ok := ctx.Value(kenginehttp.ExtraLogFieldsCtxKey).(*kenginehttp.ExtraLogFields); ok {
			extra.Add(zap.String("traceID", traceID))
		}
	}
	next := ctx.Value(nextCallCtxKey).(*nextCall)
	next.err = next.next.ServeHTTP(w, r)
}

// ServeHTTP propagates call to the by wrapped by `otelhttp` next handler.
func (ot *openTelemetryWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request, next kenginehttp.Handler) error {
	n := &nextCall{
		next: next,
		err:  nil,
	}
	ot.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), nextCallCtxKey, n)))

	return n.err
}

// cleanup flush all remaining data and shutdown a tracerProvider
func (ot *openTelemetryWrapper) cleanup(logger *zap.Logger) error {
	return globalTracerProvider.cleanupTracerProvider(logger)
}

// newResource creates a resource that describe current handler instance and merge it with a default attributes value.
func (ot *openTelemetryWrapper) newResource(
	webEngineName,
	webEngineVersion string,
) (*resource.Resource, error) {
	return resource.Merge(resource.Default(), resource.NewSchemaless(
		semconv.WebEngineName(webEngineName),
		semconv.WebEngineVersion(webEngineVersion),
	))
}

// spanNameFormatter performs the replacement of placeholders in the span name
func (ot *openTelemetryWrapper) spanNameFormatter(operation string, r *http.Request) string {
	return r.Context().Value(kengine.ReplacerCtxKey).(*kengine.Replacer).ReplaceAll(operation, "")
}
