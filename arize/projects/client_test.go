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
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/projects"
)

func testID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Project:1:" + suffix))
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
				json.NewEncoder(w).Encode(generated.ProjectList{
					Projects:   []generated.Project{{Id: "proj-1", Name: "my-project", CreatedAt: time.Now(), SpaceId: "space-1"}},
					Pagination: generated.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				limit := 10
				return c.Projects.List(ctx, projects.ListParams{Limit: &limit})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*generated.ProjectList)
				if len(resp.Projects) != 1 {
					t.Errorf("expected 1 project, got %d", len(resp.Projects))
				}
			},
		},
		{
			name: "List_NewFilters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				listFiltersQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"projects":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name, spaceID, spaceName := "prod", "sp-1", "demo"
				return c.Projects.List(ctx, projects.ListParams{
					Name: &name, SpaceId: &spaceID, SpaceName: &spaceName,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if listFiltersQuery.Get("name") != "prod" || listFiltersQuery.Get("space_id") != "sp-1" || listFiltersQuery.Get("space_name") != "demo" {
					t.Errorf("query: %v", listFiltersQuery)
				}
			},
		},
		{
			name: "Get",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(generated.Project{Id: "proj-1", Name: "my-project", CreatedAt: time.Now(), SpaceId: "space-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.Get(ctx, testID("proj-1"), "")
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				proj := got.(*generated.Project)
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
				return c.Projects.Get(ctx, testID("missing"), "")
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(generated.Project{Id: "proj-new", Name: "new-project", CreatedAt: time.Now(), SpaceId: "space-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Projects.Create(ctx, testID("space-1"), "new-project")
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				proj := got.(*generated.Project)
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
				return nil, c.Projects.Delete(ctx, testID("proj-1"), "")
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
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
