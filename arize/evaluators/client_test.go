package evaluators_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
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
// tests can assert on the type discriminator and template config without
// importing internal/generated.
type wireEvaluatorCreate struct {
	Type    string `json:"type"`
	Version struct {
		CommitMessage string `json:"commit_message"`
		CodeConfig    struct {
			Type string `json:"type"`
		} `json:"code_config"`
		TemplateConfig wireTemplateConfig `json:"template_config"`
		RemoteConfig   struct {
			IntegrationId string `json:"integration_id"`
		} `json:"remote_config"`
	} `json:"version"`
}

// wireTemplateConfig captures the template_config fields the SDK is expected to
// serialize. classification_choices is required on write, so tests assert the
// SDK sends the caller's map rather than a nil map (which encodes as null and
// the API rejects).
type wireTemplateConfig struct {
	ClassificationChoices map[string]float64 `json:"classification_choices"`
}

// wireDeleteVersions mirrors the JSON shape of the DeleteVersions request body.
type wireDeleteVersions struct {
	VersionIds []string `json:"version_ids"`
}

// wireCreateVersion mirrors the JSON shape of the CreateVersion request body.
type wireCreateVersion struct {
	CommitMessage string `json:"commit_message"`
	RemoteConfig  struct {
		IntegrationId string `json:"integration_id"`
	} `json:"remote_config"`
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
				resp := got.(*evaluators.ListEvaluators)
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
				if body.Type != "TEMPLATE" {
					t.Errorf("body type: want template, got %q", body.Type)
				}
				wantChoices := map[string]float64{"good": 1, "bad": 0}
				if got := body.Version.TemplateConfig.ClassificationChoices; !reflect.DeepEqual(got, wantChoices) {
					t.Errorf("classification_choices: want %v, got %v", wantChoices, got)
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
						Template: &evaluators.TemplateConfigInput{
							Name:                  "score",
							Template:              "{{input}}",
							ClassificationChoices: &map[string]float32{"good": 1, "bad": 0},
						},
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
				if body.Type != "CODE" {
					t.Errorf("body type: want code, got %q", body.Type)
				}
				if body.Version.CodeConfig.Type != "MANAGED" {
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
				if body.Type != "CODE" {
					t.Errorf("body type: want code, got %q", body.Type)
				}
				if body.Version.CodeConfig.Type != "CUSTOM" {
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
			name: "Create_Remote",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireEvaluatorCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Type != "REMOTE" {
					t.Errorf("body type: want REMOTE, got %q", body.Type)
				}
				if body.Version.RemoteConfig.IntegrationId != "integ-1" {
					t.Errorf("remote_config.integration_id: want integ-1, got %q", body.Version.RemoteConfig.IntegrationId)
				}
				if body.Version.CommitMessage != "initial" {
					t.Errorf("commit_message: want initial, got %q", body.Version.CommitMessage)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ev-5","name":"remote-eval","type":"REMOTE","version":{"id":"ver-r1","type":"REMOTE","remote_config":{"integration_id":"integ-1"}}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Create(ctx, evaluators.CreateRequest{
					Space: testID("sp-1"),
					Name:  "remote-eval",
					Version: evaluators.VersionConfig{
						CommitMessage: "initial",
						Remote:        &evaluators.RemoteConfigInput{IntegrationId: "integ-1"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ev := got.(*evaluators.EvaluatorWithVersion); ev.Id != "ev-5" {
					t.Errorf("unexpected id: %s", ev.Id)
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
						Template:      &evaluators.TemplateConfigInput{Name: "score", Template: "{{input}}"},
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
			name:    "Create_TemplateAndRemote",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.Create(ctx, evaluators.CreateRequest{
					Space: testID("sp-1"),
					Name:  "x",
					Version: evaluators.VersionConfig{
						CommitMessage: "initial",
						Template:      &evaluators.TemplateConfigInput{Name: "score", Template: "{{input}}"},
						Remote:        &evaluators.RemoteConfigInput{IntegrationId: "integ-1"},
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
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got := string(body["name"]); got != `"updated-eval"` {
					t.Errorf("name: want updated-eval, got %q", got)
				}
				if _, ok := body["description"]; ok {
					t.Errorf("description should be omitted when nil, got %q", body["description"])
				}
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
			name: "Update_ClearDescription",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got, ok := body["description"]; !ok || string(got) != "null" {
					t.Errorf("description: want JSON null, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"ev-1"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				empty := ""
				return c.Evaluators.Update(ctx, evaluators.UpdateRequest{Evaluator: testID("ev-1"), Description: &empty})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
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
				_, _ = w.Write([]byte(`{"evaluator_versions":[{"id":"ver-1","type":"TEMPLATE"}],"pagination":{"has_more":false}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.ListVersions(ctx, evaluators.ListVersionsRequest{Evaluator: testID("ev-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if resp := got.(*evaluators.ListEvaluatorVersions); len(resp.EvaluatorVersions) != 1 {
					t.Errorf("expected 1 version, got %d", len(resp.EvaluatorVersions))
				}
			},
		},
		{
			name: "CreateVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					TemplateConfig wireTemplateConfig `json:"template_config"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				wantChoices := map[string]float64{"good": 1, "bad": 0}
				if got := body.TemplateConfig.ClassificationChoices; !reflect.DeepEqual(got, wantChoices) {
					t.Errorf("classification_choices: want %v, got %v", wantChoices, got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ver-2","evaluator_id":"ev-1","type":"TEMPLATE"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateVersion(ctx, evaluators.CreateVersionRequest{
					Evaluator: testID("ev-1"),
					Version: evaluators.VersionConfig{
						CommitMessage: "v2",
						Template: &evaluators.TemplateConfigInput{
							Name:                  "score",
							Template:              "{{input}}",
							ClassificationChoices: &map[string]float32{"good": 1, "bad": 0},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				ver := got.(*evaluators.EvaluatorVersion)
				decoded, err := ver.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("decode version: %v", err)
				}
				tmpl, ok := decoded.(evaluators.EvaluatorVersionTemplate)
				if !ok {
					t.Fatal("expected template variant")
				}
				if tmpl.Id != "ver-2" {
					t.Errorf("unexpected version id: %s", tmpl.Id)
				}
			},
		},
		{
			name: "CreateVersion_Remote",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireCreateVersion
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.RemoteConfig.IntegrationId != "integ-2" {
					t.Errorf("remote_config.integration_id: want integ-2, got %q", body.RemoteConfig.IntegrationId)
				}
				if body.CommitMessage != "remote-v2" {
					t.Errorf("commit_message: want remote-v2, got %q", body.CommitMessage)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ver-r2","evaluator_id":"ev-5","type":"REMOTE","remote_config":{"integration_id":"integ-2"}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateVersion(ctx, evaluators.CreateVersionRequest{
					Evaluator: testID("ev-5"),
					Version: evaluators.VersionConfig{
						CommitMessage: "remote-v2",
						Remote:        &evaluators.RemoteConfigInput{IntegrationId: "integ-2"},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				ver := got.(*evaluators.EvaluatorVersion)
				decoded, err := ver.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("decode version: %v", err)
				}
				remote, ok := decoded.(evaluators.EvaluatorVersionRemote)
				if !ok {
					t.Fatalf("expected remote variant, got %T", decoded)
				}
				if remote.Id != "ver-r2" {
					t.Errorf("unexpected version id: %s", remote.Id)
				}
				if remote.RemoteConfig.IntegrationId == nil || *remote.RemoteConfig.IntegrationId != "integ-2" {
					t.Errorf("remote_config.integration_id: want integ-2, got %v", remote.RemoteConfig.IntegrationId)
				}
			},
		},
		{
			name: "GetVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"ver-1","evaluator_id":"ev-1","type":"TEMPLATE"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.GetVersion(ctx, evaluators.GetVersionRequest{VersionID: "ver-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				ver := got.(*evaluators.EvaluatorVersion)
				decoded, err := ver.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("decode version: %v", err)
				}
				tmpl, ok := decoded.(evaluators.EvaluatorVersionTemplate)
				if !ok {
					t.Fatal("expected template variant")
				}
				if tmpl.Id != "ver-1" || tmpl.EvaluatorId != "ev-1" {
					t.Errorf("unexpected version: id=%s evaluator_id=%s", tmpl.Id, tmpl.EvaluatorId)
				}
			},
		},
		{
			name: "CreateTemplateEvaluator",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireEvaluatorCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Type != "TEMPLATE" {
					t.Errorf("body type: want TEMPLATE, got %q", body.Type)
				}
				if body.Version.CommitMessage != "tmpl-v1" {
					t.Errorf("commit_message: want tmpl-v1, got %q", body.Version.CommitMessage)
				}
				wantChoices := map[string]float64{"good": 1, "bad": 0}
				if got := body.Version.TemplateConfig.ClassificationChoices; !reflect.DeepEqual(got, wantChoices) {
					t.Errorf("classification_choices: want %v, got %v", wantChoices, got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ev-10","name":"tmpl-eval"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateTemplateEvaluator(ctx, evaluators.CreateTemplateEvaluatorRequest{
					Space:         testID("sp-1"),
					Name:          "tmpl-eval",
					CommitMessage: "tmpl-v1",
					Config: evaluators.TemplateConfigInput{
						Name:                  "score",
						Template:              "{{input}}",
						ClassificationChoices: &map[string]float32{"good": 1, "bad": 0},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ev := got.(*evaluators.EvaluatorWithVersion); ev.Id != "ev-10" {
					t.Errorf("unexpected id: %s", ev.Id)
				}
			},
		},
		{
			name: "CreateCodeEvaluator",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireEvaluatorCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Type != "CODE" {
					t.Errorf("body type: want CODE, got %q", body.Type)
				}
				if body.Version.CodeConfig.Type != "MANAGED" {
					t.Errorf("code_config type: want MANAGED, got %q", body.Version.CodeConfig.Type)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ev-11","name":"code-eval-2"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateCodeEvaluator(ctx, evaluators.CreateCodeEvaluatorRequest{
					Space:         testID("sp-1"),
					Name:          "code-eval-2",
					CommitMessage: "code-v1",
					Config: evaluators.CodeConfig{
						Managed: &evaluators.ManagedCodeConfig{
							Name:             "hallucination",
							ManagedEvaluator: "hallucination",
							Variables:        []string{"input"},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ev := got.(*evaluators.EvaluatorWithVersion); ev.Id != "ev-11" {
					t.Errorf("unexpected id: %s", ev.Id)
				}
			},
		},
		{
			name: "CreateRemoteEvaluator",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireEvaluatorCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Type != "REMOTE" {
					t.Errorf("body type: want REMOTE, got %q", body.Type)
				}
				if body.Version.RemoteConfig.IntegrationId != "integ-9" {
					t.Errorf("remote_config.integration_id: want integ-9, got %q", body.Version.RemoteConfig.IntegrationId)
				}
				if body.Version.CommitMessage != "remote-v1" {
					t.Errorf("commit_message: want remote-v1, got %q", body.Version.CommitMessage)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ev-12","name":"remote-eval-2","type":"REMOTE","version":{"id":"ver-r9","type":"REMOTE","remote_config":{"integration_id":"integ-9"}}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateRemoteEvaluator(ctx, evaluators.CreateRemoteEvaluatorRequest{
					Space:         testID("sp-1"),
					Name:          "remote-eval-2",
					CommitMessage: "remote-v1",
					IntegrationID: "integ-9",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ev := got.(*evaluators.EvaluatorWithVersion); ev.Id != "ev-12" {
					t.Errorf("unexpected id: %s", ev.Id)
				}
			},
		},
		{
			name: "CreateTemplateVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					CommitMessage  string             `json:"commit_message"`
					TemplateConfig wireTemplateConfig `json:"template_config"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.CommitMessage != "tmpl-v2" {
					t.Errorf("commit_message: want tmpl-v2, got %q", body.CommitMessage)
				}
				wantChoices := map[string]float64{"pass": 1, "fail": 0}
				if got := body.TemplateConfig.ClassificationChoices; !reflect.DeepEqual(got, wantChoices) {
					t.Errorf("classification_choices: want %v, got %v", wantChoices, got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ver-t2","evaluator_id":"ev-1","type":"TEMPLATE"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateTemplateVersion(ctx, evaluators.CreateTemplateVersionRequest{
					Evaluator:     testID("ev-1"),
					CommitMessage: "tmpl-v2",
					Config: evaluators.TemplateConfigInput{
						Name:                  "score",
						Template:              "{{input}} v2",
						ClassificationChoices: &map[string]float32{"pass": 1, "fail": 0},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ver := got.(*evaluators.EvaluatorVersion); ver == nil {
					t.Fatal("expected non-nil version")
				}
			},
		},
		{
			name: "CreateCodeVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					CommitMessage string `json:"commit_message"`
					CodeConfig    struct {
						Type string `json:"type"`
					} `json:"code_config"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.CommitMessage != "code-v2" {
					t.Errorf("commit_message: want code-v2, got %q", body.CommitMessage)
				}
				if body.CodeConfig.Type != "CUSTOM" {
					t.Errorf("code_config type: want CUSTOM, got %q", body.CodeConfig.Type)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ver-c2","evaluator_id":"ev-3","type":"CODE"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateCodeVersion(ctx, evaluators.CreateCodeVersionRequest{
					Evaluator:     testID("ev-3"),
					CommitMessage: "code-v2",
					Config: evaluators.CodeConfig{
						Custom: &evaluators.CustomCodeConfig{
							Name:      "my-custom",
							Code:      "class Eval:\n    def evaluate(self, **kwargs):\n        return 1",
							Variables: []string{"input"},
						},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if ver := got.(*evaluators.EvaluatorVersion); ver == nil {
					t.Fatal("expected non-nil version")
				}
			},
		},
		{
			name: "CreateRemoteVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireCreateVersion
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.RemoteConfig.IntegrationId != "integ-10" {
					t.Errorf("remote_config.integration_id: want integ-10, got %q", body.RemoteConfig.IntegrationId)
				}
				if body.CommitMessage != "remote-v3" {
					t.Errorf("commit_message: want remote-v3, got %q", body.CommitMessage)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"ver-r3","evaluator_id":"ev-12","type":"REMOTE","remote_config":{"integration_id":"integ-10"}}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.CreateRemoteVersion(ctx, evaluators.CreateRemoteVersionRequest{
					Evaluator:     testID("ev-12"),
					CommitMessage: "remote-v3",
					IntegrationID: "integ-10",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				ver := got.(*evaluators.EvaluatorVersion)
				decoded, err := ver.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("decode version: %v", err)
				}
				remote, ok := decoded.(evaluators.EvaluatorVersionRemote)
				if !ok {
					t.Fatalf("expected remote variant, got %T", decoded)
				}
				if remote.RemoteConfig.IntegrationId == nil || *remote.RemoteConfig.IntegrationId != "integ-10" {
					t.Errorf("remote_config.integration_id: want integ-10, got %v", remote.RemoteConfig.IntegrationId)
				}
			},
		},
		{
			name: "DeleteVersions",
			handler: func(w http.ResponseWriter, r *http.Request) {
				evID := testID("ev-1")
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.URL.Path != "/v2/evaluators/"+evID+"/versions" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var body wireDeleteVersions
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if len(body.VersionIds) != 2 || body.VersionIds[0] != "ver-1" || body.VersionIds[1] != "ver-2" {
					t.Errorf("unexpected version_ids: %v", body.VersionIds)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(evaluators.DeleteEvaluatorVersions{
					Completed:            true,
					DeletedVersionIds:    []string{"ver-1", "ver-2"},
					NotDeletedVersionIds: []string{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.DeleteVersions(ctx, evaluators.DeleteVersionsRequest{
					Evaluator:  testID("ev-1"),
					VersionIDs: []string{"ver-1", "ver-2"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*evaluators.DeleteEvaluatorVersions)
				if !resp.Completed {
					t.Errorf("expected Completed=true on full success, got false")
				}
				if len(resp.DeletedVersionIds) != 2 {
					t.Errorf("expected 2 deleted ids, got %v", resp.DeletedVersionIds)
				}
				if len(resp.NotDeletedVersionIds) != 0 {
					t.Errorf("expected empty NotDeletedVersionIds, got %v", resp.NotDeletedVersionIds)
				}
			},
		},
		{
			name: "DeleteVersions_Partial",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(evaluators.DeleteEvaluatorVersions{
					Completed:            true,
					DeletedVersionIds:    []string{"ver-1"},
					NotDeletedVersionIds: []string{"ver-2"},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.DeleteVersions(ctx, evaluators.DeleteVersionsRequest{
					Evaluator:  testID("ev-1"),
					VersionIDs: []string{"ver-1", "ver-2"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*evaluators.DeleteEvaluatorVersions)
				if !resp.Completed {
					t.Errorf("expected Completed=true, got false")
				}
				if len(resp.NotDeletedVersionIds) != 1 || resp.NotDeletedVersionIds[0] != "ver-2" {
					t.Errorf("unexpected NotDeletedVersionIds: %v", resp.NotDeletedVersionIds)
				}
			},
		},
		{
			name: "DeleteVersions_ServerError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
				_, _ = w.Write([]byte(`{"title":"invalid version"}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Evaluators.DeleteVersions(ctx, evaluators.DeleteVersionsRequest{
					Evaluator:  testID("ev-1"),
					VersionIDs: []string{"ver-1"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				var bre *arize.BadRequestError
				if !errors.As(err, &bre) {
					t.Fatalf("expected *BadRequestError, got %T: %v", err, err)
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
