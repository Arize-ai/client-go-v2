package prompts_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/prompts"
)

// testID returns a base64-encoded ID so resolve.IsResourceID treats it as an
// ID and skips name-resolution lookups during tests.
func testID(prefix, suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte(prefix + ":1:" + suffix))
}

func promptID(suffix string) string { return testID("Prompt", suffix) }
func spaceID(suffix string) string  { return testID("Space", suffix) }

func newTestServer(t *testing.T, handler http.HandlerFunc) *arize.Client {
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
	return client
}

// wirePromptCreate mirrors the JSON shape of the Create request body so tests
// can assert on the wire payload without importing internal/generated.
type wirePromptCreate struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	SpaceId     string          `json:"space_id"`
	Version     json.RawMessage `json:"version"`
}

// wirePromptVersionCreate mirrors the JSON shape of the CreateVersion body.
type wirePromptVersionCreate struct {
	CommitMessage       string  `json:"commit_message"`
	Provider            string  `json:"provider"`
	InputVariableFormat *string `json:"input_variable_format,omitempty"`
	Model               *string `json:"model,omitempty"`
}

// wireLabels mirrors the JSON shape of the SetVersionLabels body.
type wireLabels struct {
	Labels []string `json:"labels"`
}

func TestPrompts(t *testing.T) {
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
				json.NewEncoder(w).Encode(prompts.PromptList{
					Prompts:    []prompts.Prompt{{Id: "p-1", Name: "my-prompt"}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.List(ctx, prompts.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*prompts.PromptList)
				if len(resp.Prompts) != 1 {
					t.Errorf("expected 1 prompt, got %d", len(resp.Prompts))
				}
			},
		},
		{
			name: "List_Filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got, want := q.Get("space_id"), spaceID("sp-1"); got != want {
					t.Errorf("space_id query: want %q, got %q", want, got)
				}
				if got := q.Get("name"); got != "rag" {
					t.Errorf("name query: want rag, got %q", got)
				}
				if got := q.Get("limit"); got != "25" {
					t.Errorf("limit query: want 25, got %q", got)
				}
				if got := q.Get("cursor"); got != "cursor-abc" {
					t.Errorf("cursor query: want cursor-abc, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.PromptList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.List(ctx, prompts.ListRequest{
					Space:  spaceID("sp-1"),
					Name:   "rag",
					Limit:  25,
					Cursor: "cursor-abc",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// A plain Space name (not a resource ID) is sent as the space_name
			// substring filter — no GET /v2/spaces resolution round-trip.
			name: "List_Filters_SpaceName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/prompts" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				q := r.URL.Query()
				if got := q.Get("space_name"); got != "demo" {
					t.Errorf("space_name query: want demo, got %q", got)
				}
				if got := q.Get("space_id"); got != "" {
					t.Errorf("space_id query: want empty, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.PromptList{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.List(ctx, prompts.ListRequest{Space: "demo"})
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
				json.NewEncoder(w).Encode(prompts.PromptWithVersion{Id: "p-1", Name: "my-prompt"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.Get(ctx, prompts.GetRequest{Prompt: promptID("p-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*prompts.PromptWithVersion).Name != "my-prompt" {
					t.Errorf("unexpected name: %s", got.(*prompts.PromptWithVersion).Name)
				}
			},
		},
		{
			name: "Get_WithVersionID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("version_id"); got != "v-9" {
					t.Errorf("version_id query: want v-9, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.PromptWithVersion{Id: "p-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.Get(ctx, prompts.GetRequest{Prompt: promptID("p-1"), VersionID: "v-9"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Get_WithLabel",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("label"); got != "production" {
					t.Errorf("label query: want production, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.PromptWithVersion{Id: "p-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.Get(ctx, prompts.GetRequest{Prompt: promptID("p-1"), Label: "production"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			// Exercises resolve.FindPromptID: a plain prompt name (not a base64
			// resource ID) plus a space ID triggers a prompts list filtered by
			// name to discover the ID, then issues the real GET.
			name: "Get_ResolvesByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v2/prompts":
					if got := r.URL.Query().Get("name"); got != "my-prompt" {
						t.Errorf("resolver list name query: want my-prompt, got %q", got)
					}
					json.NewEncoder(w).Encode(prompts.PromptList{
						Prompts:    []prompts.Prompt{{Id: promptID("p-1"), Name: "my-prompt"}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/"+promptID("p-1")):
					json.NewEncoder(w).Encode(prompts.PromptWithVersion{Id: promptID("p-1"), Name: "my-prompt"})
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.Get(ctx, prompts.GetRequest{Prompt: "my-prompt", Space: spaceID("sp-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Create",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wirePromptCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.Name != "my-prompt" {
					t.Errorf("body name: want my-prompt, got %q", body.Name)
				}
				if body.SpaceId != spaceID("sp-1") {
					t.Errorf("body space_id: want %q, got %q", spaceID("sp-1"), body.SpaceId)
				}
				if body.Description == nil || *body.Description != "desc" {
					t.Errorf("body description: want pointer to %q, got %v", "desc", body.Description)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(prompts.PromptWithVersion{Id: "p-new"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.Create(ctx, prompts.CreateRequest{
					Space:       spaceID("sp-1"),
					Name:        "my-prompt",
					Description: "desc",
					Version: prompts.PromptVersionCreate{
						CommitMessage: "init",
						Provider:      prompts.LlmProviderOpenAi,
						Messages:      []prompts.LLMMessage{},
					},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*prompts.PromptWithVersion).Id != "p-new" {
					t.Errorf("unexpected id: %s", got.(*prompts.PromptWithVersion).Id)
				}
			},
		},
		{
			name: "Update",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				var raw map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if _, ok := raw["description"]; !ok {
					t.Errorf("description should be present, got keys %v", raw)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.Prompt{Id: "p-1", Description: ptr("new desc")})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.Update(ctx, prompts.UpdateRequest{
					Prompt:      promptID("p-1"),
					Description: ptr("new desc"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				p := got.(*prompts.Prompt)
				if p.Description == nil || *p.Description != "new desc" {
					t.Errorf("unexpected description: %v", p.Description)
				}
			},
		},
		{
			// Verifies the *string + omitempty preserve contract: a nil
			// Description must be omitted from the PATCH body entirely.
			name: "Update_PreserveDescription",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var raw map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if _, ok := raw["description"]; ok {
					t.Errorf("description should be omitted when nil, got keys %v", raw)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.Prompt{Id: "p-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.Update(ctx, prompts.UpdateRequest{Prompt: promptID("p-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Delete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Prompts.Delete(ctx, prompts.DeleteRequest{Prompt: promptID("p-1")})
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
				if got := r.URL.Query().Get("limit"); got != "10" {
					t.Errorf("limit query: want 10, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.PromptVersionList{
					PromptVersions: []prompts.PromptVersion{{Id: "v-1"}},
					Pagination:     arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.ListVersions(ctx, prompts.ListVersionsRequest{
					Prompt: promptID("p-1"),
					Limit:  10,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*prompts.PromptVersionList)
				if len(resp.PromptVersions) != 1 {
					t.Errorf("expected 1 version, got %d", len(resp.PromptVersions))
				}
			},
		},
		{
			name: "CreateVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body wirePromptVersionCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.CommitMessage != "v2" {
					t.Errorf("body commit_message: want v2, got %q", body.CommitMessage)
				}
				if body.Provider != "open_ai" {
					t.Errorf("body provider: want open_ai, got %q", body.Provider)
				}
				if body.Model == nil || *body.Model != "gpt-4" {
					t.Errorf("body model: want pointer to gpt-4, got %v", body.Model)
				}
				if body.InputVariableFormat == nil || *body.InputVariableFormat != "f_string" {
					t.Errorf("body input_variable_format: want pointer to f_string, got %v", body.InputVariableFormat)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(prompts.PromptVersion{Id: "v-new"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.CreateVersion(ctx, prompts.CreateVersionRequest{
					Prompt:              promptID("p-1"),
					CommitMessage:       "v2",
					Provider:            prompts.LlmProviderOpenAi,
					Messages:            []prompts.LLMMessage{},
					InputVariableFormat: prompts.InputVariableFormatFString,
					Model:               "gpt-4",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*prompts.PromptVersion).Id != "v-new" {
					t.Errorf("unexpected id: %s", got.(*prompts.PromptVersion).Id)
				}
			},
		},
		{
			name: "GetVersion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "/v-1") {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.PromptVersion{Id: "v-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.GetVersion(ctx, prompts.GetVersionRequest{VersionID: "v-1"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*prompts.PromptVersion).Id != "v-1" {
					t.Errorf("unexpected id: %s", got.(*prompts.PromptVersion).Id)
				}
			},
		},
		{
			name: "GetVersionByLabel",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "/labels/production") {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.PromptVersion{Id: "v-prod"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.GetVersionByLabel(ctx, prompts.GetVersionByLabelRequest{
					Prompt:    promptID("p-1"),
					LabelName: "production",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.(*prompts.PromptVersion).Id != "v-prod" {
					t.Errorf("unexpected id: %s", got.(*prompts.PromptVersion).Id)
				}
			},
		},
		{
			name: "SetVersionLabels",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				var body wireLabels
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if len(body.Labels) != 2 || body.Labels[0] != "production" || body.Labels[1] != "stable" {
					t.Errorf("body labels: want [production stable], got %v", body.Labels)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(prompts.PromptVersionLabels{Labels: []string{"production", "stable"}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Prompts.SetVersionLabels(ctx, prompts.SetVersionLabelsRequest{
					VersionID: "v-1",
					Labels:    []string{"production", "stable"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*prompts.PromptVersionLabels)
				if len(resp.Labels) != 2 || resp.Labels[0] != "production" || resp.Labels[1] != "stable" {
					t.Errorf("returned labels: want [production stable], got %v", resp.Labels)
				}
			},
		},
		{
			name: "DeleteVersionLabel",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/labels/production") {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Prompts.DeleteVersionLabel(ctx, prompts.DeleteVersionLabelRequest{
					VersionID: "v-1",
					LabelName: "production",
				})
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
			client := newTestServer(t, tt.handler)
			got, err := tt.invoke(context.Background(), client)
			tt.check(t, got, err)
		})
	}
}

func ptr(s string) *string { return &s }
