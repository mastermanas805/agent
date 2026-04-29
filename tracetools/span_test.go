package tracetools

import (
	"errors"
	"slices"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

type TestOtelSpan struct {
	embedded.Span

	finished       bool
	err            error
	events         []string
	spanContext    trace.SpanContext
	statusCode     codes.Code
	statusDesc     string
	name           string
	links          []trace.Link
	attributes     []attribute.KeyValue
	tracerProvider trace.TracerProvider
}

var _ trace.Span = (*TestOtelSpan)(nil)

func (t *TestOtelSpan) End(options ...trace.SpanEndOption)            { t.finished = true }
func (t *TestOtelSpan) IsRecording() bool                             { return !t.finished }
func (t *TestOtelSpan) RecordError(err error, _ ...trace.EventOption) { t.err = err }
func (t *TestOtelSpan) SpanContext() trace.SpanContext                { return t.spanContext }
func (t *TestOtelSpan) SetName(name string)                           { t.name = name }
func (t *TestOtelSpan) TracerProvider() trace.TracerProvider          { return t.tracerProvider }
func (t *TestOtelSpan) AddLink(link trace.Link)                       { t.links = append(t.links, link) }

func (t *TestOtelSpan) SetAttributes(kv ...attribute.KeyValue) {
	t.attributes = append(t.attributes, kv...)
}

func (t *TestOtelSpan) SetStatus(code codes.Code, description string) {
	t.statusCode, t.statusDesc = code, description
}

func (t *TestOtelSpan) AddEvent(name string, _ ...trace.EventOption) {
	t.events = append(t.events, name)
}

func newTestOtelSpan() *OpenTelemetrySpan {
	return &OpenTelemetrySpan{Span: &TestOtelSpan{events: []string{}, attributes: []attribute.KeyValue{}}}
}

func TestAddAttributeToSpan_OpenTelemetry(t *testing.T) {
	t.Parallel()

	span := newTestOtelSpan()
	implSpan, ok := span.Span.(*TestOtelSpan)
	if got := ok; !got {
		t.Errorf("span.Span.(*TestOtelSpan) = %t, want true", got)
	}

	if got := len(implSpan.attributes); got != 0 {
		t.Errorf("implSpan.attributes = %v, want 0", got)
	}

	span.AddAttributes(map[string]string{"colour": "blue", "flavour": "bittersweet"})
	if got, want := implSpan.attributes, attribute.String("colour", "blue"); !slices.Contains(got, want) {
		t.Errorf("implSpan.attributes = %v, want containing %v", got, want)
	}
	if got, want := implSpan.attributes, attribute.String("flavour", "bittersweet"); !slices.Contains(got, want) {
		t.Errorf("implSpan.attributes = %v, want containing %v", got, want)
	}
}

func TestFinishWithError_OpenTelemetry(t *testing.T) {
	t.Parallel()
	err := errors.New("test error")

	span := newTestOtelSpan()
	implSpan, ok := span.Span.(*TestOtelSpan)
	if got := ok; !got {
		t.Errorf("span.Span.(*TestOtelSpan) = %t, want true", got)
	}

	span.FinishWithError(err)
	if got := implSpan.finished; !got {
		t.Errorf("implSpan.finished = %t, want true", got)
	}
	if err, want := implSpan.err, err; !errors.Is(err, want) {
		t.Errorf("implSpan.err error = %v, want %v", err, want)
	}
	if got, want := codes.Error, implSpan.statusCode; got != want {
		t.Errorf("codes.Error = %d, want %d", got, want)
	}
	if got, want := implSpan.statusDesc, "failed"; got != want {
		t.Errorf("implSpan.statusDesc = %q, want %q", got, want)
	}

	span = newTestOtelSpan()
	implSpan, ok = span.Span.(*TestOtelSpan)
	if got := ok; !got {
		t.Errorf("span.Span.(*TestOtelSpan) = %t, want true", got)
	}

	span.FinishWithError(nil)
	if got := implSpan.finished; !got {
		t.Errorf("implSpan.finished = %t, want true", got)
	}
	if err := implSpan.err; err != nil {
		t.Errorf("implSpan.err error = %v, want nil", err)
	}
	if got, want := codes.Unset, implSpan.statusCode; got != want {
		t.Errorf("codes.Unset = %d, want %d", got, want)
	}
}
