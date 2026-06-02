package users_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/users"
)

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

func TestUsersList(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "success forwards filters and returns users",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got := q.Get("email"); got != "alice" {
					t.Errorf("email = %q, want alice", got)
				}
				if got := q.Get("limit"); got != "10" {
					t.Errorf("limit = %q, want 10", got)
				}
				if got := q["status"]; len(got) != 1 || got[0] != "active" {
					t.Errorf("status = %v, want [active]", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users.UserList{
					Users: []users.User{{
						Id:    "usr-1",
						Name:  "Alice",
						Email: "alice@example.com",
					}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.List(ctx, users.ListRequest{
					Email:  "alice",
					Status: []users.UserStatus{users.UserStatusActive},
					Limit:  10,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*users.UserList)
				if len(resp.Users) != 1 || resp.Users[0].Name != "Alice" {
					t.Errorf("unexpected users: %+v", resp.Users)
				}
			},
		},
		{
			name: "forbidden maps to typed error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "forbidden", "status": 403})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.List(ctx, users.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				var fe *arize.ForbiddenError
				if !errors.As(err, &fe) {
					t.Errorf("want *ForbiddenError, got %T: %v", err, err)
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

func TestUsersGet(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "by ID hits the get endpoint",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users.User{Id: "usr-1", Name: "Alice", Email: "alice@example.com"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.Get(ctx, users.GetRequest{User: "VXNlcjoxOnVzci0x"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*users.User).Name != "Alice" {
					t.Errorf("unexpected user: %+v", got)
				}
			},
		},
		{
			name: "by email resolves then fetches",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Query().Get("email") != "" {
					_ = json.NewEncoder(w).Encode(users.UserList{
						Users:      []users.User{{Id: "VXNlcjoxOnVzci05", Name: "Bob", Email: "bob@example.com"}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
					return
				}
				if r.URL.Path != "/v2/users/VXNlcjoxOnVzci05" {
					t.Errorf("fetch path = %s, want /v2/users/VXNlcjoxOnVzci05 (resolved ID)", r.URL.Path)
				}
				_ = json.NewEncoder(w).Encode(users.User{Id: "VXNlcjoxOnVzci05", Name: "Bob", Email: "bob@example.com"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.Get(ctx, users.GetRequest{User: "bob@example.com"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*users.User).Name != "Bob" {
					t.Errorf("unexpected user: %+v", got)
				}
			},
		},
		{
			name: "email no-match returns ResourceNotFoundError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users.UserList{
					Users:      []users.User{},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.Get(ctx, users.GetRequest{User: "ghost@example.com"})
			},
			check: func(t *testing.T, got any, err error) {
				var rnfe *arize.ResourceNotFoundError
				if !errors.As(err, &rnfe) {
					t.Errorf("want *ResourceNotFoundError, got %T: %v", err, err)
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

func TestUsersCreate(t *testing.T) {
	type wireRole struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Id   string `json:"id"`
	}
	type wireCreateUser struct {
		Name        string   `json:"name"`
		Email       string   `json:"email"`
		Role        wireRole `json:"role"`
		InviteMode  string   `json:"invite_mode"`
		IsDeveloper *bool    `json:"is_developer"`
	}

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "201 created returns user; sends predefined role body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireCreateUser
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Name != "Carol" || body.Email != "carol@example.com" {
					t.Errorf("unexpected name/email: %+v", body)
				}
				if body.Role.Type != "predefined" || body.Role.Name != "member" {
					t.Errorf("unexpected role: %+v", body.Role)
				}
				if body.InviteMode != "email_link" {
					t.Errorf("invite_mode = %q, want email_link", body.InviteMode)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"usr-3","name":"Carol","email":"carol@example.com","created_at":"2026-01-01T00:00:00Z","status":"invited","is_developer":true,"invite_mode":"email_link","role":{"type":"predefined","name":"member"}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.Create(ctx, users.CreateRequest{
					Name:       "Carol",
					Email:      "carol@example.com",
					Role:       users.AssignPredefinedRole(users.UserRoleMember),
					InviteMode: users.InviteModeEmailLink,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				u := got.(*users.User)
				if u.Id != "usr-3" || u.Name != "Carol" {
					t.Errorf("unexpected user: %+v", u)
				}
			},
		},
		{
			name: "200 idempotent hit returns existing user",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":"usr-existing","name":"Carol","email":"carol@example.com","created_at":"2026-01-01T00:00:00Z","status":"invited","is_developer":true,"role":{"type":"predefined","name":"member"}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.Create(ctx, users.CreateRequest{
					Name:       "Carol",
					Email:      "carol@example.com",
					Role:       users.AssignPredefinedRole(users.UserRoleMember),
					InviteMode: users.InviteModeEmailLink,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*users.User).Id != "usr-existing" {
					t.Errorf("unexpected user: %+v", got)
				}
			},
		},
		{
			name: "unexpected non-payload 2xx returns an error, not a nil user",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// 202 carries no body the parser recognizes, so neither JSON200
				// nor JSON201 is populated. Create must fail loudly rather than
				// return (nil, nil).
				w.WriteHeader(http.StatusAccepted)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.Create(ctx, users.CreateRequest{
					Name:       "Carol",
					Email:      "carol@example.com",
					Role:       users.AssignPredefinedRole(users.UserRoleMember),
					InviteMode: users.InviteModeEmailLink,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Errorf("want error for unexpected status, got nil (user=%+v)", got)
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

func TestUsersUpdate(t *testing.T) {
	type wireUpdate struct {
		Name        *string `json:"name"`
		IsDeveloper *bool   `json:"is_developer"`
	}
	newName := "Renamed"
	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "name only sends just name and returns user",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("method = %s, want PATCH", r.Method)
				}
				var body wireUpdate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if body.Name == nil || *body.Name != "Renamed" {
					t.Errorf("name = %v, want Renamed", body.Name)
				}
				if body.IsDeveloper != nil {
					t.Errorf("is_developer should be omitted, got %v", *body.IsDeveloper)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users.User{Id: "usr-1", Name: "Renamed", Email: "a@example.com"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.Update(ctx, users.UpdateRequest{UserID: "usr-1", Name: &newName})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*users.User).Name != "Renamed" {
					t.Errorf("unexpected user: %+v", got)
				}
			},
		},
		{
			name: "is_developer only sends just the bool and returns user",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireUpdate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if body.Name != nil {
					t.Errorf("name should be omitted, got %v", *body.Name)
				}
				if body.IsDeveloper == nil || *body.IsDeveloper != false {
					t.Errorf("is_developer = %v, want false", body.IsDeveloper)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users.User{Id: "usr-1", Name: "Existing", Email: "a@example.com"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				isDev := false
				return c.Users.Update(ctx, users.UpdateRequest{UserID: "usr-1", IsDeveloper: &isDev})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name:    "no fields returns an error without calling the API",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be hit") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Users.Update(ctx, users.UpdateRequest{UserID: "usr-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Errorf("want error for empty update, got nil")
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

func TestUsersNoPayloadOps(t *testing.T) {
	tests := []struct {
		name       string
		wantMethod string
		wantPath   string
		status     int
		invoke     func(ctx context.Context, c *arize.Client) error
		wantErr    bool
	}{
		{
			name: "delete success", wantMethod: http.MethodDelete, wantPath: "/v2/users/usr-1", status: http.StatusNoContent,
			invoke: func(ctx context.Context, c *arize.Client) error {
				return c.Users.Delete(ctx, users.DeleteRequest{UserID: "usr-1"})
			},
		},
		{
			name: "resend invitation success", wantMethod: http.MethodPost, wantPath: "/v2/users/usr-1/resend-invitation", status: http.StatusNoContent,
			invoke: func(ctx context.Context, c *arize.Client) error {
				return c.Users.ResendInvitation(ctx, users.ResendInvitationRequest{UserID: "usr-1"})
			},
		},
		{
			name: "reset password success", wantMethod: http.MethodPost, wantPath: "/v2/users/usr-1/reset-password", status: http.StatusNoContent,
			invoke: func(ctx context.Context, c *arize.Client) error {
				return c.Users.ResetPassword(ctx, users.ResetPasswordRequest{UserID: "usr-1"})
			},
		},
		{
			name: "delete forbidden maps to typed error", wantMethod: http.MethodDelete, wantPath: "/v2/users/usr-1", status: http.StatusForbidden, wantErr: true,
			invoke: func(ctx context.Context, c *arize.Client) error {
				return c.Users.Delete(ctx, users.DeleteRequest{UserID: "usr-1"})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.wantMethod {
					t.Errorf("method = %s, want %s", r.Method, tt.wantMethod)
				}
				if r.URL.Path != tt.wantPath {
					t.Errorf("path = %s, want %s", r.URL.Path, tt.wantPath)
				}
				if tt.status >= 400 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.status)
					_ = json.NewEncoder(w).Encode(map[string]any{"title": "forbidden", "status": tt.status})
					return
				}
				w.WriteHeader(tt.status)
			})
			err := tt.invoke(context.Background(), client)
			if tt.wantErr {
				var fe *arize.ForbiddenError
				if !errors.As(err, &fe) {
					t.Errorf("want *ForbiddenError, got %T: %v", err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUsersBulkDelete(t *testing.T) {
	t.Run("mixed ids and emails with one missing and one failing", func(t *testing.T) {
		_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && r.URL.Query().Get("email") == "found@example.com":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users.UserList{
					Users:      []users.User{{Id: "usr-ok", Email: "found@example.com"}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			case r.Method == http.MethodGet && r.URL.Query().Get("email") == "ghost@example.com":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users.UserList{Pagination: arize.PaginationMetadata{HasMore: false}})
			case r.Method == http.MethodDelete && r.URL.Path == "/v2/users/usr-bad":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "forbidden", "status": 403})
			case r.Method == http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			}
		})

		got, err := client.Users.BulkDelete(context.Background(), users.BulkDeleteRequest{
			UserIDs: []string{"usr-bad", "usr-good"},
			Emails:  []string{"found@example.com", "ghost@example.com"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		byKey := map[string]users.DeletionStatus{}
		emailByKey := map[string]string{}
		for _, r := range got {
			key := r.UserID
			if key == "" {
				key = r.Email
			}
			byKey[key] = r.Status
			emailByKey[key] = r.Email
		}
		if byKey["usr-bad"] != users.DeletionStatusFailed {
			t.Errorf("usr-bad status = %q, want failed", byKey["usr-bad"])
		}
		if byKey["usr-good"] != users.DeletionStatusDeleted {
			t.Errorf("usr-good status = %q, want deleted", byKey["usr-good"])
		}
		if byKey["usr-ok"] != users.DeletionStatusDeleted {
			t.Errorf("usr-ok status = %q, want deleted", byKey["usr-ok"])
		}
		// usr-ok was resolved from an email, so the result carries that email.
		if emailByKey["usr-ok"] != "found@example.com" {
			t.Errorf("usr-ok email = %q, want found@example.com", emailByKey["usr-ok"])
		}
		if byKey["ghost@example.com"] != users.DeletionStatusNotFound {
			t.Errorf("ghost status = %q, want not_found", byKey["ghost@example.com"])
		}
	})

	t.Run("empty request returns an error", func(t *testing.T) {
		_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			t.Error("server should not be hit")
		})
		if _, err := client.Users.BulkDelete(context.Background(), users.BulkDeleteRequest{}); err == nil {
			t.Errorf("want error for empty bulk delete, got nil")
		}
	})

	t.Run("non-NotFound resolver error aborts the batch", func(t *testing.T) {
		_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			// Email resolution lists users; a 403 there is an auth failure, not a
			// missing user, so BulkDelete must abort rather than record NotFound.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "forbidden", "status": 403})
		})
		got, err := client.Users.BulkDelete(context.Background(), users.BulkDeleteRequest{
			Emails: []string{"alice@example.com"},
		})
		var fe *arize.ForbiddenError
		if !errors.As(err, &fe) {
			t.Errorf("want *ForbiddenError, got %T: %v", err, err)
		}
		if got != nil {
			t.Errorf("want nil results on abort, got %+v", got)
		}
	})
}
