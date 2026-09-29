package traces

import (
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// Response, list, and nested types remain aliases to generated wire shapes so
// callers can construct/assert on them without importing internal/generated.
type (
	// Trace is a single trace: its trace ID, root span ID, time bounds, and the
	// flat list of spans belonging to it.
	Trace = generated.Trace
	// ListTraces is the response envelope for Client.List.
	ListTraces = generated.ListTracesResponse

	// Span is a single span belonging to a trace. Each span has the same shape
	// and enrichment as spans returned by POST /v2/spans.
	Span = generated.Span
	// SpanContext holds the trace and span identifiers for a span.
	SpanContext = generated.SpanContext
	// SpanEvent is a timestamped event attached to a span.
	SpanEvent = generated.SpanEvent
	// SpanStatusCode is the status code of a span.
	SpanStatusCode = generated.SpanStatusCode
	// Annotation is a human annotation on a span.
	Annotation = generated.Annotation
	// AnnotatorUser is a user assigned as an annotator, identified by ID and email.
	AnnotatorUser = generated.AnnotatorUser
	// Evaluation is an evaluation result attached to a span.
	Evaluation = generated.Evaluation
	// Email is an RFC-5322 email address (alias of openapi_types.Email).
	Email = generated.Email
)

const (
	SpanStatusCodeERROR SpanStatusCode = generated.SpanStatusCodeERROR
	SpanStatusCodeOK    SpanStatusCode = generated.SpanStatusCodeOK
	SpanStatusCodeUNSET SpanStatusCode = generated.SpanStatusCodeUNSET
)

// ListRequest is the request shape for Client.List. Like spans.List, traces.List
// takes both a body (project, time range, filter) and query params (pagination);
// both halves are flattened into this single struct.
type ListRequest struct {
	// Project identifies the target project. Accepts either a project name or
	// ID.
	Project string
	// Space accepts either a space name or ID. Required when Project is a
	// name; ignored when Project is an ID.
	Space string

	// Start is the optional inclusive lower bound on span start time. When
	// zero, the server defaults to 1 week ago.
	Start time.Time
	// End is the optional exclusive upper bound on span start time. When zero,
	// the server defaults to the current time.
	End time.Time
	// Filter is an optional filter expression (SQL-like syntax, e.g.
	// `status_code = 'ERROR'`). A trace is returned when any of its spans
	// matches the filter. When empty, no filter is applied.
	Filter string

	// Limit is the optional maximum number of traces to return (max 50). When
	// zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}
