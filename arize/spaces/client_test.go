package spaces_test

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
	"github.com/Arize-ai/client-go-v2/arize/spaces"
)

// testID returns a base64-encoded ID so resolve.IsResourceID treats it as an
// ID and skips name-resolution lookups during tests.
func testID(prefix, suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte(prefix + ":1:" + suffix))
}

func spaceID(suffix string) string { return testID("Space", suffix) }
func orgID(suffix string) string   { return testID("Organization", suffix) }
func userID(suffix string) string  { return testID("User", suffix) }

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

// wireSpace mirrors the JSON shape of Create/Update request bodies so tests
// can assert on the wire payload without importing internal/generated.
type wireSpace struct {
	Name           string  `json:"name,omitempty"`
	OrganizationId string  `json:"organization_id,omitempty"`
	Description    *string `json:"description,omitempty"`
}

// wireMembershipInput mirrors the JSON shape of the AddUser request body.
type wireMembershipInput struct {
	UserId string         `json:"user_id"`
	Role   map[string]any `json:"role"`
}

func TestSpaces(t *testing.T) {
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
				json.NewEncoder(w).Encode(spaces.SpaceList{
					Spaces:     []spaces.Space{{Id: "space-1", Name: "my-space", CreatedAt: time.Now()}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.List(ctx, spaces.ListRequest{Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*spaces.SpaceList)
				if len(resp.Spaces) != 1 {
					t.Errorf("expected 1 space, got %d", len(resp.Spaces))
				}
			},
		},
		{
			name: "List_Filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got, want := q.Get("org_id"), orgID("org-1"); got != want {
					t.Errorf("org_id query: want %q, got %q", want, got)
				}
				if got := q.Get("name"); got != "demo" {
					t.Errorf("name query: want demo, got %q", got)
				}
				if got := q.Get("limit"); got != "25" {
					t.Errorf("limit query: want 25, got %q", got)
				}
				if got := q.Get("cursor"); got != "cursor-abc" {
					t.Errorf("cursor query: want cursor-abc, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spaces.SpaceList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.List(ctx, spaces.ListRequest{
					Organization: orgID("org-1"),
					Name:         "demo",
					Limit:        25,
					Cursor:       "cursor-abc",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// Exercises List's optional org-name resolution: a plain
			// Organization name triggers FindOrganizationID (GET
			// /v2/organizations) before the spaces list, and the resolved ID
			// must flow into the org_id query param.
			name: "List_ResolvesOrgByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/organizations":
					if got := r.URL.Query().Get("name"); got != "my-org" {
						t.Errorf("resolver list name query: want my-org, got %q", got)
					}
					json.NewEncoder(w).Encode(map[string]any{
						"organizations": []map[string]any{{"id": orgID("org-1"), "name": "my-org"}},
						"pagination":    map[string]any{"has_more": false},
					})
				case r.Method == http.MethodGet && r.URL.Path == "/v2/spaces":
					if got, want := r.URL.Query().Get("org_id"), orgID("org-1"); got != want {
						t.Errorf("spaces list org_id query: want %q, got %q", want, got)
					}
					json.NewEncoder(w).Encode(spaces.SpaceList{Pagination: arize.PaginationMetadata{HasMore: false}})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.List(ctx, spaces.ListRequest{Organization: "my-org"})
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
				json.NewEncoder(w).Encode(spaces.Space{Id: spaceID("space-1"), Name: "my-space", CreatedAt: time.Now()})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.Get(ctx, spaces.GetRequest{Space: spaceID("space-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*spaces.Space).Name != "my-space" {
					t.Errorf("unexpected name: %s", got.(*spaces.Space).Name)
				}
			},
		},
		{
			// Exercises resolve.FindSpaceID: when req.Space is a plain name
			// (not a base64 resource ID), the SDK first lists spaces filtered
			// by name to discover the ID, then issues the real GET. This case
			// is the only place that proves the resolver is wired in.
			name: "Get_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/spaces":
					if got := r.URL.Query().Get("name"); got != "my-space" {
						t.Errorf("resolver list name query: want my-space, got %q", got)
					}
					json.NewEncoder(w).Encode(spaces.SpaceList{
						Spaces:     []spaces.Space{{Id: spaceID("space-1"), Name: "my-space", CreatedAt: time.Now()}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/"+spaceID("space-1")):
					json.NewEncoder(w).Encode(spaces.Space{Id: spaceID("space-1"), Name: "my-space", CreatedAt: time.Now()})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.Get(ctx, spaces.GetRequest{Space: "my-space"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*spaces.Space).Name != "my-space" {
					t.Errorf("unexpected name: %s", got.(*spaces.Space).Name)
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
				return c.Spaces.Get(ctx, spaces.GetRequest{Space: spaceID("missing")})
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
				var body wireSpace
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "new-space" {
					t.Errorf("body name: want new-space, got %q", body.Name)
				}
				if body.OrganizationId != orgID("org-1") {
					t.Errorf("body organization_id: want %q, got %q", orgID("org-1"), body.OrganizationId)
				}
				if body.Description == nil || *body.Description != "purpose" {
					t.Errorf("body description: want pointer to %q, got %v", "purpose", body.Description)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(spaces.Space{Id: "space-new", Name: "new-space", CreatedAt: time.Now()})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.Create(ctx, spaces.CreateRequest{
					Name:         "new-space",
					Organization: orgID("org-1"),
					Description:  "purpose",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				s := got.(*spaces.Space)
				if s.Id != "space-new" {
					t.Errorf("unexpected id: %s", s.Id)
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
				return c.Spaces.Create(ctx, spaces.CreateRequest{
					Name:         "duplicate",
					Organization: orgID("org-1"),
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
			// Exercises Create's org-name resolution: a plain Organization name
			// triggers FindOrganizationID before the POST, and the resolved ID
			// must appear as organization_id in the request body.
			name: "Create_ResolvesByOrgName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/organizations":
					json.NewEncoder(w).Encode(map[string]any{
						"organizations": []map[string]any{{"id": orgID("org-1"), "name": "my-org"}},
						"pagination":    map[string]any{"has_more": false},
					})
				case r.Method == http.MethodPost && r.URL.Path == "/v2/spaces":
					var body wireSpace
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("decode body: %v", err)
					}
					if body.OrganizationId != orgID("org-1") {
						t.Errorf("body organization_id: want %q, got %q", orgID("org-1"), body.OrganizationId)
					}
					w.WriteHeader(201)
					json.NewEncoder(w).Encode(spaces.Space{Id: "space-new", Name: "new-space", CreatedAt: time.Now()})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.Create(ctx, spaces.CreateRequest{Name: "new-space", Organization: "my-org"})
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
				var body wireSpace
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "renamed" {
					t.Errorf("body name: want renamed, got %q", body.Name)
				}
				if body.Description == nil || *body.Description != "" {
					t.Errorf("body description: want pointer to empty string (clear), got %v", body.Description)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spaces.Space{Id: spaceID("space-1"), Name: "renamed", CreatedAt: time.Now()})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "renamed"
				cleared := ""
				return c.Spaces.Update(ctx, spaces.UpdateRequest{
					Space:       spaceID("space-1"),
					Name:        &name,
					Description: &cleared,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				s := got.(*spaces.Space)
				if s.Name != "renamed" {
					t.Errorf("unexpected name: %s", s.Name)
				}
			},
		},
		{
			// Verifies the *string + omitempty preserve contract: a nil
			// Description must be omitted from the PATCH body entirely, not sent
			// as "description":"" (which would clear it).
			name: "Update_NameOnly",
			handler: func(w http.ResponseWriter, r *http.Request) {
				raw := map[string]json.RawMessage{}
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if _, ok := raw["description"]; ok {
					t.Errorf("description should be omitted when nil, got keys %v", raw)
				}
				if _, ok := raw["name"]; !ok {
					t.Errorf("name should be present, got keys %v", raw)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spaces.Space{Id: spaceID("space-1"), Name: "renamed", CreatedAt: time.Now()})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "renamed"
				return c.Spaces.Update(ctx, spaces.UpdateRequest{Space: spaceID("space-1"), Name: &name})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// Exercises Update's space-name resolution: a plain Space name
			// triggers FindSpaceID (GET /v2/spaces) before the PATCH on the
			// resolved ID.
			name: "Update_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/spaces":
					json.NewEncoder(w).Encode(spaces.SpaceList{
						Spaces:     []spaces.Space{{Id: spaceID("space-1"), Name: "my-space", CreatedAt: time.Now()}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				case r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/"+spaceID("space-1")):
					json.NewEncoder(w).Encode(spaces.Space{Id: spaceID("space-1"), Name: "renamed", CreatedAt: time.Now()})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "renamed"
				return c.Spaces.Update(ctx, spaces.UpdateRequest{Space: "my-space", Name: &name})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
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
				name := "renamed"
				return c.Spaces.Update(ctx, spaces.UpdateRequest{Space: spaceID("missing"), Name: &name})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
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
				return nil, c.Spaces.Delete(ctx, spaces.DeleteRequest{Space: spaceID("space-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// Exercises Delete's space-name resolution: a plain Space name
			// triggers FindSpaceID before the DELETE on the resolved ID.
			name: "Delete_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/spaces":
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(spaces.SpaceList{
						Spaces:     []spaces.Space{{Id: spaceID("space-1"), Name: "my-space", CreatedAt: time.Now()}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/"+spaceID("space-1")):
					w.WriteHeader(204)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Spaces.Delete(ctx, spaces.DeleteRequest{Space: "my-space"})
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
				return nil, c.Spaces.Delete(ctx, spaces.DeleteRequest{Space: spaceID("missing")})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "AddUser_Predefined",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/users") {
					t.Errorf("expected path ending in /users, got %s", r.URL.Path)
				}
				var body wireMembershipInput
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.UserId != userID("user-1") {
					t.Errorf("body user_id: want %q, got %q", userID("user-1"), body.UserId)
				}
				if got, want := body.Role["type"], "predefined"; got != want {
					t.Errorf("body role.type: want %q, got %v", want, got)
				}
				if got, want := body.Role["name"], "admin"; got != want {
					t.Errorf("body role.name: want %q, got %v", want, got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"id":       "membership-1",
					"user_id":  userID("user-1"),
					"space_id": spaceID("space-1"),
					"role":     map[string]any{"type": "predefined", "name": "admin"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.AddUser(ctx, spaces.AddUserRequest{
					Space:  spaceID("space-1"),
					UserID: userID("user-1"),
					Role:   spaces.AssignPredefinedRole(spaces.UserSpaceRoleAdmin),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				m := got.(*spaces.SpaceMembership)
				if m.Id != "membership-1" {
					t.Errorf("unexpected membership id: %s", m.Id)
				}
				role, err := m.Role.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("decode role: %v", err)
				}
				pre, ok := role.(spaces.PredefinedSpaceRole)
				if !ok {
					t.Fatalf("expected predefined role, got discriminator mismatch")
				}
				if pre.Name != spaces.UserSpaceRoleAdmin {
					t.Errorf("role name: want admin, got %s", pre.Name)
				}
			},
		},
		{
			name: "AddUser_Custom",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireMembershipInput
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if got, want := body.Role["type"], "custom"; got != want {
					t.Errorf("body role.type: want %q, got %v", want, got)
				}
				if got, want := body.Role["id"], "custom-role-1"; got != want {
					t.Errorf("body role.id: want %q, got %v", want, got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"id":       "membership-2",
					"user_id":  userID("user-1"),
					"space_id": spaceID("space-1"),
					"role":     map[string]any{"type": "custom", "id": "custom-role-1"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.AddUser(ctx, spaces.AddUserRequest{
					Space:  spaceID("space-1"),
					UserID: userID("user-1"),
					Role:   spaces.AssignCustomRole("custom-role-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				m := got.(*spaces.SpaceMembership)
				role, err := m.Role.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("decode role: %v", err)
				}
				custom, ok := role.(spaces.CustomSpaceRole)
				if !ok {
					t.Fatalf("expected custom role, got discriminator mismatch")
				}
				if custom.Id != "custom-role-1" {
					t.Errorf("role id: want custom-role-1, got %s", custom.Id)
				}
			},
		},
		{
			// Exercises AddUser's space-name resolution: a plain Space name
			// triggers FindSpaceID before the POST to the resolved space's
			// /users subpath.
			name: "AddUser_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/spaces":
					json.NewEncoder(w).Encode(spaces.SpaceList{
						Spaces:     []spaces.Space{{Id: spaceID("space-1"), Name: "my-space", CreatedAt: time.Now()}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/"+spaceID("space-1")+"/users"):
					json.NewEncoder(w).Encode(map[string]any{
						"id":       "membership-1",
						"user_id":  userID("user-1"),
						"space_id": spaceID("space-1"),
						"role":     map[string]any{"type": "predefined", "name": "admin"},
					})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spaces.AddUser(ctx, spaces.AddUserRequest{
					Space:  "my-space",
					UserID: userID("user-1"),
					Role:   spaces.AssignPredefinedRole(spaces.UserSpaceRoleAdmin),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "RemoveUser",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/"+spaceID("space-1")+"/users/"+userID("user-1")) {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Spaces.RemoveUser(ctx, spaces.RemoveUserRequest{
					Space:  spaceID("space-1"),
					UserID: userID("user-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// Exercises RemoveUser's space-name resolution: a plain Space name
			// triggers FindSpaceID before the DELETE on the resolved space's
			// /users/<userID> subpath.
			name: "RemoveUser_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/spaces":
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(spaces.SpaceList{
						Spaces:     []spaces.Space{{Id: spaceID("space-1"), Name: "my-space", CreatedAt: time.Now()}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/"+spaceID("space-1")+"/users/"+userID("user-1")):
					w.WriteHeader(204)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Spaces.RemoveUser(ctx, spaces.RemoveUserRequest{
					Space:  "my-space",
					UserID: userID("user-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "RemoveUser_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Spaces.RemoveUser(ctx, spaces.RemoveUserRequest{
					Space:  spaceID("space-1"),
					UserID: userID("user-missing"),
				})
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

func TestRoleAccessors(t *testing.T) {
	tests := []struct {
		name       string
		role       spaces.SpaceRoleAssignment
		wantPre    bool
		preName    spaces.UserSpaceRole
		wantCustom bool
		customID   string
	}{
		{
			name:    "Predefined",
			role:    spaces.AssignPredefinedRole(spaces.UserSpaceRoleAdmin),
			wantPre: true,
			preName: spaces.UserSpaceRoleAdmin,
		},
		{
			name:       "Custom",
			role:       spaces.AssignCustomRole("custom-role-1"),
			wantCustom: true,
			customID:   "custom-role-1",
		},
		{
			name: "ZeroUnion",
			role: spaces.SpaceRoleAssignment{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := tt.role.ValueByDiscriminator()
			pre, isPre := v.(spaces.PredefinedSpaceRole)
			if isPre != tt.wantPre {
				t.Errorf("predefined: got %v, want %v", isPre, tt.wantPre)
			}
			if tt.wantPre && pre.Name != tt.preName {
				t.Errorf("predefined name: got %q, want %q", pre.Name, tt.preName)
			}
			c, isCustom := v.(spaces.CustomSpaceRole)
			if isCustom != tt.wantCustom {
				t.Errorf("custom: got %v, want %v", isCustom, tt.wantCustom)
			}
			if tt.wantCustom && c.Id != tt.customID {
				t.Errorf("custom id: got %q, want %q", c.Id, tt.customID)
			}
		})
	}
}

// TestRoleConstructors verifies the Assign* helpers build assignments that
// round-trip through the As* accessors with the right discriminator and payload.
func TestRoleConstructors(t *testing.T) {
	tests := []struct {
		name       string
		role       spaces.SpaceRoleAssignment
		wantPre    bool
		preName    spaces.UserSpaceRole
		wantCustom bool
		customID   string
	}{
		{
			name:    "AssignPredefinedRole",
			role:    spaces.AssignPredefinedRole(spaces.UserSpaceRoleMember),
			wantPre: true,
			preName: spaces.UserSpaceRoleMember,
		},
		{
			name:       "AssignCustomRole",
			role:       spaces.AssignCustomRole("custom-role-1"),
			wantCustom: true,
			customID:   "custom-role-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := tt.role.ValueByDiscriminator()
			pre, isPre := v.(spaces.PredefinedSpaceRole)
			if isPre != tt.wantPre {
				t.Errorf("predefined: got %v, want %v", isPre, tt.wantPre)
			}
			if tt.wantPre && pre.Name != tt.preName {
				t.Errorf("predefined name: got %q, want %q", pre.Name, tt.preName)
			}
			c, isCustom := v.(spaces.CustomSpaceRole)
			if isCustom != tt.wantCustom {
				t.Errorf("custom: got %v, want %v", isCustom, tt.wantCustom)
			}
			if tt.wantCustom && c.Id != tt.customID {
				t.Errorf("custom id: got %q, want %q", c.Id, tt.customID)
			}
		})
	}
}
