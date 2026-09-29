package spans

import (
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// Response, list, and nested types remain aliases to generated wire shapes so
// callers can construct/assert on them without importing internal/generated.
type (
	Span        = generated.Span
	ListSpans   = generated.ListSpansResponse
	DeleteSpans = generated.DeleteSpansResponse

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

	// AnnotateRecordInput is a single record (span, trace root span, or
	// session, depending on AnnotateRequest.Granularity) to annotate in a
	// batch, carrying the record ID and one or more annotation values.
	AnnotateRecordInput = generated.AnnotateRecordInput
	// AnnotationInput is an annotation value to set on a record, identified by
	// its annotation config name. Omitting Label/Score/Text leaves the
	// existing value unchanged.
	AnnotationInput = generated.AnnotationInput

	// Granularity selects what an AnnotateRecordInput.RecordId identifies:
	// a span, a trace (by its root span), or a session.
	Granularity = generated.RecordGranularity
)

const (
	SpanStatusCodeERROR SpanStatusCode = generated.SpanStatusCodeERROR
	SpanStatusCodeOK    SpanStatusCode = generated.SpanStatusCodeOK
	SpanStatusCodeUNSET SpanStatusCode = generated.SpanStatusCodeUNSET

	// GranularitySPAN annotates a record identified by its span ID. This is
	// the default when AnnotateRequest.Granularity is left zero.
	GranularitySPAN Granularity = generated.RecordGranularitySPAN
	// GranularityTRACE annotates a record identified by a trace's root span
	// ID. Annotating a non-root span is rejected.
	GranularityTRACE Granularity = generated.RecordGranularityTRACE
	// GranularitySESSION annotates a record identified by a session ID. The
	// annotation is written to the root span of the session's earliest trace
	// found within the lookup window.
	GranularitySESSION Granularity = generated.RecordGranularitySESSION
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
	// IncludedColumns is an optional list of full dotted column paths to return
	// for each span. Fixed span fields are always returned. When nil, all
	// available columns are returned. Cannot be used with ExcludedColumns.
	IncludedColumns []string
	// ExcludedColumns is an optional list of full dotted column paths to omit
	// from each span. Fixed span fields are always returned. When nil, no
	// columns are omitted. Cannot be used with IncludedColumns.
	ExcludedColumns []string

	// Limit is the optional maximum number of items to return (max 500). When
	// zero, the SDK applies a default of 50.
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

	// Start is the optional inclusive lower bound on span start time. When
	// zero, the server searches the full 2-year lookback window.
	Start time.Time
	// End is the optional exclusive upper bound on span start time. When zero,
	// the server searches up to the current time.
	End time.Time
}

// AnnotateRequest is the request shape for Client.Annotate.
type AnnotateRequest struct {
	// Project identifies the target project. Accepts either a project name or
	// ID.
	Project string
	// Space accepts either a space name or ID. Required when Project is a
	// name; ignored when Project is an ID.
	Space string

	// Annotations is the batch of record annotations to write. Up to 1000
	// records per request for GranularitySPAN/GranularityTRACE; up to 100 for
	// GranularitySESSION. Each entry identifies a record by its RecordId
	// (interpreted per Granularity) and carries one or more AnnotationInput
	// values; resubmitting the same annotation config name for the same
	// record overwrites the previous value, so retries do not create
	// duplicates.
	Annotations []AnnotateRecordInput

	// Start is the optional inclusive lower bound on record lookup time. When
	// zero, the server defaults to 31 days ago, or 7 days ago when
	// Granularity is GranularitySESSION.
	Start time.Time
	// End is the optional exclusive upper bound on record lookup time. When
	// zero, the server defaults to the current time.
	End time.Time

	// Granularity selects what each RecordId identifies: a span, a trace (by
	// its root span), or a session. When zero, the server defaults to
	// GranularitySPAN.
	Granularity Granularity
}
