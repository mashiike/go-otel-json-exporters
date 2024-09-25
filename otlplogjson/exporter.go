package otlplogjson

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	collogpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type options struct {
	writer    io.Writer
	marshaler protojson.MarshalOptions
	implOpts  []otlploghttp.Option
}

type Option func(*options)

// WithPrettyPrint prettifies the emitted output.
func WithPrettyPrint() Option {
	return func(o *options) {
		o.marshaler.Multiline = true
		o.marshaler.Indent = "  "
	}
}

// WithWriter sets the export stream destination.
func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.writer = w
	}
}

type Exporter struct {
	impl    *otlploghttp.Exporter
	options options
}

var _ log.Exporter = &Exporter{}

func New(ctx context.Context, opts ...Option) (*Exporter, error) {
	o := options{
		writer:    os.Stdout,
		marshaler: protojson.MarshalOptions{},
	}
	for _, opt := range opts {
		opt(&o)
	}
	e := &Exporter{
		options: o,
	}
	implOpts := append(o.implOpts, otlploghttp.WithProxy(e.httpTransportProxy))
	impl, err := otlploghttp.New(ctx, implOpts...)
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
	var svcReq collogpb.ExportLogsServiceRequest
	if err := proto.Unmarshal(bs, &svcReq); err != nil {
		return nil, err
	}
	data := logpb.LogsData{
		ResourceLogs: make([]*logpb.ResourceLogs, len(svcReq.ResourceLogs)),
	}
	copy(data.ResourceLogs, svcReq.ResourceLogs)
	bs, err = e.options.marshaler.Marshal(&data)
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(e.options.writer, string(bs))
	return nil, errExported
}

func (e *Exporter) Export(ctx context.Context, records []log.Record) error {
	err := e.impl.Export(ctx, records)
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
