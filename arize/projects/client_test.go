package projects_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/projects"
)

func projectID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Project:1:" + suffix))
}

func spaceID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Space:1:" + suffix))
}

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

// wireProject mirrors the JSON shape the API sends/receives for project create
// requests. Tests use it so they can decode request bodies without importing
// internal/generated.
type wireProject struct {
	Name    string `json:"name"`
	SpaceId string `json:"space_id"`
}

// wireProjectUpdate mirrors the JSON shape of a PATCH /v2/projects/{id} body.
type wireProjectUpdate struct {
	Name string `json:"name"`
}

func TestProjects(t *testing.T) {
	var listFiltersQuery url.Values

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(projects.ListProjects{
					Projects:   []projects.Project{{Id: "proj-1", Name: "my-project", CreatedAt: time.Now(), SpaceId: "space-1"}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.List(ctx, projects.ListRequest{Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*projects.ListProjects)
				if len(resp.Projects) != 1 {
					t.Errorf("expected 1 project, got %d", len(resp.Projects))
				}
			},
		},
		{
			name: "List_SpaceAsID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				listFiltersQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"projects":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.List(ctx, projects.ListRequest{
					Name:  "prod",
					Space: spaceID("space-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if listFiltersQuery.Get("name") != "prod" ||
					listFiltersQuery.Get("space_id") != spaceID("space-1") ||
					listFiltersQuery.Get("space_name") != "" {
					t.Errorf("query: %v", listFiltersQuery)
				}
			},
		},
		{
			name: "List_SpaceAsName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				listFiltersQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"projects":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.List(ctx, projects.ListRequest{
					Space: "demo",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if listFiltersQuery.Get("space_id") != "" ||
					listFiltersQuery.Get("space_name") != "demo" {
					t.Errorf("query: %v", listFiltersQuery)
				}
			},
		},
		{
			name: "Get",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(projects.Project{Id: "proj-1", Name: "my-project", CreatedAt: time.Now(), SpaceId: "space-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.Get(ctx, projects.GetRequest{Project: projectID("proj-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				proj := got.(*projects.Project)
				if proj.Id != "proj-1" {
					t.Errorf("unexpected id: %s", proj.Id)
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
				return c.Projects.Get(ctx, projects.GetRequest{Project: projectID("missing")})
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
				var body wireProject
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "new-project" {
					t.Errorf("body name: want new-project, got %q", body.Name)
				}
				if body.SpaceId != spaceID("space-1") {
					t.Errorf("body space_id: want %q, got %q", spaceID("space-1"), body.SpaceId)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(projects.Project{Id: "proj-new", Name: "new-project", CreatedAt: time.Now(), SpaceId: "space-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.Create(ctx, projects.CreateRequest{
					Name:  "new-project",
					Space: spaceID("space-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				proj := got.(*projects.Project)
				if proj.Id != "proj-new" {
					t.Errorf("unexpected id: %s", proj.Id)
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
				return nil, c.Projects.Delete(ctx, projects.DeleteRequest{Project: projectID("proj-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Update",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				var body wireProjectUpdate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "renamed-project" {
					t.Errorf("body name: want renamed-project, got %q", body.Name)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(projects.Project{Id: "proj-1", Name: "renamed-project", CreatedAt: time.Now(), SpaceId: "space-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.Update(ctx, projects.UpdateRequest{
					Project: projectID("proj-1"),
					Name:    "renamed-project",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				proj := got.(*projects.Project)
				if proj.Name != "renamed-project" {
					t.Errorf("unexpected name: %s", proj.Name)
				}
			},
		},
		{
			name: "Update_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.Update(ctx, projects.UpdateRequest{
					Project: projectID("missing"),
					Name:    "renamed-project",
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
			name: "Update_Conflict",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(409)
				json.NewEncoder(w).Encode(map[string]any{"title": "conflict", "status": 409})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.Update(ctx, projects.UpdateRequest{
					Project: projectID("proj-1"),
					Name:    "existing-project",
				})
			},
			check: func(t *testing.T, got any, err error) {
				var ce *arize.ConflictError
				if !errors.As(err, &ce) {
					t.Errorf("expected *ConflictError, got %T: %v", err, err)
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
