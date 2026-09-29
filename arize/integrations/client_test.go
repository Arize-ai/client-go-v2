package integrations_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/integrations"
)

// helperID encodes a fake base64 resource ID (e.g. "Integration:1:foo") so
// that resolve.IsResourceID returns true and the SDK short-circuits any
// name-to-ID lookup.
func helperID(kind, suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte(kind + ":1:" + suffix))
}

func ptr[T any](v T) *T { return &v }

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

// agentUnion builds an Integration union carrying an AGENT variant, so
// handlers can encode a realistic polymorphic response without importing
// internal/generated.
func agentUnion(t *testing.T, id, name string) integrations.Integration {
	t.Helper()
	var it integrations.Integration
	if err := it.FromAgentIntegration(integrations.AgentIntegration{
		Id:        id,
		Name:      name,
		Type:      "AGENT",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Config: integrations.AgentConfig{
			Endpoint:       "https://agent.example.com/replay",
			InputSchema:    map[string]any{"type": "object"},
			RequestPresets: []integrations.AgentRequestPreset{},
		},
		Scopings: []integrations.IntegrationScoping{},
	}); err != nil {
		t.Fatalf("build agent union: %v", err)
	}
	return it
}

// llmUnion builds an Integration union carrying an LLM variant.
func llmUnion(t *testing.T, id, name string) integrations.Integration {
	t.Helper()
	var cfg integrations.LLMConfig
	if err := cfg.FromOpenAiConfig(integrations.OpenAIConfig{
		Provider:                 "OPEN_AI",
		HasApiKey:                true,
		IsFunctionCallingEnabled: true,
	}); err != nil {
		t.Fatalf("build llm config: %v", err)
	}
	var it integrations.Integration
	if err := it.FromLlmIntegration(integrations.LLMIntegration{
		Id:        id,
		Name:      name,
		Type:      "LLM",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Config:    cfg,
		Scopings:  []integrations.IntegrationScoping{},
	}); err != nil {
		t.Fatalf("build llm union: %v", err)
	}
	return it
}

// wireCreateAgent mirrors the JSON the server receives from POST
// /v2/integrations for an agent create.
type wireCreateAgent struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Config      map[string]any `json:"config"`
	Scopings    *[]any         `json:"scopings,omitempty"`
}

// wireCreateLlm mirrors the JSON the server receives from POST
// /v2/integrations for an LLM create.
type wireCreateLlm struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Config map[string]any `json:"config"`
}

// wireUpdate mirrors the JSON the server receives from PATCH
// /v2/integrations/{id}.
type wireUpdate struct {
	Type     string         `json:"type"`
	Name     *string        `json:"name,omitempty"`
	Config   map[string]any `json:"config,omitempty"`
	Scopings *[]any         `json:"scopings,omitempty"`
}

