package spans_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/spans"
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

func TestSpans(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spans.SpanList{
					Spans:      []spans.Span{},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spans.List(ctx, spans.ListRequest{
					ProjectId: "proj-1",
				}, spans.ListParams{})
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
			name: "List not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spans.List(ctx, spans.ListRequest{ProjectId: "nonexistent"}, spans.ListParams{})
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
				return c.Spans.Delete(ctx, spans.DeleteRequest{
					ProjectId: "proj-1",
					SpanIds:   []string{"span-1"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				partial, _ := got.(*spans.SpanDeletePartial)
				if partial != nil {
					t.Errorf("expected nil partial on 204, got %v", partial)
				}
			},
		},
		{
			name: "Delete partial success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spans.SpanDeletePartial{
					DeletedSpanIds: []string{"span-1"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spans.Delete(ctx, spans.DeleteRequest{
					ProjectId: "proj-1",
					SpanIds:   []string{"span-1", "span-2"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				partial, ok := got.(*spans.SpanDeletePartial)
				if !ok || partial == nil {
					t.Fatalf("expected *SpanDeletePartial, got %T", got)
				}
				if len(partial.DeletedSpanIds) != 1 || partial.DeletedSpanIds[0] != "span-1" {
					t.Errorf("unexpected DeletedSpanIds: %v", partial.DeletedSpanIds)
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
				return c.Spans.Delete(ctx, spans.DeleteRequest{
					ProjectId: "nonexistent",
					SpanIds:   []string{"span-1"},
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
