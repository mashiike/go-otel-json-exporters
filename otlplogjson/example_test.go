package otlplogjson_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mashiike/go-otel-json-exporters/otlplogjson"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InstallExportPipeline(ctx context.Context) (*sdklog.LoggerProvider, error) {
	exporter, err := otlplogjson.New(ctx, otlplogjson.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating otlp json exporter: %w", err)
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(exporter),
		),
		sdklog.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("go-otlp-json-exporter-example"),
			semconv.ServiceVersion("0.0.1"),
		)),
	)
	return loggerProvider, nil
}

func Example() {
	ctx := context.Background()
	// Registers a tracer Provider globally.
	loggerProvider, err := InstallExportPipeline(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := loggerProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	logger := loggerProvider.Logger(
		"github.com/instrumentron",
		otellog.WithInstrumentationVersion("0.1.0"),
		otellog.WithSchemaURL(semconv.SchemaURL),
	)
	var record otellog.Record
	record.SetTimestamp(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
	record.SetObservedTimestamp(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
	record.SetSeverity(otellog.SeverityInfo)
	record.SetBody(otellog.StringValue("hello world"))
	logger.Emit(ctx, record)
	//{
	//  "resourceLogs":  [
	//    {
	//    "resource":  {
	//      "attributes":  [
	//      {
	//        "key":  "service.name",
	//        "value":  {
	//        "stringValue":  "go-otlp-json-exporter-example"
	//        }
	//      },
	//      {
	//        "key":  "service.version",
	//        "value":  {
	//        "stringValue":  "0.0.1"
	//        }
	//      }
	//      ]
	//    },
	//    "scopeLogs":  [
	//      {
	//      "scope":  {
	//        "name":  "github.com/instrumentron",
	//        "version":  "0.1.0"
	//      },
	//      "logRecords":  [
	//        {
	//        "timeUnixNano":  "1609459200000000000",
	//        "observedTimeUnixNano":  "1609459200000000000",
	//        "severityNumber":  "SEVERITY_NUMBER_INFO",
	//        "body":  {
	//          "stringValue":  "hello world"
	//        }
	//        }
	//      ],
	//      "schemaUrl":  "https://opentelemetry.io/schemas/1.26.0"
	//      }
	//    ],
	//    "schemaUrl":  "https://opentelemetry.io/schemas/1.26.0"
	//    }
	//  ]
	//}
}
