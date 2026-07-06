package experiments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// ErrNoRuns is returned by Create when the request carries no runs. An
// experiment must have at least one run.
var ErrNoRuns = errors.New("experiments: cannot create an experiment without runs")

// Client provides access to the Arize Experiments API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// AppendRuns appends new runs to an existing experiment.
//
// Between 1 and 1000 runs may be appended per request. Each run must include
// ExampleId and Output. Additional user-defined fields can be set via
// AdditionalProperties. The response includes the updated experiment and the
// generated IDs for the inserted runs in input order.
func (c *Client) AppendRuns(ctx context.Context, req AppendRunsRequest) (*ExperimentWithRunIds, error) {
	prerelease.Warn("experiments.append_runs", prerelease.Beta)
	body := generated.InsertExperimentRunsBody{
		ExperimentRuns: req.ExperimentRuns,
	}
	resp, err := c.gen.ExperimentsRunsInsertWithResponse(ctx, req.ExperimentID, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// List returns a paginated list of experiments. req.Dataset, when non-empty,
// accepts a dataset name or ID and restricts results to that dataset.
func (c *Client) List(ctx context.Context, req ListRequest) (*ExperimentList, error) {
	prerelease.Warn("experiments.list", prerelease.Alpha)
	params := generated.ExperimentsListParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrIfSet(req.Limit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	if req.Dataset != "" {
		datasetID, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
		if err != nil {
			return nil, err
		}
		params.DatasetId = &datasetID
	}
	resp, err := c.gen.ExperimentsListWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single experiment, resolving by name or ID. Dataset is
// required when Experiment is a name; Space is required when Dataset is also
// passed as a name.
func (c *Client) Get(ctx context.Context, req GetRequest) (*Experiment, error) {
	prerelease.Warn("experiments.get", prerelease.Alpha)
	id, err := resolve.FindExperimentID(ctx, c.gen, req.Experiment, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.ExperimentsGetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new experiment, resolving the parent dataset by name or ID.
// Runs are transformed from user-named columns (per TaskFields and
// EvaluatorColumns) into the wire format the API expects. An empty req.Runs
// returns ErrNoRuns.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*Experiment, error) {
	prerelease.Warn("experiments.create", prerelease.Alpha)
	if len(req.Runs) == 0 {
		return nil, ErrNoRuns
	}
	runs, err := buildExperimentRuns(req.Runs, req.TaskFields, req.EvaluatorColumns)
	if err != nil {
		return nil, err
	}
	datasetID, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.ExperimentsCreateWithResponse(ctx, generated.ExperimentsCreateJSONRequestBody{
		DatasetId:      datasetID,
		Name:           req.Name,
		ExperimentRuns: runs,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// buildExperimentRuns transforms user-shaped run records into the wire format.
// It validates required columns, encodes non-string outputs as JSON, and
// renames evaluator result columns to the `eval.<name>.<field>` schema.
func buildExperimentRuns(runs []map[string]any, tf TaskFields, evals map[string]EvaluatorFields) ([]ExperimentRunCreate, error) {
	if tf.ExampleID == "" {
		return nil, fmt.Errorf("experiments: TaskFields.ExampleID is required")
	}
	if tf.Output == "" {
		return nil, fmt.Errorf("experiments: TaskFields.Output is required")
	}
	for name, ef := range evals {
		if ef.Score == "" && ef.Label == "" {
			return nil, fmt.Errorf("experiments: evaluator %q: at least Score or Label must be set", name)
		}
	}
	out := make([]ExperimentRunCreate, 0, len(runs))
	for i, run := range runs {
		rec := make(map[string]any, len(run))
		maps.Copy(rec, run)
		rawID, ok := rec[tf.ExampleID]
		if !ok {
			return nil, fmt.Errorf("experiments: run %d missing column %q for ExampleID", i, tf.ExampleID)
		}
		idStr, ok := rawID.(string)
		if !ok {
			return nil, fmt.Errorf("experiments: run %d column %q must be a string, got %T", i, tf.ExampleID, rawID)
		}
		rawOutput, ok := rec[tf.Output]
		if !ok {
			return nil, fmt.Errorf("experiments: run %d missing column %q for Output", i, tf.Output)
		}
		outStr, err := encodeOutput(rawOutput)
		if err != nil {
			return nil, fmt.Errorf("experiments: run %d output: %w", i, err)
		}
		delete(rec, tf.ExampleID)
		delete(rec, tf.Output)
		for evName, ef := range evals {
			if err := remapEvaluatorColumns(rec, i, evName, ef); err != nil {
				return nil, err
			}
		}
		out = append(out, ExperimentRunCreate{
			ExampleId:            idStr,
			Output:               outStr,
			AdditionalProperties: rec,
		})
	}
	return out, nil
}

func remapEvaluatorColumns(rec map[string]any, runIdx int, evName string, ef EvaluatorFields) error {
	move := func(src, dstSuffix string) {
		if src == "" {
			return
		}
		if v, ok := rec[src]; ok {
			rec["eval."+evName+"."+dstSuffix] = v
			delete(rec, src)
		}
	}
	move(ef.Score, "score")
	move(ef.Label, "label")
	move(ef.Explanation, "explanation")
	for metaKey, colName := range ef.Metadata {
		src := colName
		if src == "" {
			src = metaKey
		}
		v, ok := rec[src]
		if !ok {
			return fmt.Errorf("experiments: run %d evaluator %q metadata column %q not found", runIdx, evName, src)
		}
		rec["eval."+evName+".metadata."+metaKey] = v
		delete(rec, src)
	}
	return nil
}

func encodeOutput(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Delete removes an experiment, resolving by name or ID. Dataset is required
// when Experiment is a name; Space is required when Dataset is also passed
// as a name.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("experiments.delete", prerelease.Alpha)
	id, err := resolve.FindExperimentID(ctx, c.gen, req.Experiment, req.Dataset, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.ExperimentsDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ListRuns returns a paginated list of runs for an experiment, resolving the
// experiment by name or ID.
func (c *Client) ListRuns(ctx context.Context, req ListRunsRequest) (*ExperimentRunsList, error) {
	prerelease.Warn("experiments.list_runs", prerelease.Alpha)
	id, err := resolve.FindExperimentID(ctx, c.gen, req.Experiment, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	params := generated.ExperimentsRunsListParams{
		Limit:  optfields.PtrIfSet(req.Limit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ExperimentsRunsListWithResponse(ctx, id, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
