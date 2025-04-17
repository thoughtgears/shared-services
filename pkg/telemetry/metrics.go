package telemetry

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var counter metric.Int64Counter // nolint:unused

func (o *Otel) InitCounter(ctx context.Context) func(context.Context) error {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			resource.Default().SchemaURL(),
			semconv.ServiceName(o.ServiceName),
		),
	)
	if err != nil {
		log.Fatalf("Error creating resource: %v", err)
	}

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Error creating exporter: %s", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(r),
	)

	meter := provider.Meter(fmt.Sprintf("%s/%s", o.DomainName, o.ServiceName))
	counter, err = meter.Int64Counter(fmt.Sprintf("%s/%s/requests", o.DomainName, o.ServiceName))
	if err != nil {
		log.Fatalf("Error creating counter: %s", err)
	}

	return provider.Shutdown
}
