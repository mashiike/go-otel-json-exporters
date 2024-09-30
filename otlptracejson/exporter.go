package otlptracejson

import (
	"context"
	"io"
	"os"

	"github.com/mashiike/go-otlp-helper/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type options struct {
	writer io.Writer
	enc    *otlp.JSONEncoder
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

// New constructs a new Exporter and starts it.
func New(ctx context.Context, opts ...Option) (*otlptrace.Exporter, error) {
	return otlptrace.New(ctx, NewClient(opts...))
}

// NewUnstarted constructs a new Exporter and does not start it.
func NewUnstarted(opts ...Option) *otlptrace.Exporter {
	return otlptrace.NewUnstarted(NewClient(opts...))
}

type Client struct {
	options options
}

var _ otlptrace.Client = &Client{}

func NewClient(opts ...Option) *Client {
	o := options{
		writer: os.Stdout,
		enc:    otlp.NewJSONEncoder(os.Stdout),
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &Client{
		options: o,
	}
}

func (c *Client) Start(ctx context.Context) error {
	return nil
}

func (c *Client) Stop(ctx context.Context) error {
	return nil
}

func (c *Client) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {
	data := &tracepb.TracesData{
		ResourceSpans: protoSpans,
	}

	if err := c.options.enc.Encode(data); err != nil {
		return err
	}
	if _, err := c.options.writer.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}
