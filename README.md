# go-otel-json-exporters

This repository contains a set of OpenTelemetry exporters that output JSON. The exporters are written in Go and are intended to be used with the OpenTelemetry Go SDK.

## Usage 

```go
package main

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

func main() {
	ctx := context.Background()
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
```

export to stdout as following:

```json
{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {
              "stringValue": "go-otlp-json-exporter-example"
            }
          },
          {
            "key": "service.version",
            "value": {
              "stringValue": "0.0.1"
            }
          }
        ]
      },
      "scopeSpans": [
        {
          "scope": {
            "name": "github.com/instrumentron",
            "version": "0.1.0"
          },
          "spans": [
            {
              "traceId": "kIhUbvcWbS0zGEIndbzBNw==",
              "spanId": "ZdpumOL1uDY=",
              "parentSpanId": "9Z+0Ai5UZf4=",
              "flags": 256,
              "name": "Multiplication",
              "kind": "SPAN_KIND_INTERNAL",
              "startTimeUnixNano": "1727244736674966000",
              "endTimeUnixNano": "1727244736674966500",
              "status": {}
            },
            {
              "traceId": "kIhUbvcWbS0zGEIndbzBNw==",
              "spanId": "RFSMEkYEda4=",
              "parentSpanId": "9Z+0Ai5UZf4=",
              "flags": 256,
              "name": "Multiplication",
              "kind": "SPAN_KIND_INTERNAL",
              "startTimeUnixNano": "1727244736674970000",
              "endTimeUnixNano": "1727244736674970375",
              "status": {}
            },
            {
              "traceId": "kIhUbvcWbS0zGEIndbzBNw==",
              "spanId": "jCiV+gi78B8=",
              "parentSpanId": "9Z+0Ai5UZf4=",
              "flags": 256,
              "name": "Addition",
              "kind": "SPAN_KIND_INTERNAL",
              "startTimeUnixNano": "1727244736674971000",
              "endTimeUnixNano": "1727244736674971458",
              "status": {}
            },
            {
              "traceId": "kIhUbvcWbS0zGEIndbzBNw==",
              "spanId": "9Z+0Ai5UZf4=",
              "flags": 256,
              "name": "Calculation",
              "kind": "SPAN_KIND_INTERNAL",
              "startTimeUnixNano": "1727244736674951000",
              "endTimeUnixNano": "1727244736675025541",
              "status": {}
            }
          ],
          "schemaUrl": "https://opentelemetry.io/schemas/1.26.0"
        }
      ],
      "schemaUrl": "https://opentelemetry.io/schemas/1.26.0"
    }
  ]
}
```

metrics and logs exporters same as trace exporter.
see details in [otlpmetricjson/example_test.go](./otlpmetricjson/example_test.go) and [otlplogjson/example_test.go](./otlplogjson/example_test.go)

## License

MIT
