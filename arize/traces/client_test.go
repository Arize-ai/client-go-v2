package traces_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/traces"
)

func projectID(suffix string) string {
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

// wireTracesList mirrors the JSON shape the API receives for traces.List
// request bodies. Tests use it so they can decode request bodies without
// importing internal/generated.
type wireTracesList struct {
	ProjectId string     `json:"project_id"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Filter    *string    `json:"filter,omitempty"`
}

func TestTraces(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wireTracesList
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.ProjectId != projectID("proj-1") {
					t.Errorf("body project_id: want %q, got %q", projectID("proj-1"), body.ProjectId)
				}
				if body.Filter == nil || *body.Filter != "status_code = 'ERROR'" {
					t.Errorf("body filter: %v", body.Filter)
				}
				if r.URL.Query().Get("limit") != "50" {
					t.Errorf("query limit: %q", r.URL.Query().Get("limit"))
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(traces.ListTraces{
					Traces:     []traces.Trace{},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Traces.List(ctx, traces.ListRequest{
					Project: projectID("proj-1"),
					Filter:  "status_code = 'ERROR'",
					Limit:   50,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == nil {
					t.Error("expected non-nil response")
				}
			},
		},
		{
			name: "List_DefaultLimit",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("limit") != "50" {
					t.Errorf("query limit: want %q, got %q", "50", r.URL.Query().Get("limit"))
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(traces.ListTraces{
					Traces:     []traces.Trace{},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Traces.List(ctx, traces.ListRequest{
					Project: projectID("proj-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == nil {
					t.Error("expected non-nil response")
				}
			},
		},
		{
			name: "List_WithCursor",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("cursor"); got != "next-page" {
					t.Errorf("query cursor: want %q, got %q", "next-page", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(traces.ListTraces{
					Traces:     []traces.Trace{},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Traces.List(ctx, traces.ListRequest{
					Project: projectID("proj-1"),
					Cursor:  "next-page",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == nil {
					t.Error("expected non-nil response")
				}
			},
		},
		{
			name: "List_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Traces.List(ctx, traces.ListRequest{Project: projectID("nonexistent")})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "List_ByProjectName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/projects":
					// Project name resolution: return a project whose name
					// matches the request so the resolver yields its ID.
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]any{
						"projects": []map[string]any{
							{"id": projectID("proj-1"), "name": "my-project"},
						},
						"pagination": map[string]any{"has_more": false},
					})
				case "/v2/traces":
					if r.Method != http.MethodPost {
						t.Errorf("expected POST, got %s", r.Method)
					}
					var body wireTracesList
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("decode body: %v", err)
					}
					if body.ProjectId != projectID("proj-1") {
						t.Errorf("body project_id: want %q, got %q", projectID("proj-1"), body.ProjectId)
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(traces.ListTraces{
						Traces:     []traces.Trace{},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				default:
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Traces.List(ctx, traces.ListRequest{
					Project: "my-project",
					Space:   "my-space",
					Limit:   25,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == nil {
					t.Error("expected non-nil response")
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
