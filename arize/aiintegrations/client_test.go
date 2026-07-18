package aiintegrations_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/aiintegrations"
)

// helperID encodes a fake base64 resource ID (e.g. "AiIntegration:1:foo") so
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

// wireCreateBody mirrors the JSON shape the server receives from POST
// /v2/ai-integrations so tests can assert on what the SDK sent without
// importing internal/generated. The scopings and provider_metadata
// payloads are inspected via the raw updateRawMap/createRawMap instead.
type wireCreateBody struct {
	Name                   string             `json:"name"`
	Provider               string             `json:"provider"`
	APIKey                 *string            `json:"api_key,omitempty"`
	BaseURL                *string            `json:"base_url,omitempty"`
	ModelNames             *[]string          `json:"model_names,omitempty"`
	Headers                *map[string]string `json:"headers,omitempty"`
	EnableDefaultModels    *bool              `json:"enable_default_models,omitempty"`
	FunctionCallingEnabled *bool              `json:"function_calling_enabled,omitempty"`
	AuthType               *string            `json:"auth_type,omitempty"`
}

// wireUpdateBody mirrors the JSON shape the server receives from PATCH
// /v2/ai-integrations/{id}. Only fields with `omitempty` appear; absence
// means "preserve existing". The scopings and provider_metadata payloads
// are inspected via the raw updateRawMap instead.
type wireUpdateBody struct {
	Name                   *string            `json:"name,omitempty"`
	Provider               *string            `json:"provider,omitempty"`
	APIKey                 *string            `json:"api_key,omitempty"`
	BaseURL                *string            `json:"base_url,omitempty"`
	ModelNames             *[]string          `json:"model_names,omitempty"`
	Headers                *map[string]string `json:"headers,omitempty"`
	EnableDefaultModels    *bool              `json:"enable_default_models,omitempty"`
	FunctionCallingEnabled *bool              `json:"function_calling_enabled,omitempty"`
	AuthType               *string            `json:"auth_type,omitempty"`
}

