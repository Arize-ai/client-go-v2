package experiments

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response and nested types are aliases to generated wire shapes so callers
// can construct/assert on them without importing internal/generated.
type (
	Experiment           = generated.Experiment
	ListExperiments      = generated.ListExperimentsResponse
	ExperimentRun        = generated.ExperimentRun
	ExperimentRunInput   = generated.ExperimentRunInput
	ListExperimentRuns   = generated.ListExperimentRunsResponse
	ExperimentWithRunIds = generated.ExperimentWithRunIds
)

// TaskFields names the columns in each Run map that hold the dataset example
// ID and the task output. Output is always required; ExampleID is required only
// for a dataset-backed experiment.
type TaskFields struct {
	// ExampleID is the column name in each run record holding the dataset
	// example ID (a string matching an example in the dataset/version).
	// Required when the experiment is associated with a dataset. Leave empty
	// for an experiment created in a space, whose runs reference no example;
	// no example_id is then sent.
	ExampleID string
	// Output is the column name in each run record holding the task output.
	// String values are sent as-is; any other type is JSON-encoded.
	Output string
}

// EvaluatorFields names the columns in each Run map that hold an evaluator's
// results. At least one of Score or Label must be set.
//
// Score, Label, and Explanation are best-effort: when a named column is absent
// from a given run record, that field is silently omitted for that run (not an
// error). Metadata columns, by contrast, are required — a named metadata column
// missing from any run record makes Create return an error.
type EvaluatorFields struct {
	// Score is the column name holding the evaluator score. Optional; silently
	// skipped for any run record that lacks the named column.
	Score string
	// Label is the column name holding the evaluator label. Optional; silently
	// skipped for any run record that lacks the named column.
	Label string
	// Explanation is the column name holding the evaluator explanation. Optional;
	// silently skipped for any run record that lacks the named column.
	Explanation string
	// Metadata maps a result metadata key to the column name in each run record
	// holding its value. An empty value means "use the metadata key as the
	// column name". Optional, but each named metadata column must be present in
	// every run record or Create returns an error.
	Metadata map[string]string
}

// ListRequest holds optional filters for listing experiments.
type ListRequest struct {
	// Dataset is an optional name-or-ID filter. When non-empty, only
	// experiments belonging to this dataset are returned.
	Dataset string
	// Space is an optional name-or-ID filter. When non-empty and Dataset is
	// empty, only experiments in this space are returned — both those
	// associated with a dataset and those without one. When Dataset is also
	// set, Space only resolves Dataset when Dataset is a name. When both are
	// empty, experiments across all accessible spaces are returned.
	Space string
	// Name is an optional case-insensitive substring filter on the experiment
	// name. When empty, results are not filtered by name.
	Name string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the server applies its default page size.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// GetRequest identifies the experiment to fetch.
type GetRequest struct {
	// Experiment accepts either an experiment name or ID.
	Experiment string
	// Dataset accepts either a dataset name or ID and scopes the search for
	// Experiment to that dataset. Provide Dataset or Space when Experiment is a
	// name; both are ignored when Experiment is an ID.
	Dataset string
	// Space accepts either a space name or ID. When Dataset is empty, it scopes
	// the search for Experiment to that space — the only way to resolve the name
	// of an experiment with no dataset. A name matching more than one experiment
	// in the space yields AmbiguousNameError, since names are unique per dataset
	// and, separately, per space among experiments with no dataset. Also
	// required when Dataset is passed as a name.
	Space string
}

// CreateRequest describes a new experiment. Each entry of Runs is a row of
// user data; TaskFields tells the SDK which columns hold the per-run
// example_id and output. EvaluatorColumns optionally remaps evaluator result
// columns to the wire format (e.g. `eval.<name>.score`).
type CreateRequest struct {
	// Dataset accepts either a dataset name or ID, and associates the
	// experiment with that dataset. Provide Space instead to create the
	// experiment directly in a space with no dataset; passing neither returns
	// ErrNoDatasetOrSpace.
	Dataset string
	// Space accepts either a space name or ID. When Dataset is empty, this is
	// the space to create the experiment in. When Dataset is a name, this
	// resolves it; ignored when Dataset is an ID.
	Space string
	// Name is the experiment's name. Required. Must be unique within the
	// dataset, or within the space when there is no dataset.
	Name string
	// Runs is the list of raw experiment run records. Required (one or more);
	// an empty Runs returns ErrNoRuns.
	Runs []map[string]any
	// TaskFields names the example_id and output columns in Runs. Required.
	// TaskFields.ExampleID itself is required only when Dataset is set.
	TaskFields TaskFields
	// EvaluatorColumns optionally maps evaluator name → its result column
	// names in Runs. Mapped columns are renamed to `eval.<evaluator>.score`
	// and friends in the wire payload. Optional.
	EvaluatorColumns map[string]EvaluatorFields
}

// DeleteRequest identifies the experiment to delete.
type DeleteRequest struct {
	// Experiment accepts either an experiment name or ID.
	Experiment string
	// Dataset accepts either a dataset name or ID and scopes the search for
	// Experiment to that dataset. Provide Dataset or Space when Experiment is a
	// name; both are ignored when Experiment is an ID.
	Dataset string
	// Space accepts either a space name or ID. When Dataset is empty, it scopes
	// the search for Experiment to that space — the only way to resolve the name
	// of an experiment with no dataset. A name matching more than one experiment
	// in the space yields AmbiguousNameError, since names are unique per dataset
	// and, separately, per space among experiments with no dataset. Also
	// required when Dataset is passed as a name.
	Space string
}

// ListRunsRequest identifies the experiment (resolved by name or ID) and
// pagination options for listing its runs.
type ListRunsRequest struct {
	// Experiment accepts either an experiment name or ID.
	Experiment string
	// Dataset accepts either a dataset name or ID and scopes the search for
	// Experiment to that dataset. Provide Dataset or Space when Experiment is a
	// name; both are ignored when Experiment is an ID.
	Dataset string
	// Space accepts either a space name or ID. When Dataset is empty, it scopes
	// the search for Experiment to that space — the only way to resolve the name
	// of an experiment with no dataset. A name matching more than one experiment
	// in the space yields AmbiguousNameError, since names are unique per dataset
	// and, separately, per space among experiments with no dataset. Also
	// required when Dataset is passed as a name.
	Space string
	// Limit is the optional maximum number of runs to return (max 500). When
	// zero, the server applies its default page size.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// AppendRunsRequest is the request shape for Client.AppendRuns.
type AppendRunsRequest struct {
	// ExperimentID is the ID of the experiment to append runs to. Required.
	ExperimentID string

	// ExperimentRuns is the list of runs to append. Between 1 and 1000 runs
	// per request. Each run must include Output. ExampleId is required only
	// when the target experiment is associated with a dataset; leave it nil
	// for an experiment that has none. The server rejects the mismatch.
	ExperimentRuns []ExperimentRunInput
}
