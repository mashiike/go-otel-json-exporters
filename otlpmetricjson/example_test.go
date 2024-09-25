package otlpmetricjson_test

import (
	"context"
	"fmt"
	"log"

	"github.com/mashiike/go-otel-json-exporters/otlpmetricjson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InstallExportPipeline(ctx context.Context) (func(context.Context) error, error) {
	exporter, err := otlpmetricjson.New(ctx, otlpmetricjson.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating otlp json exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter),
		),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("go-otlp-json-exporter-example"),
			semconv.ServiceVersion("0.0.1"),
		)),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown, nil
}

func Example() {
	ctx := context.Background()
	// Registers a tracer Provider globally.
	shutdown, err := InstallExportPipeline(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	meter := otel.GetMeterProvider().Meter(
		"github.com/instrumentron",
		metric.WithInstrumentationVersion("0.1.0"),
		metric.WithSchemaURL(semconv.SchemaURL),
	)

	counter, _ := meter.Int64Counter("Counter")
	counter.Add(ctx, 1)
	counter.Add(ctx, 2)
}
