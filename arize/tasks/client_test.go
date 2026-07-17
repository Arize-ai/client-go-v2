package tasks_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/tasks"
)

// testID returns a base64-encoded ID so resolve.IsResourceID treats it as an
// ID and skips name-resolution lookups during tests.
func testID(prefix, suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte(prefix + ":1:" + suffix))
}

func taskID(suffix string) string    { return testID("Task", suffix) }
func taskRunID(suffix string) string { return testID("TaskRun", suffix) }
func spaceID(suffix string) string   { return testID("Space", suffix) }
func projectID(suffix string) string { return testID("Project", suffix) }
func datasetID(suffix string) string { return testID("Dataset", suffix) }

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *arize.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := arize.NewClient(arize.Config{
		APIKey:    "test-key",
		APIHost:   srv.Listener.Addr().String(),
		APIScheme: "http",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return srv, client
}

// wireEvaluator mirrors the JSON shape of an evaluator entry in create and
// update request bodies.
type wireEvaluator struct {
	EvaluatorId    string            `json:"evaluator_id"`
	QueryFilter    *string           `json:"query_filter"`
	ColumnMappings map[string]string `json:"column_mappings"`
}

// wireCreateEval mirrors the JSON shape of the CreateEvaluationTask request
// body.
type wireCreateEval struct {
	Type          string          `json:"type"`
	Name          string          `json:"name"`
	ProjectId     *string         `json:"project_id"`
	DatasetId     *string         `json:"dataset_id"`
	ExperimentIds []string        `json:"experiment_ids"`
	Evaluators    []wireEvaluator `json:"evaluators"`
	SamplingRate  *float64        `json:"sampling_rate"`
	IsContinuous  *bool           `json:"is_continuous"`
	QueryFilter   *string         `json:"query_filter"`
}

// wireCreateRunExp mirrors the JSON shape of the CreateRunExperimentTask
// request body.
type wireCreateRunExp struct {
	Type             string         `json:"type"`
	Name             string         `json:"name"`
	DatasetId        string         `json:"dataset_id"`
	RunConfiguration map[string]any `json:"run_configuration"`
}

// wireUpdateEval mirrors the JSON shape of the evaluation-task Update body.
type wireUpdateEval struct {
	Name         *string          `json:"name"`
	SamplingRate *float64         `json:"sampling_rate"`
	IsContinuous *bool            `json:"is_continuous"`
	QueryFilter  *string          `json:"query_filter"`
	Evaluators   *[]wireEvaluator `json:"evaluators"`
}

// wireUpdateRunExp mirrors the JSON shape of the run_experiment Update body.
type wireUpdateRunExp struct {
	Name             *string        `json:"name"`
	RunConfiguration map[string]any `json:"run_configuration"`
}

// wireTriggerEval mirrors the JSON shape of the evaluation-task TriggerRun
// body.
type wireTriggerEval struct {
	DataStartTime       *time.Time `json:"data_start_time"`
	DataEndTime         *time.Time `json:"data_end_time"`
	MaxSpans            *int       `json:"max_spans"`
	OverrideEvaluations *bool      `json:"override_evaluations"`
	ExperimentIds       []string   `json:"experiment_ids"`
}

// wireTriggerRunExp mirrors the JSON shape of the run_experiment TriggerRun
// body.
type wireTriggerRunExp struct {
	ExperimentName    string            `json:"experiment_name"`
	DatasetVersionId  *string           `json:"dataset_version_id"`
	ExampleIds        []string          `json:"example_ids"`
	MaxExamples       *int              `json:"max_examples"`
	TracingMetadata   map[string]string `json:"tracing_metadata"`
	EvaluationTaskIds []string          `json:"evaluation_task_ids"`
}

// serveTask writes a Task with the given ID and type, for handlers backing
// the GET that Update and TriggerRun issue to dispatch on the task type.
func serveTask(t *testing.T, w http.ResponseWriter, id string, taskType tasks.TaskType) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks.Task{
		Id:        id,
		Name:      "my-task",
		Type:      taskType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); err != nil {
		t.Errorf("encode task: %v", err)
	}
}

