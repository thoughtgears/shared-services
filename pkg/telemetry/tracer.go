package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func (o *Otel) InitTracer(ctx context.Context) func(context.Context) error {
	resources, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(o.ServiceName),
			semconv.OTelScopeName(semconv.TelemetrySDKLanguageGo.Value.AsString()),
			semconv.OTelScopeVersion("1.24.2"),
		),
	)
	if err != nil {
		log.Printf("Could not set resources: %v", err)
	}

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)

	return exporter.Shutdown
}
