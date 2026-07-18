package spans_test

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
	"github.com/Arize-ai/client-go-v2/arize/spans"
)

func projectID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Project:1:" + suffix))
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

// wireSpansList mirrors the JSON shape the API receives for spans.List request
// bodies. Tests use it so they can decode request bodies without importing
// internal/generated.
type wireSpansList struct {
	ProjectId string     `json:"project_id"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Filter    *string    `json:"filter,omitempty"`
}

// wireDeleteSpans mirrors the JSON shape the API receives for spans.Delete
// request bodies.
type wireDeleteSpans struct {
	ProjectId string   `json:"project_id"`
	SpanIds   []string `json:"span_ids"`
}

// wireAnnotateSpans mirrors the JSON shape the API receives for
// spans.Annotate request bodies.
type wireAnnotateSpans struct {
	ProjectId   string               `json:"project_id"`
	Annotations []wireAnnotateRecord `json:"annotations"`
	StartTime   *time.Time           `json:"start_time,omitempty"`
	EndTime     *time.Time           `json:"end_time,omitempty"`
}

type wireAnnotateRecord struct {
	RecordId string                `json:"record_id"`
	Values   []wireAnnotationInput `json:"values"`
}

type wireAnnotationInput struct {
	Name  string   `json:"name"`
	Label *string  `json:"label,omitempty"`
	Score *float64 `json:"score,omitempty"`
	Text  *string  `json:"text,omitempty"`
}

func TestSpans(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wireSpansList
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.ProjectId != projectID("proj-1") {
					t.Errorf("body project_id: want %q, got %q", projectID("proj-1"), body.ProjectId)
				}
				if body.Filter == nil || *body.Filter != "status_code = 'ERROR'" {
					t.Errorf("body filter: %v", body.Filter)
				}
				if r.URL.Query().Get("limit") != "25" {
					t.Errorf("query limit: %q", r.URL.Query().Get("limit"))
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spans.ListSpans{
					Spans:      []spans.Span{},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spans.List(ctx, spans.ListRequest{
					Project: projectID("proj-1"),
					Filter:  "status_code = 'ERROR'",
					Limit:   25,
				})
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
			name: "List_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spans.List(ctx, spans.ListRequest{Project: projectID("nonexistent")})
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
				var body wireDeleteSpans
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.ProjectId != projectID("proj-1") {
					t.Errorf("body project_id: want %q, got %q", projectID("proj-1"), body.ProjectId)
				}
				if len(body.SpanIds) != 1 || body.SpanIds[0] != "span-1" {
					t.Errorf("body span_ids: %v", body.SpanIds)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spans.DeleteSpans{
					Completed:         true,
					DeletedSpanIds:    []string{"span-1"},
					NotDeletedSpanIds: []string{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spans.Delete(ctx, spans.DeleteRequest{
					Project: projectID("proj-1"),
					SpanIDs: []string{"span-1"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				result, ok := got.(*spans.DeleteSpans)
				if !ok || result == nil {
					t.Fatalf("expected *DeleteSpans, got %T", got)
				}
				if !result.Completed {
					t.Errorf("expected Completed=true on full success, got false")
				}
				if len(result.DeletedSpanIds) != 1 || result.DeletedSpanIds[0] != "span-1" {
					t.Errorf("unexpected DeletedSpanIds: %v", result.DeletedSpanIds)
				}
				if len(result.NotDeletedSpanIds) != 0 {
					t.Errorf("expected empty NotDeletedSpanIds, got %v", result.NotDeletedSpanIds)
				}
			},
		},
		{
			name: "Delete_Incomplete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spans.DeleteSpans{
					Completed:         false,
					DeletedSpanIds:    []string{"span-1"},
					NotDeletedSpanIds: []string{"span-2"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Spans.Delete(ctx, spans.DeleteRequest{
					Project: projectID("proj-1"),
					SpanIDs: []string{"span-1", "span-2"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				result, ok := got.(*spans.DeleteSpans)
				if !ok || result == nil {
					t.Fatalf("expected *DeleteSpans, got %T", got)
				}
				if result.Completed {
					t.Errorf("expected Completed=false on incomplete, got true")
				}
				if len(result.DeletedSpanIds) != 1 || result.DeletedSpanIds[0] != "span-1" {
					t.Errorf("unexpected DeletedSpanIds: %v", result.DeletedSpanIds)
				}
				if len(result.NotDeletedSpanIds) != 1 || result.NotDeletedSpanIds[0] != "span-2" {
					t.Errorf("unexpected NotDeletedSpanIds: %v", result.NotDeletedSpanIds)
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
				return c.Spans.Delete(ctx, spans.DeleteRequest{
					Project: projectID("nonexistent"),
					SpanIDs: []string{"span-1"},
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
			name: "Annotate",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wireAnnotateSpans
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.ProjectId != projectID("proj-1") {
					t.Errorf("body project_id: want %q, got %q", projectID("proj-1"), body.ProjectId)
				}
				if len(body.Annotations) != 1 || body.Annotations[0].RecordId != "span-1" {
					t.Errorf("body annotations: %+v", body.Annotations)
				}
				if len(body.Annotations[0].Values) != 1 || body.Annotations[0].Values[0].Name != "Correctness" {
					t.Errorf("body annotation values: %+v", body.Annotations[0].Values)
				}
				if body.Annotations[0].Values[0].Label == nil || *body.Annotations[0].Values[0].Label != "correct" {
					t.Errorf("body annotation label: %v", body.Annotations[0].Values[0].Label)
				}
				if body.StartTime == nil {
					t.Errorf("expected start_time to be forwarded")
				}
				if body.EndTime == nil {
					t.Errorf("expected end_time to be forwarded")
				}
				w.WriteHeader(202)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				label := "correct"
				return nil, c.Spans.Annotate(ctx, spans.AnnotateRequest{
					Project: projectID("proj-1"),
					Annotations: []spans.AnnotateRecordInput{
						{
							RecordId: "span-1",
							Values: []spans.AnnotationInput{
								{Name: "Correctness", Label: &label},
							},
						},
					},
					Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Annotate_DefaultWindow",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// When Start/End are nil the SDK must omit them so the
				// server applies its default 31-day lookup window.
				var body wireAnnotateSpans
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.StartTime != nil {
					t.Errorf("expected start_time omitted, got %v", body.StartTime)
				}
				if body.EndTime != nil {
					t.Errorf("expected end_time omitted, got %v", body.EndTime)
				}
				w.WriteHeader(202)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Spans.Annotate(ctx, spans.AnnotateRequest{
					Project: projectID("proj-1"),
					Annotations: []spans.AnnotateRecordInput{
						{RecordId: "span-1", Values: []spans.AnnotationInput{{Name: "Correctness"}}},
					},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Annotate_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Spans.Annotate(ctx, spans.AnnotateRequest{
					Project: projectID("proj-1"),
					Annotations: []spans.AnnotateRecordInput{
						{RecordId: "missing-span", Values: []spans.AnnotationInput{{Name: "Correctness"}}},
					},
				})
			},
			check: func(t *testing.T, _ any, err error) {
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
