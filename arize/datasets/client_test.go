package datasets_test

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
	"github.com/Arize-ai/client-go-v2/arize/datasets"
)

// testID returns a base64-encoded ID so resolve.IsResourceID treats it as an
// ID and skips name-resolution lookups during tests.
func testID(prefix, suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte(prefix + ":1:" + suffix))
}

func datasetID(suffix string) string { return testID("Dataset", suffix) }
func spaceID(suffix string) string   { return testID("Space", suffix) }

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

// wireCreate mirrors the JSON shape of the Create request body.
type wireCreate struct {
	SpaceId  string           `json:"space_id"`
	Name     string           `json:"name"`
	Examples []map[string]any `json:"examples"`
}

// wireInsert mirrors the JSON shape of the AppendExamples request body.
type wireInsert struct {
	Examples []map[string]any `json:"examples"`
}

// wireUpdate mirrors the JSON shape of the Update (rename) request body.
type wireUpdate struct {
	Name string `json:"name"`
}

// wireDeleteExamples mirrors the JSON shape of the DeleteExamples request body.
type wireDeleteExamples struct {
	DatasetVersionId string   `json:"dataset_version_id"`
	ExampleIds       []string `json:"example_ids"`
}

// wireAnnotate mirrors the JSON shape of the AnnotateExamples request body.
type wireAnnotate struct {
	Annotations []struct {
		RecordId string `json:"record_id"`
		Values   []struct {
			Name  string   `json:"name"`
			Score *float64 `json:"score,omitempty"`
		} `json:"values"`
	} `json:"annotations"`
}