func TestIntegrations(t *testing.T) {
	// Handler constructors decode the request body, run the caller-supplied
	// assert closure inside the handler (where the request is in scope, per
	// the AGENTS.md rule: never assert via shared captured state), and respond
	// with a canned integration. The `check` closure validates only the
	// returned value and error.
	agentCreateHandler := func(id, name string, assert func(body wireCreateAgent)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var body wireCreateAgent
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
			assert(body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(agentUnion(t, id, name))
		}
	}
	llmCreateHandler := func(id, name string, assert func(body wireCreateLlm)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var body wireCreateLlm
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
			assert(body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(llmUnion(t, id, name))
		}
	}
	agentUpdateHandler := func(id, name string, assert func(body wireUpdate, raw map[string]any)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var body wireUpdate
			rawBytes, _ := io.ReadAll(r.Body)
			raw := map[string]any{}
			_ = json.Unmarshal(rawBytes, &body)
			_ = json.Unmarshal(rawBytes, &raw)
			assert(body, raw)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(agentUnion(t, id, name))
		}
	}
	llmUpdateHandler := func(id, name string, assert func(body wireUpdate, raw map[string]any)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var body wireUpdate
			rawBytes, _ := io.ReadAll(r.Body)
			raw := map[string]any{}
			_ = json.Unmarshal(rawBytes, &body)
			_ = json.Unmarshal(rawBytes, &raw)
			assert(body, raw)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(llmUnion(t, id, name))
		}
	}

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List_LLM_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("type"); got != "LLM" {
					t.Errorf("type query: got %q, want %q", got, "LLM")
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(integrations.ListIntegrations{
					Integrations: []integrations.Integration{llmUnion(t, "int-1", "my-llm")},
					Pagination:   arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.List(ctx, integrations.ListRequest{Type: integrations.IntegrationTypeLLM, Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*integrations.ListIntegrations)
				if len(resp.Integrations) != 1 {
					t.Fatalf("expected 1 integration, got %d", len(resp.Integrations))
				}
				llm, err := resp.Integrations[0].AsLlmIntegration()
				if err != nil {
					t.Fatalf("unwrap llm: %v", err)
				}
				if llm.Name != "my-llm" {
					t.Errorf("name: got %q, want %q", llm.Name, "my-llm")
				}
			},
		},
		{
			name: "List_Agent_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("type"); got != "AGENT" {
					t.Errorf("type query: got %q, want %q", got, "AGENT")
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(integrations.ListIntegrations{
					Integrations: []integrations.Integration{agentUnion(t, "int-a", "my-agent")},
					Pagination:   arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.List(ctx, integrations.ListRequest{Type: integrations.IntegrationTypeAgent})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*integrations.ListIntegrations)
				agent, err := resp.Integrations[0].AsAgentIntegration()
				if err != nil {
					t.Fatalf("unwrap agent: %v", err)
				}
				if agent.Config.Endpoint == "" {
					t.Error("expected agent config endpoint to be populated")
				}
			},
		},
		{
			name: "List_Untyped_Merged",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if _, ok := r.URL.Query()["type"]; ok {
					t.Errorf("type query param should be omitted, got %q", r.URL.Query().Get("type"))
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(integrations.ListIntegrations{
					Integrations: []integrations.Integration{
						llmUnion(t, "int-1", "my-llm"),
						agentUnion(t, "int-a", "my-agent"),
					},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.List(ctx, integrations.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*integrations.ListIntegrations)
				if len(resp.Integrations) != 2 {
					t.Fatalf("expected 2 integrations, got %d", len(resp.Integrations))
				}
				var types []string
				for _, it := range resp.Integrations {
					d, err := it.Discriminator()
					if err != nil {
						t.Fatalf("discriminator: %v", err)
					}
					types = append(types, d)
				}
				if types[0] != "LLM" || types[1] != "AGENT" {
					t.Errorf("discriminators: got %v, want [LLM AGENT]", types)
				}
			},
		},
		{
			name: "List_QueryParams",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got := q.Get("name"); got != "openai" {
					t.Errorf("name: got %q, want %q", got, "openai")
				}
				if got := q.Get("space_id"); got != helperID("Space", "demo") {
					t.Errorf("space_id: got %q", got)
				}
				if got := q.Get("limit"); got != "25" {
					t.Errorf("limit: got %q, want %q", got, "25")
				}
				if got := q.Get("cursor"); got != "next-page-token" {
					t.Errorf("cursor: got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(integrations.ListIntegrations{
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.List(ctx, integrations.ListRequest{
					Type:   integrations.IntegrationTypeLLM,
					Name:   "openai",
					Space:  helperID("Space", "demo"),
					Limit:  25,
					Cursor: "next-page-token",
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "List_DefaultLimit_SpaceNameSubstring",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if got := q.Get("limit"); got != "50" {
					t.Errorf("default limit: got %q, want %q", got, "50")
				}
				if got := q.Get("space_name"); got != "prod" {
					t.Errorf("space_name: got %q, want %q", got, "prod")
				}
				if got := q.Get("space_id"); got != "" {
					t.Errorf("space_id should be empty when filtering by name, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(integrations.ListIntegrations{
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.List(ctx, integrations.ListRequest{
					Type:  integrations.IntegrationTypeAgent,
					Space: "prod", // bare name → space_name substring filter
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Get_Agent_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(agentUnion(t, "int-a", "my-agent"))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.Get(ctx, integrations.GetRequest{
					Integration: helperID("Integration", "int-a"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*integrations.Integration)
				agent, err := resp.AsAgentIntegration()
				if err != nil {
					t.Fatalf("unwrap agent: %v", err)
				}
				if agent.Name != "my-agent" {
					t.Errorf("name: got %q, want %q", agent.Name, "my-agent")
				}
			},
		},
		{
			name: "Get_Llm_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(llmUnion(t, "int-1", "my-llm"))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.Get(ctx, integrations.GetRequest{
					Integration: helperID("Integration", "int-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*integrations.Integration)
				llm, err := resp.AsLlmIntegration()
				if err != nil {
					t.Fatalf("unwrap llm: %v", err)
				}
				oa, err := llm.Config.AsOpenAiConfig()
				if err != nil {
					t.Fatalf("unwrap openai config: %v", err)
				}
				if oa.Provider != "OPEN_AI" {
					t.Errorf("provider: got %q, want %q", oa.Provider, "OPEN_AI")
				}
			},
		},
		{
			name: "Get_ByName_ResolvesViaList",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case strings.HasSuffix(r.URL.Path, "/v2/integrations"):
					q := r.URL.Query()
					if q.Get("name") != "my-agent" {
						t.Errorf("list query name: got %q, want %q", q.Get("name"), "my-agent")
					}
					if q.Get("type") != "AGENT" {
						t.Errorf("list query type: got %q, want %q", q.Get("type"), "AGENT")
					}
					_ = json.NewEncoder(w).Encode(integrations.ListIntegrations{
						Integrations: []integrations.Integration{agentUnion(t, "int-resolved", "my-agent")},
						Pagination:   arize.PaginationMetadata{HasMore: false},
					})
				case strings.Contains(r.URL.Path, "/v2/integrations/"):
					if !strings.HasSuffix(r.URL.Path, "/int-resolved") {
						t.Errorf("get path: got %q, want suffix /int-resolved", r.URL.Path)
					}
					_ = json.NewEncoder(w).Encode(agentUnion(t, "int-resolved", "my-agent"))
				default:
					t.Errorf("unexpected request path: %q", r.URL.Path)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.Get(ctx, integrations.GetRequest{
					Integration: "my-agent", // bare name → triggers list-and-match
					Type:        integrations.IntegrationTypeAgent,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*integrations.Integration)
				agent, err := resp.AsAgentIntegration()
				if err != nil {
					t.Fatalf("unwrap agent: %v", err)
				}
				if agent.Id != "int-resolved" {
					t.Errorf("id: got %q, want %q", agent.Id, "int-resolved")
				}
			},
		},
		{
			name: "Get_ByName_MissingType",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called when type is missing for name resolution")
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.Get(ctx, integrations.GetRequest{Integration: "my-agent"})
			},
			check: func(t *testing.T, _ any, err error) {
				if err == nil {
					t.Fatal("expected error resolving a name without a type, got nil")
				}
			},
		},
		{
			name: "Get_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.Get(ctx, integrations.GetRequest{
					Integration: helperID("Integration", "missing"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "CreateAgent_Success",
			handler: agentCreateHandler("int-new", "new-agent", func(body wireCreateAgent) {
				if body.Type != "AGENT" {
					t.Errorf("type: got %q, want %q", body.Type, "AGENT")
				}
				if body.Name != "new-agent" {
					t.Errorf("name: got %q, want %q", body.Name, "new-agent")
				}
				if body.Description == nil || *body.Description != "my agent" {
					t.Errorf("description: got %v, want %q", body.Description, "my agent")
				}
				if body.Config["endpoint"] != "https://agent.example.com/replay" {
					t.Errorf("config.endpoint: got %v", body.Config["endpoint"])
				}
				if _, ok := body.Config["input_schema"]; !ok {
					t.Error("config.input_schema missing from body")
				}
				// Unset optional collections omitted.
				if _, ok := body.Config["headers"]; ok {
					t.Error("config.headers: expected omitted")
				}
				if body.Scopings != nil {
					t.Errorf("scopings: expected omitted, got %v", body.Scopings)
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateAgent(ctx, integrations.CreateAgentRequest{
					Name:        "new-agent",
					Endpoint:    "https://agent.example.com/replay",
					InputSchema: map[string]any{"type": "object"},
					Description: "my agent",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*integrations.Integration)
				if _, err := resp.AsAgentIntegration(); err != nil {
					t.Fatalf("unwrap agent: %v", err)
				}
			},
		},
		{
			name: "CreateAgent_WithPresetsAndHeaders",
			handler: agentCreateHandler("int-p", "agent-presets", func(body wireCreateAgent) {
				hdrs, ok := body.Config["headers"].(map[string]any)
				if !ok || hdrs["X-Token"] != "abc" {
					t.Errorf("config.headers: got %v", body.Config["headers"])
				}
				presets, ok := body.Config["request_presets"].([]any)
				if !ok || len(presets) != 1 {
					t.Errorf("config.request_presets: got %v", body.Config["request_presets"])
					return
				}
				p0 := presets[0].(map[string]any)
				if p0["name"] != "default" {
					t.Errorf("preset name: got %v", p0["name"])
				}
				if p0["description"] != "the default preset" {
					t.Errorf("preset description: got %v", p0["description"])
				}
				if body.Scopings == nil || len(*body.Scopings) != 1 {
					t.Errorf("scopings: got %v", body.Scopings)
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateAgent(ctx, integrations.CreateAgentRequest{
					Name:        "agent-presets",
					Endpoint:    "https://agent.example.com/replay",
					InputSchema: map[string]any{"type": "object"},
					Headers:     map[string]string{"X-Token": "abc"},
					RequestPresets: []integrations.AgentRequestPresetInput{{
						Name:        "default",
						Config:      map[string]any{"foo": "bar"},
						Description: "the default preset",
					}},
					Scopings: []integrations.IntegrationScoping{{OrganizationId: ptr("org-1")}},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "CreateLlm_OpenAI_Success",
			handler: llmCreateHandler("int-oa", "openai", func(body wireCreateLlm) {
				if body.Type != "LLM" {
					t.Errorf("type: got %q, want %q", body.Type, "LLM")
				}
				if body.Config["provider"] != "OPEN_AI" {
					t.Errorf("config.provider: got %v, want OPEN_AI", body.Config["provider"])
				}
				if body.Config["api_key"] != "sk-secret" {
					t.Errorf("config.api_key: got %v", body.Config["api_key"])
				}
				// IsFunctionCallingEnabled left nil → omitted.
				if _, ok := body.Config["is_function_calling_enabled"]; ok {
					t.Error("config.is_function_calling_enabled: expected omitted")
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "openai",
					Config: integrations.CreateLLMConfig{
						OpenAI: &integrations.CreateOpenAIConfig{APIKey: "sk-secret"},
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
			name: "CreateLlm_Anthropic_FunctionCallingDisabled",
			handler: llmCreateHandler("int-an", "anthropic", func(body wireCreateLlm) {
				if body.Config["provider"] != "ANTHROPIC" {
					t.Errorf("config.provider: got %v, want ANTHROPIC", body.Config["provider"])
				}
				v, ok := body.Config["is_function_calling_enabled"]
				if !ok {
					t.Error("is_function_calling_enabled missing; expected false")
				} else if v != false {
					t.Errorf("is_function_calling_enabled: got %v, want false", v)
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "anthropic",
					Config: integrations.CreateLLMConfig{
						Anthropic: &integrations.CreateAnthropicConfig{
							APIKey:                 "sk-ant",
							FunctionCallingEnabled: ptr(false),
						},
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
			name: "CreateLlm_Gemini_Success",
			handler: llmCreateHandler("int-gm", "gemini", func(body wireCreateLlm) {
				if body.Config["provider"] != "GEMINI" {
					t.Errorf("config.provider: got %v, want GEMINI", body.Config["provider"])
				}
				if body.Config["api_key"] != "sk-gem" {
					t.Errorf("config.api_key: got %v", body.Config["api_key"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "gemini",
					Config: integrations.CreateLLMConfig{
						Gemini: &integrations.CreateGeminiConfig{APIKey: "sk-gem"},
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
			name: "CreateLlm_Custom_Success",
			handler: llmCreateHandler("int-cu", "custom", func(body wireCreateLlm) {
				if body.Config["provider"] != "CUSTOM" {
					t.Errorf("config.provider: got %v, want CUSTOM", body.Config["provider"])
				}
				if body.Config["base_url"] != "https://llm.example.com/v1" {
					t.Errorf("config.base_url: got %v", body.Config["base_url"])
				}
				if body.Config["api_key"] != "sk-custom" {
					t.Errorf("config.api_key: got %v", body.Config["api_key"])
				}
				hdrs, ok := body.Config["headers"].(map[string]any)
				if !ok || hdrs["X-Org"] != "acme" {
					t.Errorf("config.headers: got %v", body.Config["headers"])
				}
				if body.Config["is_default_models_enabled"] != true {
					t.Errorf("config.is_default_models_enabled: got %v, want true", body.Config["is_default_models_enabled"])
				}
				names, ok := body.Config["model_names"].([]any)
				if !ok || len(names) != 1 || names[0] != "my-model" {
					t.Errorf("config.model_names: got %v", body.Config["model_names"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "custom",
					Config: integrations.CreateLLMConfig{
						Custom: &integrations.CreateCustomConfig{
							BaseURL:                "https://llm.example.com/v1",
							APIKey:                 "sk-custom",
							Headers:                map[string]string{"X-Org": "acme"},
							ModelNames:             []string{"my-model"},
							IsDefaultModelsEnabled: ptr(true),
						},
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
			name: "CreateLlm_VertexAI_Success",
			handler: llmCreateHandler("int-vx", "vertex", func(body wireCreateLlm) {
				if body.Config["provider"] != "VERTEX_AI" {
					t.Errorf("config.provider: got %v, want VERTEX_AI", body.Config["provider"])
				}
				if body.Config["project_id"] != "my-proj" {
					t.Errorf("config.project_id: got %v", body.Config["project_id"])
				}
				if body.Config["location"] != "us-central1" {
					t.Errorf("config.location: got %v", body.Config["location"])
				}
				if body.Config["project_access_label"] != "arize-access" {
					t.Errorf("config.project_access_label: got %v", body.Config["project_access_label"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "vertex",
					Config: integrations.CreateLLMConfig{
						VertexAI: &integrations.CreateVertexAIConfig{
							ProjectID:          "my-proj",
							Location:           "us-central1",
							ProjectAccessLabel: "arize-access",
						},
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
			name: "CreateLlm_NvidiaNIM_Success",
			handler: llmCreateHandler("int-nv", "nvidia", func(body wireCreateLlm) {
				if body.Config["provider"] != "NVIDIA_NIM" {
					t.Errorf("config.provider: got %v, want NVIDIA_NIM", body.Config["provider"])
				}
				if body.Config["base_url"] != "https://nim.example.com/v1" {
					t.Errorf("config.base_url: got %v", body.Config["base_url"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "nvidia",
					Config: integrations.CreateLLMConfig{
						NvidiaNIM: &integrations.CreateNvidiaNIMConfig{
							BaseURL:                "https://nim.example.com/v1",
							IsDefaultModelsEnabled: ptr(true),
						},
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
			name: "CreateLlm_LiteLLM_Success",
			handler: llmCreateHandler("int-ll", "litellm", func(body wireCreateLlm) {
				if body.Config["provider"] != "LITELLM" {
					t.Errorf("config.provider: got %v, want LITELLM", body.Config["provider"])
				}
				if body.Config["base_url"] != "https://litellm.internal:4000" {
					t.Errorf("config.base_url: got %v", body.Config["base_url"])
				}
				if body.Config["api_key"] != "sk-litellm-x" {
					t.Errorf("config.api_key: got %v", body.Config["api_key"])
				}
				if _, present := body.Config["is_default_models_enabled"]; present {
					t.Errorf("config.is_default_models_enabled must not be sent for LITELLM")
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "litellm",
					Config: integrations.CreateLLMConfig{
						LiteLLM: &integrations.CreateLiteLLMConfig{
							BaseURL:    "https://litellm.internal:4000",
							APIKey:     "sk-litellm-x",
							ModelNames: []string{"team-gpt-4o"},
						},
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
			name: "CreateLlm_Fireworks_Success",
			handler: llmCreateHandler("int-fw", "fireworks", func(body wireCreateLlm) {
				if body.Config["provider"] != "FIREWORKS" {
					t.Errorf("config.provider: got %v, want FIREWORKS", body.Config["provider"])
				}
				if body.Config["api_key"] != "fw-key-x" {
					t.Errorf("config.api_key: got %v", body.Config["api_key"])
				}
				if body.Config["is_default_models_enabled"] != true {
					t.Errorf("config.is_default_models_enabled: got %v, want true", body.Config["is_default_models_enabled"])
				}
				if _, present := body.Config["base_url"]; present {
					t.Errorf("config.base_url must not be sent for FIREWORKS")
				}
				if _, present := body.Config["headers"]; present {
					t.Errorf("config.headers must not be sent for FIREWORKS")
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "fireworks",
					Config: integrations.CreateLLMConfig{
						Fireworks: &integrations.CreateFireworksConfig{
							APIKey:                 "fw-key-x",
							ModelNames:             []string{"accounts/fireworks/models/llama-v3p1-8b-instruct"},
							IsDefaultModelsEnabled: ptr(true),
						},
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
			name: "CreateLlm_TogetherAi_Success",
			handler: llmCreateHandler("int-ta", "together-ai", func(body wireCreateLlm) {
				if body.Config["provider"] != "TOGETHER_AI" {
					t.Errorf("config.provider: got %v, want TOGETHER_AI", body.Config["provider"])
				}
				if body.Config["api_key"] != "together-key-x" {
					t.Errorf("config.api_key: got %v", body.Config["api_key"])
				}
				if body.Config["is_default_models_enabled"] != true {
					t.Errorf("config.is_default_models_enabled: got %v, want true", body.Config["is_default_models_enabled"])
				}
				if _, present := body.Config["base_url"]; present {
					t.Errorf("config.base_url must not be sent for TOGETHER_AI")
				}
				if _, present := body.Config["headers"]; present {
					t.Errorf("config.headers must not be sent for TOGETHER_AI")
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "together-ai",
					Config: integrations.CreateLLMConfig{
						TogetherAi: &integrations.CreateTogetherAiConfig{
							APIKey:                 "together-key-x",
							ModelNames:             []string{"meta-llama/Llama-3.3-70B-Instruct-Turbo"},
							IsDefaultModelsEnabled: ptr(true),
						},
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
			name: "CreateLlm_Bedrock_DefaultAuth",
			handler: llmCreateHandler("int-bd", "bedrock-default", func(body wireCreateLlm) {
				if body.Config["provider"] != "AWS_BEDROCK" {
					t.Errorf("config.provider: got %v, want AWS_BEDROCK", body.Config["provider"])
				}
				if body.Config["is_default_models_enabled"] != true {
					t.Errorf("config.is_default_models_enabled: got %v, want true", body.Config["is_default_models_enabled"])
				}
				auth, ok := body.Config["auth"].(map[string]any)
				if !ok {
					t.Errorf("config.auth: got %v", body.Config["auth"])
					return
				}
				if auth["auth_type"] != "DEFAULT" {
					t.Errorf("auth.auth_type: got %v, want DEFAULT", auth["auth_type"])
				}
				if auth["role_arn"] != "arn:aws:iam::123:role/arize" {
					t.Errorf("auth.role_arn: got %v", auth["role_arn"])
				}
				if auth["external_id"] != "ext-1" {
					t.Errorf("auth.external_id: got %v", auth["external_id"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "bedrock-default",
					Config: integrations.CreateLLMConfig{
						AWSBedrock: &integrations.CreateAWSBedrockConfig{
							Auth: integrations.CreateAWSBedrockAuth{
								Default: &integrations.CreateAWSBedrockDefaultAuth{
									RoleARN:    "arn:aws:iam::123:role/arize",
									ExternalID: "ext-1",
								},
							},
							IsDefaultModelsEnabled: ptr(true),
						},
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
			name: "CreateLlm_Bedrock_BearerTokenAuth",
			handler: llmCreateHandler("int-bb", "bedrock-bearer", func(body wireCreateLlm) {
				auth, ok := body.Config["auth"].(map[string]any)
				if !ok {
					t.Errorf("config.auth: got %v", body.Config["auth"])
					return
				}
				if auth["auth_type"] != "BEARER_TOKEN" {
					t.Errorf("auth.auth_type: got %v, want BEARER_TOKEN", auth["auth_type"])
				}
				if auth["api_key"] != "bedrock-token" {
					t.Errorf("auth.api_key: got %v", auth["api_key"])
				}
				names, ok := body.Config["model_names"].([]any)
				if !ok || len(names) != 1 || names[0] != "anthropic.claude-v2" {
					t.Errorf("config.model_names: got %v", body.Config["model_names"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "bedrock-bearer",
					Config: integrations.CreateLLMConfig{
						AWSBedrock: &integrations.CreateAWSBedrockConfig{
							Auth: integrations.CreateAWSBedrockAuth{
								BearerToken: &integrations.CreateAWSBedrockBearerTokenAuth{
									APIKey: "bedrock-token",
								},
							},
							ModelNames: []string{"anthropic.claude-v2"},
						},
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
			name: "CreateLlm_Bedrock_ProxyWithHeadersAuth",
			handler: llmCreateHandler("int-bp", "bedrock-proxy", func(body wireCreateLlm) {
				auth, ok := body.Config["auth"].(map[string]any)
				if !ok {
					t.Errorf("config.auth: got %v", body.Config["auth"])
					return
				}
				if auth["auth_type"] != "PROXY_WITH_HEADERS" {
					t.Errorf("auth.auth_type: got %v, want PROXY_WITH_HEADERS", auth["auth_type"])
				}
				if auth["base_url"] != "https://proxy.example.com" {
					t.Errorf("auth.base_url: got %v", auth["base_url"])
				}
				hdrs, ok := auth["headers"].(map[string]any)
				if !ok || hdrs["X-Proxy"] != "yes" {
					t.Errorf("auth.headers: got %v", auth["headers"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "bedrock-proxy",
					Config: integrations.CreateLLMConfig{
						AWSBedrock: &integrations.CreateAWSBedrockConfig{
							Auth: integrations.CreateAWSBedrockAuth{
								ProxyWithHeaders: &integrations.CreateAWSBedrockProxyWithHeadersAuth{
									BaseURL: "https://proxy.example.com",
									Headers: map[string]string{"X-Proxy": "yes"},
								},
							},
							IsDefaultModelsEnabled: ptr(true),
						},
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
			name:    "CreateLlm_NoProviderConfig",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{Name: "empty"})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, integrations.ErrInvalidProviderConfig) {
					t.Errorf("expected ErrInvalidProviderConfig, got %v", err)
				}
			},
		},
		{
			name:    "CreateLlm_MultipleProviderConfigs",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "two",
					Config: integrations.CreateLLMConfig{
						OpenAI:    &integrations.CreateOpenAIConfig{APIKey: "a"},
						Anthropic: &integrations.CreateAnthropicConfig{APIKey: "b"},
					},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, integrations.ErrInvalidProviderConfig) {
					t.Errorf("expected ErrInvalidProviderConfig, got %v", err)
				}
			},
		},
		{
			name:    "CreateLlm_Bedrock_InvalidAuth",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "bedrock-noauth",
					Config: integrations.CreateLLMConfig{
						AWSBedrock: &integrations.CreateAWSBedrockConfig{IsDefaultModelsEnabled: ptr(true)},
					},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, integrations.ErrInvalidBedrockAuth) {
					t.Errorf("expected ErrInvalidBedrockAuth, got %v", err)
				}
			},
		},
		{
			name:    "CreateLlm_Bedrock_MultipleAuthModes",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
					Name: "bedrock-multiauth",
					Config: integrations.CreateLLMConfig{
						AWSBedrock: &integrations.CreateAWSBedrockConfig{
							Auth: integrations.CreateAWSBedrockAuth{
								Default: &integrations.CreateAWSBedrockDefaultAuth{
									RoleARN: "arn:aws:iam::123:role/arize",
								},
								BearerToken: &integrations.CreateAWSBedrockBearerTokenAuth{
									APIKey: "bedrock-token",
								},
							},
							IsDefaultModelsEnabled: ptr(true),
						},
					},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, integrations.ErrInvalidBedrockAuth) {
					t.Errorf("expected ErrInvalidBedrockAuth, got %v", err)
				}
			},
		},
		{
			name: "CreateAgent_Conflict",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "conflict", "status": 409})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.CreateAgent(ctx, integrations.CreateAgentRequest{
					Name:        "dup",
					Endpoint:    "https://x.example.com",
					InputSchema: map[string]any{},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				var ce *arize.ConflictError
				if !errors.As(err, &ce) {
					t.Errorf("expected *ConflictError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "UpdateAgent_NameOnly",
			handler: agentUpdateHandler("int-a", "renamed", func(body wireUpdate, raw map[string]any) {
				if body.Type != "AGENT" {
					t.Errorf("type: got %q, want %q", body.Type, "AGENT")
				}
				if body.Name == nil || *body.Name != "renamed" {
					t.Errorf("name: got %v, want %q", body.Name, "renamed")
				}
				// No config fields set → config omitted entirely.
				if _, present := raw["config"]; present {
					t.Error("config: expected omitted when no config fields patched")
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateAgent(ctx, integrations.UpdateAgentRequest{
					Integration: helperID("Integration", "int-a"),
					Name:        ptr("renamed"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "UpdateAgent_ConfigFields",
			handler: agentUpdateHandler("int-a", "my-agent", func(body wireUpdate, _ map[string]any) {
				if body.Config == nil {
					t.Error("config: expected present when config fields patched")
					return
				}
				if body.Config["endpoint"] != "https://new.example.com/replay" {
					t.Errorf("config.endpoint: got %v", body.Config["endpoint"])
				}
				presets, ok := body.Config["request_presets"].([]any)
				if !ok || len(presets) != 1 {
					t.Errorf("config.request_presets: got %v", body.Config["request_presets"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateAgent(ctx, integrations.UpdateAgentRequest{
					Integration: helperID("Integration", "int-a"),
					Endpoint:    ptr("https://new.example.com/replay"),
					RequestPresets: &[]integrations.AgentRequestPresetInput{{
						Name:   "p1",
						Config: map[string]any{"a": 1},
					}},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "UpdateAgent_ClearDescription",
			handler: agentUpdateHandler("int-a", "my-agent", func(_ wireUpdate, raw map[string]any) {
				v, present := raw["description"]
				if !present {
					t.Error("description: expected present (explicit null), got omitted")
				}
				if v != nil {
					t.Errorf("description: got %v, want null", v)
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateAgent(ctx, integrations.UpdateAgentRequest{
					Integration: helperID("Integration", "int-a"),
					Description: ptr(""), // &"" → clear
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name:    "UpdateAgent_NoFields",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateAgent(ctx, integrations.UpdateAgentRequest{
					Integration: helperID("Integration", "int-a"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, integrations.ErrNoUpdateFields) {
					t.Errorf("expected ErrNoUpdateFields, got %v", err)
				}
			},
		},
		{
			name: "UpdateLlm_RotateKey",
			handler: llmUpdateHandler("int-1", "my-llm", func(body wireUpdate, _ map[string]any) {
				if body.Type != "LLM" {
					t.Errorf("type: got %q, want %q", body.Type, "LLM")
				}
				if body.Config == nil {
					t.Error("config: expected present")
					return
				}
				if body.Config["api_key"] != "sk-rotated" {
					t.Errorf("config.api_key: got %v", body.Config["api_key"])
				}
				if body.Config["is_function_calling_enabled"] != false {
					t.Errorf("config.is_function_calling_enabled: got %v, want false", body.Config["is_function_calling_enabled"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration:            helperID("Integration", "int-1"),
					APIKey:                 ptr("sk-rotated"),
					FunctionCallingEnabled: ptr(false),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "UpdateLlm_BedrockAuth",
			handler: llmUpdateHandler("int-bd", "bedrock", func(body wireUpdate, _ map[string]any) {
				if body.Config == nil {
					t.Error("config: expected present")
					return
				}
				auth, ok := body.Config["auth"].(map[string]any)
				if !ok {
					t.Errorf("config.auth: got %v", body.Config["auth"])
					return
				}
				if auth["auth_type"] != "DEFAULT" {
					t.Errorf("auth.auth_type: got %v, want DEFAULT", auth["auth_type"])
				}
				if auth["role_arn"] != "arn:aws:iam::999:role/new" {
					t.Errorf("auth.role_arn: got %v", auth["role_arn"])
				}
				names, ok := body.Config["model_names"].([]any)
				if !ok || len(names) != 1 || names[0] != "anthropic.claude-3" {
					t.Errorf("config.model_names: got %v", body.Config["model_names"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration: helperID("Integration", "int-bd"),
					Auth: &integrations.CreateAWSBedrockAuth{
						Default: &integrations.CreateAWSBedrockDefaultAuth{
							RoleARN: "arn:aws:iam::999:role/new",
						},
					},
					ModelNames: &[]string{"anthropic.claude-3"},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name:    "UpdateLlm_BedrockAuth_Invalid",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration: helperID("Integration", "int-bd"),
					Auth:        &integrations.CreateAWSBedrockAuth{},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, integrations.ErrInvalidBedrockAuth) {
					t.Errorf("expected ErrInvalidBedrockAuth, got %v", err)
				}
			},
		},
		{
			name: "UpdateLlm_CustomFields",
			handler: llmUpdateHandler("int-cu", "custom", func(body wireUpdate, _ map[string]any) {
				if body.Config == nil {
					t.Error("config: expected present")
					return
				}
				if body.Config["base_url"] != "https://new.example.com/v1" {
					t.Errorf("config.base_url: got %v", body.Config["base_url"])
				}
				hdrs, ok := body.Config["headers"].(map[string]any)
				if !ok || hdrs["X-New"] != "1" {
					t.Errorf("config.headers: got %v", body.Config["headers"])
				}
				if body.Config["is_default_models_enabled"] != false {
					t.Errorf("config.is_default_models_enabled: got %v, want false", body.Config["is_default_models_enabled"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration:            helperID("Integration", "int-cu"),
					BaseURL:                ptr("https://new.example.com/v1"),
					Headers:                &map[string]string{"X-New": "1"},
					IsDefaultModelsEnabled: ptr(false),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "UpdateLlm_HeadersClear",
			handler: llmUpdateHandler("int-cu", "custom", func(body wireUpdate, _ map[string]any) {
				if body.Config == nil {
					t.Error("config: expected present")
					return
				}
				v, present := body.Config["headers"]
				if !present {
					t.Error("config.headers: expected present (explicit null), got omitted")
				}
				if v != nil {
					t.Errorf("config.headers: got %v, want null", v)
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				var cleared map[string]string
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration: helperID("Integration", "int-cu"),
					Headers:     &cleared, // pointer to nil map → JSON null
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "UpdateLlm_ClearApiKey",
			handler: llmUpdateHandler("int-1", "my-llm", func(body wireUpdate, _ map[string]any) {
				if body.Config == nil {
					t.Error("config: expected present")
					return
				}
				v, present := body.Config["api_key"]
				if !present {
					t.Error("config.api_key: expected present (explicit null), got omitted")
				}
				if v != nil {
					t.Errorf("config.api_key: got %v, want null", v)
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration: helperID("Integration", "int-1"),
					APIKey:      ptr(""), // &"" → clear
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "UpdateLlm_ClearBaseURL",
			handler: llmUpdateHandler("int-cu", "custom", func(body wireUpdate, _ map[string]any) {
				if body.Config == nil {
					t.Error("config: expected present")
					return
				}
				v, present := body.Config["base_url"]
				if !present {
					t.Error("config.base_url: expected present (explicit null), got omitted")
				}
				if v != nil {
					t.Errorf("config.base_url: got %v, want null", v)
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration: helperID("Integration", "int-cu"),
					BaseURL:     ptr(""), // &"" → clear
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "UpdateLlm_VertexAIFields",
			handler: llmUpdateHandler("int-vx", "vertex", func(body wireUpdate, _ map[string]any) {
				if body.Config == nil {
					t.Error("config: expected present")
					return
				}
				if body.Config["project_id"] != "new-proj" {
					t.Errorf("config.project_id: got %v", body.Config["project_id"])
				}
				if body.Config["location"] != "us-east1" {
					t.Errorf("config.location: got %v", body.Config["location"])
				}
				if body.Config["project_access_label"] != "new-label" {
					t.Errorf("config.project_access_label: got %v", body.Config["project_access_label"])
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration:        helperID("Integration", "int-vx"),
					ProjectID:          ptr("new-proj"),
					Location:           ptr("us-east1"),
					ProjectAccessLabel: ptr("new-label"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "UpdateLlm_NameOnly",
			handler: llmUpdateHandler("int-1", "renamed-llm", func(body wireUpdate, raw map[string]any) {
				if body.Name == nil || *body.Name != "renamed-llm" {
					t.Errorf("name: got %v, want %q", body.Name, "renamed-llm")
				}
				if _, present := raw["config"]; present {
					t.Error("config: expected omitted when no config fields patched")
				}
			}),
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration: helperID("Integration", "int-1"),
					Name:        ptr("renamed-llm"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name:    "UpdateLlm_NoFields",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
					Integration: helperID("Integration", "int-1"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, integrations.ErrNoUpdateFields) {
					t.Errorf("expected ErrNoUpdateFields, got %v", err)
				}
			},
		},
		{
			name: "Delete_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				wantSuffix := "/" + helperID("Integration", "int-1")
				if !strings.HasSuffix(r.URL.Path, wantSuffix) {
					t.Errorf("delete path: got %q, want suffix %q", r.URL.Path, wantSuffix)
				}
				w.WriteHeader(http.StatusNoContent)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Integrations.Delete(ctx, integrations.DeleteRequest{
					Integration: helperID("Integration", "int-1"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "Delete_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Integrations.Delete(ctx, integrations.DeleteRequest{
					Integration: helperID("Integration", "missing"),
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
			client := newTestServer(t, tt.handler)
			got, err := tt.invoke(context.Background(), client)
			tt.check(t, got, err)
		})
	}
}
