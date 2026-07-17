package tasks

import (
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// Response, list, and nested types remain aliases to the generated wire
// shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	// Task is a single task: an automated job that evaluates data or runs
	// experiments.
	Task = generated.Task

	// TaskList is the cursor-paginated list response shape.
	TaskList = generated.ListTasksResponse

	// TaskEvaluator is an evaluator attached to an evaluation task, as
	// returned on Task.Evaluators.
	TaskEvaluator = generated.TaskEvaluator

	// TaskRun is a single async run of a task, created by TriggerRun.
	TaskRun = generated.TaskRun

	// TaskRunList is the cursor-paginated run list response shape.
	TaskRunList = generated.ListTaskRunsResponse

	// TaskType is the task kind: TEMPLATE_EVALUATION, CODE_EVALUATION, or
	// RUN_EXPERIMENT.
	TaskType = generated.TaskType

	// TaskRunStatus is the lifecycle status of a task run.
	TaskRunStatus = generated.TaskRunStatus

	// RunConfiguration is the experiment execution configuration for a
	// RUN_EXPERIMENT task. It is a oneOf: populate exactly one variant with
	// FromLlmGenerationRunConfig / FromTemplateEvaluationRunConfig and read
	// the active variant with ValueByDiscriminator and a type switch over
	// LLMGenerationRunConfig / TemplateEvaluationRunConfig.
	RunConfiguration = generated.RunConfiguration

	// LLMGenerationRunConfig is the llm_generation variant of a
	// RunConfiguration: runs an LLM over each dataset example.
	LLMGenerationRunConfig = generated.LlmGenerationRunConfig

	// TemplateEvaluationRunConfig is the TEMPLATE_EVALUATION variant of a
	// RunConfiguration: evaluates dataset examples with a prompt template.
	TemplateEvaluationRunConfig = generated.TemplateEvaluationRunConfig

	// LLMMessage is a single message in an LLMGenerationRunConfig.
	LLMMessage = generated.LLMMessage

	// InputVariableFormat is the placeholder syntax used by run-config
	// messages (f_string or mustache).
	InputVariableFormat = generated.InputVariableFormat

	// InvocationParams holds LLM invocation parameters (temperature, etc.).
	InvocationParams = generated.InvocationParams

	// ToolConfig holds the tool configuration for an LLMGenerationRunConfig.
	ToolConfig = generated.ToolConfig
)

const (
	TaskTypeTemplateEvaluation TaskType = generated.TaskTypeTEMPLATEEVALUATION
	TaskTypeCodeEvaluation     TaskType = generated.TaskTypeCODEEVALUATION
	TaskTypeRunExperiment      TaskType = generated.TaskTypeRUNEXPERIMENT

	TaskRunStatusPending   TaskRunStatus = generated.TaskRunStatusPENDING
	TaskRunStatusRunning   TaskRunStatus = generated.TaskRunStatusRUNNING
	TaskRunStatusCompleted TaskRunStatus = generated.TaskRunStatusCOMPLETED
	TaskRunStatusFailed    TaskRunStatus = generated.TaskRunStatusFAILED
	TaskRunStatusCancelled TaskRunStatus = generated.TaskRunStatusCANCELLED
)

// EvaluatorInput attaches an evaluator to an evaluation task.
type EvaluatorInput struct {
	// EvaluatorID is the evaluator's ID (base64). Required; duplicates are
	// not allowed within one task.
	EvaluatorID string
	// QueryFilter is an optional per-evaluator query filter, combined with
	// the task-level filter (AND). When empty, no per-evaluator filter is
	// applied.
	QueryFilter string
	// ColumnMappings optionally maps evaluator template variable names to
	// data source column names. When nil, no mappings are sent.
	ColumnMappings map[string]string
}

