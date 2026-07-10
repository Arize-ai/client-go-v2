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

// wireAnnotationConfigCreate mirrors the JSON shape of the Create request body so tests
// can assert on the wire payload without importing internal/generated.
type wireAnnotationConfigCreate struct {
	Name                  string                           `json:"name"`
	SpaceId               string                           `json:"space_id"`
	AnnotationConfigType  string                           `json:"annotation_config_type"`
	MinimumScore          *float64                         `json:"minimum_score,omitempty"`
	MaximumScore          *float64                         `json:"maximum_score,omitempty"`
	Values                []wireCategoricalAnnotationValue `json:"values,omitempty"`
	OptimizationDirection *string                          `json:"optimization_direction,omitempty"`
}

// wireCategoricalAnnotationValue mirrors the JSON shape of a categorical value.
type wireCategoricalAnnotationValue struct {
	Label string   `json:"label"`
	Score *float64 `json:"score,omitempty"`
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
	var listQueries url.Values

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
			name: "CreateCategorical (typed)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireAnnotationConfigCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.AnnotationConfigType != "categorical" {
					t.Errorf("body annotation_config_type: want categorical, got %v", body.AnnotationConfigType)
				}
				if body.Name != "thumbs" {
					t.Errorf("body name: want thumbs, got %v", body.Name)
				}
				if body.SpaceId != spaceID("sp-1") {
					t.Errorf("body space_id: want %s, got %v", spaceID("sp-1"), body.SpaceId)
				}
				if len(body.Values) != 2 {
					t.Errorf("body values: want 2 entries, got %v", body.Values)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_, _ = w.Write([]byte(`{}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationConfigs.CreateCategorical(ctx, annotationconfigs.CreateCategoricalRequest{
					Space: spaceID("sp-1"),
					Name:  "thumbs",
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
			},
		},
		{
			name: "CreateContinuous (typed)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireAnnotationConfigCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.AnnotationConfigType != "continuous" {
					t.Errorf("body annotation_config_type: want continuous, got %v", body.AnnotationConfigType)
				}
				if body.Name != "score" {
					t.Errorf("body name: want score, got %v", body.Name)
				}
				if body.SpaceId != spaceID("sp-1") {
					t.Errorf("body space_id: want %s, got %v", spaceID("sp-1"), body.SpaceId)
				}
				if body.MinimumScore == nil || *body.MinimumScore != 0 {
					t.Errorf("body minimum_score: want 0, got %v", body.MinimumScore)
				}
				if body.MaximumScore == nil || *body.MaximumScore != 5 {
					t.Errorf("body maximum_score: want 5, got %v", body.MaximumScore)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_, _ = w.Write([]byte(`{}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationConfigs.CreateContinuous(ctx, annotationconfigs.CreateContinuousRequest{
					Space:        spaceID("sp-1"),
					Name:         "score",
					MinimumScore: 0,
					MaximumScore: 5,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "CreateFreeform (typed)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireAnnotationConfigCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.AnnotationConfigType != "freeform" {
					t.Errorf("body annotation_config_type: want freeform, got %v", body.AnnotationConfigType)
				}
				if body.Name != "notes" {
					t.Errorf("body name: want notes, got %v", body.Name)
				}
				if body.SpaceId != spaceID("sp-1") {
					t.Errorf("body space_id: want %s, got %v", spaceID("sp-1"), body.SpaceId)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_, _ = w.Write([]byte(`{}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AnnotationConfigs.CreateFreeform(ctx, annotationconfigs.CreateFreeformRequest{
					Space: spaceID("sp-1"),
					Name:  "notes",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
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
			name: "Update categorical",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				if want := "/v2/annotation-configs/" + testID("cfg-1"); r.URL.Path != want {
					t.Errorf("unexpected path: want %s, got %s", want, r.URL.Path)
				}
				var body map[string]any
				_ = json.NewDecoder(r.Body).Decode(&body)
				if body["annotation_config_type"] != "categorical" {
					t.Errorf("body annotation_config_type: want categorical, got %v", body["annotation_config_type"])
				}
				if body["name"] != "renamed" {
					t.Errorf("body name: want renamed, got %v", body["name"])
				}
				values, ok := body["values"].([]any)
				if !ok || len(values) != 2 {
					t.Errorf("body values: want 2 entries, got %v", body["values"])
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "renamed"
				return c.AnnotationConfigs.UpdateCategorical(ctx, annotationconfigs.UpdateCategoricalRequest{
					AnnotationConfig: testID("cfg-1"),
					Name:             &name,
					Values: &[]annotationconfigs.CategoricalAnnotationValue{
						{Label: "up"},
						{Label: "down"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*annotationconfigs.AnnotationConfig) == nil {
					t.Error("expected non-nil response")
				}
			},
		},
		{
			name: "Update continuous",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				var body map[string]any
				_ = json.NewDecoder(r.Body).Decode(&body)
				if body["annotation_config_type"] != "continuous" {
					t.Errorf("body annotation_config_type: want continuous, got %v", body["annotation_config_type"])
				}
				if body["maximum_score"] != float64(10) {
					t.Errorf("body maximum_score: want 10, got %v", body["maximum_score"])
				}
				if _, ok := body["name"]; ok {
					t.Errorf("body name should be omitted when nil, got %v", body["name"])
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				maxScore := 10.0
				return c.AnnotationConfigs.UpdateContinuous(ctx, annotationconfigs.UpdateContinuousRequest{
					AnnotationConfig: testID("cfg-1"),
					MaximumScore:     &maxScore,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Update freeform",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				var body map[string]any
				_ = json.NewDecoder(r.Body).Decode(&body)
				if body["annotation_config_type"] != "freeform" {
					t.Errorf("body annotation_config_type: want freeform, got %v", body["annotation_config_type"])
				}
				if body["name"] != "renamed" {
					t.Errorf("body name: want renamed, got %v", body["name"])
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				name := "renamed"
				return c.AnnotationConfigs.UpdateFreeform(ctx, annotationconfigs.UpdateFreeformRequest{
					AnnotationConfig: testID("cfg-1"),
					Name:             &name,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*annotationconfigs.AnnotationConfig) == nil {
					t.Error("expected non-nil response")
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