func templateRunConfiguration(t *testing.T) tasks.RunConfiguration {
	t.Helper()
	var rc tasks.RunConfiguration
	if err := rc.FromTemplateEvaluationRunConfig(tasks.TemplateEvaluationRunConfig{
		AiIntegrationId:    "ai-1",
		Template:           "Is {{output}} correct?",
		ProvideExplanation: true,
	}); err != nil {
		t.Fatalf("build run configuration: %v", err)
	}
	return rc
}

func TestTasks(t *testing.T) {
	tID := taskID("task-1")
	runID := taskRunID("run-1")

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/tasks" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if got := r.URL.Query().Get("limit"); got != "50" {
					t.Errorf("limit query: want default 50, got %q", got)
				}
				next := "cursor-next"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tasks.TaskList{
					Tasks:      []tasks.Task{{Id: tID, Name: "my-task", Type: tasks.TaskTypeTemplateEvaluation}},
					Pagination: arize.PaginationMetadata{HasMore: true, NextCursor: &next},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.List(ctx, tasks.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*tasks.TaskList)
				if len(resp.Tasks) != 1 || resp.Tasks[0].Name != "my-task" {
					t.Errorf("unexpected tasks: %+v", resp.Tasks)
				}
				if !resp.Pagination.HasMore || resp.Pagination.NextCursor == nil || *resp.Pagination.NextCursor != "cursor-next" {
					t.Errorf("unexpected pagination: %+v", resp.Pagination)
				}
			},
		},
		{
			name: "List_Filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got, want := q.Get("space_id"), spaceID("sp-1"); got != want {
					t.Errorf("space_id query: want %q, got %q", want, got)
				}
				if q.Get("space_name") != "" {
					t.Errorf("space_name should be empty for an ID, got %q", q.Get("space_name"))
				}
				if got, want := q.Get("project_id"), projectID("p-1"); got != want {
					t.Errorf("project_id query: want %q, got %q", want, got)
				}
				if got, want := q.Get("dataset_id"), datasetID("ds-1"); got != want {
					t.Errorf("dataset_id query: want %q, got %q", want, got)
				}
				if got := q.Get("name"); got != "eval" {
					t.Errorf("name query: want eval, got %q", got)
				}
				if got := q.Get("type"); got != "TEMPLATE_EVALUATION" {
					t.Errorf("type query: want TEMPLATE_EVALUATION, got %q", got)
				}
				if got := q.Get("limit"); got != "25" {
					t.Errorf("limit query: want 25, got %q", got)
				}
				if got := q.Get("cursor"); got != "cursor-abc" {
					t.Errorf("cursor query: want cursor-abc, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tasks.TaskList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.List(ctx, tasks.ListRequest{
					Space:   spaceID("sp-1"),
					Project: projectID("p-1"),
					Dataset: datasetID("ds-1"),
					Name:    "eval",
					Type:    tasks.TaskTypeTemplateEvaluation,
					Limit:   25,
					Cursor:  "cursor-abc",
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
				if r.URL.Path != "/v2/tasks/"+tID {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				serveTask(t, w, tID, tasks.TaskTypeTemplateEvaluation)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.Get(ctx, tasks.GetRequest{Task: tID})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*tasks.Task).Name != "my-task" {
					t.Errorf("unexpected name: %s", got.(*tasks.Task).Name)
				}
			},
		},
		{
			// Exercises resolve.FindTaskID: a plain Task name (with a Space)
			// first lists tasks filtered by name to discover the ID, then
			// issues the real GET.
			name: "Get_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks":
					if got := r.URL.Query().Get("name"); got != "my-task" {
						t.Errorf("resolver list name query: want my-task, got %q", got)
					}
					if got, want := r.URL.Query().Get("space_id"), spaceID("sp-1"); got != want {
						t.Errorf("resolver list space_id query: want %q, got %q", want, got)
					}
					json.NewEncoder(w).Encode(tasks.TaskList{
						Tasks:      []tasks.Task{{Id: tID, Name: "my-task", Type: tasks.TaskTypeTemplateEvaluation}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID:
					serveTask(t, w, tID, tasks.TaskTypeTemplateEvaluation)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.Get(ctx, tasks.GetRequest{Task: "my-task", Space: spaceID("sp-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*tasks.Task).Id != tID {
					t.Errorf("unexpected id: %s", got.(*tasks.Task).Id)
				}
			},
		},
		{
			name: "Get_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.Get(ctx, tasks.GetRequest{Task: taskID("missing")})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "CreateEvaluationTask_ProjectBased",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v2/tasks" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				var body wireCreateEval
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Type != "TEMPLATE_EVALUATION" {
					t.Errorf("body type: want TEMPLATE_EVALUATION, got %q", body.Type)
				}
				if body.Name != "eval-task" {
					t.Errorf("body name: want eval-task, got %q", body.Name)
				}
				if body.ProjectId == nil || *body.ProjectId != projectID("p-1") {
					t.Errorf("body project_id: want %q, got %v", projectID("p-1"), body.ProjectId)
				}
				if body.DatasetId != nil {
					t.Errorf("body dataset_id should be omitted, got %v", *body.DatasetId)
				}
				if len(body.Evaluators) != 1 || body.Evaluators[0].EvaluatorId != "ev-1" {
					t.Errorf("unexpected evaluators: %+v", body.Evaluators)
				}
				if body.Evaluators[0].QueryFilter == nil || *body.Evaluators[0].QueryFilter != "attributes.llm" {
					t.Errorf("unexpected evaluator query_filter: %v", body.Evaluators[0].QueryFilter)
				}
				if body.Evaluators[0].ColumnMappings["question"] != "attributes.input" {
					t.Errorf("unexpected evaluator column_mappings: %v", body.Evaluators[0].ColumnMappings)
				}
				if body.SamplingRate == nil || *body.SamplingRate != 0.5 {
					t.Errorf("body sampling_rate: want 0.5, got %v", body.SamplingRate)
				}
				if body.IsContinuous == nil || !*body.IsContinuous {
					t.Errorf("body is_continuous: want true, got %v", body.IsContinuous)
				}
				if body.QueryFilter == nil || *body.QueryFilter != "status_code = 'OK'" {
					t.Errorf("body query_filter: want status_code filter, got %v", body.QueryFilter)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				serveTask(t, w, tID, tasks.TaskTypeTemplateEvaluation)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.CreateEvaluationTask(ctx, tasks.CreateEvaluationTaskRequest{
					Name: "eval-task",
					Type: tasks.TaskTypeTemplateEvaluation,
					Evaluators: []tasks.EvaluatorInput{{
						EvaluatorID:    "ev-1",
						QueryFilter:    "attributes.llm",
						ColumnMappings: map[string]string{"question": "attributes.input"},
					}},
					Project:      projectID("p-1"),
					SamplingRate: 0.5,
					IsContinuous: true,
					QueryFilter:  "status_code = 'OK'",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*tasks.Task).Id != tID {
					t.Errorf("unexpected id: %s", got.(*tasks.Task).Id)
				}
			},
		},
		{
			name: "CreateEvaluationTask_DatasetBased",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireCreateEval
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Type != "CODE_EVALUATION" {
					t.Errorf("body type: want code_evaluation, got %q", body.Type)
				}
				if body.DatasetId == nil || *body.DatasetId != datasetID("ds-1") {
					t.Errorf("body dataset_id: want %q, got %v", datasetID("ds-1"), body.DatasetId)
				}
				if body.ProjectId != nil {
					t.Errorf("body project_id should be omitted, got %v", *body.ProjectId)
				}
				if len(body.ExperimentIds) != 1 || body.ExperimentIds[0] != "exp-1" {
					t.Errorf("unexpected experiment_ids: %v", body.ExperimentIds)
				}
				if body.SamplingRate != nil {
					t.Errorf("body sampling_rate should be omitted, got %v", *body.SamplingRate)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				serveTask(t, w, tID, tasks.TaskTypeCodeEvaluation)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.CreateEvaluationTask(ctx, tasks.CreateEvaluationTaskRequest{
					Name:          "code-task",
					Type:          tasks.TaskTypeCodeEvaluation,
					Evaluators:    []tasks.EvaluatorInput{{EvaluatorID: "ev-1"}},
					Dataset:       datasetID("ds-1"),
					ExperimentIDs: []string{"exp-1"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "CreateRunExperimentTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v2/tasks" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				var body wireCreateRunExp
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Type != "RUN_EXPERIMENT" {
					t.Errorf("body type: want run_experiment, got %q", body.Type)
				}
				if body.Name != "exp-task" {
					t.Errorf("body name: want exp-task, got %q", body.Name)
				}
				if body.DatasetId != datasetID("ds-1") {
					t.Errorf("body dataset_id: want %q, got %q", datasetID("ds-1"), body.DatasetId)
				}
				if got := body.RunConfiguration["experiment_type"]; got != "TEMPLATE_EVALUATION" {
					t.Errorf("run_configuration experiment_type: want TEMPLATE_EVALUATION, got %v", got)
				}
				if got := body.RunConfiguration["template"]; got != "Is {{output}} correct?" {
					t.Errorf("run_configuration template: got %v", got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				serveTask(t, w, tID, tasks.TaskTypeRunExperiment)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.CreateRunExperimentTask(ctx, tasks.CreateRunExperimentTaskRequest{
					Name:             "exp-task",
					Dataset:          datasetID("ds-1"),
					RunConfiguration: templateRunConfiguration(t),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*tasks.Task).Type != tasks.TaskTypeRunExperiment {
					t.Errorf("unexpected type: %s", got.(*tasks.Task).Type)
				}
			},
		},
		{
			name: "Update_EvaluationTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID:
					serveTask(t, w, tID, tasks.TaskTypeTemplateEvaluation)
				case r.Method == http.MethodPatch && r.URL.Path == "/v2/tasks/"+tID:
					var body wireUpdateEval
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					if body.Name == nil || *body.Name != "renamed" {
						t.Errorf("body name: want renamed, got %v", body.Name)
					}
					if body.SamplingRate == nil || *body.SamplingRate != 0.25 {
						t.Errorf("body sampling_rate: want 0.25, got %v", body.SamplingRate)
					}
					// A pointer to "" clears the filter: it must be present
					// (and empty) on the wire, not omitted.
					if body.QueryFilter == nil || *body.QueryFilter != "" {
						t.Errorf("body query_filter: want present and empty, got %v", body.QueryFilter)
					}
					if body.Evaluators == nil || len(*body.Evaluators) != 1 || (*body.Evaluators)[0].EvaluatorId != "ev-2" {
						t.Errorf("unexpected evaluators: %v", body.Evaluators)
					}
					serveTask(t, w, tID, tasks.TaskTypeTemplateEvaluation)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "renamed"
				rate := float32(0.25)
				clearFilter := ""
				return c.Tasks.Update(ctx, tasks.UpdateRequest{
					Task:         tID,
					Name:         &name,
					SamplingRate: &rate,
					QueryFilter:  &clearFilter,
					Evaluators:   []tasks.EvaluatorInput{{EvaluatorID: "ev-2"}},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Update_RunExperimentTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID:
					serveTask(t, w, tID, tasks.TaskTypeRunExperiment)
				case r.Method == http.MethodPatch && r.URL.Path == "/v2/tasks/"+tID:
					var body wireUpdateRunExp
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					if body.Name == nil || *body.Name != "renamed" {
						t.Errorf("body name: want renamed, got %v", body.Name)
					}
					if got := body.RunConfiguration["experiment_type"]; got != "TEMPLATE_EVALUATION" {
						t.Errorf("run_configuration experiment_type: want TEMPLATE_EVALUATION, got %v", got)
					}
					serveTask(t, w, tID, tasks.TaskTypeRunExperiment)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "renamed"
				rc := templateRunConfiguration(t)
				return c.Tasks.Update(ctx, tasks.UpdateRequest{
					Task:             tID,
					Name:             &name,
					RunConfiguration: &rc,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// Update rejects an empty patch client-side, before any request.
			name: "Update_NoFields",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.Update(ctx, tasks.UpdateRequest{Task: tID})
			},
			check: func(t *testing.T, got any, err error) {
				if !errors.Is(err, tasks.ErrNoUpdateFields) {
					t.Fatalf("expected ErrNoUpdateFields, got %T: %v", err, err)
				}
			},
		},
		{
			// RunConfiguration on an evaluation task fails after the type
			// lookup, without issuing the PATCH.
			name: "Update_RunConfigurationOnEvaluationTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID {
					serveTask(t, w, tID, tasks.TaskTypeCodeEvaluation)
					return
				}
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				rc := templateRunConfiguration(t)
				return c.Tasks.Update(ctx, tasks.UpdateRequest{Task: tID, RunConfiguration: &rc})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil || !strings.Contains(err.Error(), "RunConfiguration") {
					t.Fatalf("expected RunConfiguration type-mismatch error, got %v", err)
				}
			},
		},
		{
			// Evaluation-only fields on a run_experiment task fail after the
			// type lookup, without issuing the PATCH.
			name: "Update_QueryFilterOnRunExperimentTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID {
					serveTask(t, w, tID, tasks.TaskTypeRunExperiment)
					return
				}
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				filter := "status_code = 'OK'"
				return c.Tasks.Update(ctx, tasks.UpdateRequest{Task: tID, QueryFilter: &filter})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil || !strings.Contains(err.Error(), "only apply to evaluation tasks") {
					t.Fatalf("expected evaluation-only type-mismatch error, got %v", err)
				}
			},
		},
		{
			name: "Delete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.URL.Path != "/v2/tasks/"+tID {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Tasks.Delete(ctx, tasks.DeleteRequest{Task: tID})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "TriggerRun_EvaluationTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID:
					serveTask(t, w, tID, tasks.TaskTypeTemplateEvaluation)
				case r.Method == http.MethodPost && r.URL.Path == "/v2/tasks/"+tID+"/trigger":
					var body wireTriggerEval
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					wantStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
					if body.DataStartTime == nil || !body.DataStartTime.Equal(wantStart) {
						t.Errorf("body data_start_time: want %s, got %v", wantStart, body.DataStartTime)
					}
					if body.DataEndTime != nil {
						t.Errorf("body data_end_time should be omitted, got %v", *body.DataEndTime)
					}
					if body.MaxSpans == nil || *body.MaxSpans != 500 {
						t.Errorf("body max_spans: want 500, got %v", body.MaxSpans)
					}
					if body.OverrideEvaluations == nil || !*body.OverrideEvaluations {
						t.Errorf("body override_evaluations: want true, got %v", body.OverrideEvaluations)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(201)
					json.NewEncoder(w).Encode(tasks.TaskRun{Id: runID, TaskId: tID, Status: tasks.TaskRunStatusPending})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{
					Task:                tID,
					DataStartTime:       time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
					MaxSpans:            500,
					OverrideEvaluations: true,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				run := got.(*tasks.TaskRun)
				if run.Status != tasks.TaskRunStatusPending {
					t.Errorf("unexpected status: %s", run.Status)
				}
			},
		},
		{
			name: "TriggerRun_RunExperimentTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID:
					serveTask(t, w, tID, tasks.TaskTypeRunExperiment)
				case r.Method == http.MethodPost && r.URL.Path == "/v2/tasks/"+tID+"/trigger":
					var body wireTriggerRunExp
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					if body.ExperimentName != "run-1" {
						t.Errorf("body experiment_name: want run-1, got %q", body.ExperimentName)
					}
					if body.DatasetVersionId == nil || *body.DatasetVersionId != "v-1" {
						t.Errorf("body dataset_version_id: want v-1, got %v", body.DatasetVersionId)
					}
					if body.MaxExamples == nil || *body.MaxExamples != 10 {
						t.Errorf("body max_examples: want 10, got %v", body.MaxExamples)
					}
					if body.TracingMetadata["env"] != "test" {
						t.Errorf("unexpected tracing_metadata: %v", body.TracingMetadata)
					}
					if len(body.EvaluationTaskIds) != 1 || body.EvaluationTaskIds[0] != taskID("eval-1") {
						t.Errorf("unexpected evaluation_task_ids: %v", body.EvaluationTaskIds)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(201)
					json.NewEncoder(w).Encode(tasks.TaskRun{Id: runID, TaskId: tID, Status: tasks.TaskRunStatusPending})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{
					Task:              tID,
					ExperimentName:    "run-1",
					DatasetVersionID:  "v-1",
					MaxExamples:       10,
					TracingMetadata:   map[string]string{"env": "test"},
					EvaluationTaskIDs: []string{taskID("eval-1")},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// ExperimentName is required for run_experiment tasks; the error
			// fires after the type lookup, without issuing the trigger.
			name: "TriggerRun_MissingExperimentName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID {
					serveTask(t, w, tID, tasks.TaskTypeRunExperiment)
					return
				}
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{Task: tID})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil || !strings.Contains(err.Error(), "ExperimentName is required") {
					t.Fatalf("expected ExperimentName-required error, got %v", err)
				}
			},
		},
		{
			name: "TriggerRun_ExampleIDsAndMaxExamples",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID {
					serveTask(t, w, tID, tasks.TaskTypeRunExperiment)
					return
				}
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{
					Task:           tID,
					ExperimentName: "run-1",
					ExampleIDs:     []string{"ex-1"},
					MaxExamples:    10,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
					t.Fatalf("expected mutual-exclusivity error, got %v", err)
				}
			},
		},
		{
			// Evaluation-only fields on a run_experiment task fail after the
			// type lookup, without issuing the trigger.
			name: "TriggerRun_EvalFieldsOnRunExperimentTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID {
					serveTask(t, w, tID, tasks.TaskTypeRunExperiment)
					return
				}
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{
					Task:           tID,
					ExperimentName: "run-1",
					MaxSpans:       100,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil || !strings.Contains(err.Error(), "only apply to evaluation tasks") {
					t.Fatalf("expected evaluation-only type-mismatch error, got %v", err)
				}
			},
		},
		{
			// Run-experiment-only fields on an evaluation task fail after the
			// type lookup, without issuing the trigger.
			name: "TriggerRun_RunExperimentFieldsOnEvaluationTask",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/v2/tasks/"+tID {
					serveTask(t, w, tID, tasks.TaskTypeTemplateEvaluation)
					return
				}
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{
					Task:           tID,
					ExperimentName: "run-1",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil || !strings.Contains(err.Error(), "only apply to") {
					t.Fatalf("expected run-experiment-only type-mismatch error, got %v", err)
				}
			},
		},
		{
			name: "ListRuns",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/tasks/"+tID+"/runs" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				q := r.URL.Query()
				if got := q.Get("status"); got != "COMPLETED" {
					t.Errorf("status query: want completed, got %q", got)
				}
				if got := q.Get("limit"); got != "50" {
					t.Errorf("limit query: want default 50, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tasks.TaskRunList{
					TaskRuns:   []tasks.TaskRun{{Id: runID, TaskId: tID, Status: tasks.TaskRunStatusCompleted}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.ListRuns(ctx, tasks.ListRunsRequest{
					Task:   tID,
					Status: tasks.TaskRunStatusCompleted,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*tasks.TaskRunList)
				if len(resp.TaskRuns) != 1 || resp.TaskRuns[0].Id != runID {
					t.Errorf("unexpected runs: %+v", resp.TaskRuns)
				}
			},
		},
		{
			name: "GetRun",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/task-runs/"+runID {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tasks.TaskRun{Id: runID, TaskId: tID, Status: tasks.TaskRunStatusRunning})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.GetRun(ctx, tasks.GetRunRequest{RunID: runID})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*tasks.TaskRun).Status != tasks.TaskRunStatusRunning {
					t.Errorf("unexpected status: %s", got.(*tasks.TaskRun).Status)
				}
			},
		},
		{
			name: "CancelRun",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/v2/task-runs/"+runID+"/cancel" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tasks.TaskRun{Id: runID, TaskId: tID, Status: tasks.TaskRunStatusCancelled})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Tasks.CancelRun(ctx, tasks.CancelRunRequest{RunID: runID})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*tasks.TaskRun).Status != tasks.TaskRunStatusCancelled {
					t.Errorf("unexpected status: %s", got.(*tasks.TaskRun).Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, tt.handler)
			got, err := tt.invoke(context.Background(), client)
			tt.check(t, got, err)
		})
	}
}

// TestTasks_CreateEvaluationTask_Validation covers the client-side checks
// that reject a create before any request is sent.
func TestTasks_CreateEvaluationTask_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     tasks.CreateEvaluationTaskRequest
		wantErr string
	}{
		{
			name: "MissingName",
			req: tasks.CreateEvaluationTaskRequest{
				Type:       tasks.TaskTypeTemplateEvaluation,
				Evaluators: []tasks.EvaluatorInput{{EvaluatorID: "ev-1"}},
				Project:    projectID("p-1"),
			},
			wantErr: "Name is required",
		},
		{
			name: "WrongType",
			req: tasks.CreateEvaluationTaskRequest{
				Name:       "t",
				Type:       tasks.TaskTypeRunExperiment,
				Evaluators: []tasks.EvaluatorInput{{EvaluatorID: "ev-1"}},
				Project:    projectID("p-1"),
			},
			wantErr: "CreateRunExperimentTask",
		},
		{
			name: "NoEvaluators",
			req: tasks.CreateEvaluationTaskRequest{
				Name:    "t",
				Type:    tasks.TaskTypeTemplateEvaluation,
				Project: projectID("p-1"),
			},
			wantErr: "at least one evaluator",
		},
		{
			name: "BothProjectAndDataset",
			req: tasks.CreateEvaluationTaskRequest{
				Name:       "t",
				Type:       tasks.TaskTypeTemplateEvaluation,
				Evaluators: []tasks.EvaluatorInput{{EvaluatorID: "ev-1"}},
				Project:    projectID("p-1"),
				Dataset:    datasetID("ds-1"),
			},
			wantErr: "exactly one of Project or Dataset",
		},
		{
			name: "NeitherProjectNorDataset",
			req: tasks.CreateEvaluationTaskRequest{
				Name:       "t",
				Type:       tasks.TaskTypeTemplateEvaluation,
				Evaluators: []tasks.EvaluatorInput{{EvaluatorID: "ev-1"}},
			},
			wantErr: "exactly one of Project or Dataset",
		},
		{
			name: "DatasetWithoutExperimentIDs",
			req: tasks.CreateEvaluationTaskRequest{
				Name:       "t",
				Type:       tasks.TaskTypeTemplateEvaluation,
				Evaluators: []tasks.EvaluatorInput{{EvaluatorID: "ev-1"}},
				Dataset:    datasetID("ds-1"),
			},
			wantErr: "at least one entry in ExperimentIDs",
		},
		{
			name: "ProjectWithExperimentIDs",
			req: tasks.CreateEvaluationTaskRequest{
				Name:          "t",
				Type:          tasks.TaskTypeTemplateEvaluation,
				Evaluators:    []tasks.EvaluatorInput{{EvaluatorID: "ev-1"}},
				Project:       projectID("p-1"),
				ExperimentIDs: []string{"exp-1"},
			},
			wantErr: "only applies to dataset-based tasks",
		},
		{
			name: "DatasetWithSamplingRate",
			req: tasks.CreateEvaluationTaskRequest{
				Name:          "t",
				Type:          tasks.TaskTypeTemplateEvaluation,
				Evaluators:    []tasks.EvaluatorInput{{EvaluatorID: "ev-1"}},
				Dataset:       datasetID("ds-1"),
				ExperimentIDs: []string{"exp-1"},
				SamplingRate:  0.5,
			},
			wantErr: "only apply to project-based tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			})
			_, err := client.Tasks.CreateEvaluationTask(context.Background(), tt.req)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

// TestTasks_CreateRunExperimentTask_Validation covers the client-side checks
// that reject a create before any request is sent.
func TestTasks_CreateRunExperimentTask_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     tasks.CreateRunExperimentTaskRequest
		wantErr string
	}{
		{
			name: "MissingName",
			req: tasks.CreateRunExperimentTaskRequest{
				Dataset:          datasetID("ds-1"),
				RunConfiguration: templateRunConfiguration(t),
			},
			wantErr: "Name is required",
		},
		{
			name: "MissingRunConfiguration",
			req: tasks.CreateRunExperimentTaskRequest{
				Name:    "t",
				Dataset: datasetID("ds-1"),
			},
			wantErr: "RunConfiguration is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			})
			_, err := client.Tasks.CreateRunExperimentTask(context.Background(), tt.req)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestTasks_WaitForRun(t *testing.T) {
	runID := taskRunID("run-1")

	tests := []struct {
		name     string
		statuses []tasks.TaskRunStatus // returned in order; the last repeats
		req      tasks.WaitForRunRequest
		ctx      func() (context.Context, context.CancelFunc)
		check    func(t *testing.T, got *tasks.TaskRun, err error)
	}{
		{
			name:     "PollsUntilTerminal",
			statuses: []tasks.TaskRunStatus{tasks.TaskRunStatusPending, tasks.TaskRunStatusRunning, tasks.TaskRunStatusCompleted},
			req:      tasks.WaitForRunRequest{RunID: runID, PollInterval: time.Millisecond},
			check: func(t *testing.T, got *tasks.TaskRun, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.Status != tasks.TaskRunStatusCompleted {
					t.Errorf("unexpected status: %s", got.Status)
				}
			},
		},
		{
			name:     "ReturnsFailedRun",
			statuses: []tasks.TaskRunStatus{tasks.TaskRunStatusFailed},
			req:      tasks.WaitForRunRequest{RunID: runID, PollInterval: time.Millisecond},
			check: func(t *testing.T, got *tasks.TaskRun, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.Status != tasks.TaskRunStatusFailed {
					t.Errorf("unexpected status: %s", got.Status)
				}
			},
		},
		{
			name:     "Timeout",
			statuses: []tasks.TaskRunStatus{tasks.TaskRunStatusPending},
			req:      tasks.WaitForRunRequest{RunID: runID, PollInterval: time.Millisecond, Timeout: 5 * time.Millisecond},
			check: func(t *testing.T, got *tasks.TaskRun, err error) {
				if !errors.Is(err, tasks.ErrWaitTimeout) {
					t.Fatalf("expected ErrWaitTimeout, got %T: %v", err, err)
				}
			},
		},
		{
			// Negative polling parameters are rejected before any request.
			name:     "NegativePollInterval",
			statuses: []tasks.TaskRunStatus{tasks.TaskRunStatusPending},
			req:      tasks.WaitForRunRequest{RunID: runID, PollInterval: -time.Second},
			check: func(t *testing.T, got *tasks.TaskRun, err error) {
				if err == nil || !strings.Contains(err.Error(), "PollInterval must be positive") {
					t.Fatalf("expected PollInterval validation error, got %v", err)
				}
			},
		},
		{
			name:     "NegativeTimeout",
			statuses: []tasks.TaskRunStatus{tasks.TaskRunStatusPending},
			req:      tasks.WaitForRunRequest{RunID: runID, PollInterval: time.Millisecond, Timeout: -time.Second},
			check: func(t *testing.T, got *tasks.TaskRun, err error) {
				if err == nil || !strings.Contains(err.Error(), "Timeout must be positive") {
					t.Fatalf("expected Timeout validation error, got %v", err)
				}
			},
		},
		{
			name:     "ContextCancelled",
			statuses: []tasks.TaskRunStatus{tasks.TaskRunStatusPending},
			req:      tasks.WaitForRunRequest{RunID: runID, PollInterval: time.Second},
			ctx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 10*time.Millisecond)
			},
			check: func(t *testing.T, got *tasks.TaskRun, err error) {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected context.DeadlineExceeded, got %T: %v", err, err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Polls arrive sequentially, so the unguarded counter is safe.
			poll := 0
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/task-runs/"+runID {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				status := tt.statuses[min(poll, len(tt.statuses)-1)]
				poll++
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tasks.TaskRun{Id: runID, Status: status})
			})
			ctx := context.Background()
			if tt.ctx != nil {
				var cancel context.CancelFunc
				ctx, cancel = tt.ctx()
				defer cancel()
			}
			got, err := client.Tasks.WaitForRun(ctx, tt.req)
			tt.check(t, got, err)
		})
	}
}
