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

// ErrNoDatasetOrSpace is returned by Create when neither Dataset nor Space is
// set. An experiment is created either against a dataset or directly in a
// space, so exactly one of them is required.
var ErrNoDatasetOrSpace = errors.New("experiments: either Dataset or Space is required")

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
// Output; ExampleId is required only when the target experiment is associated
// with a dataset, and must be left nil when it has none. Additional
// user-defined fields can be set via AdditionalProperties. The response
// includes the updated experiment and the generated IDs for the inserted runs
// in input order.
func (c *Client) AppendRuns(ctx context.Context, req AppendRunsRequest) (*ExperimentWithRunIds, error) {
	prerelease.Warn("experiments.append_runs", prerelease.Beta)
	body := generated.InsertExperimentRunsRequest{
		ExperimentRuns: req.ExperimentRuns,
	}
	resp, err := c.gen.InsertExperimentRunsWithResponse(ctx, req.ExperimentID, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// List returns a paginated list of experiments, narrowed by whichever scope is
// given:
//
//   - req.Dataset — only experiments run on that dataset.
//   - req.Space — every experiment in that space, both those associated with a
//     dataset and those without one.
//   - neither — every experiment across all spaces the caller can read.
//
// Both accept a name or an ID. Passing both applies the narrower dataset scope,
// with req.Space used only to resolve a dataset name.
func (c *Client) List(ctx context.Context, req ListRequest) (*ListExperiments, error) {
	prerelease.Warn("experiments.list", prerelease.Beta)
	params := generated.ListExperimentsParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrIfSet(req.Limit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	// The endpoint rejects dataset_id and space_id together, so send at most
	// one: dataset is the narrower scope and wins, leaving space to resolve a
	// dataset name.
	if req.Dataset != "" {
		datasetID, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
		if err != nil {
			return nil, err
		}
		params.DatasetId = &datasetID
	} else if req.Space != "" {
		spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
		if err != nil {
			return nil, err
		}
		params.SpaceId = &spaceID
	}
	resp, err := c.gen.ListExperimentsWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single experiment, resolving by name or ID. Either Dataset or
// Space is required when Experiment is a name; Space is also required when
// Dataset is passed as a name. Resolving by Space alone returns
// AmbiguousNameError when the name matches more than one experiment in that
// space — pass Dataset or an experiment ID to disambiguate.
func (c *Client) Get(ctx context.Context, req GetRequest) (*Experiment, error) {
	prerelease.Warn("experiments.get", prerelease.Beta)
	id, err := resolve.FindExperimentID(ctx, c.gen, req.Experiment, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetExperimentWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new experiment, either against a dataset or directly in a
// space. Exactly one of req.Dataset or req.Space is required; both are accepted
// as a name or an ID, and passing neither returns ErrNoDatasetOrSpace. Runs are
// transformed from user-named columns (per TaskFields and EvaluatorColumns)
// into the wire format the API expects. An empty req.Runs returns ErrNoRuns.
//
// TaskFields.ExampleID is required only for a dataset-backed experiment, whose
// runs reference that dataset's examples. An experiment created in a space has
// no examples to reference, so its runs carry only an output.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*Experiment, error) {
	prerelease.Warn("experiments.create", prerelease.Beta)
	if len(req.Runs) == 0 {
		return nil, ErrNoRuns
	}
	if req.Dataset == "" && req.Space == "" {
		return nil, ErrNoDatasetOrSpace
	}
	// The endpoint rejects dataset_id and space_id together, so send exactly
	// one: dataset is the narrower scope and wins, leaving space to resolve a
	// dataset name.
	var datasetID, spaceID *string
	if req.Dataset != "" {
		id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
		if err != nil {
			return nil, err
		}
		datasetID = &id
	} else {
		id, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
		if err != nil {
			return nil, err
		}
		spaceID = &id
	}
	runs, err := buildExperimentRuns(req.Runs, req.TaskFields, req.EvaluatorColumns, datasetID != nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.CreateExperimentWithResponse(ctx, generated.CreateExperimentJSONRequestBody{
		DatasetId:      datasetID,
		SpaceId:        spaceID,
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
//
// requireExampleID mirrors the server's own per-run validation: an example ID
// links a run to a dataset example, so it's required only for a dataset-backed
// experiment. When false, TaskFields.ExampleID may be left empty and no
// example_id is emitted; naming a column anyway still maps it.
func buildExperimentRuns(runs []map[string]any, tf TaskFields, evals map[string]EvaluatorFields, requireExampleID bool) ([]ExperimentRunInput, error) {
	if requireExampleID && tf.ExampleID == "" {
		return nil, fmt.Errorf("experiments: TaskFields.ExampleID is required for a dataset-backed experiment")
	}
	if tf.Output == "" {
		return nil, fmt.Errorf("experiments: TaskFields.Output is required")
	}
	for name, ef := range evals {
		if ef.Score == "" && ef.Label == "" {
			return nil, fmt.Errorf("experiments: evaluator %q: at least Score or Label must be set", name)
		}
	}
	out := make([]ExperimentRunInput, 0, len(runs))
	for i, run := range runs {
		rec := make(map[string]any, len(run))
		maps.Copy(rec, run)
		// An experiment with no dataset has no example to point at, so the
		// column is only looked up when the caller named one.
		var exampleID *string
		if tf.ExampleID != "" {
			rawID, ok := rec[tf.ExampleID]
			if !ok {
				return nil, fmt.Errorf("experiments: run %d missing column %q for ExampleID", i, tf.ExampleID)
			}
			idStr, ok := rawID.(string)
			if !ok {
				return nil, fmt.Errorf("experiments: run %d column %q must be a string, got %T", i, tf.ExampleID, rawID)
			}
			exampleID = &idStr
		}
		rawOutput, ok := rec[tf.Output]
		if !ok {
			return nil, fmt.Errorf("experiments: run %d missing column %q for Output", i, tf.Output)
		}
		outStr, err := encodeOutput(rawOutput)
		if err != nil {
			return nil, fmt.Errorf("experiments: run %d output: %w", i, err)
		}
		if tf.ExampleID != "" {
			delete(rec, tf.ExampleID)
		}
		delete(rec, tf.Output)
		for evName, ef := range evals {
			if err := remapEvaluatorColumns(rec, i, evName, ef); err != nil {
				return nil, err
			}
		}
		out = append(out, ExperimentRunInput{
			ExampleId:            exampleID,
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

// Delete removes an experiment, resolving by name or ID. Either Dataset or
// Space is required when Experiment is a name; Space is also required when
// Dataset is passed as a name. Resolving by Space alone returns
// AmbiguousNameError when the name matches more than one experiment in that
// space — pass Dataset or an experiment ID to disambiguate.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("experiments.delete", prerelease.Beta)
	id, err := resolve.FindExperimentID(ctx, c.gen, req.Experiment, req.Dataset, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.DeleteExperimentWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ListRuns returns a paginated list of runs for an experiment, resolving the
// experiment by name or ID. Either Dataset or Space is required when Experiment
// is a name. Runs of an experiment with no dataset carry no example ID.
func (c *Client) ListRuns(ctx context.Context, req ListRunsRequest) (*ListExperimentRuns, error) {
	prerelease.Warn("experiments.list_runs", prerelease.Beta)
	id, err := resolve.FindExperimentID(ctx, c.gen, req.Experiment, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	params := generated.ListExperimentRunsParams{
		Limit:  optfields.PtrIfSet(req.Limit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListExperimentRunsWithResponse(ctx, id, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
