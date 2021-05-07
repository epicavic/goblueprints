package trace

import (
	"fmt"
	"io"
)

// tracer is a type that writes to an io.Writer interface
type tracer struct {
	out io.Writer
}

// Trace writes the arguments to tracer's io.Writer interface
func (t *tracer) Trace(a ...interface{}) {
	fmt.Fprintln(t.out, a...)
}

// nilTracer
type nilTracer struct{}

// Trace for a nilTracer does nothing.
func (t *nilTracer) Trace(a ...interface{}) {}

// Tracer is the interface that describes an object capable of
// tracing events throughout code.
type Tracer interface {
	Trace(...interface{})
}

// New creates a new Tracer that will write the output to
// the specified io.Writer interface
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

// Off creates a Tracer that will ignore calls to Trace.
func Off() Tracer {
	return &nilTracer{}
}