// ListRequest holds optional filters for listing tasks.
type ListRequest struct {
	// Space is an optional name-or-ID filter. When non-empty, only tasks
	// belonging to this space are returned; when empty, tasks across all
	// accessible spaces are returned.
	Space string
	// Name is an optional case-insensitive substring filter on the task
	// name. When empty, results are not filtered by name.
	Name string
	// Project accepts either a project name or ID. When non-empty, only
	// tasks attached to this project are returned. Space is required when
	// Project is a name; ignored when Project is an ID.
	Project string
	// Dataset accepts either a dataset name or ID. When non-empty, only
	// tasks attached to this dataset are returned. Space is required when
	// Dataset is a name; ignored when Dataset is an ID.
	Dataset string
	// Type is an optional filter on the task type. When empty, results are
	// not filtered by type.
	Type TaskType
	// Limit is the optional maximum number of items to return (max 100).
	// When zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// GetRequest identifies the task to fetch.
type GetRequest struct {
	// Task accepts either a task name or ID.
	Task string
	// Space accepts either a space name or ID. Required when Task is a
	// name; ignored when Task is an ID.
	Space string
}

// CreateEvaluationTaskRequest describes a new TEMPLATE_EVALUATION or
// CODE_EVALUATION task. Exactly one of Project or Dataset must be set.
type CreateEvaluationTaskRequest struct {
	// Name is the task's name (must be unique within the space).
	Name string
	// Type is the task type: TaskTypeTemplateEvaluation or
	// TaskTypeCodeEvaluation.
	Type TaskType
	// Evaluators are the evaluators to attach. At least one is required.
	Evaluators []EvaluatorInput
	// Project accepts either a project name or ID. Required when Dataset is
	// not set; mutually exclusive with Dataset. Space is required when
	// Project is a name; ignored when Project is an ID.
	Project string
	// Dataset accepts either a dataset name or ID. Required when Project is
	// not set; mutually exclusive with Project. Space is required when
	// Dataset is a name; ignored when Dataset is an ID.
	Dataset string
	// Space accepts either a space name or ID. Required when Project or
	// Dataset is a name; ignored when they are IDs.
	Space string
	// ExperimentIDs are experiment IDs (base64) to evaluate. Required when
	// Dataset is set (at least one entry); must be empty for project-based
	// tasks.
	ExperimentIDs []string
	// SamplingRate is the optional sampling rate between 0 and 1. Only
	// supported on project-based tasks. When zero, no sampling rate is sent
	// and the server applies its default.
	SamplingRate float32
	// IsContinuous optionally makes the task run continuously on incoming
	// data. Only supported on project-based tasks. When false, the field is
	// omitted and the server default (false) applies.
	IsContinuous bool
	// QueryFilter is an optional task-level query filter applied to all
	// evaluated data (AND-ed with per-evaluator filters). When empty, no
	// filter is applied.
	QueryFilter string
}

// CreateRunExperimentTaskRequest describes a new RUN_EXPERIMENT task.
type CreateRunExperimentTaskRequest struct {
	// Name is the task's name (must be unique within the space).
	Name string
	// Dataset accepts either a dataset name or ID and identifies the
	// dataset the experiments run against.
	Dataset string
	// Space accepts either a space name or ID. Required when Dataset is a
	// name; ignored when Dataset is an ID.
	Space string
	// RunConfiguration is the experiment execution configuration. Populate
	// exactly one variant via its FromLlmGenerationRunConfig or
	// FromTemplateEvaluationRunConfig method.
	RunConfiguration RunConfiguration
}

// UpdateRequest carries the target task and the patch fields. The SDK first
// fetches the task to determine its type and rejects fields that don't apply
// to it: Name applies to all tasks; SamplingRate, IsContinuous, QueryFilter,
// and Evaluators apply only to evaluation tasks; RunConfiguration applies
// only to RUN_EXPERIMENT tasks. At least one patch field must be set.
type UpdateRequest struct {
	// Task accepts either a task name or ID.
	Task string
	// Space accepts either a space name or ID. Required when Task is a
	// name; ignored when Task is an ID.
	Space string
	// Name is optional. When non-nil, sets a new name for the task; when
	// nil, the existing name is preserved.
	Name *string
	// SamplingRate is optional (evaluation tasks on projects only). When
	// non-nil, sets a new sampling rate between 0 and 1; when nil, the
	// existing rate is preserved.
	SamplingRate *float32
	// IsContinuous is optional (evaluation tasks on projects only). When
	// non-nil, sets whether the task runs continuously; when nil, the
	// existing value is preserved.
	IsContinuous *bool
	// QueryFilter is optional (evaluation tasks only). When non-nil, sets a
	// new task-level query filter (pass a pointer to an empty string to
	// clear the existing filter); when nil, the existing filter is
	// preserved.
	QueryFilter *string
	// Evaluators is optional (evaluation tasks only). When non-empty,
	// replaces the entire evaluator list (at least one evaluator is
	// required by the API); when nil, the existing evaluators are
	// preserved.
	Evaluators []EvaluatorInput
	// RunConfiguration is optional (RUN_EXPERIMENT tasks only). When
	// non-nil, fully replaces the stored run configuration; when nil, the
	// existing configuration is preserved.
	RunConfiguration *RunConfiguration
}

// DeleteRequest identifies the task to delete.
type DeleteRequest struct {
	// Task accepts either a task name or ID.
	Task string
	// Space accepts either a space name or ID. Required when Task is a
	// name; ignored when Task is an ID.
	Space string
}

// TriggerRunRequest carries the target task and the run parameters. The SDK
// first fetches the task to determine its type and rejects fields that don't
// apply to it: DataStartTime, DataEndTime, MaxSpans, OverrideEvaluations, and
// ExperimentIDs apply only to evaluation tasks; ExperimentName,
// DatasetVersionID, ExampleIDs, MaxExamples, TracingMetadata, and
// EvaluationTaskIDs apply only to RUN_EXPERIMENT tasks.
type TriggerRunRequest struct {
	// Task accepts either a task name or ID.
	Task string
	// Space accepts either a space name or ID. Required when Task is a
	// name; ignored when Task is an ID.
	Space string

	// Evaluation-task fields (TEMPLATE_EVALUATION / CODE_EVALUATION).

	// DataStartTime is the optional start of the data window to evaluate.
	// When zero, the server defaults to the task's last run time (required
	// by the server on the first run). Project-based tasks only.
	DataStartTime time.Time
	// DataEndTime is the optional end of the data window to evaluate. When
	// zero, the server defaults to now. Project-based tasks only.
	DataEndTime time.Time
	// MaxSpans is the optional maximum number of spans to process. When
	// zero, the server applies a default of 10000.
	MaxSpans int
	// OverrideEvaluations optionally re-evaluates data that already has
	// evaluation labels. When false, the field is omitted and the server
	// default (false) applies.
	OverrideEvaluations bool
	// ExperimentIDs are optional experiment IDs (base64) to run against.
	// Dataset-based evaluation tasks only. When nil, the task's configured
	// experiments are used.
	ExperimentIDs []string

	// Run-experiment-task fields.

	// ExperimentName is the display name for the experiment to be created
	// (must be unique within the dataset). Required for RUN_EXPERIMENT
	// tasks.
	ExperimentName string
	// DatasetVersionID is the optional dataset version (base64) to run
	// against. When empty, the server uses the latest version.
	DatasetVersionID string
	// ExampleIDs are optional specific example IDs to run against. Mutually
	// exclusive with MaxExamples. When both are zero, all examples are
	// used.
	ExampleIDs []string
	// MaxExamples is the optional maximum number of examples to run, in
	// dataset order. Mutually exclusive with ExampleIDs. When both are
	// zero, all examples are used.
	MaxExamples int
	// TracingMetadata is optional arbitrary key-value metadata. Providing
	// it enables tracing for the run. When nil, tracing is not enabled.
	TracingMetadata map[string]string
	// EvaluationTaskIDs are optional task IDs (base64) of evaluation tasks
	// to trigger after the experiment run completes. When nil, no follow-up
	// tasks are triggered.
	EvaluationTaskIDs []string
}

// ListRunsRequest identifies the task and holds optional filters for listing
// its runs.
type ListRunsRequest struct {
	// Task accepts either a task name or ID.
	Task string
	// Space accepts either a space name or ID. Required when Task is a
	// name; ignored when Task is an ID.
	Space string
	// Status is an optional filter on run status. When empty, results are
	// not filtered by status.
	Status TaskRunStatus
	// Limit is the optional maximum number of items to return (max 100).
	// When zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// GetRunRequest identifies the task run to fetch.
type GetRunRequest struct {
	// RunID is the task run's ID (base64).
	RunID string
}

// CancelRunRequest identifies the task run to cancel.
type CancelRunRequest struct {
	// RunID is the task run's ID (base64).
	RunID string
}

// WaitForRunRequest identifies the task run to wait for and the polling
// parameters.
type WaitForRunRequest struct {
	// RunID is the task run's ID (base64).
	RunID string
	// PollInterval is the optional time between status polls. When zero,
	// the SDK applies a default of 5 seconds.
	PollInterval time.Duration
	// Timeout is the optional maximum time to wait for the run to reach a
	// terminal state. When zero, the SDK applies a default of 10 minutes.
	// On expiry, WaitForRun returns an error wrapping ErrWaitTimeout.
	Timeout time.Duration
}
