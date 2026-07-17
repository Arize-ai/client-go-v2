package resourcerestrictions_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/resourcerestrictions"
)

// wireResourceRestrictionCreate mirrors the API JSON body for creating a
// resource restriction. Tests decode incoming request bodies into this struct
// rather than importing the generated package.
type wireResourceRestrictionCreate struct {
	ResourceID string `json:"resource_id"`
}

type wireResourceRestriction struct {
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// wireResourceRestrictionList mirrors the API JSON list response envelope.
type wireResourceRestrictionList struct {
	ResourceRestrictions []wireResourceRestriction `json:"resource_restrictions"`
	Pagination           wirePagination            `json:"pagination"`
}

type wirePagination struct {
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor,omitempty"`
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

func TestResourceRestrictions(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if got := r.URL.Query().Get("resource_type"); got != "PROJECT" {
					t.Errorf("query resource_type: want PROJECT, got %q", got)
				}
				if got := r.URL.Query().Get("limit"); got != "10" {
					t.Errorf("query limit: want 10, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(wireResourceRestrictionList{
					ResourceRestrictions: []wireResourceRestriction{
						{ResourceID: "proj-1", ResourceType: "PROJECT", CreatedAt: time.Now()},
						{ResourceID: "proj-2", ResourceType: "PROJECT", CreatedAt: time.Now()},
					},
					Pagination: wirePagination{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.ResourceRestrictions.List(ctx, resourcerestrictions.ListRequest{
					ResourceType: resourcerestrictions.ResourceRestrictionTypePROJECT,
					Limit:        10,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				list := got.(*resourcerestrictions.ResourceRestrictionList)
				if len(list.ResourceRestrictions) != 2 {
					t.Fatalf("want 2 restrictions, got %d", len(list.ResourceRestrictions))
				}
				if list.ResourceRestrictions[0].ResourceId != "proj-1" {
					t.Errorf("unexpected resource_id: %s", list.ResourceRestrictions[0].ResourceId)
				}
			},
		},
		{
			name: "List unauthorized",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(401)
				json.NewEncoder(w).Encode(map[string]any{"title": "unauthorized", "status": 401})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.ResourceRestrictions.List(ctx, resourcerestrictions.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				var ue *arize.UnauthorizedError
				if !errors.As(err, &ue) {
					t.Errorf("expected *UnauthorizedError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Restrict success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wireResourceRestrictionCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.ResourceID != "proj-1" {
					t.Errorf("body resource_id: want proj-1, got %q", body.ResourceID)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(wireResourceRestriction{
					ResourceID:   "proj-1",
					ResourceType: "PROJECT",
					CreatedAt:    time.Now(),
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.ResourceRestrictions.Restrict(ctx, resourcerestrictions.RestrictRequest{
					ResourceID: "proj-1",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				rr := got.(*resourcerestrictions.ResourceRestriction)
				if rr.ResourceId != "proj-1" {
					t.Errorf("unexpected resource_id: %s", rr.ResourceId)
				}
			},
		},
		{
			name: "Restrict bad request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(400)
				json.NewEncoder(w).Encode(map[string]any{"title": "bad request", "status": 400})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.ResourceRestrictions.Restrict(ctx, resourcerestrictions.RestrictRequest{
					ResourceID: "not-a-project",
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
			name: "Unrestrict success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.ResourceRestrictions.Unrestrict(ctx, resourcerestrictions.UnrestrictRequest{
					ResourceID: "proj-1",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Unrestrict not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.ResourceRestrictions.Unrestrict(ctx, resourcerestrictions.UnrestrictRequest{
					ResourceID: "nonexistent",
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
