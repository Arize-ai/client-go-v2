package annotationqueues

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, and nested input types remain aliases to the generated wire
// shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	AnnotationQueue               = generated.AnnotationQueue
	ListAnnotationQueues          = generated.ListAnnotationQueuesResponse
	AnnotationQueueRecord         = generated.AnnotationQueueRecord
	ListAnnotationQueueRecords    = generated.ListAnnotationQueueRecordsResponse
	CreateAnnotationQueueRecord   = generated.CreateAnnotationQueueRecordResponse
	AnnotateAnnotationQueueRecord = generated.AnnotateAnnotationQueueRecordResponse
	AssignAnnotationQueueRecord   = generated.AssignAnnotationQueueRecordResponse

	// Email is an annotator or assignee email address.
	Email = generated.Email

	// AnnotationInput is one annotation value: a required Name plus an optional
	// Score, Label, or Text.
	AnnotationInput = generated.AnnotationInput

	// AnnotationQueueRecordInput is a record source to add to a queue — either a
	// set of dataset examples or a set of spans. Build one with
	// NewExampleRecordSource or NewSpanRecordSource.
	AnnotationQueueRecordInput = generated.AnnotationQueueRecordInput
	// AnnotationQueueExampleRecordInput adds dataset examples to a queue.
	AnnotationQueueExampleRecordInput = generated.AnnotationQueueExampleRecordInput
	// AnnotationQueueSpanRecordInput adds spans (by project and time range) to a queue.
	AnnotationQueueSpanRecordInput = generated.AnnotationQueueSpanRecordInput

	// AssignmentMethod controls how queue records are assigned to annotators.
	AssignmentMethod = generated.AssignmentMethod
)

const (
	// AssignmentMethodAll assigns every record to every annotator.
	AssignmentMethodAll AssignmentMethod = generated.AssignmentMethodALL
	// AssignmentMethodRandom assigns each record randomly to one annotator.
	AssignmentMethodRandom AssignmentMethod = generated.AssignmentMethodRANDOM
)

// NewExampleRecordSource builds a record source that adds dataset examples to a
// queue, setting the record-type discriminator for the caller.
func NewExampleRecordSource(in AnnotationQueueExampleRecordInput) (AnnotationQueueRecordInput, error) {
	in.RecordType = generated.AnnotationQueueExampleRecordInputRecordTypeEXAMPLE
	var src AnnotationQueueRecordInput
	if err := src.FromAnnotationQueueExampleRecordInput(in); err != nil {
		return AnnotationQueueRecordInput{}, err
	}
	return src, nil
}

// NewSpanRecordSource builds a record source that adds spans to a queue, setting
// the record-type discriminator for the caller.
func NewSpanRecordSource(in AnnotationQueueSpanRecordInput) (AnnotationQueueRecordInput, error) {
	in.RecordType = generated.AnnotationQueueSpanRecordInputRecordTypeSPAN
	var src AnnotationQueueRecordInput
	if err := src.FromAnnotationQueueSpanRecordInput(in); err != nil {
		return AnnotationQueueRecordInput{}, err
	}
	return src, nil
}

// ListRequest holds optional filters for listing annotation queues.
type ListRequest struct {
	// Space is an optional name-or-ID filter. When non-empty, only queues in
	// that space are returned; when empty, queues across all accessible spaces
	// are returned.
	Space string
	// Name is an optional case-insensitive substring filter on the queue name.
	// When empty, results are not filtered by name.
	Name string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the server applies its default page size.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous response's
	// pagination.next_cursor. When empty, results start from the first page.
	Cursor string
}

// GetRequest identifies the annotation queue to fetch.
type GetRequest struct {
	// AnnotationQueue accepts either a queue name or ID.
	AnnotationQueue string
	// Space accepts either a space name or ID. Required when AnnotationQueue is
	// a name; ignored when AnnotationQueue is an ID.
	Space string
}

