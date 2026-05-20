package spans

import (
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// Response, list, and nested types remain aliases to generated wire shapes so
// callers can construct/assert on them without importing internal/generated.
type (
	Span              = generated.Span
	SpanList          = generated.SpanList
	SpanDeletePartial = generated.SpanDeletePartial

	// SpanStatusCode is the status code of a span.
	SpanStatusCode = generated.SpanStatusCode
	// SpanContext holds the trace and span identifiers for a span.
	SpanContext = generated.SpanContext
	// SpanEvent is a timestamped event attached to a span.
	SpanEvent = generated.SpanEvent
	// Annotation is a human annotation on a span.
	Annotation = generated.Annotation
	// AnnotatorUser is a user assigned as an annotator, identified by ID and email.
	AnnotatorUser = generated.AnnotatorUser
	// Evaluation is an evaluation result attached to a span.
	Evaluation = generated.Evaluation
	// Email is an RFC-5322 email address (alias of openapi_types.Email).
	Email = generated.Email

	// AnnotateRecordInput is a single record (span) to annotate in a batch,
	// carrying the span ID and one or more annotation values.
	AnnotateRecordInput = generated.AnnotateRecordInput
	// AnnotationInput is an annotation value to set on a record, identified by
	// its annotation config name. Omitting Label/Score/Text leaves the
	// existing value unchanged.
	AnnotationInput = generated.AnnotationInput
)

const (
	SpanStatusCodeERROR SpanStatusCode = generated.SpanStatusCodeERROR
	SpanStatusCodeOK    SpanStatusCode = generated.SpanStatusCodeOK
	SpanStatusCodeUNSET SpanStatusCode = generated.SpanStatusCodeUNSET
)

// ListRequest is the request shape for Client.List. Unlike other list methods
// in this SDK, spans.List takes both a body (filter, time range) and query
// params (pagination); both halves are flattened into this single struct.
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
	// `status_code = 'ERROR'`). When empty, no filter is applied.
	Filter string

	// Limit is the optional maximum number of items to return (max 500). When
	// zero, the server applies its default page size.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// DeleteRequest is the request shape for Client.Delete.
type DeleteRequest struct {
	// Project identifies the target project. Accepts either a project name or
	// ID.
	Project string
	// Space accepts either a space name or ID. Required when Project is a
	// name; ignored when Project is an ID.
	Space string

	// SpanIDs is the list of span IDs to delete (maximum 5000).
	SpanIDs []string
}

// AnnotateRequest is the request shape for Client.Annotate.
type AnnotateRequest struct {
	// Project identifies the target project. Accepts either a project name or
	// ID.
	Project string
	// Space accepts either a space name or ID. Required when Project is a
	// name; ignored when Project is an ID.
	Space string

	// Annotations is the batch of span annotations to write. Up to 1000 spans
	// per request. Each entry identifies a span by its RecordId and carries
	// one or more AnnotationInput values; resubmitting the same annotation
	// config name for the same span overwrites the previous value, so retries
	// do not create duplicates.
	Annotations []AnnotateRecordInput

	// Start is the optional inclusive lower bound on span start time used when
	// looking up the spans to annotate. When zero, the server defaults to 31
	// days ago.
	Start time.Time
	// End is the optional exclusive upper bound on span start time used when
	// looking up the spans to annotate. When zero, the server defaults to the
	// current time.
	End time.Time
}
