package evaluators_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/evaluators"
)

func testID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Evaluator:1:" + suffix))
}

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

// wireEvaluatorCreate mirrors the JSON shape of the Create request body so
// tests can assert on the type discriminator without importing
// internal/generated.
type wireEvaluatorCreate struct {
	Type    string `json:"type"`
	Version struct {
		CodeConfig struct {
			Type string `json:"type"`
		} `json:"code_config"`
	} `json:"version"`
}

func TestEvaluators(t *testing.T) {
	updateName := "updated-eval"

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
				_, _ = w.Write([]byte(`{"evaluators":[{"id":"ev-1","name":"a"}],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.List(ctx, evaluators.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*evaluators.EvaluatorList)
				if len(resp.Evaluators) != 1 || resp.Evaluators[0].Id != "ev-1" {
					t.Errorf("unexpected list: %+v", resp.Evaluators)
				}
			},
		},
		{
			name: "List_Filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got, want := q.Get("space_id"), testID("sp-1"); got != want {
					t.Errorf("space_id query: want %q, got %q", want, got)
				}
				if got := q.Get("name"); got != "score" {
					t.Errorf("name query: want score, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"evaluators":[],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.List(ctx, evaluators.ListRequest{Space: testID("sp-1"), Name: "score"})
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
				_, _ = w.Write([]byte(`{"id":"ev-1","name":"my-eval"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Get(ctx, evaluators.GetRequest{Evaluator: testID("ev-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ev := got.(*evaluators.EvaluatorWithVersion); ev.Id != "ev-1" {
					t.Errorf("unexpected id: %s", ev.Id)
				}
			},
		},
		{
			name: "Get_WithVersionID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("version_id"); got != "ver-7" {
					t.Errorf("version_id query: want ver-7, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"ev-1"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Get(ctx, evaluators.GetRequest{Evaluator: testID("ev-1"), VersionID: "ver-7"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Create_Template",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireEvaluatorCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Type != "template" {
					t.Errorf("body type: want template, got %q", body.Type)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ev-2","name":"new-eval"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Create(ctx, evaluators.CreateRequest{
					Space: testID("sp-1"),
					Name:  "new-eval",
					Version: evaluators.VersionConfig{
						CommitMessage: "initial",
						Template:      &evaluators.TemplateConfig{Name: "score", Template: "{{input}}"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ev := got.(*evaluators.EvaluatorWithVersion); ev.Id != "ev-2" {
					t.Errorf("unexpected id: %s", ev.Id)
				}
			},
		},
		{
			name: "Create_CodeManaged",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireEvaluatorCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Type != "code" {
					t.Errorf("body type: want code, got %q", body.Type)
				}
				if body.Version.CodeConfig.Type != "managed" {
					t.Errorf("code_config type: want managed, got %q", body.Version.CodeConfig.Type)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ev-3","name":"code-eval"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Create(ctx, evaluators.CreateRequest{
					Space: testID("sp-1"),
					Name:  "code-eval",
					Version: evaluators.VersionConfig{
						CommitMessage: "initial",
						Code: &evaluators.CodeConfig{
							Managed: &evaluators.ManagedCodeConfig{
								Name:             "hallucination",
								ManagedEvaluator: "hallucination",
								Variables:        []string{"input"},
							},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Create_CodeCustom",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireEvaluatorCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Type != "code" {
					t.Errorf("body type: want code, got %q", body.Type)
				}
				if body.Version.CodeConfig.Type != "custom" {
					t.Errorf("code_config type: want custom, got %q", body.Version.CodeConfig.Type)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ev-4","name":"custom-eval"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Create(ctx, evaluators.CreateRequest{
					Space: testID("sp-1"),
					Name:  "custom-eval",
					Version: evaluators.VersionConfig{
						CommitMessage: "initial",
						Code: &evaluators.CodeConfig{
							Custom: &evaluators.CustomCodeConfig{
								Name:      "my-custom",
								Code:      "class Eval:\n    def evaluate(self, **kwargs):\n        return 1",
								Variables: []string{"input"},
							},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ev := got.(*evaluators.EvaluatorWithVersion); ev.Id != "ev-4" {
					t.Errorf("unexpected id: %s", ev.Id)
				}
			},
		},
		{
			name:    "Create_NoConfig",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Create(ctx, evaluators.CreateRequest{Space: testID("sp-1"), Name: "x"})
			},
			check: func(t *testing.T, got any, err error) {
				if err == nil {
					t.Fatal("expected error for missing version config")
				}
			},
		},
		{
			name:    "Create_TemplateAndCode",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Create(ctx, evaluators.CreateRequest{
					Space: testID("sp-1"),
					Name:  "x",
					Version: evaluators.VersionConfig{
						CommitMessage: "initial",
						Template:      &evaluators.TemplateConfig{Name: "score", Template: "{{input}}"},
						Code: &evaluators.CodeConfig{
							Managed: &evaluators.ManagedCodeConfig{
								Name:             "hallucination",
								ManagedEvaluator: "hallucination",
								Variables:        []string{"input"},
							},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if !errors.Is(err, evaluators.ErrConflictingVersionConfig) {
					t.Fatalf("want ErrConflictingVersionConfig, got %v", err)
				}
			},
		},
		{
			name:    "Create_ManagedAndCustom",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Create(ctx, evaluators.CreateRequest{
					Space: testID("sp-1"),
					Name:  "x",
					Version: evaluators.VersionConfig{
						CommitMessage: "initial",
						Code: &evaluators.CodeConfig{
							Managed: &evaluators.ManagedCodeConfig{
								Name:             "hallucination",
								ManagedEvaluator: "hallucination",
								Variables:        []string{"input"},
							},
							Custom: &evaluators.CustomCodeConfig{
								Name:      "my-custom",
								Code:      "class Eval:\n    def evaluate(self, **kwargs):\n        return 1",
								Variables: []string{"input"},
							},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if !errors.Is(err, evaluators.ErrConflictingCodeConfig) {
					t.Fatalf("want ErrConflictingCodeConfig, got %v", err)
				}
			},
		},
		{
			name: "Update",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"ev-1","name":"updated-eval"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Update(ctx, evaluators.UpdateRequest{Evaluator: testID("ev-1"), Name: &updateName})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ev := got.(*evaluators.Evaluator); ev.Name != updateName {
					t.Errorf("unexpected name: %s", ev.Name)
				}
			},
		},
		{
			name:    "Update_NoFields",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Update(ctx, evaluators.UpdateRequest{Evaluator: testID("ev-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if !errors.Is(err, evaluators.ErrNoUpdateFields) {
					t.Fatalf("want ErrNoUpdateFields, got %v", err)
				}
			},
		},
		{
			name: "Delete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Evaluators.Delete(ctx, evaluators.DeleteRequest{Evaluator: testID("ev-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "ListVersions",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"evaluator_versions":[{"id":"ver-1","type":"template"}],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.ListVersions(ctx, evaluators.ListVersionsRequest{Evaluator: testID("ev-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if resp := got.(*evaluators.EvaluatorVersionList); len(resp.EvaluatorVersions) != 1 {
					t.Errorf("expected 1 version, got %d", len(resp.EvaluatorVersions))
				}
			},
		},
		{
			name: "CreateVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ver-2","evaluator_id":"ev-1","type":"template"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateVersion(ctx, evaluators.CreateVersionRequest{
					Evaluator: testID("ev-1"),
					Version: evaluators.VersionConfig{
						CommitMessage: "v2",
						Template:      &evaluators.TemplateConfig{Name: "score", Template: "{{input}}"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				ver := got.(*evaluators.EvaluatorVersion)
				tmpl, ok := evaluators.AsTemplate(*ver)
				if !ok {
					t.Fatal("expected template variant")
				}
				if tmpl.Id != "ver-2" {
					t.Errorf("unexpected version id: %s", tmpl.Id)
				}
			},
		},
		{
			name: "GetVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"ver-1","evaluator_id":"ev-1","type":"template"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.GetVersion(ctx, evaluators.GetVersionRequest{VersionID: "ver-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				ver := got.(*evaluators.EvaluatorVersion)
				tmpl, ok := evaluators.AsTemplate(*ver)
				if !ok {
					t.Fatal("expected template variant")
				}
				if tmpl.Id != "ver-1" || tmpl.EvaluatorId != "ev-1" {
					t.Errorf("unexpected version: id=%s evaluator_id=%s", tmpl.Id, tmpl.EvaluatorId)
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
