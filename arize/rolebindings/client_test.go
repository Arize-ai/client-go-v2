package rolebindings_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/rolebindings"
)

// wireRoleBinding mirrors the API JSON shape for a role-binding create/update body,
// so tests can decode incoming requests without importing internal/generated.
type wireRoleBinding struct {
	ResourceId   string                                  `json:"resource_id,omitempty"`
	ResourceType rolebindings.RoleBindingResourceType    `json:"resource_type,omitempty"`
	RoleId       string                                  `json:"role_id,omitempty"`
	UserId       string                                  `json:"user_id,omitempty"`
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

func TestRoleBindings(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "Get success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(rolebindings.RoleBinding{
					Id:           "binding-1",
					ResourceId:   "space-1",
					ResourceType: rolebindings.RoleBindingResourceTypeSPACE,
					RoleId:       "role-1",
					UserId:       "user-1",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.RoleBindings.Get(ctx, rolebindings.GetRequest{RoleBindingID: "binding-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*rolebindings.RoleBinding).Id != "binding-1" {
					t.Errorf("unexpected id: %s", got.(*rolebindings.RoleBinding).Id)
				}
			},
		},
		{
			name: "Get not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.RoleBindings.Get(ctx, rolebindings.GetRequest{RoleBindingID: "nonexistent"})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Create success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wireRoleBinding
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.RoleId != "role-1" || body.UserId != "user-1" ||
					body.ResourceId != "space-1" || body.ResourceType != rolebindings.RoleBindingResourceTypeSPACE {
					t.Errorf("unexpected body: %+v", body)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(rolebindings.RoleBinding{
					Id:           "binding-1",
					ResourceId:   "space-1",
					ResourceType: rolebindings.RoleBindingResourceTypeSPACE,
					RoleId:       "role-1",
					UserId:       "user-1",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.RoleBindings.Create(ctx, rolebindings.CreateRequest{
					ResourceID:   "space-1",
					ResourceType: rolebindings.RoleBindingResourceTypeSPACE,
					RoleID:       "role-1",
					UserID:       "user-1",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*rolebindings.RoleBinding).Id != "binding-1" {
					t.Errorf("unexpected id: %s", got.(*rolebindings.RoleBinding).Id)
				}
			},
		},
		{
			name: "Create bad request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(400)
				json.NewEncoder(w).Encode(map[string]any{"title": "invalid resource_type", "status": 400})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.RoleBindings.Create(ctx, rolebindings.CreateRequest{
					ResourceID:   "dataset-1",
					ResourceType: rolebindings.RoleBindingResourceType("DATASET"),
					RoleID:       "role-1",
					UserID:       "user-1",
				})
			},
			check: func(t *testing.T, got any, err error) {
				var be *arize.BadRequestError
				if !errors.As(err, &be) {
					t.Errorf("expected *BadRequestError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Create conflict",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(409)
				json.NewEncoder(w).Encode(map[string]any{"title": "conflict", "status": 409})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.RoleBindings.Create(ctx, rolebindings.CreateRequest{
					ResourceID:   "space-1",
					ResourceType: rolebindings.RoleBindingResourceTypeSPACE,
					RoleID:       "role-1",
					UserID:       "user-1",
				})
			},
			check: func(t *testing.T, got any, err error) {
				var ce *arize.ConflictError
				if !errors.As(err, &ce) {
					t.Errorf("expected *ConflictError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Update success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				var body wireRoleBinding
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.RoleId != "role-2" {
					t.Errorf("body role_id: want role-2, got %q", body.RoleId)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(rolebindings.RoleBinding{
					Id:           "binding-1",
					ResourceId:   "space-1",
					ResourceType: rolebindings.RoleBindingResourceTypeSPACE,
					RoleId:       "role-2",
					UserId:       "user-1",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.RoleBindings.Update(ctx, rolebindings.UpdateRequest{
					RoleBindingID: "binding-1",
					RoleID:        "role-2",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*rolebindings.RoleBinding).RoleId != "role-2" {
					t.Errorf("unexpected role_id: %s", got.(*rolebindings.RoleBinding).RoleId)
				}
			},
		},
		{
			name: "Update not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.RoleBindings.Update(ctx, rolebindings.UpdateRequest{
					RoleBindingID: "nonexistent",
					RoleID:        "role-2",
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
			name: "Delete success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.RoleBindings.Delete(ctx, rolebindings.DeleteRequest{RoleBindingID: "binding-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Delete not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.RoleBindings.Delete(ctx, rolebindings.DeleteRequest{RoleBindingID: "nonexistent"})
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