func TestDatasets(t *testing.T) {
	dsID := datasetID("ds-1")

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/datasets" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.DatasetList{
					Datasets:   []datasets.Dataset{{Id: "ds-1", Name: "my-dataset"}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.List(ctx, datasets.ListRequest{Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*datasets.DatasetList)
				if len(resp.Datasets) != 1 {
					t.Errorf("expected 1 dataset, got %d", len(resp.Datasets))
				}
			},
		},
		{
			name: "List_Filters_SpaceID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got, want := q.Get("space_id"), spaceID("sp-1"); got != want {
					t.Errorf("space_id query: want %q, got %q", want, got)
				}
				if q.Get("space_name") != "" {
					t.Errorf("space_name should be empty for an ID, got %q", q.Get("space_name"))
				}
				if got := q.Get("name"); got != "eval" {
					t.Errorf("name query: want eval, got %q", got)
				}
				if got := q.Get("limit"); got != "25" {
					t.Errorf("limit query: want 25, got %q", got)
				}
				if got := q.Get("cursor"); got != "cursor-abc" {
					t.Errorf("cursor query: want cursor-abc, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.DatasetList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.List(ctx, datasets.ListRequest{
					Space:  spaceID("sp-1"),
					Name:   "eval",
					Limit:  25,
					Cursor: "cursor-abc",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// A plain Space name flows into space_name (no resolution; the
			// list endpoint filters by name directly).
			name: "List_Filters_SpaceName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got := q.Get("space_name"); got != "demo" {
					t.Errorf("space_name query: want demo, got %q", got)
				}
				if q.Get("space_id") != "" {
					t.Errorf("space_id should be empty for a name, got %q", q.Get("space_id"))
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.DatasetList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.List(ctx, datasets.ListRequest{Space: "demo"})
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
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.Dataset{Id: dsID, Name: "my-dataset"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.Get(ctx, datasets.GetRequest{Dataset: dsID})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*datasets.Dataset).Name != "my-dataset" {
					t.Errorf("unexpected name: %s", got.(*datasets.Dataset).Name)
				}
			},
		},
		{
			// Exercises resolve.FindDatasetID: a plain Dataset name (with a
			// Space) first lists datasets filtered by name to discover the ID,
			// then issues the real GET.
			name: "Get_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/datasets":
					if got := r.URL.Query().Get("name"); got != "my-dataset" {
						t.Errorf("resolver list name query: want my-dataset, got %q", got)
					}
					if got, want := r.URL.Query().Get("space_id"), spaceID("sp-1"); got != want {
						t.Errorf("resolver list space_id query: want %q, got %q", want, got)
					}
					json.NewEncoder(w).Encode(datasets.DatasetList{
						Datasets:   []datasets.Dataset{{Id: dsID, Name: "my-dataset"}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/"+dsID):
					json.NewEncoder(w).Encode(datasets.Dataset{Id: dsID, Name: "my-dataset"})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.Get(ctx, datasets.GetRequest{Dataset: "my-dataset", Space: spaceID("sp-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*datasets.Dataset).Name != "my-dataset" {
					t.Errorf("unexpected name: %s", got.(*datasets.Dataset).Name)
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
				return c.Datasets.Get(ctx, datasets.GetRequest{Dataset: datasetID("missing")})
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
				var body wireCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "new-ds" {
					t.Errorf("body name: want new-ds, got %q", body.Name)
				}
				if body.SpaceId != spaceID("space-1") {
					t.Errorf("body space_id: want %q, got %q", spaceID("space-1"), body.SpaceId)
				}
				if len(body.Examples) != 1 || body.Examples[0]["input"] != "hello" {
					t.Errorf("unexpected examples: %v", body.Examples)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(datasets.Dataset{Id: "ds-new", Name: "new-ds"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.Create(ctx, datasets.CreateRequest{
					Space: spaceID("space-1"),
					Name:  "new-ds",
					Examples: []datasets.DatasetExampleCreate{
						{"input": "hello"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				ds := got.(*datasets.Dataset)
				if ds.Id != "ds-new" {
					t.Errorf("unexpected id: %s", ds.Id)
				}
			},
		},
		{
			// Exercises Create's space-name resolution: a plain Space name
			// triggers FindSpaceID before the POST, and the resolved ID must
			// appear as space_id in the request body.
			name: "Create_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/spaces":
					json.NewEncoder(w).Encode(map[string]any{
						"spaces":     []map[string]any{{"id": spaceID("sp-1"), "name": "demo", "created_at": time.Now()}},
						"pagination": map[string]any{"has_more": false},
					})
				case r.Method == http.MethodPost && r.URL.Path == "/v2/datasets":
					var body wireCreate
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("decode body: %v", err)
					}
					if body.SpaceId != spaceID("sp-1") {
						t.Errorf("body space_id: want %q, got %q", spaceID("sp-1"), body.SpaceId)
					}
					w.WriteHeader(201)
					json.NewEncoder(w).Encode(datasets.Dataset{Id: "ds-new", Name: "new-ds"})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.Create(ctx, datasets.CreateRequest{
					Space: "demo",
					Name:  "new-ds",
					Examples: []datasets.DatasetExampleCreate{
						{"input": "hello"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Delete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Datasets.Delete(ctx, datasets.DeleteRequest{Dataset: dsID})
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Datasets.Delete(ctx, datasets.DeleteRequest{Dataset: datasetID("missing")})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "ListExamples",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/v2/datasets/"+dsID+"/examples" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if got := r.URL.Query().Get("dataset_version_id"); got != "v-1" {
					t.Errorf("dataset_version_id query: want v-1, got %s", got)
				}
				if got := r.URL.Query().Get("limit"); got != "10" {
					t.Errorf("limit query: want 10, got %s", got)
				}
				exampleID := "ex-1"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.DatasetExampleList{
					Examples:   []datasets.DatasetExample{{Id: &exampleID}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.ListExamples(ctx, datasets.ListExamplesRequest{
					Dataset:          dsID,
					Limit:            10,
					DatasetVersionID: "v-1",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*datasets.DatasetExampleList)
				if len(resp.Examples) != 1 {
					t.Errorf("expected 1 example, got %d", len(resp.Examples))
				}
			},
		},
		{
			name: "ListExamples_Cursor",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("cursor"); got != "tok-123" {
					t.Errorf("cursor query: want tok-123, got %s", got)
				}
				exampleID := "ex-2"
				nextCursor := "tok-456"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.DatasetExampleList{
					Examples: []datasets.DatasetExample{{Id: &exampleID}},
					Pagination: arize.PaginationMetadata{
						HasMore:    true,
						NextCursor: &nextCursor,
					},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.ListExamples(ctx, datasets.ListExamplesRequest{
					Dataset: dsID,
					Cursor:  "tok-123",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*datasets.DatasetExampleList)
				if !resp.Pagination.HasMore {
					t.Errorf("expected HasMore true")
				}
				if resp.Pagination.NextCursor == nil || *resp.Pagination.NextCursor != "tok-456" {
					t.Errorf("expected NextCursor tok-456, got %v", resp.Pagination.NextCursor)
				}
			},
		},
		{
			name: "Update",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				if r.URL.Path != "/v2/datasets/"+dsID {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var body wireUpdate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Name != "renamed-ds" {
					t.Errorf("body name: want renamed-ds, got %q", body.Name)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.Dataset{Id: dsID, Name: "renamed-ds"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.Update(ctx, datasets.UpdateRequest{
					Dataset: dsID,
					Name:    "renamed-ds",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*datasets.Dataset).Name != "renamed-ds" {
					t.Errorf("unexpected name: %s", got.(*datasets.Dataset).Name)
				}
			},
		},
		{
			name: "AppendExamples",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/v2/datasets/"+dsID+"/examples" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if got := r.URL.Query().Get("dataset_version_id"); got != "v-1" {
					t.Errorf("dataset_version_id query: want v-1, got %s", got)
				}
				var body wireInsert
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if len(body.Examples) != 1 {
					t.Fatalf("expected 1 example in body, got %d", len(body.Examples))
				}
				if v := body.Examples[0]["input"]; v != "hello" {
					t.Errorf("expected input=hello, got %v", v)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(datasets.DatasetExamplesInserted{
					DatasetVersionId: "v-1",
					ExampleIds:       []string{"ex-new"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.AppendExamples(ctx, datasets.AppendExamplesRequest{
					Dataset:          dsID,
					DatasetVersionID: "v-1",
					Examples: []datasets.DatasetExampleCreate{
						{"input": "hello", "output": "world"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*datasets.DatasetExamplesInserted)
				if resp.DatasetVersionId != "v-1" {
					t.Errorf("unexpected version: %s", resp.DatasetVersionId)
				}
				if len(resp.ExampleIds) != 1 || resp.ExampleIds[0] != "ex-new" {
					t.Errorf("unexpected example ids: %v", resp.ExampleIds)
				}
			},
		},
		{
			name: "AppendExamples_ServerError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
				_, _ = w.Write([]byte(`{"title":"invalid example"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.AppendExamples(ctx, datasets.AppendExamplesRequest{
					Dataset:  datasetID("ds-1"),
					Examples: []datasets.DatasetExampleCreate{{"input": "x"}},
				})
			},
			check: func(t *testing.T, got any, err error) {
				var bre *arize.BadRequestError
				if !errors.As(err, &bre) {
					t.Fatalf("expected *BadRequestError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "DeleteExamples",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.URL.Path != "/v2/datasets/"+dsID+"/examples" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var body wireDeleteExamples
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.DatasetVersionId != "v-1" {
					t.Errorf("body dataset_version_id: want v-1, got %q", body.DatasetVersionId)
				}
				if len(body.ExampleIds) != 2 || body.ExampleIds[0] != "ex-1" || body.ExampleIds[1] != "ex-2" {
					t.Errorf("unexpected example_ids: %v", body.ExampleIds)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.DatasetExamplesDeleteResult{
					Completed:            true,
					DeletedExampleIds:    []string{"ex-1", "ex-2"},
					NotDeletedExampleIds: []string{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.DeleteExamples(ctx, datasets.DeleteExamplesRequest{
					Dataset:          dsID,
					DatasetVersionID: "v-1",
					ExampleIDs:       []string{"ex-1", "ex-2"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*datasets.DatasetExamplesDeleteResult)
				if !resp.Completed {
					t.Errorf("expected Completed=true on full success, got false")
				}
				if len(resp.DeletedExampleIds) != 2 {
					t.Errorf("expected 2 deleted ids, got %v", resp.DeletedExampleIds)
				}
				if len(resp.NotDeletedExampleIds) != 0 {
					t.Errorf("expected empty NotDeletedExampleIds, got %v", resp.NotDeletedExampleIds)
				}
			},
		},
		{
			// A 200 with Completed=false reports a partial delete that the
			// caller should retry; not-deleted IDs are surfaced back.
			name: "DeleteExamples_Incomplete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(datasets.DatasetExamplesDeleteResult{
					Completed:            false,
					DeletedExampleIds:    []string{"ex-1"},
					NotDeletedExampleIds: []string{"ex-2"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.DeleteExamples(ctx, datasets.DeleteExamplesRequest{
					Dataset:          dsID,
					DatasetVersionID: "v-1",
					ExampleIDs:       []string{"ex-1", "ex-2"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*datasets.DatasetExamplesDeleteResult)
				if resp.Completed {
					t.Errorf("expected Completed=false on incomplete, got true")
				}
				if len(resp.NotDeletedExampleIds) != 1 || resp.NotDeletedExampleIds[0] != "ex-2" {
					t.Errorf("unexpected NotDeletedExampleIds: %v", resp.NotDeletedExampleIds)
				}
			},
		},
		{
			name: "DeleteExamples_ServerError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
				_, _ = w.Write([]byte(`{"title":"invalid version"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.DeleteExamples(ctx, datasets.DeleteExamplesRequest{
					Dataset:          datasetID("ds-1"),
					DatasetVersionID: "v-1",
					ExampleIDs:       []string{"ex-1"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				var bre *arize.BadRequestError
				if !errors.As(err, &bre) {
					t.Fatalf("expected *BadRequestError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "AnnotateExamples",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/v2/datasets/"+dsID+"/examples/annotate" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var body wireAnnotate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if len(body.Annotations) != 1 {
					t.Fatalf("expected 1 annotation, got %d", len(body.Annotations))
				}
				if body.Annotations[0].RecordId != "ex-1" {
					t.Errorf("record_id: want ex-1, got %q", body.Annotations[0].RecordId)
				}
				if len(body.Annotations[0].Values) != 1 || body.Annotations[0].Values[0].Name != "quality" {
					t.Errorf("unexpected values: %+v", body.Annotations[0].Values)
				}
				w.WriteHeader(202)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				score := 0.9
				return nil, c.Datasets.AnnotateExamples(ctx, datasets.AnnotateExamplesRequest{
					Dataset: dsID,
					Annotations: []datasets.AnnotateRecordInput{
						{
							RecordId: "ex-1",
							Values: []datasets.AnnotationInput{
								{Name: "quality", Score: &score},
							},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// Create rejects an empty example set client-side, before any
			// request is sent, matching the Python SDK.
			name: "Create_NoExamples",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Datasets.Create(ctx, datasets.CreateRequest{
					Space: spaceID("space-1"),
					Name:  "empty-ds",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if !errors.Is(err, datasets.ErrNoExamples) {
					t.Fatalf("expected ErrNoExamples, got %T: %v", err, err)
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
