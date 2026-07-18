package annotationqueues_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/annotationqueues"
)

func testID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("AnnotationQueue:1:" + suffix))
}

func spaceID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Space:1:" + suffix))
}

func newTestClient(t *testing.T, h http.HandlerFunc) *arize.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := arize.NewClient(arize.Config{APIKey: "key", APIHost: srv.Listener.Addr().String(), APIScheme: "http"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// wireCreate mirrors the JSON shape of the create request body so tests can
// assert what the SDK sent without importing internal/generated.
type wireCreate struct {
	SpaceId             string            `json:"space_id"`
	Name                string            `json:"name"`
	Instructions        *string           `json:"instructions"`
	AssignmentMethod    *string           `json:"assignment_method"`
	AnnotatorEmails     []string          `json:"annotator_emails"`
	AnnotationConfigIds []string          `json:"annotation_config_ids"`
	RecordSources       []json.RawMessage `json:"record_sources"`
}

func TestAnnotationQueues(t *testing.T) {
	updatedName := "updated-queue"

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
				json.NewEncoder(w).Encode(annotationqueues.ListAnnotationQueues{
					AnnotationQueues: []annotationqueues.AnnotationQueue{{Id: "aq-1", Name: "queue-1"}},
					Pagination:       arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.List(ctx, annotationqueues.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*annotationqueues.ListAnnotationQueues)
				if len(resp.AnnotationQueues) != 1 {
					t.Errorf("expected 1 annotation queue, got %d", len(resp.AnnotationQueues))
				}
				if resp.AnnotationQueues[0].Id != "aq-1" {
					t.Errorf("unexpected id: %s", resp.AnnotationQueues[0].Id)
				}
			},
		},
		{
			name: "List_Filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				// "demo" is not a resource ID, so it resolves to a space-name filter.
				if q.Get("name") != "labels" || q.Get("space_name") != "demo" || q.Get("limit") != "25" {
					t.Errorf("query: %v", q)
				}
				if q.Has("space_id") {
					t.Errorf("space_id should be unset for a name filter: %v", q)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"annotation_queues":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.List(ctx, annotationqueues.ListRequest{
					Space: "demo", Name: "labels", Limit: 25,
				})
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
				json.NewEncoder(w).Encode(annotationqueues.AnnotationQueue{Id: "aq-1", Name: "queue-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.Get(ctx, annotationqueues.GetRequest{AnnotationQueue: testID("aq-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				q := got.(*annotationqueues.AnnotationQueue)
				if q.Id != "aq-1" {
					t.Errorf("unexpected id: %s", q.Id)
				}
			},
		},
		{
			name: "Create",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var createBody wireCreate
				if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
					t.Fatalf("decode create body: %v", err)
				}
				if createBody.Name != "new-queue" {
					t.Errorf("name: %q", createBody.Name)
				}
				if createBody.Instructions == nil || *createBody.Instructions != "be thorough" {
					t.Errorf("instructions: %v", createBody.Instructions)
				}
				if createBody.AssignmentMethod == nil || *createBody.AssignmentMethod != "RANDOM" {
					t.Errorf("assignment_method: %v", createBody.AssignmentMethod)
				}
				if len(createBody.AnnotatorEmails) != 1 || createBody.AnnotatorEmails[0] != "annotator@example.com" {
					t.Errorf("annotator_emails: %v", createBody.AnnotatorEmails)
				}
				if len(createBody.AnnotationConfigIds) != 1 || createBody.AnnotationConfigIds[0] != "cfg-1" {
					t.Errorf("annotation_config_ids: %v", createBody.AnnotationConfigIds)
				}
				if len(createBody.RecordSources) != 1 {
					t.Errorf("record_sources: %v", createBody.RecordSources)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(annotationqueues.AnnotationQueue{Id: "aq-2", Name: "new-queue"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				src, err := annotationqueues.NewExampleRecordSource(annotationqueues.AnnotationQueueExampleRecordInput{
					DatasetId: "ds-1",
				})
				if err != nil {
					return nil, err
				}
				return c.AnnotationQueues.Create(ctx, annotationqueues.CreateRequest{
					Space:               spaceID("sp-1"),
					Name:                "new-queue",
					Instructions:        "be thorough",
					AssignmentMethod:    annotationqueues.AssignmentMethodRandom,
					AnnotatorEmails:     []annotationqueues.Email{"annotator@example.com"},
					AnnotationConfigIDs: []string{"cfg-1"},
					RecordSources:       []annotationqueues.AnnotationQueueRecordInput{src},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				q := got.(*annotationqueues.AnnotationQueue)
				if q.Id != "aq-2" {
					t.Errorf("unexpected id: %s", q.Id)
				}
			},
		},
		{
			name: "Update",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(annotationqueues.AnnotationQueue{Id: "aq-1", Name: updatedName})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.Update(ctx, annotationqueues.UpdateRequest{
					AnnotationQueue: testID("aq-1"),
					Name:            &updatedName,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				q := got.(*annotationqueues.AnnotationQueue)
				if q.Name != updatedName {
					t.Errorf("unexpected name: %s", q.Name)
				}
			},
		},
		{
			name: "Delete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.AnnotationQueues.Delete(ctx, annotationqueues.DeleteRequest{AnnotationQueue: testID("aq-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "ListRecords",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(annotationqueues.ListAnnotationQueueRecords{
					Records:    []annotationqueues.AnnotationQueueRecord{{Id: "rec-1", AnnotationQueueId: "aq-1"}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.ListRecords(ctx, annotationqueues.ListRecordsRequest{AnnotationQueue: testID("aq-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*annotationqueues.ListAnnotationQueueRecords)
				if len(resp.Records) != 1 {
					t.Errorf("expected 1 record, got %d", len(resp.Records))
				}
				if resp.Records[0].Id != "rec-1" {
					t.Errorf("unexpected record id: %s", resp.Records[0].Id)
				}
			},
		},
		{
			name: "AddRecords",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(annotationqueues.CreateAnnotationQueueRecord{
					RecordSources: []annotationqueues.AnnotationQueueRecord{{Id: "rec-1", AnnotationQueueId: "aq-1"}},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.AddRecords(ctx, annotationqueues.AddRecordsRequest{
					AnnotationQueue: testID("aq-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				result := got.(*annotationqueues.CreateAnnotationQueueRecord)
				if len(result.RecordSources) != 1 {
					t.Errorf("expected 1 record source, got %d", len(result.RecordSources))
				}
			},
		},
		{
			name: "AddRecords_201",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(annotationqueues.CreateAnnotationQueueRecord{
					RecordSources: []annotationqueues.AnnotationQueueRecord{{Id: "rec-2", AnnotationQueueId: "aq-1"}},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.AddRecords(ctx, annotationqueues.AddRecordsRequest{
					AnnotationQueue: testID("aq-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				result := got.(*annotationqueues.CreateAnnotationQueueRecord)
				if len(result.RecordSources) != 1 {
					t.Errorf("expected 1 record source, got %d", len(result.RecordSources))
				}
				if result.RecordSources[0].Id != "rec-2" {
					t.Errorf("unexpected record id: %s", result.RecordSources[0].Id)
				}
			},
		},
		{
			name: "DeleteRecords",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.AnnotationQueues.DeleteRecords(ctx, annotationqueues.DeleteRecordsRequest{
					AnnotationQueue: testID("aq-1"),
					RecordIDs:       []string{"rec-1"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Annotate",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(annotationqueues.AnnotateAnnotationQueueRecord{
					Id:                "rec-1",
					AnnotationQueueId: "aq-1",
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.Annotate(ctx, annotationqueues.AnnotateRequest{
					AnnotationQueue: testID("aq-1"),
					RecordID:        "rec-1",
					Annotations:     []annotationqueues.AnnotationInput{{Name: "quality"}},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				result := got.(*annotationqueues.AnnotateAnnotationQueueRecord)
				if result.Id != "rec-1" {
					t.Errorf("unexpected id: %s", result.Id)
				}
			},
		},
		{
			name: "Assign",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(annotationqueues.AssignAnnotationQueueRecord{
					Id:                "rec-1",
					AnnotationQueueId: "aq-1",
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.Assign(ctx, annotationqueues.AssignRequest{
					AnnotationQueue:    testID("aq-1"),
					RecordID:           "rec-1",
					AssignedUserEmails: []annotationqueues.Email{"annotator@example.com"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				result := got.(*annotationqueues.AssignAnnotationQueueRecord)
				if result.Id != "rec-1" {
					t.Errorf("unexpected id: %s", result.Id)
				}
			},
		},
		{
			name: "Error_422",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(422)
				_, _ = w.Write([]byte(`{"title":"Unprocessable Entity","detail":"invalid","status":422}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationQueues.List(ctx, annotationqueues.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var apiErr *arize.UnprocessableEntityError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *arize.UnprocessableEntityError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != 422 {
					t.Errorf("status: %d", apiErr.StatusCode)
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
