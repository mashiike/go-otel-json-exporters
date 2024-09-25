package otlptracejson_test

import (
	"context"
	"fmt"
	"log"

	"github.com/mashiike/go-otel-json-exporters/otlptracejson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func InstallExportPipeline(ctx context.Context) (func(context.Context) error, error) {
	exporter, err := otlptracejson.New(ctx, otlptracejson.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating otlp json exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("go-otlp-json-exporter-example"),
			semconv.ServiceVersion("0.0.1"),
		)),
	)
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
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

	tracer := otel.GetTracerProvider().Tracer(
		"github.com/instrumentron",
		trace.WithInstrumentationVersion("0.1.0"),
		trace.WithSchemaURL(semconv.SchemaURL),
	)

	add := func(ctx context.Context, x, y int64) int64 {
		_, span := tracer.Start(ctx, "Addition")
		defer span.End()

		return x + y
	}

	multiply := func(ctx context.Context, x, y int64) int64 {
		_, span := tracer.Start(ctx, "Multiplication")
		defer span.End()

		return x * y
	}

	ctx, span := tracer.Start(ctx, "Calculation")
	defer span.End()
	ans := multiply(ctx, 2, 2)
	ans = multiply(ctx, ans, 10)
	ans = add(ctx, ans, 2)
	log.Println("the answer is", ans)
}
