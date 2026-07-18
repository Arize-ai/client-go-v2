package datasets

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, and nested example types remain aliases to the generated
// wire shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	Dataset             = generated.Dataset
	ListDatasets        = generated.ListDatasetsResponse
	DatasetExample      = generated.DatasetExample
	ListDatasetExamples = generated.ListDatasetExamplesResponse

	// CreateDatasetExampleInput is a new example's arbitrary user-defined fields.
	CreateDatasetExampleInput = generated.CreateDatasetExampleInput

	// UpdateDatasetExampleInput is an existing example ID plus arbitrary
	// user-defined fields to update.
	UpdateDatasetExampleInput = generated.UpdateDatasetExampleInput

	// DatasetExamplesInserted reports the dataset version the examples were
	// written to and the server-assigned IDs of the inserted examples.
	InsertDatasetExamples = generated.InsertDatasetExamplesResponse

	// UpdateDatasetExamplesResponse reports the dataset version the examples were
	// written to and the IDs of the updated examples.
	UpdateDatasetExamplesResponse = generated.UpdateDatasetExamplesResponse

	// DatasetExamplesDeleteResult reports the outcome of a batch example
	// delete. Completed indicates whether a retry is needed; DeletedExampleIds
	// are the IDs confirmed deleted; NotDeletedExampleIds are requested IDs
	// that were not deleted (not found in the selected version, or not reached
	// when Completed is false).
	DeleteDatasetExamples = generated.DeleteDatasetExamplesResponse

	// AnnotateRecordInput is a single dataset example to annotate, identified
	// by its example ID (RecordId) plus the annotation Values to set.
	AnnotateRecordInput = generated.AnnotateRecordInput
	// AnnotationInput is one annotation value: a required Name plus an optional
	// Score, Label, or Text.
	AnnotationInput = generated.AnnotationInput
)

// ListRequest holds optional filters for listing datasets.
type ListRequest struct {
	// Space is an optional name-or-ID filter. When non-empty, only datasets
	// belonging to this space are returned; when empty, datasets across all
	// accessible spaces are returned.
	Space string
	// Name is an optional case-insensitive substring filter on the dataset
	// name. When empty, results are not filtered by name.
	Name string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// GetRequest identifies the dataset to fetch.
type GetRequest struct {
	// Dataset accepts either a dataset name or ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
}

// CreateRequest describes a new dataset.
type CreateRequest struct {
	// Space accepts either a space name or ID and identifies the parent space
	// for the new dataset.
	Space string
	// Name is the dataset's name (must be unique within the space).
	Name string
	// Examples are the initial examples for the dataset. At least one example
	// is required; creating an empty dataset returns ErrNoExamples.
	Examples []CreateDatasetExampleInput
}

// UpdateRequest renames a dataset.
type UpdateRequest struct {
	// Dataset accepts either a dataset name or ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
	// Name is the dataset's new name (must be unique within the space).
	Name string
}

// DeleteRequest identifies the dataset to delete.
type DeleteRequest struct {
	// Dataset accepts either a dataset name or ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
}

// ListExamplesRequest identifies the dataset (resolved by name or ID) and
// pagination/filter options for listing its examples.
type ListExamplesRequest struct {
	// Dataset accepts either a dataset name or ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
	// Limit is the optional maximum number of items to return (max 500). When
	// zero, the SDK applies a default of 50.
	Limit int
	// DatasetVersionID is the optional version to list examples from. When
	// empty, the server uses the dataset's latest version.
	DatasetVersionID string
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// AppendExamplesRequest identifies the dataset and the examples to append.
type AppendExamplesRequest struct {
	// Dataset accepts either a dataset name or ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
	// DatasetVersionID is the optional version to append to. When empty, the
	// server uses the dataset's latest version.
	DatasetVersionID string
	// Examples are the examples to append.
	Examples []CreateDatasetExampleInput
}

// UpdateDatasetExamplesRequest identifies the dataset and examples to update.
type UpdateDatasetExamplesRequest struct {
	// Dataset accepts either a dataset name or ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
	// DatasetVersionID is the optional version to update. When empty, the
	// server uses the dataset's latest version.
	DatasetVersionID string
	// Examples are existing examples to update. Each example must include Id;
	// arbitrary user-defined fields may be supplied through
	// AdditionalProperties.
	Examples []UpdateDatasetExampleInput
	// NewVersion is an optional name for a new dataset version. When empty,
	// the selected version is updated in place.
	NewVersion string
}

// DeleteExamplesRequest identifies the dataset and the examples to delete.
type DeleteExamplesRequest struct {
	// Dataset accepts either a dataset name or ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
	// DatasetVersionID is the version to delete the examples from. Required;
	// examples are removed in place from this version and no new version is
	// created.
	DatasetVersionID string
	// ExampleIDs are the IDs of the examples to delete (up to 1000 per
	// request).
	ExampleIDs []string
}

// AnnotateExamplesRequest identifies the dataset and the example annotations
// to write.
type AnnotateExamplesRequest struct {
	// Dataset accepts either a dataset name or ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
	// Annotations are the per-example annotations to upsert (up to 1000 per
	// request). Annotations are keyed by annotation config name for each
	// example; re-annotating the same example with the same config name
	// overwrites the previous value.
	Annotations []AnnotateRecordInput
}
