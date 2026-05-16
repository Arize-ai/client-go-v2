package spans

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	Span              = generated.Span
	SpanList          = generated.SpanList
	SpanDeletePartial = generated.SpanDeletePartial
	ListRequest       = generated.ListSpansRequest
	DeleteRequest     = generated.DeleteSpansRequest
	ListParams        = generated.SpansListParams

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
)

const (
	SpanStatusCodeERROR SpanStatusCode = generated.SpanStatusCodeERROR
	SpanStatusCodeOK    SpanStatusCode = generated.SpanStatusCodeOK
	SpanStatusCodeUNSET SpanStatusCode = generated.SpanStatusCodeUNSET
)
