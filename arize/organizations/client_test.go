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

// wireOrganization mirrors the JSON shape the API sends/receives. Tests use it
// so they can both decode request bodies and assert on the JSON keys without
// importing internal/generated.
type wireOrganization struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// wireMembershipInput mirrors the JSON shape sent by AddUser. The role field
// is decoded as a raw map so tests can assert on the discriminated-union
// payload without depending on the generated wrapper.
type wireMembershipInput struct {
	UserId string         `json:"user_id"`
	Role   map[string]any `json:"role"`
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
				return c.Organizations.List(ctx, organizations.ListRequest{Limit: 10})
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
				return c.Organizations.List(ctx, organizations.ListRequest{Name: "acme"})
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
			name: "Get",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(organizations.Organization{Id: testID("org-1"), Name: "my-org", CreatedAt: time.Now()})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Organizations.Get(ctx, organizations.GetRequest{Organization: testID("org-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*organizations.Organization).Name != "my-org" {
					t.Errorf("unexpected name: %s", got.(*organizations.Organization).Name)
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
				return c.Organizations.Get(ctx, organizations.GetRequest{Organization: testID("missing")})
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
				var body wireOrganization
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
				var body wireOrganization
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "renamed" {
					t.Errorf("body name: want renamed, got %q", body.Name)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(organizations.Organization{Id: testID("org-1"), Name: "renamed", CreatedAt: time.Now()})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "renamed"
				return c.Organizations.Update(ctx, organizations.UpdateRequest{
					Organization: testID("org-1"),
					Name:         &name,
				})
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
				name := "renamed"
				return c.Organizations.Update(ctx, organizations.UpdateRequest{
					Organization: testID("missing"),
					Name:         &name,
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
			name: "Delete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Organizations.Delete(ctx, organizations.DeleteRequest{
					Organization: testID("org-1"),
				})
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
				return nil, c.Organizations.Delete(ctx, organizations.DeleteRequest{
					Organization: testID("missing"),
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
			name: "AddUser",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wireMembershipInput
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.UserId != "user-1" {
					t.Errorf("body user_id: want user-1, got %q", body.UserId)
				}
				if got, want := body.Role["type"], "predefined"; got != want {
					t.Errorf("body role.type: want %q, got %v", want, got)
				}
				if got, want := body.Role["name"], "admin"; got != want {
					t.Errorf("body role.name: want %q, got %v", want, got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"id":              "membership-1",
					"user_id":         "user-1",
					"organization_id": testID("org-1"),
					"role":            map[string]any{"type": "predefined", "name": "admin"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Organizations.AddUser(ctx, organizations.AddUserRequest{
					Organization: testID("org-1"),
					UserID:       "user-1",
					Role:         organizations.PredefinedOrgRole{Name: organizations.OrganizationRoleAdmin},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				m := got.(*organizations.OrganizationMembership)
				if m.Id != "membership-1" {
					t.Errorf("unexpected membership id: %s", m.Id)
				}
				if m.UserId != "user-1" {
					t.Errorf("unexpected user id: %s", m.UserId)
				}
				role, err := m.Role.AsOrganizationPredefinedRoleAssignment()
				if err != nil {
					t.Fatalf("decode role: %v", err)
				}
				if role.Name != organizations.OrganizationRoleAdmin {
					t.Errorf("role name: want admin, got %s", role.Name)
				}
			},
		},
		{
			name: "RemoveUser",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Organizations.RemoveUser(ctx, organizations.RemoveUserRequest{
					Organization: testID("org-1"),
					UserID:       "user-1",
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
				return nil, c.Organizations.RemoveUser(ctx, organizations.RemoveUserRequest{
					Organization: testID("org-1"),
					UserID:       "user-missing",
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
	predefined := func() organizations.OrganizationRoleAssignment {
		var r organizations.OrganizationRoleAssignment
		if err := r.FromOrganizationPredefinedRoleAssignment(organizations.PredefinedOrgRole{
			Name: organizations.OrganizationRoleAdmin,
		}); err != nil {
			t.Fatalf("seed predefined: %v", err)
		}
		return r
	}
	custom := func() organizations.OrganizationRoleAssignment {
		var r organizations.OrganizationRoleAssignment
		if err := r.FromOrganizationCustomRoleAssignment(organizations.CustomOrgRole{
			Id: "custom-role-1",
		}); err != nil {
			t.Fatalf("seed custom: %v", err)
		}
		return r
	}

	tests := []struct {
		name       string
		role       organizations.OrganizationRoleAssignment
		wantPre    bool
		preName    organizations.OrganizationRole
		wantCustom bool
		customID   string
	}{
		{
			name:    "Predefined",
			role:    predefined(),
			wantPre: true,
			preName: organizations.OrganizationRoleAdmin,
		},
		{
			name:       "Custom",
			role:       custom(),
			wantCustom: true,
			customID:   "custom-role-1",
		},
		{
			name: "ZeroUnion",
			role: organizations.OrganizationRoleAssignment{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pre, ok := organizations.AsPredefined(tt.role)
			if ok != tt.wantPre {
				t.Errorf("AsPredefined ok: got %v, want %v", ok, tt.wantPre)
			}
			if tt.wantPre && pre.Name != tt.preName {
				t.Errorf("AsPredefined name: got %q, want %q", pre.Name, tt.preName)
			}
			c, ok := organizations.AsCustom(tt.role)
			if ok != tt.wantCustom {
				t.Errorf("AsCustom ok: got %v, want %v", ok, tt.wantCustom)
			}
			if tt.wantCustom && c.Id != tt.customID {
				t.Errorf("AsCustom id: got %q, want %q", c.Id, tt.customID)
			}
		})
	}
}
