package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var logger zerolog.Logger
var tracer trace.Tracer

func initTracer() (*tracesdk.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger:14268/api/traces")))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("test-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

type TraceHook struct{}

func (h TraceHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	span := trace.SpanFromContext(e.GetCtx())
	if span.SpanContext().IsValid() {
		e.Str("trace_id", span.SpanContext().TraceID().String())
		e.Str("span_id", span.SpanContext().SpanID().String())
	}
}

func main() {
	logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger = logger.Hook(TraceHook{})

	tp, err := initTracer()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize tracer")
	}
	defer tp.Shutdown(context.Background())

	tracer = tp.Tracer("test-tracer")

	http.HandleFunc("/", handler)
	logger.Info().Msg("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handler")
	defer span.End()

	logger.Info().Ctx(ctx).Msg("Handling request")

	time.Sleep(100 * time.Millisecond) // Simulate some work

	resp, err := makeDownstreamRequest(ctx)
	if err != nil {
		logger.Error().Ctx(ctx).Err(err).Msg("Downstream request failed")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Info().Ctx(ctx).Msg("Request handled successfully")
	_, _ = fmt.Fprint(w, resp)
}

func makeDownstreamRequest(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "downstream-request")
	defer span.End()

	logger.Info().Ctx(ctx).Msg("Making downstream request")

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	logger.Info().Ctx(ctx).Msg("Downstream request completed")
	return string(body), nil
}
