package apikeys_test

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
	"github.com/Arize-ai/client-go-v2/arize/apikeys"
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

func TestAPIKeys(t *testing.T) {
	// seenKeyType and seenStatus are shared between the handler and check closures of
	// "List new filters". Tests run sequentially so this is safe.
	var seenKeyType, seenStatus string

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(apikeys.APIKeyList{
					ApiKeys: []apikeys.APIKeyRedacted{{
						Id:              "key-1",
						Name:            "my-key",
						CreatedAt:       time.Now(),
						CreatedByUserId: "user-1",
						KeyType:         apikeys.APIKeyTypeUser,
						RedactedKey:     "ak-abc...xyz",
						Status:          apikeys.APIKeyStatusActive,
					}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.List(ctx, apikeys.ListRequest{Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*apikeys.APIKeyList)
				if len(resp.ApiKeys) != 1 {
					t.Errorf("expected 1 key, got %d", len(resp.ApiKeys))
				}
			},
		},
		{
			name: "List new filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				seenKeyType = r.URL.Query().Get("key_type")
				seenStatus = r.URL.Query().Get("status")
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(apikeys.APIKeyList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.List(ctx, apikeys.ListRequest{
					KeyType: apikeys.APIKeyTypeService,
					Status:  apikeys.APIKeyStatusRevoked,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if seenKeyType != "SERVICE" {
					t.Errorf("key_type query: want service, got %q", seenKeyType)
				}
				if seenStatus != "REVOKED" {
					t.Errorf("status query: want revoked, got %q", seenStatus)
				}
			},
		},
		{
			name: "List default limit",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("limit"); got != "50" {
					t.Errorf("expected limit=50, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(apikeys.APIKeyList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.List(ctx, apikeys.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "List error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(403)
				json.NewEncoder(w).Encode(map[string]any{"title": "forbidden", "status": 403})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.List(ctx, apikeys.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				var fe *arize.ForbiddenError
				if !errors.As(err, &fe) {
					t.Errorf("expected *ForbiddenError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Revoke success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.APIKeys.Revoke(ctx, apikeys.RevokeRequest{APIKeyID: "key-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Revoke not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.APIKeys.Revoke(ctx, apikeys.RevokeRequest{APIKeyID: "nonexistent"})
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(apikeys.APIKey{
					Id:              "key-2",
					Name:            "new-key",
					Key:             "ak-secret",
					KeyType:         apikeys.APIKeyTypeUser,
					Status:          apikeys.APIKeyStatusActive,
					CreatedAt:       time.Now(),
					CreatedByUserId: "user-1",
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Create(ctx, apikeys.CreateRequest{Name: "new-key"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*apikeys.APIKey)
				if resp.Id != "key-2" {
					t.Errorf("expected Id %q, got %q", "key-2", resp.Id)
				}
				if resp.Key == "" {
					t.Error("expected non-empty Key")
				}
			},
		},
		{
			name: "Create error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]any{"title": "bad request", "status": 400})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Create(ctx, apikeys.CreateRequest{Name: "valid-name"})
			},
			check: func(t *testing.T, got any, err error) {
				var be *arize.BadRequestError
				if !errors.As(err, &be) {
					t.Errorf("expected *BadRequestError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "CreateServiceKey success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(apikeys.APIKey{
					Id:              "key-3",
					Name:            "svc-key",
					Key:             "ak-service-secret",
					KeyType:         apikeys.APIKeyTypeService,
					Status:          apikeys.APIKeyStatusActive,
					CreatedAt:       time.Now(),
					CreatedByUserId: "user-1",
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				spaceID := base64.StdEncoding.EncodeToString([]byte("Space:test-space-id"))
				return c.APIKeys.CreateServiceKey(ctx, apikeys.CreateServiceKeyRequest{
					Name:      "svc-key",
					Space:     spaceID,
					SpaceRole: apikeys.APIKeySpaceRoleMember,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*apikeys.APIKey)
				if resp.KeyType != apikeys.APIKeyTypeService {
					t.Errorf("expected service key type, got %v", resp.KeyType)
				}
				if resp.Key == "" {
					t.Error("expected non-empty Key")
				}
			},
		},
		{
			name:    "CreateServiceKey validation: missing space",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.CreateServiceKey(ctx, apikeys.CreateServiceKeyRequest{Name: "svc-key", Space: ""})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Fatal("expected error for missing space, got nil")
				}
			},
		},
		{
			name: "Refresh success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(apikeys.APIKey{
					Id:              "key-1",
					Name:            "rotated-key",
					Key:             "ak-rotated",
					KeyType:         apikeys.APIKeyTypeUser,
					Status:          apikeys.APIKeyStatusActive,
					CreatedAt:       time.Now(),
					CreatedByUserId: "user-1",
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Refresh(ctx, apikeys.RefreshRequest{APIKeyID: "key-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*apikeys.APIKey)
				if resp.Key == "" {
					t.Error("expected non-empty Key on refresh")
				}
			},
		},
		{
			name: "Refresh not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Refresh(ctx, apikeys.RefreshRequest{APIKeyID: "nonexistent"})
			},
			check: func(t *testing.T, got any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Refresh with grace period sends correct wire body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var wire struct {
					GracePeriodSeconds *int `json:"grace_period_seconds"`
					ExpiresAt          any  `json:"expires_at"`
				}
				if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				if wire.GracePeriodSeconds == nil {
					t.Error("expected grace_period_seconds in request body, got nil")
				} else if *wire.GracePeriodSeconds != 3600 {
					t.Errorf("grace_period_seconds: want 3600, got %d", *wire.GracePeriodSeconds)
				}
				if wire.ExpiresAt != nil {
					t.Errorf("expires_at should be omitted when zero, got %v", wire.ExpiresAt)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(apikeys.APIKey{
					Id: "key-1", Key: "ak-new", KeyType: apikeys.APIKeyTypeUser,
					Status: apikeys.APIKeyStatusActive, CreatedAt: time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Refresh(ctx, apikeys.RefreshRequest{
					APIKeyID:           "key-1",
					GracePeriodSeconds: 3600,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Refresh zero grace period omits field from wire body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Decode into a raw map so we can distinguish missing key from zero value.
				var wire map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				if _, present := wire["grace_period_seconds"]; present {
					t.Error("grace_period_seconds should be omitted when GracePeriodSeconds is 0")
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(apikeys.APIKey{
					Id: "key-1", Key: "ak-new", KeyType: apikeys.APIKeyTypeUser,
					Status: apikeys.APIKeyStatusActive, CreatedAt: time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Refresh(ctx, apikeys.RefreshRequest{APIKeyID: "key-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Refresh with grace period and expires_at sends both fields",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var wire struct {
					GracePeriodSeconds *int    `json:"grace_period_seconds"`
					ExpiresAt          *string `json:"expires_at"`
				}
				if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				if wire.GracePeriodSeconds == nil {
					t.Error("expected grace_period_seconds in request body, got nil")
				} else if *wire.GracePeriodSeconds != 86400 {
					t.Errorf("grace_period_seconds: want 86400 (max), got %d", *wire.GracePeriodSeconds)
				}
				if wire.ExpiresAt == nil {
					t.Error("expected expires_at in request body, got nil")
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(apikeys.APIKey{
					Id: "key-1", Key: "ak-new", KeyType: apikeys.APIKeyTypeUser,
					Status: apikeys.APIKeyStatusActive, CreatedAt: time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Refresh(ctx, apikeys.RefreshRequest{
					APIKeyID:           "key-1",
					ExpiresAt:          time.Now().Add(30 * 24 * time.Hour),
					GracePeriodSeconds: 86400,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Refresh with grace period propagates server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]any{"title": "grace_period_seconds exceeds maximum", "status": 400})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Refresh(ctx, apikeys.RefreshRequest{
					APIKeyID:           "key-1",
					GracePeriodSeconds: 999999,
				})
			},
			check: func(t *testing.T, got any, err error) {
				var be *arize.BadRequestError
				if !errors.As(err, &be) {
					t.Errorf("expected *BadRequestError, got %T: %v", err, err)
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
