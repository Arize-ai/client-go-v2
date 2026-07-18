package auditlogs_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/auditlogs"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *arize.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := arize.NewClient(arize.Config{
		APIKey: "test-key", APIHost: srv.Listener.Addr().String(), APIScheme: "http",
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestAuditLogs(t *testing.T) {
	userID := "VXNlcjoxMjM0NQ=="

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("limit"); got != "50" {
					t.Errorf("limit query: want %q, got %q", "50", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"logs":[{"id":"QXVkaXRMb2c6MQ==","user_id":"VXNlcjoxMjM0NQ==","ip":"1.2.3.4","operation_type":"QUERY","created_at":"2026-05-18T12:00:00Z"}],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AuditLogs.List(ctx, auditlogs.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*auditlogs.ListAuditLogs)
				if len(resp.Logs) != 1 || resp.Logs[0].Id != "QXVkaXRMb2c6MQ==" {
					t.Errorf("unexpected list: %+v", resp.Logs)
				}
				if resp.Logs[0].OperationType != auditlogs.AuditLogOperationTypeQUERY {
					t.Errorf("unexpected operation type: %s", resp.Logs[0].OperationType)
				}
			},
		},
		{
			name: "List_WithFilters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got := q.Get("user_id"); got != userID {
					t.Errorf("user_id query: want %q, got %q", userID, got)
				}
				if got := q.Get("operation_type"); got != "MUTATION" {
					t.Errorf("operation_type query: want MUTATION, got %q", got)
				}
				if got := q.Get("limit"); got != "10" {
					t.Errorf("limit query: want 10, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"logs":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AuditLogs.List(ctx, auditlogs.ListRequest{
					UserID:        userID,
					OperationType: auditlogs.AuditLogOperationTypeMUTATION,
					Limit:         10,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "List_WithTimeFilters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				wantStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
				if got := r.URL.Query().Get("start_time"); got != wantStart {
					t.Errorf("start_time: want %q, got %q", wantStart, got)
				}
				wantEnd := time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC).Format(time.RFC3339)
				if got := r.URL.Query().Get("end_time"); got != wantEnd {
					t.Errorf("end_time: want %q, got %q", wantEnd, got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"logs":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AuditLogs.List(ctx, auditlogs.ListRequest{
					StartTime: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "List_Pagination",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("cursor"); got != "cursor-prev" {
					t.Errorf("cursor query: want %q, got %q", "cursor-prev", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"logs":[{"id":"QXVkaXRMb2c6Mg==","user_id":"VXNlcjoxMjM0NQ==","ip":"5.6.7.8","operation_type":"SUBSCRIPTION","created_at":"2026-05-19T10:00:00Z"}],"pagination":{"has_more":true,"next_cursor":"cursor-abc"}}`))

			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AuditLogs.List(ctx, auditlogs.ListRequest{Cursor: "cursor-prev"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*auditlogs.ListAuditLogs)
				if !resp.Pagination.HasMore {
					t.Error("expected has_more=true")
				}
				if resp.Pagination.NextCursor == nil || *resp.Pagination.NextCursor != "cursor-abc" {
					t.Errorf("unexpected next_cursor: %v", resp.Pagination.NextCursor)
				}
			},
		},
		{
			name: "List_Forbidden",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"title":"Forbidden","status":403,"detail":"account admin required"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AuditLogs.List(ctx, auditlogs.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Fatal("expected error for 403")
				}
				var fe *arize.ForbiddenError
				if !errors.As(err, &fe) {
					t.Errorf("expected *ForbiddenError, got %T: %v", err, err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, tt.handler)
			got, err := tt.invoke(context.Background(), client)
			tt.check(t, got, err)
		})
	}
}
