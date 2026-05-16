package apikeys_test

import (
	"context"
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

func TestApiKeys(t *testing.T) {
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
				json.NewEncoder(w).Encode(apikeys.ApiKeyList{
					ApiKeys: []apikeys.ApiKey{{
						Id:              "key-1",
						Name:            "my-key",
						CreatedAt:       time.Now(),
						CreatedByUserId: "user-1",
						KeyType:         apikeys.ApiKeyKeyTypeUser,
						RedactedKey:     "ak-abc...xyz",
						Status:          apikeys.ApiKeyStatusActive,
					}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				limit := 10
				return c.APIKeys.List(ctx, apikeys.ListParams{Limit: &limit})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*apikeys.ApiKeyList)
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
				json.NewEncoder(w).Encode(apikeys.ApiKeyList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				kt := apikeys.ListParamsKeyTypeService
				st := apikeys.ApiKeyStatusDeleted
				return c.APIKeys.List(ctx, apikeys.ListParams{
					KeyType: &kt,
					Status:  &st,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if seenKeyType != "service" {
					t.Errorf("key_type query: want service, got %q", seenKeyType)
				}
				if seenStatus != "deleted" {
					t.Errorf("status query: want deleted, got %q", seenStatus)
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
				return c.APIKeys.List(ctx, apikeys.ListParams{})
			},
			check: func(t *testing.T, got any, err error) {
				var fe *arize.ForbiddenError
				if !errors.As(err, &fe) {
					t.Errorf("expected *ForbiddenError, got %T: %v", err, err)
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
				return nil, c.APIKeys.Delete(ctx, "key-1")
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
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
				json.NewEncoder(w).Encode(apikeys.ApiKeyCreated{
					Id:              "key-2",
					Name:            "new-key",
					Key:             "ak-secret",
					RedactedKey:     "ak-sec...ret",
					KeyType:         apikeys.ApiKeyCreatedKeyTypeUser,
					Status:          apikeys.ApiKeyStatusActive,
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
				resp := got.(*apikeys.ApiKeyCreated)
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
				return c.APIKeys.Create(ctx, apikeys.CreateRequest{Name: ""})
			},
			check: func(t *testing.T, got any, err error) {
				var be *arize.BadRequestError
				if !errors.As(err, &be) {
					t.Errorf("expected *BadRequestError, got %T: %v", err, err)
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
				json.NewEncoder(w).Encode(apikeys.ApiKeyCreated{
					Id:              "key-1",
					Name:            "rotated-key",
					Key:             "ak-rotated",
					RedactedKey:     "ak-rot...ted",
					KeyType:         apikeys.ApiKeyCreatedKeyTypeUser,
					Status:          apikeys.ApiKeyStatusActive,
					CreatedAt:       time.Now(),
					CreatedByUserId: "user-1",
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.APIKeys.Refresh(ctx, "key-1", apikeys.RefreshRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*apikeys.ApiKeyCreated)
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
				return c.APIKeys.Refresh(ctx, "nonexistent", apikeys.RefreshRequest{})
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
