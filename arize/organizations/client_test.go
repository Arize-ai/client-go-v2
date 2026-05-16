package organizations_test

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
	"github.com/Arize-ai/client-go-v2/arize/organizations"
)

func testID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Organization:1:" + suffix))
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

func TestOrganizations(t *testing.T) {
	var seenName string
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
				json.NewEncoder(w).Encode(organizations.OrganizationList{
					Organizations: []organizations.Organization{{Id: "org-1", Name: "my-org", CreatedAt: time.Now()}},
					Pagination:    arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				limit := 10
				return c.Organizations.List(ctx, organizations.ListParams{Limit: &limit})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*organizations.OrganizationList)
				if len(resp.Organizations) != 1 {
					t.Errorf("expected 1 organization, got %d", len(resp.Organizations))
				}
			},
		},
		{
			name: "List_NameFilter",
			handler: func(w http.ResponseWriter, r *http.Request) {
				seenName = r.URL.Query().Get("name")
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(organizations.OrganizationList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "acme"
				return c.Organizations.List(ctx, organizations.ListParams{Name: &name})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if seenName != "acme" {
					t.Errorf("name query: want acme, got %q", seenName)
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
				return c.Organizations.Get(ctx, testID("missing"))
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
				var body organizations.CreateRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "new-org" {
					t.Errorf("body name: want new-org, got %q", body.Name)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(organizations.Organization{Id: "org-new", Name: "new-org", CreatedAt: time.Now()})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Organizations.Create(ctx, organizations.CreateRequest{
					Name: "new-org",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				org := got.(*organizations.Organization)
				if org.Id != "org-new" {
					t.Errorf("unexpected id: %s", org.Id)
				}
			},
		},
		{
			name: "Create_Conflict",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(409)
				json.NewEncoder(w).Encode(map[string]any{"title": "conflict", "status": 409})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Organizations.Create(ctx, organizations.CreateRequest{Name: "duplicate"})
			},
			check: func(t *testing.T, got any, err error) {
				var ce *arize.ConflictError
				if !errors.As(err, &ce) {
					t.Errorf("expected *ConflictError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Update",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				var body organizations.UpdateRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name == nil || *body.Name != "renamed" {
					t.Errorf("body name: want renamed, got %v", body.Name)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(organizations.Organization{Id: testID("org-1"), Name: "renamed", CreatedAt: time.Now()})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				newName := "renamed"
				return c.Organizations.Update(ctx, testID("org-1"), organizations.UpdateRequest{Name: &newName})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				org := got.(*organizations.Organization)
				if org.Name != "renamed" {
					t.Errorf("unexpected name: %s", org.Name)
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
				newName := "renamed"
				return c.Organizations.Update(ctx, testID("missing"), organizations.UpdateRequest{Name: &newName})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
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
