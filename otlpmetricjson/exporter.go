package otlpmetricjson

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/mashiike/go-otlp-helper/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	colmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"
)

type options struct {
	writer   io.Writer
	enc      *otlp.JSONEncoder
	implOpts []otlpmetrichttp.Option
}

type Option func(*options)

// WithPrettyPrint prettifies the emitted output.
func WithPrettyPrint() Option {
	return func(o *options) {
		o.enc.SetIndent("  ")
	}
}

// WithWriter sets the export stream destination.
func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.writer = w
		o.enc = otlp.NewJSONEncoder(w)
	}
}

// WithTemporalitySelector sets the temporality selector for the exporter
func WithTemporalitySelector(selector metric.TemporalitySelector) Option {
	return func(o *options) {
		o.implOpts = append(o.implOpts, otlpmetrichttp.WithTemporalitySelector(selector))
	}
}

// WithAggregationSelector sets the aggregation selector for the exporter
func WithAggregationSelector(selector metric.AggregationSelector) Option {
	return func(o *options) {
		o.implOpts = append(o.implOpts, otlpmetrichttp.WithAggregationSelector(selector))
	}
}

type Exporter struct {
	impl    *otlpmetrichttp.Exporter
	options options
}

var _ metric.Exporter = &Exporter{}

func New(ctx context.Context, opts ...Option) (*Exporter, error) {
	o := options{
		writer: os.Stdout,
		enc:    otlp.NewJSONEncoder(os.Stdout),
	}
	for _, opt := range opts {
		opt(&o)
	}
	e := &Exporter{
		options: o,
	}
	implOpts := append(o.implOpts, otlpmetrichttp.WithProxy(e.httpTransportProxy))
	impl, err := otlpmetrichttp.New(ctx, implOpts...)
	if err != nil {
		return nil, err
	}
	e.impl = impl
	return e, nil
}

var errExported = errors.New("exported")

func (e *Exporter) httpTransportProxy(req *http.Request) (*url.URL, error) {
	bs, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	var svcReq colmetricpb.ExportMetricsServiceRequest
	if err := proto.Unmarshal(bs, &svcReq); err != nil {
		return nil, err
	}
	rm := svcReq.GetResourceMetrics()
	data := metricpb.MetricsData{
		ResourceMetrics: make([]*metricpb.ResourceMetrics, len(rm)),
	}
	copy(data.ResourceMetrics, rm)
	if err := e.options.enc.Encode(&data); err != nil {
		return nil, err
	}
	if _, err := e.options.writer.Write([]byte("\n")); err != nil {
		return nil, err
	}
	return nil, errExported
}

func (e *Exporter) Temporality(kind metric.InstrumentKind) metricdata.Temporality {
	return e.impl.Temporality(kind)
}
func (e *Exporter) Aggregation(kind metric.InstrumentKind) metric.Aggregation {
	return e.impl.Aggregation(kind)
}

func (e *Exporter) Export(ctx context.Context, rms *metricdata.ResourceMetrics) error {
	err := e.impl.Export(ctx, rms)
	if errors.Is(err, errExported) {
		return nil
	}
	return err
}
func (e *Exporter) ForceFlush(ctx context.Context) error {
	err := e.impl.ForceFlush(ctx)
	if errors.Is(err, errExported) {
		return nil
	}
	return err
}

func (e *Exporter) Shutdown(context.Context) error {
	err := e.impl.Shutdown(context.Background())
	if errors.Is(err, errExported) {
		return nil
	}
	return err
}
