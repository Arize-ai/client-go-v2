package annotationconfigs_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/annotationconfigs"
)

func testID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("AnnotationConfig:1:" + suffix))
}

func spaceID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Space:1:" + suffix))
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

func TestAnnotationConfigs(t *testing.T) {
	var (
		createBody  map[string]any
		listQueries url.Values
	)

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/annotation-configs" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"annotationconfigs":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationConfigs.List(ctx, annotationconfigs.ListRequest{Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*annotationconfigs.AnnotationConfigList) == nil {
					t.Error("expected non-nil response")
				}
			},
		},
		{
			name: "List with filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				listQueries = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"annotationconfigs":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationConfigs.List(ctx, annotationconfigs.ListRequest{
					Name:  "thumbs",
					Space: spaceID("sp-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if listQueries.Get("name") != "thumbs" {
					t.Errorf("query name: want thumbs, got %q", listQueries.Get("name"))
				}
				if listQueries.Get("space_id") != spaceID("sp-1") {
					t.Errorf("query space_id: want %s, got %q", spaceID("sp-1"), listQueries.Get("space_id"))
				}
			},
		},
		{
			name: "Create categorical",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				_ = json.NewDecoder(r.Body).Decode(&createBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_, _ = w.Write([]byte(`{}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationConfigs.Create(ctx, annotationconfigs.CreateRequest{
					Space: spaceID("sp-1"),
					Name:  "thumbs",
					Type:  annotationconfigs.AnnotationConfigTypeCategorical,
					Values: []annotationconfigs.CategoricalAnnotationValue{
						{Label: "up"},
						{Label: "down"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if createBody["annotation_config_type"] != "categorical" {
					t.Errorf("body annotation_config_type: want categorical, got %v", createBody["annotation_config_type"])
				}
				if createBody["name"] != "thumbs" {
					t.Errorf("body name: want thumbs, got %v", createBody["name"])
				}
				if createBody["space_id"] != spaceID("sp-1") {
					t.Errorf("body space_id: want %s, got %v", spaceID("sp-1"), createBody["space_id"])
				}
				values, ok := createBody["values"].([]any)
				if !ok || len(values) != 2 {
					t.Errorf("body values: want 2 entries, got %v", createBody["values"])
				}
			},
		},
		{
			name: "Create continuous",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&createBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_, _ = w.Write([]byte(`{}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationConfigs.Create(ctx, annotationconfigs.CreateRequest{
					Space:        spaceID("sp-1"),
					Name:         "score",
					Type:         annotationconfigs.AnnotationConfigTypeContinuous,
					MinimumScore: 0,
					MaximumScore: 5,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if createBody["annotation_config_type"] != "continuous" {
					t.Errorf("body annotation_config_type: want continuous, got %v", createBody["annotation_config_type"])
				}
				if createBody["minimum_score"] != float64(0) {
					t.Errorf("body minimum_score: want 0, got %v", createBody["minimum_score"])
				}
				if createBody["maximum_score"] != float64(5) {
					t.Errorf("body maximum_score: want 5, got %v", createBody["maximum_score"])
				}
			},
		},
		{
			name: "Get not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationConfigs.Get(ctx, annotationconfigs.GetRequest{AnnotationConfig: testID("missing")})
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
				return nil, c.AnnotationConfigs.Delete(ctx, annotationconfigs.DeleteRequest{AnnotationConfig: testID("cfg-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
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
