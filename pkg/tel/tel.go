package tel

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	OTLPEndpoint   string
}

type Telemetry struct {
	Logger   *slog.Logger
	tracer   trace.Tracer
	meter    metric.Meter
	shutdown []func(context.Context) error
}

func New(ctx context.Context, cfg Config) (*Telemetry, error) {
	res := newResource(cfg)

	tp, err := newTracerProvider(ctx, cfg.OTLPEndpoint, res)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tp)

	mp, err := newMeterProvider(ctx, cfg.OTLPEndpoint, res)
	if err != nil {
		return nil, err
	}
	otel.SetMeterProvider(mp)

	lp, err := newLoggerProvider(ctx, cfg.OTLPEndpoint, res)
	if err != nil {
		return nil, err
	}
	global.SetLoggerProvider(lp)

	logger := slog.New(otelslog.NewHandler(cfg.ServiceName, otelslog.WithLoggerProvider(lp)))

	return &Telemetry{
		Logger: logger,
		tracer: otel.Tracer(cfg.ServiceName),
		meter:  otel.Meter(cfg.ServiceName),
		shutdown: []func(context.Context) error{
			tp.Shutdown,
			mp.Shutdown,
			lp.Shutdown,
		},
	}, nil
}

// NewNoopTelemetry returns a Telemetry backed by the default (no-op) global providers.
// Useful for tests and fallback scenarios.
func NewNoopTelemetry() *Telemetry {
	return &Telemetry{
		Logger: slog.Default(),
		tracer: otel.Tracer("noop"),
		meter:  otel.Meter("noop"),
	}
}

func (t *Telemetry) Meter() metric.Meter  { return t.meter }
func (t *Telemetry) Tracer() trace.Tracer { return t.tracer }

//nolint:spancheck
func (t *Telemetry) TraceStart(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName)
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error
	for _, fn := range t.shutdown {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

func newResource(cfg Config) *resource.Resource {
	hostname, _ := os.Hostname()
	return resource.NewSchemaless(
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.HostName(hostname),
	)
}

func newTracerProvider(ctx context.Context, endpoint string, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	), nil
}

func newMeterProvider(ctx context.Context, endpoint string, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
		sdkmetric.WithResource(res),
	), nil
}

func newLoggerProvider(ctx context.Context, endpoint string, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	exp, err := otlploggrpc.New(ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("create log exporter: %w", err)
	}
	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),
		sdklog.WithResource(res),
	), nil
}
