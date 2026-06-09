package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// ErrNoUpdateFields is returned by Update when the request carries no patch
// fields. At least one field must be set.
var ErrNoUpdateFields = errors.New("tasks: update requires at least one field to change")

// ErrWaitTimeout is returned (wrapped) by WaitForRun when the run does not
// reach a terminal state within the configured timeout.
var ErrWaitTimeout = errors.New("tasks: timed out waiting for run to reach a terminal state")

const (
	defaultPollInterval = 5 * time.Second
	defaultWaitTimeout  = 10 * time.Minute
)

// Client provides access to the Arize Tasks API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of tasks. req.Space, when non-empty, accepts
// a space name or ID and restricts results to that space. req.Project and
// req.Dataset, when non-empty, accept a name or ID and restrict results to
// tasks attached to that project or dataset.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*TaskList, error) {
	prerelease.Warn("tasks.list", prerelease.Alpha)
	params := generated.TasksListParams{
		Name:   optfields.PtrIfSet(req.Name),
		Type:   optfields.PtrIfSet(req.Type),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	if req.Project != "" {
		projectID, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
		if err != nil {
			return nil, err
		}
		params.ProjectId = &projectID
	}
	if req.Dataset != "" {
		datasetID, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
		if err != nil {
			return nil, err
		}
		params.DatasetId = &datasetID
	}
	resp, err := c.gen.TasksListWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single task. req.Task accepts a name or ID; req.Space is
// required when req.Task is a name.
func (c *Client) Get(
	ctx context.Context,
	req GetRequest,
) (*Task, error) {
	prerelease.Warn("tasks.get", prerelease.Alpha)
	id, err := resolve.FindTaskID(ctx, c.gen, req.Task, req.Space)
	if err != nil {
		return nil, err
	}
	return c.getByID(ctx, id)
}

// CreateEvaluationTask creates a new template_evaluation or code_evaluation
// task and returns it. Exactly one of req.Project or req.Dataset must be set
// (name or ID; req.Space is required when either is a name). At least one
// evaluator is required. Dataset-based tasks require at least one entry in
// req.ExperimentIDs; req.SamplingRate and req.IsContinuous apply only to
// project-based tasks.
func (c *Client) CreateEvaluationTask(
	ctx context.Context,
	req CreateEvaluationTaskRequest,
) (*Task, error) {
	prerelease.Warn("tasks.create_evaluation_task", prerelease.Alpha)
	if req.Name == "" {
		return nil, errors.New("tasks: Name is required")
	}
	if req.Type != TaskTypeTemplateEvaluation && req.Type != TaskTypeCodeEvaluation {
		return nil, fmt.Errorf("tasks: Type must be %q or %q (for %q tasks use CreateRunExperimentTask), got %q",
			TaskTypeTemplateEvaluation, TaskTypeCodeEvaluation, TaskTypeRunExperiment, req.Type)
	}
	if len(req.Evaluators) == 0 {
		return nil, errors.New("tasks: at least one evaluator is required")
	}
	if (req.Project == "") == (req.Dataset == "") {
		return nil, errors.New("tasks: exactly one of Project or Dataset must be set")
	}
	if req.Dataset != "" && len(req.ExperimentIDs) == 0 {
		return nil, errors.New("tasks: dataset-based tasks require at least one entry in ExperimentIDs")
	}
	if req.Project != "" && len(req.ExperimentIDs) > 0 {
		return nil, errors.New("tasks: ExperimentIDs only applies to dataset-based tasks")
	}
	if req.Dataset != "" && (req.SamplingRate != 0 || req.IsContinuous) {
		return nil, errors.New("tasks: SamplingRate and IsContinuous only apply to project-based tasks")
	}

	var projectID, datasetID *string
	if req.Project != "" {
		id, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
		if err != nil {
			return nil, err
		}
		projectID = &id
	}
	if req.Dataset != "" {
		id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
		if err != nil {
			return nil, err
		}
		datasetID = &id
	}

	var experimentIDs *[]string
	if len(req.ExperimentIDs) > 0 {
		experimentIDs = &req.ExperimentIDs
	}

	var body generated.CreateTaskRequestBody
	switch req.Type {
	case TaskTypeTemplateEvaluation:
		if err := body.FromCreateTemplateEvaluationTaskRequest(generated.CreateTemplateEvaluationTaskRequest{
			Type:          generated.CreateTemplateEvaluationTaskRequestTypeTemplateEvaluation,
			Name:          req.Name,
			ProjectId:     projectID,
			DatasetId:     datasetID,
			ExperimentIds: experimentIDs,
			Evaluators:    evaluatorInputs(req.Evaluators),
			SamplingRate:  optfields.PtrIfSet(req.SamplingRate),
			IsContinuous:  optfields.PtrIfSet(req.IsContinuous),
			QueryFilter:   optfields.PtrIfSet(req.QueryFilter),
		}); err != nil {
			return nil, fmt.Errorf("tasks: build template_evaluation body: %w", err)
		}
	case TaskTypeCodeEvaluation:
		if err := body.FromCreateCodeEvaluationTaskRequest(generated.CreateCodeEvaluationTaskRequest{
			Type:          generated.CreateCodeEvaluationTaskRequestTypeCodeEvaluation,
			Name:          req.Name,
			ProjectId:     projectID,
			DatasetId:     datasetID,
			ExperimentIds: experimentIDs,
			Evaluators:    evaluatorInputs(req.Evaluators),
			SamplingRate:  optfields.PtrIfSet(req.SamplingRate),
			IsContinuous:  optfields.PtrIfSet(req.IsContinuous),
			QueryFilter:   optfields.PtrIfSet(req.QueryFilter),
		}); err != nil {
			return nil, fmt.Errorf("tasks: build code_evaluation body: %w", err)
		}
	}
	return c.create(ctx, body)
}

// CreateRunExperimentTask creates a new run_experiment task and returns it.
// req.Dataset accepts a name or ID; req.Space is required when req.Dataset
// is a name. req.RunConfiguration must hold exactly one variant, populated
// via FromLlmGenerationRunConfig or FromTemplateEvaluationRunConfig.
func (c *Client) CreateRunExperimentTask(
	ctx context.Context,
	req CreateRunExperimentTaskRequest,
) (*Task, error) {
	prerelease.Warn("tasks.create_run_experiment_task", prerelease.Alpha)
	if req.Name == "" {
		return nil, errors.New("tasks: Name is required")
	}
	if _, err := req.RunConfiguration.Discriminator(); err != nil {
		return nil, errors.New("tasks: RunConfiguration is required; populate it with FromLlmGenerationRunConfig or FromTemplateEvaluationRunConfig")
	}
	datasetID, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	var body generated.CreateTaskRequestBody
	if err := body.FromCreateRunExperimentTaskRequest(generated.CreateRunExperimentTaskRequest{
		Type:             generated.CreateRunExperimentTaskRequestTypeRunExperiment,
		Name:             req.Name,
		DatasetId:        datasetID,
		RunConfiguration: req.RunConfiguration,
	}); err != nil {
		return nil, fmt.Errorf("tasks: build run_experiment body: %w", err)
	}
	return c.create(ctx, body)
}

// Update updates an existing task and returns it. req.Task accepts a name or
// ID; req.Space is required when req.Task is a name. The SDK fetches the
// task first to determine its type: Name applies to all tasks;
// SamplingRate, IsContinuous, QueryFilter, and Evaluators apply only to
// evaluation tasks; RunConfiguration applies only to run_experiment tasks.
// Leave a patch field nil to preserve its current value; a request with no
// patch fields returns ErrNoUpdateFields.
func (c *Client) Update(
	ctx context.Context,
	req UpdateRequest,
) (*Task, error) {
	prerelease.Warn("tasks.update", prerelease.Alpha)
	if req.Name == nil && req.SamplingRate == nil && req.IsContinuous == nil &&
		req.QueryFilter == nil && len(req.Evaluators) == 0 && req.RunConfiguration == nil {
		return nil, ErrNoUpdateFields
	}
	id, err := resolve.FindTaskID(ctx, c.gen, req.Task, req.Space)
	if err != nil {
		return nil, err
	}
	task, err := c.getByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var body generated.UpdateTaskRequestBody
	switch task.Type {
	case TaskTypeTemplateEvaluation, TaskTypeCodeEvaluation:
		if req.RunConfiguration != nil {
			return nil, fmt.Errorf("tasks: RunConfiguration only applies to %q tasks, task %q is %q",
				TaskTypeRunExperiment, task.Name, task.Type)
		}
		patch := generated.UpdateEvaluationTaskRequest{
			Name:         req.Name,
			SamplingRate: req.SamplingRate,
			IsContinuous: req.IsContinuous,
			QueryFilter:  req.QueryFilter,
		}
		if len(req.Evaluators) > 0 {
			evals := evaluatorInputs(req.Evaluators)
			patch.Evaluators = &evals
		}
		if err := body.FromUpdateEvaluationTaskRequest(patch); err != nil {
			return nil, fmt.Errorf("tasks: build evaluation update body: %w", err)
		}
	case TaskTypeRunExperiment:
		if req.SamplingRate != nil || req.IsContinuous != nil || req.QueryFilter != nil || len(req.Evaluators) > 0 {
			return nil, fmt.Errorf("tasks: SamplingRate, IsContinuous, QueryFilter, and Evaluators only apply to evaluation tasks, task %q is %q",
				task.Name, task.Type)
		}
		if err := body.FromUpdateRunExperimentTaskRequest(generated.UpdateRunExperimentTaskRequest{
			Name:             req.Name,
			RunConfiguration: req.RunConfiguration,
		}); err != nil {
			return nil, fmt.Errorf("tasks: build run_experiment update body: %w", err)
		}
	default:
		return nil, fmt.Errorf("tasks: unknown task type %q", task.Type)
	}

	// Marshal the union by hand: the generated TasksUpdateJSONRequestBody
	// named type does not carry the union's MarshalJSON method and would
	// serialize as {}.
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tasks: marshal update body: %w", err)
	}
	resp, err := c.gen.TasksUpdateWithBodyWithResponse(ctx, id, "application/json", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete irreversibly removes a task and all its associated resources
// (runs, configurations, etc.). req.Task accepts a name or ID; req.Space is
// required when req.Task is a name.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("tasks.delete", prerelease.Alpha)
	id, err := resolve.FindTaskID(ctx, c.gen, req.Task, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.TasksDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// TriggerRun triggers a new run of a task and returns it (initially in
// pending status). req.Task accepts a name or ID; req.Space is required when
// req.Task is a name. The SDK fetches the task first to determine its type:
// DataStartTime, DataEndTime, MaxSpans, OverrideEvaluations, and
// ExperimentIDs apply only to evaluation tasks; ExperimentName (required),
// DatasetVersionID, ExampleIDs, MaxExamples, TracingMetadata, and
// EvaluationTaskIDs apply only to run_experiment tasks.
func (c *Client) TriggerRun(
	ctx context.Context,
	req TriggerRunRequest,
) (*TaskRun, error) {
	prerelease.Warn("tasks.trigger_run", prerelease.Alpha)
	id, err := resolve.FindTaskID(ctx, c.gen, req.Task, req.Space)
	if err != nil {
		return nil, err
	}
	task, err := c.getByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var body generated.TriggerTaskRunRequestBody
	switch task.Type {
	case TaskTypeTemplateEvaluation, TaskTypeCodeEvaluation:
		if req.ExperimentName != "" || req.DatasetVersionID != "" || len(req.ExampleIDs) > 0 ||
			req.MaxExamples != 0 || req.TracingMetadata != nil || len(req.EvaluationTaskIDs) > 0 {
			return nil, fmt.Errorf("tasks: ExperimentName, DatasetVersionID, ExampleIDs, MaxExamples, TracingMetadata, and EvaluationTaskIDs only apply to %q tasks, task %q is %q",
				TaskTypeRunExperiment, req.Task, task.Type)
		}
		var experimentIDs *[]string
		if len(req.ExperimentIDs) > 0 {
			experimentIDs = &req.ExperimentIDs
		}
		if err := body.FromTriggerEvaluationTaskRunRequest(generated.TriggerEvaluationTaskRunRequest{
			DataStartTime:       optfields.PtrIfSet(req.DataStartTime),
			DataEndTime:         optfields.PtrIfSet(req.DataEndTime),
			MaxSpans:            optfields.PtrIfSet(req.MaxSpans),
			OverrideEvaluations: optfields.PtrIfSet(req.OverrideEvaluations),
			ExperimentIds:       experimentIDs,
		}); err != nil {
			return nil, fmt.Errorf("tasks: build evaluation trigger body: %w", err)
		}
	case TaskTypeRunExperiment:
		if !req.DataStartTime.IsZero() || !req.DataEndTime.IsZero() || req.MaxSpans != 0 ||
			req.OverrideEvaluations || len(req.ExperimentIDs) > 0 {
			return nil, fmt.Errorf("tasks: DataStartTime, DataEndTime, MaxSpans, OverrideEvaluations, and ExperimentIDs only apply to evaluation tasks, task %q is %q",
				req.Task, task.Type)
		}
		if req.ExperimentName == "" {
			return nil, fmt.Errorf("tasks: ExperimentName is required to trigger a %q task", TaskTypeRunExperiment)
		}
		if len(req.ExampleIDs) > 0 && req.MaxExamples != 0 {
			return nil, errors.New("tasks: ExampleIDs and MaxExamples are mutually exclusive")
		}
		var exampleIDs *[]string
		if len(req.ExampleIDs) > 0 {
			exampleIDs = &req.ExampleIDs
		}
		var evaluationTaskIDs *[]string
		if len(req.EvaluationTaskIDs) > 0 {
			evaluationTaskIDs = &req.EvaluationTaskIDs
		}
		var tracingMetadata *map[string]string
		if req.TracingMetadata != nil {
			tracingMetadata = &req.TracingMetadata
		}
		if err := body.FromTriggerRunExperimentTaskRunRequest(generated.TriggerRunExperimentTaskRunRequest{
			ExperimentName:    req.ExperimentName,
			DatasetVersionId:  optfields.PtrIfSet(req.DatasetVersionID),
			ExampleIds:        exampleIDs,
			MaxExamples:       optfields.PtrIfSet(req.MaxExamples),
			TracingMetadata:   tracingMetadata,
			EvaluationTaskIds: evaluationTaskIDs,
		}); err != nil {
			return nil, fmt.Errorf("tasks: build run_experiment trigger body: %w", err)
		}
	default:
		return nil, fmt.Errorf("tasks: unknown task type %q", task.Type)
	}

	// Marshal the union by hand: the generated TasksTriggerRunJSONRequestBody
	// named type does not carry the union's MarshalJSON method and would
	// serialize as {}.
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tasks: marshal trigger body: %w", err)
	}
	resp, err := c.gen.TasksTriggerRunWithBodyWithResponse(ctx, id, "application/json", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// ListRuns returns a paginated list of a task's runs, newest first.
// req.Task accepts a name or ID; req.Space is required when req.Task is a
// name.
func (c *Client) ListRuns(
	ctx context.Context,
	req ListRunsRequest,
) (*TaskRunList, error) {
	prerelease.Warn("tasks.list_runs", prerelease.Alpha)
	id, err := resolve.FindTaskID(ctx, c.gen, req.Task, req.Space)
	if err != nil {
		return nil, err
	}
	params := generated.TaskRunsListParams{
		Status: optfields.PtrIfSet(req.Status),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.TaskRunsListWithResponse(ctx, id, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetRun returns a single task run with its current status and statistics.
// Use it to poll a run triggered by TriggerRun (or use WaitForRun).
func (c *Client) GetRun(
	ctx context.Context,
	req GetRunRequest,
) (*TaskRun, error) {
	prerelease.Warn("tasks.get_run", prerelease.Alpha)
	return c.getRunByID(ctx, req.RunID)
}

// CancelRun cancels a pending or running task run and returns it.
func (c *Client) CancelRun(
	ctx context.Context,
	req CancelRunRequest,
) (*TaskRun, error) {
	prerelease.Warn("tasks.cancel_run", prerelease.Alpha)
	resp, err := c.gen.TaskRunsCancelWithResponse(ctx, req.RunID)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// WaitForRun polls a task run until it reaches a terminal state (completed,
// failed, or cancelled) and returns it. It polls every req.PollInterval
// (default 5s) for up to req.Timeout (default 10m); on expiry it returns an
// error wrapping ErrWaitTimeout. Cancelling ctx stops the wait with ctx's
// error.
func (c *Client) WaitForRun(
	ctx context.Context,
	req WaitForRunRequest,
) (*TaskRun, error) {
	prerelease.Warn("tasks.wait_for_run", prerelease.Alpha)
	interval := req.PollInterval
	if interval == 0 {
		interval = defaultPollInterval
	}
	timeout := req.Timeout
	if timeout == 0 {
		timeout = defaultWaitTimeout
	}
	if interval < 0 {
		return nil, fmt.Errorf("tasks: PollInterval must be positive, got %s", interval)
	}
	if timeout < 0 {
		return nil, fmt.Errorf("tasks: Timeout must be positive, got %s", timeout)
	}
	deadline := time.Now().Add(timeout)
	for {
		run, err := c.getRunByID(ctx, req.RunID)
		if err != nil {
			return nil, err
		}
		switch run.Status {
		case TaskRunStatusCompleted, TaskRunStatusFailed, TaskRunStatusCancelled:
			return run, nil
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("tasks: run %s still %s after %s: %w",
				req.RunID, run.Status, timeout, ErrWaitTimeout)
		}
		// Clamp the final wait to the deadline so the last poll lands on it
		// rather than giving up a full interval early (matches Python).
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(min(interval, remaining)):
		}
	}
}

// getByID fetches a task by its resolved ID. Used by Get, and by Update and
// TriggerRun to determine the task type before dispatching on it.
func (c *Client) getByID(ctx context.Context, id string) (*Task, error) {
	resp, err := c.gen.TasksGetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// getRunByID fetches a task run by ID. Shared by GetRun and WaitForRun so
// polling doesn't depend on the public method.
func (c *Client) getRunByID(ctx context.Context, runID string) (*TaskRun, error) {
	resp, err := c.gen.TaskRunsGetWithResponse(ctx, runID)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// create posts an assembled CreateTaskRequestBody union and returns the
// created task. The union is marshaled by hand because the generated
// TasksCreateJSONRequestBody named type does not carry the union's
// MarshalJSON method and would serialize as {}.
func (c *Client) create(ctx context.Context, body generated.CreateTaskRequestBody) (*Task, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tasks: marshal create body: %w", err)
	}
	resp, err := c.gen.TasksCreateWithBodyWithResponse(ctx, "application/json", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// evaluatorInput is the generated bodies' inline evaluator shape, shared by
// the create and update evaluation-task requests.
type evaluatorInput = struct {
	// ColumnMappings Maps evaluator template variable names to data source column names.
	ColumnMappings *map[string]string `json:"column_mappings,omitempty"`

	// EvaluatorId Evaluator identifier (base64). Duplicates are not allowed.
	EvaluatorId string `json:"evaluator_id"`

	// QueryFilter Per-evaluator query filter. Combined with the task-level filter (AND).
	QueryFilter *string `json:"query_filter,omitempty"`
}

// evaluatorInputs translates the public EvaluatorInput slice into the
// generated bodies' inline evaluator shape.
func evaluatorInputs(in []EvaluatorInput) []evaluatorInput {
	out := make([]evaluatorInput, 0, len(in))
	for _, e := range in {
		var columnMappings *map[string]string
		if e.ColumnMappings != nil {
			m := e.ColumnMappings
			columnMappings = &m
		}
		out = append(out, evaluatorInput{
			EvaluatorId:    e.EvaluatorID,
			QueryFilter:    optfields.PtrIfSet(e.QueryFilter),
			ColumnMappings: columnMappings,
		})
	}
	return out
}