func TestAIIntegrations(t *testing.T) {
	// Captured by handlers and inspected in `check` closures. Subtests must
	// run sequentially — these are shared across all table entries. Do NOT
	// add t.Parallel() to the subtests without first moving the captures
	// into each subtest's scope (or threading them through the table struct).
	var (
		listQuery    url.Values
		createBody   wireCreateBody
		updateBody   wireUpdateBody
		updateRawMap map[string]any
		deletePath   string
	)

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.ListAIIntegrations{
					AiIntegrations: []aiintegrations.AIIntegration{{
						Id:                     "int-1",
						Name:                   "my-integration",
						Provider:               aiintegrations.AIIntegrationProviderAnthropic,
						AuthType:               aiintegrations.AIIntegrationAuthTypeDefault,
						CreatedAt:              time.Now(),
						UpdatedAt:              time.Now(),
						CreatedByUserId:        "user-1",
						EnableDefaultModels:    false,
						FunctionCallingEnabled: true,
						HasApiKey:              true,
						Scopings:               []aiintegrations.AIIntegrationScoping{},
					}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.List(ctx, aiintegrations.ListRequest{Limit: 10})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*aiintegrations.ListAIIntegrations)
				if len(resp.AiIntegrations) != 1 {
					t.Errorf("expected 1 integration, got %d", len(resp.AiIntegrations))
				}
			},
		},
		{
			name: "List_QueryParams",
			handler: func(w http.ResponseWriter, r *http.Request) {
				listQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.ListAIIntegrations{
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.List(ctx, aiintegrations.ListRequest{
					Name:   "openai",
					Space:  helperID("Space", "demo"), // ID short-circuits resolver
					Limit:  25,
					Cursor: "next-page-token",
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if listQuery.Get("name") != "openai" {
					t.Errorf("name: got %q, want %q", listQuery.Get("name"), "openai")
				}
				if listQuery.Get("space_id") != helperID("Space", "demo") {
					t.Errorf("space_id: got %q, want %q", listQuery.Get("space_id"), helperID("Space", "demo"))
				}
				if listQuery.Get("limit") != "25" {
					t.Errorf("limit: got %q, want %q", listQuery.Get("limit"), "25")
				}
				if listQuery.Get("cursor") != "next-page-token" {
					t.Errorf("cursor: got %q, want %q", listQuery.Get("cursor"), "next-page-token")
				}
			},
		},
		{
			name: "List_SpaceName_Substring",
			handler: func(w http.ResponseWriter, r *http.Request) {
				listQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.ListAIIntegrations{
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.List(ctx, aiintegrations.ListRequest{
					Space: "prod", // bare name → space_name substring filter
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if listQuery.Get("space_name") != "prod" {
					t.Errorf("space_name: got %q, want %q", listQuery.Get("space_name"), "prod")
				}
				if listQuery.Get("space_id") != "" {
					t.Errorf("space_id should be empty when filtering by name, got %q", listQuery.Get("space_id"))
				}
			},
		},
		{
			name: "List_DefaultLimit",
			handler: func(w http.ResponseWriter, r *http.Request) {
				listQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.ListAIIntegrations{
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.List(ctx, aiintegrations.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if listQuery.Get("limit") != "50" {
					t.Errorf("default limit: got %q, want %q", listQuery.Get("limit"), "50")
				}
				if listQuery.Get("name") != "" {
					t.Errorf("unexpected name filter: %q", listQuery.Get("name"))
				}
			},
		},
		{
			name: "Get_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id:              "int-1",
					Name:            "my-integration",
					Provider:        aiintegrations.AIIntegrationProviderOpenAI,
					AuthType:        aiintegrations.AIIntegrationAuthTypeDefault,
					HasApiKey:       true,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
					CreatedByUserId: "user-1",
					Scopings:        []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Get(ctx, aiintegrations.GetRequest{
					Integration: helperID("AiIntegration", "int-1"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*aiintegrations.AIIntegration)
				if resp.Name != "my-integration" {
					t.Errorf("name: got %q, want %q", resp.Name, "my-integration")
				}
				if resp.Provider != aiintegrations.AIIntegrationProviderOpenAI {
					t.Errorf("provider: got %q, want %q", resp.Provider, aiintegrations.AIIntegrationProviderOpenAI)
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
				return c.AIIntegrations.Get(ctx, aiintegrations.GetRequest{
					Integration: helperID("AiIntegration", "missing"),
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
			name: "Get_ByName_ResolvesViaList",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				// The resolver's paginated list call. The path exactly equals
				// /v2/ai-integrations (no trailing id segment).
				case strings.HasSuffix(r.URL.Path, "/v2/ai-integrations"):
					q := r.URL.Query()
					if q.Get("name") != "my-integration" {
						t.Errorf("list query name: got %q, want %q", q.Get("name"), "my-integration")
					}
					if q.Get("space_id") != helperID("Space", "demo") {
						t.Errorf("list query space_id: got %q, want %q", q.Get("space_id"), helperID("Space", "demo"))
					}
					_ = json.NewEncoder(w).Encode(aiintegrations.ListAIIntegrations{
						AiIntegrations: []aiintegrations.AIIntegration{{
							Id:       "int-resolved",
							Name:     "my-integration",
							Provider: aiintegrations.AIIntegrationProviderOpenAI,
							AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
							Scopings: []aiintegrations.AIIntegrationScoping{},
						}},
						Pagination: arize.PaginationMetadata{HasMore: false},
					})
				// The Get-by-id call that follows the resolution.
				case strings.Contains(r.URL.Path, "/v2/ai-integrations/"):
					if !strings.HasSuffix(r.URL.Path, "/int-resolved") {
						t.Errorf("get path: got %q, want suffix /int-resolved", r.URL.Path)
					}
					_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
						Id: "int-resolved", Name: "my-integration",
						Provider: aiintegrations.AIIntegrationProviderOpenAI,
						AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
						Scopings: []aiintegrations.AIIntegrationScoping{},
					})
				default:
					t.Errorf("unexpected request path: %q", r.URL.Path)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Get(ctx, aiintegrations.GetRequest{
					Integration: "my-integration", // bare name → triggers list-and-match
					Space:       helperID("Space", "demo"),
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*aiintegrations.AIIntegration)
				if resp.Id != "int-resolved" {
					t.Errorf("id: got %q, want %q", resp.Id, "int-resolved")
				}
			},
		},
		{
			name: "Create_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				createBody = wireCreateBody{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &createBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id:              "int-new",
					Name:            "my-new-integration",
					Provider:        aiintegrations.AIIntegrationProviderAnthropic,
					AuthType:        aiintegrations.AIIntegrationAuthTypeDefault,
					HasApiKey:       true,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
					CreatedByUserId: "user-1",
					Scopings:        []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:                   "my-new-integration",
					Provider:               aiintegrations.AIIntegrationProviderAnthropic,
					APIKey:                 "secret-key",
					DisableFunctionCalling: true,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				resp := got.(*aiintegrations.AIIntegration)
				if resp.Id != "int-new" {
					t.Errorf("id: got %q, want %q", resp.Id, "int-new")
				}
				if createBody.Name != "my-new-integration" {
					t.Errorf("name: got %q, want %q", createBody.Name, "my-new-integration")
				}
				if createBody.Provider != string(aiintegrations.AIIntegrationProviderAnthropic) {
					t.Errorf("provider: got %q, want %q", createBody.Provider, aiintegrations.AIIntegrationProviderAnthropic)
				}
				if createBody.APIKey == nil || *createBody.APIKey != "secret-key" {
					t.Errorf("api_key: got %v, want %q", createBody.APIKey, "secret-key")
				}
				// DisableFunctionCalling=true should send function_calling_enabled=false.
				if createBody.FunctionCallingEnabled == nil || *createBody.FunctionCallingEnabled != false {
					t.Errorf("function_calling_enabled: got %v, want &false", createBody.FunctionCallingEnabled)
				}
				// enable_default_models is always sent. Caller didn't set it
				// here, so Go zero (false) lands on the wire.
				if createBody.EnableDefaultModels == nil || *createBody.EnableDefaultModels != false {
					t.Errorf("enable_default_models: got %v, want &false (always sent)", createBody.EnableDefaultModels)
				}
				// Unset fields should be omitted entirely.
				if createBody.BaseURL != nil {
					t.Errorf("base_url: expected omitted, got %q", *createBody.BaseURL)
				}
			},
		},
		{
			name: "Create_OmitsZeroFields",
			handler: func(w http.ResponseWriter, r *http.Request) {
				createBody = wireCreateBody{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &createBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-min", Name: "minimal", Provider: aiintegrations.AIIntegrationProviderCustom,
					AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:     "minimal",
					Provider: aiintegrations.AIIntegrationProviderCustom,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// Caller set nothing optional → wire body has only name +
				// provider + enable_default_models (always sent so we don't
				// inherit a future change to the server-side default).
				if createBody.APIKey != nil || createBody.BaseURL != nil ||
					createBody.FunctionCallingEnabled != nil ||
					createBody.AuthType != nil || createBody.ModelNames != nil ||
					createBody.Headers != nil {
					t.Errorf("expected only name+provider+enable_default_models on the wire, got %+v", createBody)
				}
				if createBody.EnableDefaultModels == nil || *createBody.EnableDefaultModels != false {
					t.Errorf("enable_default_models: got %v, want &false (always sent)", createBody.EnableDefaultModels)
				}
			},
		},
		{
			name: "Create_NonNilEmptyCollections",
			handler: func(w http.ResponseWriter, r *http.Request) {
				updateRawMap = map[string]any{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &updateRawMap)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-empty", Name: "empty-collections",
					Provider: aiintegrations.AIIntegrationProviderCustom,
					AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:       "empty-collections",
					Provider:   aiintegrations.AIIntegrationProviderCustom,
					ModelNames: []string{},                              // non-nil empty → explicit []
					Headers:    map[string]string{},                     // non-nil empty → explicit {}
					Scopings:   []aiintegrations.AIIntegrationScoping{}, // non-nil empty → explicit []
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// Non-nil empty collections land on the wire as explicit empty
				// values (consistent !=nil rule for ModelNames, Headers, and
				// Scopings — matches Python's pydantic model_fields_set).
				if v, ok := updateRawMap["model_names"]; !ok {
					t.Error("model_names missing from body; expected []")
				} else if arr, ok := v.([]any); !ok || len(arr) != 0 {
					t.Errorf("model_names: got %v (%T), want []", v, v)
				}
				if v, ok := updateRawMap["headers"]; !ok {
					t.Error("headers missing from body; expected {}")
				} else if obj, ok := v.(map[string]any); !ok || len(obj) != 0 {
					t.Errorf("headers: got %v (%T), want {}", v, v)
				}
				if v, ok := updateRawMap["scopings"]; !ok {
					t.Error("scopings missing from body; expected []")
				} else if arr, ok := v.([]any); !ok || len(arr) != 0 {
					t.Errorf("scopings: got %v (%T), want []", v, v)
				}
			},
		},
		{
			name: "Create_BadRequest",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "bad request", "status": 400})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:     "broken",
					Provider: aiintegrations.AIIntegrationProviderOpenAI,
				})
			},
			check: func(t *testing.T, got any, err error) {
				var bre *arize.BadRequestError
				if !errors.As(err, &bre) {
					t.Errorf("expected *BadRequestError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Create_UnprocessableEntity",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"title":  "validation failed",
					"status": 422,
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:     "bedrock-missing-metadata",
					Provider: aiintegrations.AIIntegrationProviderAWSBedrock,
					// ProviderMetadata intentionally omitted — server returns 422.
				})
			},
			check: func(t *testing.T, _ any, err error) {
				var uee *arize.UnprocessableEntityError
				if !errors.As(err, &uee) {
					t.Errorf("expected *UnprocessableEntityError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Update_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				updateBody = wireUpdateBody{}
				updateRawMap = map[string]any{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &updateBody)
				_ = json.Unmarshal(body, &updateRawMap)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-1", Name: "renamed", Provider: aiintegrations.AIIntegrationProviderOpenAI,
					AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				newName := "renamed"
				return c.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
					Integration: helperID("AiIntegration", "int-1"),
					Name:        &newName,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updateBody.Name == nil || *updateBody.Name != "renamed" {
					t.Errorf("name: got %v, want %q", updateBody.Name, "renamed")
				}
				// Unset PATCH fields must be omitted entirely from the wire body —
				// not encoded as null — so the server preserves them.
				for _, k := range []string{"api_key", "base_url", "model_names", "headers",
					"enable_default_models", "function_calling_enabled", "auth_type",
					"provider", "provider_metadata", "scopings"} {
					if _, present := updateRawMap[k]; present {
						t.Errorf("%s: expected omitted, but key is present in body", k)
					}
				}
			},
		},
		{
			name: "Update_ClearNullableField",
			handler: func(w http.ResponseWriter, r *http.Request) {
				updateBody = wireUpdateBody{}
				updateRawMap = map[string]any{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &updateBody)
				_ = json.Unmarshal(body, &updateRawMap)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-1", Name: "x", Provider: aiintegrations.AIIntegrationProviderOpenAI,
					AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				empty := ""
				return c.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
					Integration: helperID("AiIntegration", "int-1"),
					APIKey:      &empty,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// &"" must land as `"api_key": null` on the wire — the
				// OpenAPI "Pass null to remove" signal that the server
				// recognizes as the clear semantic.
				v, present := updateRawMap["api_key"]
				if !present {
					t.Fatalf("api_key key missing from body; expected JSON null")
				}
				if v != nil {
					t.Errorf("api_key clear: got %v (%T), want JSON null", v, v)
				}
			},
		},
		{
			name: "Delete_Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				deletePath = r.URL.Path
				w.WriteHeader(http.StatusNoContent)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.AIIntegrations.Delete(ctx, aiintegrations.DeleteRequest{
					Integration: helperID("AiIntegration", "int-1"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				wantSuffix := "/" + helperID("AiIntegration", "int-1")
				if !strings.HasSuffix(deletePath, wantSuffix) {
					t.Errorf("delete path: got %q, want suffix %q", deletePath, wantSuffix)
				}
			},
		},
		{
			name: "Create_AllOptionalFields",
			handler: func(w http.ResponseWriter, r *http.Request) {
				createBody = wireCreateBody{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &createBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-full", Name: "full", Provider: aiintegrations.AIIntegrationProviderOpenAI,
					AuthType: aiintegrations.AIIntegrationAuthTypeBearerToken,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:                "full",
					Provider:            aiintegrations.AIIntegrationProviderOpenAI,
					APIKey:              "sk-test",
					BaseURL:             "https://proxy.example.com",
					ModelNames:          []string{"gpt-4o", "gpt-4o-mini"},
					Headers:             map[string]string{"X-Org": "acme"},
					EnableDefaultModels: true,
					AuthType:            aiintegrations.AIIntegrationAuthTypeBearerToken,
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if createBody.BaseURL == nil || *createBody.BaseURL != "https://proxy.example.com" {
					t.Errorf("base_url: got %v, want %q", createBody.BaseURL, "https://proxy.example.com")
				}
				if createBody.ModelNames == nil || len(*createBody.ModelNames) != 2 {
					t.Errorf("model_names: got %v, want 2 entries", createBody.ModelNames)
				}
				if createBody.Headers == nil || (*createBody.Headers)["X-Org"] != "acme" {
					t.Errorf("headers: got %v, want X-Org=acme", createBody.Headers)
				}
				if createBody.EnableDefaultModels == nil || !*createBody.EnableDefaultModels {
					t.Errorf("enable_default_models: got %v, want &true", createBody.EnableDefaultModels)
				}
				if createBody.AuthType == nil || *createBody.AuthType != string(aiintegrations.AIIntegrationAuthTypeBearerToken) {
					t.Errorf("auth_type: got %v, want %q", createBody.AuthType, aiintegrations.AIIntegrationAuthTypeBearerToken)
				}
				// DisableFunctionCalling not set → field omitted entirely so the server
				// keeps its default of true.
				if createBody.FunctionCallingEnabled != nil {
					t.Errorf("function_calling_enabled: expected omitted, got %v", *createBody.FunctionCallingEnabled)
				}
			},
		},
		{
			name: "Update_AllPatchableFields",
			handler: func(w http.ResponseWriter, r *http.Request) {
				updateBody = wireUpdateBody{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &updateBody)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-1", Name: "patched", Provider: aiintegrations.AIIntegrationProviderAnthropic,
					AuthType: aiintegrations.AIIntegrationAuthTypeBearerToken,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				newName := "patched"
				newAPIKey := "sk-rotated"
				newBaseURL := "https://proxy2.example.com"
				newProvider := aiintegrations.AIIntegrationProviderAnthropic
				newAuth := aiintegrations.AIIntegrationAuthTypeBearerToken
				newModels := []string{"claude-3-5-sonnet"}
				newHeaders := map[string]string{"X-Tenant": "demo"}
				enableDefaults := true
				fnCalling := false
				return c.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
					Integration:            helperID("AiIntegration", "int-1"),
					Name:                   &newName,
					Provider:               &newProvider,
					APIKey:                 &newAPIKey,
					BaseURL:                &newBaseURL,
					ModelNames:             &newModels,
					Headers:                &newHeaders,
					EnableDefaultModels:    &enableDefaults,
					FunctionCallingEnabled: &fnCalling,
					AuthType:               &newAuth,
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updateBody.Name == nil || *updateBody.Name != "patched" {
					t.Errorf("name: got %v, want %q", updateBody.Name, "patched")
				}
				if updateBody.Provider == nil || *updateBody.Provider != string(aiintegrations.AIIntegrationProviderAnthropic) {
					t.Errorf("provider: got %v, want %q", updateBody.Provider, aiintegrations.AIIntegrationProviderAnthropic)
				}
				if updateBody.APIKey == nil || *updateBody.APIKey != "sk-rotated" {
					t.Errorf("api_key: got %v, want %q", updateBody.APIKey, "sk-rotated")
				}
				if updateBody.BaseURL == nil || *updateBody.BaseURL != "https://proxy2.example.com" {
					t.Errorf("base_url: got %v, want %q", updateBody.BaseURL, "https://proxy2.example.com")
				}
				if updateBody.ModelNames == nil || len(*updateBody.ModelNames) != 1 || (*updateBody.ModelNames)[0] != "claude-3-5-sonnet" {
					t.Errorf("model_names: got %v, want [claude-3-5-sonnet]", updateBody.ModelNames)
				}
				if updateBody.Headers == nil || (*updateBody.Headers)["X-Tenant"] != "demo" {
					t.Errorf("headers: got %v, want X-Tenant=demo", updateBody.Headers)
				}
				if updateBody.EnableDefaultModels == nil || !*updateBody.EnableDefaultModels {
					t.Errorf("enable_default_models: got %v, want &true", updateBody.EnableDefaultModels)
				}
				if updateBody.FunctionCallingEnabled == nil || *updateBody.FunctionCallingEnabled {
					t.Errorf("function_calling_enabled: got %v, want &false", updateBody.FunctionCallingEnabled)
				}
				if updateBody.AuthType == nil || *updateBody.AuthType != string(aiintegrations.AIIntegrationAuthTypeBearerToken) {
					t.Errorf("auth_type: got %v, want %q", updateBody.AuthType, aiintegrations.AIIntegrationAuthTypeBearerToken)
				}
			},
		},
		{
			name: "Update_ClearHeaders",
			handler: func(w http.ResponseWriter, r *http.Request) {
				updateRawMap = map[string]any{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &updateRawMap)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-1", Name: "x", Provider: aiintegrations.AIIntegrationProviderOpenAI,
					AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				empty := map[string]string{}
				return c.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
					Integration: helperID("AiIntegration", "int-1"),
					Headers:     &empty,
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// &map{} must land as `"headers": null` on the wire — the
				// OpenAPI "Pass null to remove" signal that the server
				// recognizes as the clear semantic. (Empty object would
				// store {} on the server, not clear.)
				h, present := updateRawMap["headers"]
				if !present {
					t.Fatal("headers key missing from body; expected JSON null")
				}
				if h != nil {
					t.Errorf("headers clear: got %v (%T), want JSON null", h, h)
				}
			},
		},
		{
			name: "Create_WithAWSProviderMetadata",
			handler: func(w http.ResponseWriter, r *http.Request) {
				updateRawMap = map[string]any{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &updateRawMap)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-aws", Name: "bedrock",
					Provider: aiintegrations.AIIntegrationProviderAWSBedrock,
					AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:     "bedrock",
					Provider: aiintegrations.AIIntegrationProviderAWSBedrock,
					ProviderMetadata: &aiintegrations.ProviderMetadata{
						AWS: &aiintegrations.AWSProviderMetadata{
							RoleArn:    "arn:aws:iam::123456789012:role/MyRole",
							ExternalId: ptr("ext-abc"),
						},
					},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				pm, ok := updateRawMap["provider_metadata"].(map[string]any)
				if !ok {
					t.Fatalf("provider_metadata: got %v (%T), want object", updateRawMap["provider_metadata"], updateRawMap["provider_metadata"])
				}
				if pm["kind"] != "AWS" {
					t.Errorf("provider_metadata.kind: got %v, want %q", pm["kind"], "AWS")
				}
				if pm["role_arn"] != "arn:aws:iam::123456789012:role/MyRole" {
					t.Errorf("provider_metadata.role_arn: got %v", pm["role_arn"])
				}
				if pm["external_id"] != "ext-abc" {
					t.Errorf("provider_metadata.external_id: got %v", pm["external_id"])
				}
			},
		},
		{
			name: "Create_WithGCPProviderMetadata",
			handler: func(w http.ResponseWriter, r *http.Request) {
				updateRawMap = map[string]any{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &updateRawMap)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-gcp", Name: "vertex",
					Provider: aiintegrations.AIIntegrationProviderVertexAI,
					AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:     "vertex",
					Provider: aiintegrations.AIIntegrationProviderVertexAI,
					ProviderMetadata: &aiintegrations.ProviderMetadata{
						GCP: &aiintegrations.GCPProviderMetadata{
							ProjectId:          "my-project",
							Location:           "us-central1",
							ProjectAccessLabel: "demo",
						},
					},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				pm, ok := updateRawMap["provider_metadata"].(map[string]any)
				if !ok {
					t.Fatalf("provider_metadata: got %v (%T), want object", updateRawMap["provider_metadata"], updateRawMap["provider_metadata"])
				}
				if pm["kind"] != "GCP" {
					t.Errorf("provider_metadata.kind: got %v, want %q", pm["kind"], "GCP")
				}
				if pm["project_id"] != "my-project" {
					t.Errorf("provider_metadata.project_id: got %v", pm["project_id"])
				}
				if pm["location"] != "us-central1" {
					t.Errorf("provider_metadata.location: got %v", pm["location"])
				}
			},
		},
		{
			name:    "Create_ProviderMetadata_MutuallyExclusive",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
					Name:     "broken",
					Provider: aiintegrations.AIIntegrationProviderAWSBedrock,
					ProviderMetadata: &aiintegrations.ProviderMetadata{
						AWS: &aiintegrations.AWSProviderMetadata{RoleArn: "arn:x"},
						GCP: &aiintegrations.GCPProviderMetadata{ProjectId: "p"},
					},
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err == nil {
					t.Fatal("expected error for AWS+GCP both set, got nil")
				}
			},
		},
		{
			name: "Update_ClearProviderMetadata",
			handler: func(w http.ResponseWriter, r *http.Request) {
				updateRawMap = map[string]any{}
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &updateRawMap)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(aiintegrations.AIIntegration{
					Id: "int-1", Name: "x", Provider: aiintegrations.AIIntegrationProviderOpenAI,
					AuthType: aiintegrations.AIIntegrationAuthTypeDefault,
					Scopings: []aiintegrations.AIIntegrationScoping{},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
					Integration:      helperID("AiIntegration", "int-1"),
					ProviderMetadata: &aiintegrations.ProviderMetadata{}, // zero-value wrapper → clear
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				v, present := updateRawMap["provider_metadata"]
				if !present {
					t.Fatal("provider_metadata key missing from body; expected JSON null")
				}
				if v != nil {
					t.Errorf("provider_metadata clear: got %v (%T), want JSON null", v, v)
				}
			},
		},
		{
			name:    "Update_NoFields",
			handler: func(w http.ResponseWriter, r *http.Request) { t.Error("server should not be called") },
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
					Integration: helperID("AiIntegration", "int-1"),
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, aiintegrations.ErrNoUpdateFields) {
					t.Errorf("expected ErrNoUpdateFields, got %v", err)
				}
			},
		},
		{
			name: "Update_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				newName := "x"
				return c.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
					Integration: helperID("AiIntegration", "missing"),
					Name:        &newName,
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
			name: "Delete_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "not found", "status": 404})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.AIIntegrations.Delete(ctx, aiintegrations.DeleteRequest{
					Integration: helperID("AiIntegration", "missing"),
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
