package roles_test

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
	"github.com/Arize-ai/client-go-v2/arize/roles"
)

// wireRoleUpdate mirrors the JSON shape of generated.UpdateRoleRequest so tests can
// decode PATCH bodies without importing internal/generated. All fields are
// pointers so a missing field decodes to nil and an explicit `null` or zero
// value can be distinguished from "absent".
type wireRoleUpdate struct {
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	Permissions *[]roles.Permission `json:"permissions,omitempty"`
}

// wireRoleCreate mirrors the JSON shape of generated.CreateRoleRequest.
type wireRoleCreate struct {
	Name        string             `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Permissions []roles.Permission `json:"permissions,omitempty"`
}

func testID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Role:1:" + suffix))
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

// rolesListPage encodes a single page of the roles list endpoint.
func rolesListPage(w http.ResponseWriter, items []roles.Role, nextCursor *string) {
	w.Header().Set("Content-Type", "application/json")
	pagination := arize.PaginationMetadata{HasMore: nextCursor != nil, NextCursor: nextCursor}
	_ = json.NewEncoder(w).Encode(roles.ListRoles{Roles: items, Pagination: pagination})
}

func TestRoles(t *testing.T) {
	yes, no := true, false
	newName := "renamed-role"
	emptyDesc := ""

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("limit"); got != "10" {
					t.Errorf("limit query: want 10, got %q", got)
				}
				if got := r.URL.Query().Get("is_predefined"); got != "" {
					t.Errorf("is_predefined query: want absent, got %q", got)
				}
				rolesListPage(w, []roles.Role{{
					Id:          "role-1",
					Name:        "my-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}}, nil)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.List(ctx, roles.ListRequest{Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*roles.ListRoles)
				if len(resp.Roles) != 1 {
					t.Errorf("expected 1 role, got %d", len(resp.Roles))
				}
			},
		},
		{
			name: "List filters by IsPredefined=true",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("is_predefined"); got != "true" {
					t.Errorf("is_predefined query: want true, got %q", got)
				}
				// SDK applies a default of 50 when Limit is omitted.
				if got := r.URL.Query().Get("limit"); got != "50" {
					t.Errorf("limit default: want 50, got %q", got)
				}
				rolesListPage(w, nil, nil)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.List(ctx, roles.ListRequest{IsPredefined: &yes})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "List filters by IsPredefined=false",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("is_predefined"); got != "false" {
					t.Errorf("is_predefined query: want false, got %q", got)
				}
				rolesListPage(w, nil, nil)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.List(ctx, roles.ListRequest{IsPredefined: &no})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "List forwards Cursor",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("cursor"); got != "abc" {
					t.Errorf("cursor query: want abc, got %q", got)
				}
				rolesListPage(w, nil, nil)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.List(ctx, roles.ListRequest{Cursor: "abc"})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Get success by ID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(roles.Role{
					Id:          testID("role-1"),
					Name:        "my-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Get(ctx, roles.GetRequest{Role: testID("role-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*roles.Role).Name != "my-role" {
					t.Errorf("unexpected name: %s", got.(*roles.Role).Name)
				}
			},
		},
		{
			name: "Get not found by ID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Get(ctx, roles.GetRequest{Role: testID("missing")})
			},
			check: func(t *testing.T, _ any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Get not found by name across pages",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/roles" {
					t.Errorf("expected only list calls, got %s %s", r.Method, r.URL.Path)
					return
				}
				// Page 1 has "other-role"; page 2 has "another-role". The resolver
				// must walk both pages before returning ResourceNotFoundError.
				if r.URL.Query().Get("cursor") == "" {
					nextCursor := "page2"
					rolesListPage(w, []roles.Role{{
						Id:          testID("other-role"),
						Name:        "other-role",
						Permissions: []roles.Permission{roles.Permissions.ProjectRead},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}}, &nextCursor)
					return
				}
				rolesListPage(w, []roles.Role{{
					Id:          testID("another-role"),
					Name:        "another-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}}, nil)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Get(ctx, roles.GetRequest{Role: "missing-role"})
			},
			check: func(t *testing.T, _ any, err error) {
				var rnf *arize.ResourceNotFoundError
				if !errors.As(err, &rnf) {
					t.Errorf("expected *ResourceNotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Create success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wireRoleCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "new-role" {
					t.Errorf("body name: want new-role, got %q", body.Name)
				}
				if len(body.Permissions) != 1 || body.Permissions[0] != roles.Permissions.ProjectRead {
					t.Errorf("body permissions: %v", body.Permissions)
				}
				if body.Description != nil {
					t.Errorf("body description: want nil, got %v", body.Description)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(roles.Role{
					Id:          "role-new",
					Name:        "new-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Create(ctx, roles.CreateRequest{
					Name:        "new-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*roles.Role).Id != "role-new" {
					t.Errorf("unexpected id: %s", got.(*roles.Role).Id)
				}
			},
		},
		{
			name: "Create bad request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "invalid permission", "status": 400})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Create(ctx, roles.CreateRequest{
					Name:        "new-role",
					Permissions: []roles.Permission{"NOT_A_PERMISSION"},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				var be *arize.BadRequestError
				if !errors.As(err, &be) {
					t.Errorf("expected *BadRequestError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Update success by ID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				var body wireRoleUpdate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name == nil || *body.Name != "renamed-role" {
					t.Errorf("body name: want renamed-role, got %v", body.Name)
				}
				if body.Description != nil {
					t.Errorf("body description: want nil, got %v", body.Description)
				}
				if body.Permissions != nil {
					t.Errorf("body permissions: want nil, got %v", body.Permissions)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(roles.Role{
					Id:          testID("role-1"),
					Name:        "renamed-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Update(ctx, roles.UpdateRequest{
					Role: testID("role-1"),
					Name: &newName,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*roles.Role).Name != "renamed-role" {
					t.Errorf("unexpected name: %s", got.(*roles.Role).Name)
				}
			},
		},
		{
			name: "Update clears Description",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireRoleUpdate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				// nil vs "" must be distinguishable on the wire: passing &"" clears.
				if body.Description == nil || *body.Description != "" {
					t.Errorf("body description: want &\"\", got %v", body.Description)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(roles.Role{
					Id:          testID("role-1"),
					Name:        "my-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Update(ctx, roles.UpdateRequest{
					Role:        testID("role-1"),
					Description: &emptyDesc,
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Update conflict",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "name already exists", "status": 409})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Update(ctx, roles.UpdateRequest{
					Role: testID("role-1"),
					Name: &newName,
				})
			},
			check: func(t *testing.T, _ any, err error) {
				var ce *arize.ConflictError
				if !errors.As(err, &ce) {
					t.Errorf("expected *ConflictError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Delete success by ID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(http.StatusNoContent)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Roles.Delete(ctx, roles.DeleteRequest{Role: testID("role-1")})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Delete forbidden (predefined role)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "predefined roles cannot be deleted", "status": 403})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Roles.Delete(ctx, roles.DeleteRequest{Role: testID("role-1")})
			},
			check: func(t *testing.T, _ any, err error) {
				var fe *arize.ForbiddenError
				if !errors.As(err, &fe) {
					t.Errorf("expected *ForbiddenError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Get success by name",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v2/roles" {
					rolesListPage(w, []roles.Role{{
						Id:          testID("role-1"),
						Name:        "my-role",
						Permissions: []roles.Permission{roles.Permissions.ProjectRead},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}}, nil)
					return
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				_ = json.NewEncoder(w).Encode(roles.Role{
					Id:          testID("role-1"),
					Name:        "my-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Get(ctx, roles.GetRequest{Role: "my-role"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*roles.Role).Name != "my-role" {
					t.Errorf("unexpected name: %s", got.(*roles.Role).Name)
				}
			},
		},
		{
			name: "Update success by name",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v2/roles" {
					rolesListPage(w, []roles.Role{{
						Id:          testID("role-1"),
						Name:        "my-role",
						Permissions: []roles.Permission{roles.Permissions.ProjectRead},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}}, nil)
					return
				}
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				var body wireRoleUpdate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name == nil || *body.Name != "renamed-role" {
					t.Errorf("body name: want renamed-role, got %v", body.Name)
				}
				_ = json.NewEncoder(w).Encode(roles.Role{
					Id:          testID("role-1"),
					Name:        "renamed-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Update(ctx, roles.UpdateRequest{
					Role: "my-role",
					Name: &newName,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*roles.Role).Name != "renamed-role" {
					t.Errorf("unexpected name: %s", got.(*roles.Role).Name)
				}
			},
		},
		{
			name: "Delete success by name",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v2/roles" {
					rolesListPage(w, []roles.Role{{
						Id:          testID("role-1"),
						Name:        "my-role",
						Permissions: []roles.Permission{roles.Permissions.ProjectRead},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}}, nil)
					return
				}
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(http.StatusNoContent)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Roles.Delete(ctx, roles.DeleteRequest{Role: "my-role"})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Get success by name across multiple pages",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v2/roles" {
					// Page 1: doesn't contain "my-role", advertises a next cursor.
					if r.URL.Query().Get("cursor") == "" {
						nextCursor := "page2"
						rolesListPage(w, []roles.Role{{
							Id:          testID("other-role"),
							Name:        "other-role",
							Permissions: []roles.Permission{roles.Permissions.ProjectRead},
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						}}, &nextCursor)
						return
					}
					// Page 2: contains the role we want.
					rolesListPage(w, []roles.Role{{
						Id:          testID("role-1"),
						Name:        "my-role",
						Permissions: []roles.Permission{roles.Permissions.ProjectRead},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}}, nil)
					return
				}
				// Single-role GET for the resolved ID.
				_ = json.NewEncoder(w).Encode(roles.Role{
					Id:          testID("role-1"),
					Name:        "my-role",
					Permissions: []roles.Permission{roles.Permissions.ProjectRead},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Roles.Get(ctx, roles.GetRequest{Role: "my-role"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*roles.Role).Name != "my-role" {
					t.Errorf("unexpected name: %s", got.(*roles.Role).Name)
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
