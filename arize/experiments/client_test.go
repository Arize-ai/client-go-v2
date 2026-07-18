package experiments_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/experiments"
)

func testID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Experiment:1:" + suffix))
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *arize.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := arize.NewClient(arize.Config{
		APIKey:    "test-key",
		APIHost:   srv.Listener.Addr().String(),
		APIScheme: "http",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

func TestExperiments(t *testing.T) {
	const experimentID = "exp-abc123"
	expID := testID("exp-1")
	// A base64 "Dataset:..." value so resolve.FindDatasetID treats it as an ID
	// (short-circuits, no lookup call) when used as a list filter.
	datasetID := base64.StdEncoding.EncodeToString([]byte("Dataset:1:ds-1"))

	// State captured across handler + check for body-inspection tests.
	var createTransformsBody map[string]any
	var createStringOutputBody map[string]any
	var createMetadataDefaultsBody map[string]any

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "AppendRuns success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				expectedPath := "/v2/experiments/" + experimentID + "/runs"
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %q, got %q", expectedPath, r.URL.Path)
				}

				// Decode body to verify structure
				var wireBody struct {
					ExperimentRuns []struct {
						ExampleId string `json:"example_id"`
						Output    string `json:"output"`
					} `json:"experiment_runs"`
				}
				if err := json.NewDecoder(r.Body).Decode(&wireBody); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				if len(wireBody.ExperimentRuns) != 2 {
					t.Errorf("expected 2 runs, got %d", len(wireBody.ExperimentRuns))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(experiments.ExperimentWithRunIds{
					RunIds: []string{"run-1", "run-2"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.AppendRuns(ctx, experiments.AppendRunsRequest{
					ExperimentID: experimentID,
					ExperimentRuns: []experiments.ExperimentRunInput{
						{ExampleId: "ex-1", Output: "result-1"},
						{ExampleId: "ex-2", Output: "result-2"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*experiments.ExperimentWithRunIds)
				if len(resp.RunIds) != 2 {
					t.Errorf("expected 2 run IDs, got %d", len(resp.RunIds))
				}
			},
		},
		{
			name: "AppendRuns not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.AppendRuns(ctx, experiments.AppendRunsRequest{
					ExperimentID:   "nonexistent",
					ExperimentRuns: []experiments.ExperimentRunInput{{ExampleId: "ex-1", Output: "out"}},
				})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "AppendRuns unprocessable entity",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(w).Encode(map[string]any{"title": "unprocessable entity", "status": 422})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.AppendRuns(ctx, experiments.AppendRunsRequest{
					ExperimentID:   experimentID,
					ExperimentRuns: []experiments.ExperimentRunInput{{ExampleId: "ex-1", Output: "out"}},
				})
			},
			check: func(t *testing.T, got any, err error) {
				var upe *arize.UnprocessableEntityError
				if !errors.As(err, &upe) {
					t.Errorf("expected *UnprocessableEntityError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "List",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/experiments" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(experiments.ListExperiments{
					Experiments: []experiments.Experiment{
						{Id: "exp-1", Name: "my-experiment", DatasetId: "ds-1", DatasetVersionId: "dv-1"},
					},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.List(ctx, experiments.ListRequest{Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*experiments.ListExperiments)
				if len(resp.Experiments) != 1 {
					t.Errorf("expected 1 experiment, got %d", len(resp.Experiments))
				}
				if resp.Experiments[0].Id != "exp-1" {
					t.Errorf("unexpected id: %s", resp.Experiments[0].Id)
				}
			},
		},
		{
			name: "List_WithCursorAndDatasetID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("cursor") != "next-cursor" {
					t.Errorf("expected cursor query param, got: %s", r.URL.Query().Get("cursor"))
				}
				if r.URL.Query().Get("dataset_id") != datasetID {
					t.Errorf("expected dataset_id query param %q, got: %s", datasetID, r.URL.Query().Get("dataset_id"))
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(experiments.ListExperiments{
					Experiments: []experiments.Experiment{},
					Pagination:  arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.List(ctx, experiments.ListRequest{
					Cursor:  "next-cursor",
					Dataset: datasetID,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Get",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/experiments/"+expID {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(experiments.Experiment{
					Id: expID, Name: "my-experiment", DatasetId: "ds-1", DatasetVersionId: "dv-1",
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Get(ctx, experiments.GetRequest{Experiment: expID})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				exp := got.(*experiments.Experiment)
				if exp.Id != expID {
					t.Errorf("unexpected id: %s", exp.Id)
				}
			},
		},
		{
			name: "Get_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Get(ctx, experiments.GetRequest{Experiment: testID("missing")})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Create",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/v2/experiments" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(experiments.Experiment{
					Id: "exp-new", Name: "new-experiment", DatasetId: "ds-1", DatasetVersionId: "dv-1",
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset: testID("ds-1"),
					Name:    "new-experiment",
					Runs: []map[string]any{
						{"id": "ex-1", "result": "result"},
					},
					TaskFields: experiments.TaskFields{ExampleID: "id", Output: "result"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				exp := got.(*experiments.Experiment)
				if exp.Id != "exp-new" {
					t.Errorf("unexpected id: %s", exp.Id)
				}
			},
		},
		{
			name: "Create_TransformsTaskAndEvaluatorColumns",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&createTransformsBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(experiments.Experiment{Id: "exp-x", Name: "x", DatasetId: "ds-1", DatasetVersionId: "dv-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset: testID("ds-1"),
					Name:    "x",
					Runs: []map[string]any{
						{
							"row_id":    "ex-1",
							"task_out":  map[string]any{"answer": "42"},
							"q_score":   0.9,
							"q_label":   "good",
							"q_explain": "because",
							"my_meta":   "v1",
							"extra":     "kept",
						},
					},
					TaskFields: experiments.TaskFields{ExampleID: "row_id", Output: "task_out"},
					EvaluatorColumns: map[string]experiments.EvaluatorFields{
						"quality": {
							Score: "q_score", Label: "q_label", Explanation: "q_explain",
							Metadata: map[string]string{"version": "my_meta"},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				runs, ok := createTransformsBody["experiment_runs"].([]any)
				if !ok || len(runs) != 1 {
					t.Fatalf("expected 1 run, got %v", createTransformsBody["experiment_runs"])
				}
				run := runs[0].(map[string]any)
				if run["example_id"] != "ex-1" {
					t.Errorf("example_id: want ex-1, got %v", run["example_id"])
				}
				if run["output"] != `{"answer":"42"}` {
					t.Errorf("output: want JSON-encoded map, got %v", run["output"])
				}
				if run["eval.quality.score"] != 0.9 {
					t.Errorf("eval.quality.score: want 0.9, got %v", run["eval.quality.score"])
				}
				if run["eval.quality.label"] != "good" {
					t.Errorf("eval.quality.label: want good, got %v", run["eval.quality.label"])
				}
				if run["eval.quality.explanation"] != "because" {
					t.Errorf("eval.quality.explanation: want because, got %v", run["eval.quality.explanation"])
				}
				if run["eval.quality.metadata.version"] != "v1" {
					t.Errorf("eval.quality.metadata.version: want v1, got %v", run["eval.quality.metadata.version"])
				}
				if run["extra"] != "kept" {
					t.Errorf("extra: want kept, got %v", run["extra"])
				}
				for _, k := range []string{"row_id", "task_out", "q_score", "q_label", "q_explain", "my_meta"} {
					if _, ok := run[k]; ok {
						t.Errorf("expected user column %q to be remapped, but it remains in body", k)
					}
				}
			},
		},
		{
			name: "Create_StringOutputPassedThrough",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&createStringOutputBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(experiments.Experiment{Id: "exp-x"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset: testID("ds-1"),
					Name:    "x",
					Runs: []map[string]any{
						{"id": "ex-1", "out": "plain string"},
					},
					TaskFields: experiments.TaskFields{ExampleID: "id", Output: "out"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				run := createStringOutputBody["experiment_runs"].([]any)[0].(map[string]any)
				if run["output"] != "plain string" {
					t.Errorf("string output should pass through, got %v", run["output"])
				}
			},
		},
		{
			name: "Create_MissingTaskFieldsErrors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called when validation fails")
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset: testID("ds-1"),
					Name:    "x",
					Runs:    []map[string]any{{"id": "ex-1"}},
					// TaskFields intentionally empty
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Fatal("expected error for missing TaskFields")
				}
			},
		},
		{
			name: "Create_MissingColumnInRunErrors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called when validation fails")
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset:    testID("ds-1"),
					Name:       "x",
					Runs:       []map[string]any{{"id": "ex-1"}}, // missing "out"
					TaskFields: experiments.TaskFields{ExampleID: "id", Output: "out"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Fatal("expected error for missing Output column")
				}
			},
		},
		{
			name: "Create_UnmarshalableOutputErrors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called when marshaling fails")
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset:    testID("ds-1"),
					Name:       "x",
					Runs:       []map[string]any{{"id": "ex-1", "out": make(chan int)}},
					TaskFields: experiments.TaskFields{ExampleID: "id", Output: "out"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Fatal("expected marshal error for channel value")
				}
			},
		},
		{
			name: "Create_NoRunsErrors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called when there are no runs")
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset:    testID("ds-1"),
					Name:       "x",
					Runs:       nil,
					TaskFields: experiments.TaskFields{ExampleID: "id", Output: "out"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if !errors.Is(err, experiments.ErrNoRuns) {
					t.Fatalf("expected ErrNoRuns, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Create_EvaluatorWithoutScoreOrLabelErrors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called when validation fails")
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset:    testID("ds-1"),
					Name:       "x",
					Runs:       []map[string]any{{"id": "ex-1", "out": "ok"}},
					TaskFields: experiments.TaskFields{ExampleID: "id", Output: "out"},
					EvaluatorColumns: map[string]experiments.EvaluatorFields{
						"quality": {Explanation: "expl"}, // no Score or Label
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Fatal("expected error when evaluator has neither Score nor Label")
				}
			},
		},
		{
			name: "Create_MetadataDefaultsToKey",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&createMetadataDefaultsBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(experiments.Experiment{Id: "exp-x"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.Create(ctx, experiments.CreateRequest{
					Dataset: testID("ds-1"),
					Name:    "x",
					Runs: []map[string]any{
						{"id": "ex-1", "out": "ok", "score": 1.0, "version": "v1"},
					},
					TaskFields: experiments.TaskFields{ExampleID: "id", Output: "out"},
					EvaluatorColumns: map[string]experiments.EvaluatorFields{
						"quality": {
							Score:    "score",
							Metadata: map[string]string{"version": ""}, // empty -> use key as column
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				run := createMetadataDefaultsBody["experiment_runs"].([]any)[0].(map[string]any)
				if run["eval.quality.metadata.version"] != "v1" {
					t.Errorf("eval.quality.metadata.version: want v1, got %v", run["eval.quality.metadata.version"])
				}
			},
		},
		{
			name: "Delete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.URL.Path != "/v2/experiments/"+expID {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusNoContent)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Experiments.Delete(ctx, experiments.DeleteRequest{Experiment: expID})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Delete_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Experiments.Delete(ctx, experiments.DeleteRequest{Experiment: testID("missing")})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "ListRuns",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/experiments/"+expID+"/runs" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				out := "output-1"
				json.NewEncoder(w).Encode(experiments.ListExperimentRuns{
					ExperimentRuns: []experiments.ExperimentRun{
						{Output: &out},
					},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Experiments.ListRuns(ctx, experiments.ListRunsRequest{Experiment: expID, Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				runs := got.(*experiments.ListExperimentRuns)
				if len(runs.ExperimentRuns) != 1 {
					t.Errorf("expected 1 run, got %d", len(runs.ExperimentRuns))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, tt.handler)
			got, err := tt.invoke(context.Background(), client)
			tt.check(t, got, err)
		})
	}
}