// CreateRequest describes a new annotation queue.
type CreateRequest struct {
	// Space accepts either a space name or ID and identifies the parent space.
	Space string
	// Name is the annotation queue's name (must be unique within the space).
	Name string
	// AnnotationConfigIDs are optional IDs of annotation configs to associate
	// with the queue. When empty, no configs are associated.
	AnnotationConfigIDs []string
	// AnnotatorEmails are optional annotator email addresses to assign. When
	// empty, no annotators are assigned.
	AnnotatorEmails []Email
	// AssignmentMethod is the optional strategy for assigning records to
	// annotators. When empty, the server applies its default (AssignmentMethodAll).
	AssignmentMethod AssignmentMethod
	// Instructions is optional guidance shown to annotators. When empty, the
	// queue is created without instructions.
	Instructions string
	// RecordSources are optional records to add on creation (max 2 sources).
	// Build entries with NewExampleRecordSource or NewSpanRecordSource. When
	// empty, the queue is created with no records.
	RecordSources []AnnotationQueueRecordInput
}

// UpdateRequest identifies the queue to update and the fields to patch.
type UpdateRequest struct {
	// AnnotationQueue accepts either a queue name or ID.
	AnnotationQueue string
	// Space accepts either a space name or ID. Required when AnnotationQueue is
	// a name; ignored when AnnotationQueue is an ID.
	Space string
	// Name, when non-nil, sets a new queue name; when nil, the existing name is
	// preserved.
	Name *string
	// Instructions, when non-nil, sets new annotator instructions (a pointer to
	// the empty string clears them); when nil, existing instructions are
	// preserved.
	Instructions *string
	// AnnotatorEmails, when non-nil, replaces the full annotator list; when nil,
	// the existing annotators are preserved.
	AnnotatorEmails *[]Email
	// AnnotationConfigIDs, when non-nil, replaces the full annotation-config
	// list; when nil, the existing configs are preserved.
	AnnotationConfigIDs *[]string
}

// DeleteRequest identifies the queue to delete.
type DeleteRequest struct {
	// AnnotationQueue accepts either a queue name or ID.
	AnnotationQueue string
	// Space accepts either a space name or ID. Required when AnnotationQueue is
	// a name; ignored when AnnotationQueue is an ID.
	Space string
}

// ListRecordsRequest identifies the queue (resolved by name or ID) and
// pagination options for listing its records.
type ListRecordsRequest struct {
	// AnnotationQueue accepts either a queue name or ID.
	AnnotationQueue string
	// Space accepts either a space name or ID. Required when AnnotationQueue is
	// a name; ignored when AnnotationQueue is an ID.
	Space string
	// Limit is the optional maximum number of items to return. When zero, the
	// server applies its default page size.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous response.
	// When empty, results start from the first page.
	Cursor string
}

// AddRecordsRequest identifies the queue and the record sources to add.
type AddRecordsRequest struct {
	// AnnotationQueue accepts either a queue name or ID.
	AnnotationQueue string
	// Space accepts either a space name or ID. Required when AnnotationQueue is
	// a name; ignored when AnnotationQueue is an ID.
	Space string
	// RecordSources are the records to add (max 2 sources). Build entries with
	// NewExampleRecordSource or NewSpanRecordSource.
	RecordSources []AnnotationQueueRecordInput
}

// DeleteRecordsRequest identifies the queue and the record IDs to remove.
type DeleteRecordsRequest struct {
	// AnnotationQueue accepts either a queue name or ID.
	AnnotationQueue string
	// Space accepts either a space name or ID. Required when AnnotationQueue is
	// a name; ignored when AnnotationQueue is an ID.
	Space string
	// RecordIDs are the strict IDs of the records to remove.
	RecordIDs []string
}

// AnnotateRequest identifies the queue and record to annotate and the
// annotations to submit.
type AnnotateRequest struct {
	// AnnotationQueue accepts either a queue name or ID.
	AnnotationQueue string
	// Space accepts either a space name or ID. Required when AnnotationQueue is
	// a name; ignored when AnnotationQueue is an ID.
	Space string
	// RecordID is the strict ID of the record to annotate (no name resolution).
	RecordID string
	// Annotations are the annotation values to submit.
	Annotations []AnnotationInput
}

// AssignRequest identifies the queue and record to assign and the users to
// assign to it.
type AssignRequest struct {
	// AnnotationQueue accepts either a queue name or ID.
	AnnotationQueue string
	// Space accepts either a space name or ID. Required when AnnotationQueue is
	// a name; ignored when AnnotationQueue is an ID.
	Space string
	// RecordID is the strict ID of the record to assign (no name resolution).
	RecordID string
	// AssignedUserEmails are the email addresses of the users to assign.
	AssignedUserEmails []Email
}
