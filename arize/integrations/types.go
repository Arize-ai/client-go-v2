package integrations

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, config, and nested types are aliases to the generated wire
// shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	// Integration is a single integration returned by the API: a polymorphic,
	// type-tagged union over the LLM and agent variants. Read the active
	// variant with Discriminator() and ValueByDiscriminator, or unwrap
	// directly with AsLlmIntegration / AsAgentIntegration after checking the
	// discriminator.
	Integration = generated.Integration

	// ListIntegrations is the cursor-paginated list response shape. Each entry
	// in Integrations is a polymorphic Integration union.
	ListIntegrations = generated.ListIntegrationsResponse

	// AgentIntegration is the type=AGENT variant of an Integration: a
	// customer-hosted HTTPS endpoint plus a JSON Schema describing the request
	// payload. Obtain it via Integration.AsAgentIntegration.
	AgentIntegration = generated.AgentIntegration

	// LLMIntegration is the type=LLM variant of an Integration: a
	// model-provider integration (e.g. OpenAI, Anthropic). Obtain it via
	// Integration.AsLlmIntegration.
	LLMIntegration = generated.LlmIntegration

	// AgentConfig is the configuration for an agent integration: the endpoint,
	// request-payload JSON Schema, and named request presets.
	AgentConfig = generated.AgentConfig

	// AgentRequestPreset is a named, reusable request payload embedded in an
	// agent integration's config.request_presets.
	AgentRequestPreset = generated.AgentRequestPreset

	// LLMConfig is the per-provider config union on LLMIntegration.Config,
	// discriminated by provider. Unwrap the active variant with the matching
	// As<Provider>Config method: AsOpenAiConfig, AsAnthropicConfig,
	// AsGeminiConfig, AsAwsBedrockConfig, AsCustomConfig, AsVertexAiConfig,
	// AsNvidiaNimConfig, AsLiteLlmConfig, AsFireworksConfig, or
	// AsTogetherAiConfig.
	LLMConfig = generated.LlmConfig

	// OpenAIConfig is the OPEN_AI variant of an LLMConfig.
	OpenAIConfig = generated.OpenAiConfig

	// AnthropicConfig is the ANTHROPIC variant of an LLMConfig.
	AnthropicConfig = generated.AnthropicConfig

	// GeminiConfig is the GEMINI variant of an LLMConfig.
	GeminiConfig = generated.GeminiConfig

	// AWSBedrockConfig is the AWS_BEDROCK variant of an LLMConfig. Its Auth
	// field is an AWSBedrockAuth union; unwrap it with the matching
	// As<Mode>Auth method.
	AWSBedrockConfig = generated.AwsBedrockConfig

	// CustomConfig is the CUSTOM (OpenAI-compatible endpoint) variant of an
	// LLMConfig.
	CustomConfig = generated.CustomConfig

	// VertexAIConfig is the VERTEX_AI variant of an LLMConfig.
	VertexAIConfig = generated.VertexAiConfig

	// NvidiaNIMConfig is the NVIDIA_NIM variant of an LLMConfig.
	NvidiaNIMConfig = generated.NvidiaNimConfig

	// LiteLLMConfig is the LITELLM variant of an LLMConfig.
	LiteLLMConfig = generated.LiteLlmConfig

	// FireworksConfig is the FIREWORKS variant of an LLMConfig.
	FireworksConfig = generated.FireworksConfig

	// TogetherAiConfig is the TOGETHER_AI variant of an LLMConfig.
	TogetherAiConfig = generated.TogetherAiConfig

	// AWSBedrockAuth is the auth union on AWSBedrockConfig.Auth, discriminated
	// by auth_type. Unwrap the active variant with AsAwsBedrockDefaultAuth,
	// AsAwsBedrockBearerTokenAuth, or AsAwsBedrockProxyWithHeadersAuth.
	AWSBedrockAuth = generated.AwsBedrockAuth

	// AWSBedrockDefaultAuth is the DEFAULT (role-assumption) variant of an
	// AWSBedrockAuth.
	AWSBedrockDefaultAuth = generated.AwsBedrockDefaultAuth

	// AWSBedrockBearerTokenAuth is the BEARER_TOKEN variant of an
	// AWSBedrockAuth.
	AWSBedrockBearerTokenAuth = generated.AwsBedrockBearerTokenAuth

	// AWSBedrockProxyWithHeadersAuth is the PROXY_WITH_HEADERS variant of an
	// AWSBedrockAuth.
	AWSBedrockProxyWithHeadersAuth = generated.AwsBedrockProxyWithHeadersAuth

	// IntegrationScoping is a single visibility rule (organization and/or
	// space) applied to an integration. Account-wide when the list is empty.
	IntegrationScoping = generated.IntegrationScoping

	// IntegrationType is the integration category (LLM or AGENT). It selects
	// the shape of the integration's config.
	IntegrationType = generated.IntegrationType

	// LLMProvider is the model vendor backing an LLM integration.
	LLMProvider = generated.LlmIntegrationProvider
)

const (
	// IntegrationTypeLLM identifies a model-provider integration (e.g. OpenAI).
	IntegrationTypeLLM IntegrationType = generated.IntegrationTypeLLM
	// IntegrationTypeAgent identifies a customer-hosted agent integration.
	IntegrationTypeAgent IntegrationType = generated.IntegrationTypeAGENT

	// LLMProviderOpenAI is the OPEN_AI provider discriminator on a read config.
	LLMProviderOpenAI LLMProvider = generated.LlmIntegrationProviderOPENAI
	// LLMProviderAnthropic is the ANTHROPIC provider discriminator.
	LLMProviderAnthropic LLMProvider = generated.LlmIntegrationProviderANTHROPIC
	// LLMProviderGemini is the GEMINI provider discriminator.
	LLMProviderGemini LLMProvider = generated.LlmIntegrationProviderGEMINI
	// LLMProviderAWSBedrock is the AWS_BEDROCK provider discriminator.
	LLMProviderAWSBedrock LLMProvider = generated.LlmIntegrationProviderAWSBEDROCK
	// LLMProviderCustom is the CUSTOM provider discriminator.
	LLMProviderCustom LLMProvider = generated.LlmIntegrationProviderCUSTOM
	// LLMProviderVertexAI is the VERTEX_AI provider discriminator.
	LLMProviderVertexAI LLMProvider = generated.LlmIntegrationProviderVERTEXAI
	// LLMProviderNvidiaNIM is the NVIDIA_NIM provider discriminator.
	LLMProviderNvidiaNIM LLMProvider = generated.LlmIntegrationProviderNVIDIANIM
	// LLMProviderLiteLLM is the LITELLM provider discriminator.
	LLMProviderLiteLLM LLMProvider = generated.LlmIntegrationProviderLITELLM
	// LLMProviderFireworks is the FIREWORKS provider discriminator.
	LLMProviderFireworks LLMProvider = generated.LlmIntegrationProviderFIREWORKS
	// LLMProviderTogetherAi is the TOGETHER_AI provider discriminator.
	LLMProviderTogetherAi LLMProvider = generated.LlmIntegrationProviderTOGETHERAI
)

// AgentRequestPresetInput is the write shape for an agent request preset on
// CreateAgentRequest.RequestPresets and UpdateAgentRequest.RequestPresets.
// Server-generated fields (id, created_at, updated_at) are not accepted on
// input.
type AgentRequestPresetInput struct {
	// Name is the preset name, unique within the integration (length 1-255).
	// Required.
	Name string
	// Config is the partial request body validated against the parent
	// integration's input_schema with `required` dropped. Required.
	Config map[string]any
	// Description is an optional preset description (length 0-1024). When
	// empty, no description is sent.
	Description string
}

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// Type, when non-empty, filters the list to a single integration category
	// (IntegrationTypeLLM or IntegrationTypeAgent). When empty, integrations
	// of every type are returned; each item carries its `type` for
	// client-side discrimination.
	Type IntegrationType
	// Space, when non-empty, filters results by space. If the value is a
	// base64-encoded resource ID it is treated as a space ID (exact match);
	// otherwise it is used as a case-insensitive substring filter on the space
	// name (may match multiple spaces). When empty, no space filtering is
	// applied.
	Space string
	// Name, when non-empty, filters integrations to those whose name contains
	// the given case-insensitive substring. When empty, no name filtering is
	// applied.
	Name string
	// Limit is the optional maximum number of items to return. When zero, the
	// SDK applies a default of 50. Server max is 100.
	Limit int
	// Cursor is the optional opaque pagination cursor returned from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

// GetRequest is the request shape for Client.Get.
type GetRequest struct {
	// Integration is the integration's name or ID. Required.
	Integration string
	// Type is the integration category. Required when Integration is a name
	// (names are unique per account and type, so the type is needed to resolve
	// a name); ignored when Integration is an ID.
	Type IntegrationType
	// Space is an optional space name or ID used to narrow name resolution.
	// Ignored when Integration is an ID.
	Space string
}

// DeleteRequest is the request shape for Client.Delete.
type DeleteRequest struct {
	// Integration is the integration's name or ID. Required.
	Integration string
	// Type is the integration category. Required when Integration is a name;
	// ignored when Integration is an ID.
	Type IntegrationType
	// Space is an optional space name or ID used to narrow name resolution.
	// Ignored when Integration is an ID.
	Space string
}

// CreateAgentRequest is the request shape for Client.CreateAgent. Integration
// names are unique within the account per type.
type CreateAgentRequest struct {
	// Name is the integration name. Required.
	Name string
	// Endpoint is the HTTPS endpoint URL Arize calls for replay. Required.
	// Validated server-side for SSRF (must resolve to a public address).
	Endpoint string
	// InputSchema is the JSON Schema (Draft-07) the endpoint's request body
	// conforms to. Required.
	InputSchema map[string]any
	// Description is an optional human-readable description. When empty, no
	// description is sent.
	Description string
	// Headers is an optional cleartext header map sent to the endpoint.
	// Encrypted at rest; never returned in responses. When nil, no headers are
	// sent.
	Headers map[string]string
	// RequestPresets is an optional list of initial request presets. When nil,
	// no presets are created.
	RequestPresets []AgentRequestPresetInput
	// Scopings is an optional list of visibility rules. When nil, the server
	// applies its default of account-wide visibility.
	Scopings []IntegrationScoping
}

// CreateLLMRequest is the request shape for Client.CreateLLM. Integration
// names are unique within the account per type.
type CreateLLMRequest struct {
	// Name is the integration name. Required.
	Name string
	// Config is the provider-discriminated create config. Exactly one of its
	// per-provider fields must be set; that field selects the provider.
	// Required.
	Config CreateLLMConfig
	// Scopings is an optional list of visibility rules. When nil, the server
	// applies its default of account-wide visibility.
	Scopings []IntegrationScoping
}

// CreateLLMConfig is the provider-discriminated create config passed to
// Client.CreateLLM. Set exactly one of the per-provider fields; the SDK maps
// the set field onto the wire union and fills in the provider discriminator.
// Setting zero or more than one field is an error (ErrInvalidProviderConfig).
type CreateLLMConfig struct {
	// OpenAI selects the OPEN_AI provider.
	OpenAI *CreateOpenAIConfig
	// Anthropic selects the ANTHROPIC provider.
	Anthropic *CreateAnthropicConfig
	// Gemini selects the GEMINI provider.
	Gemini *CreateGeminiConfig
	// AWSBedrock selects the AWS_BEDROCK provider.
	AWSBedrock *CreateAWSBedrockConfig
	// Custom selects the CUSTOM (OpenAI-compatible endpoint) provider.
	Custom *CreateCustomConfig
	// VertexAI selects the VERTEX_AI provider.
	VertexAI *CreateVertexAIConfig
	// NvidiaNIM selects the NVIDIA_NIM provider.
	NvidiaNIM *CreateNvidiaNIMConfig
	// LiteLLM selects the LITELLM provider.
	LiteLLM *CreateLiteLLMConfig
	// Fireworks selects the FIREWORKS provider.
	Fireworks *CreateFireworksConfig
	// TogetherAi selects the TOGETHER_AI provider.
	TogetherAi *CreateTogetherAiConfig
}

// CreateOpenAIConfig is the OPEN_AI create config. The provider discriminator
// is set by the SDK.
type CreateOpenAIConfig struct {
	// APIKey is the provider API key. Write-only (never returned). Required.
	APIKey string
	// FunctionCallingEnabled, when non-nil, toggles function/tool calling.
	// Optional; nil leaves the server default (true).
	FunctionCallingEnabled *bool
}

// CreateAnthropicConfig is the ANTHROPIC create config. The provider
// discriminator is set by the SDK.
type CreateAnthropicConfig struct {
	// APIKey is the provider API key. Write-only (never returned). Required.
	APIKey string
	// FunctionCallingEnabled, when non-nil, toggles function/tool calling.
	// Optional; nil leaves the server default (true).
	FunctionCallingEnabled *bool
}

// CreateGeminiConfig is the GEMINI create config. The provider discriminator
// is set by the SDK.
type CreateGeminiConfig struct {
	// APIKey is the provider API key. Write-only (never returned). Required.
	APIKey string
	// FunctionCallingEnabled, when non-nil, toggles function/tool calling.
	// Optional; nil leaves the server default (true).
	FunctionCallingEnabled *bool
}

// CreateCustomConfig is the CUSTOM (OpenAI-compatible endpoint) create config.
// The provider discriminator is set by the SDK.
type CreateCustomConfig struct {
	// BaseURL is the endpoint URL requests are sent to (HTTPS). Required.
	BaseURL string
	// APIKey is the endpoint API key. Write-only (never returned). Optional;
	// empty is omitted.
	APIKey string
	// Headers is an optional cleartext header map sent to the endpoint.
	// Write-only; only header names are returned on read. When nil, no headers
	// are sent.
	Headers map[string]string
	// ModelNames, when non-nil, sets the custom model names to make available.
	// Optional; nil leaves the server default (none).
	ModelNames []string
	// IsDefaultModelsEnabled, when non-nil, toggles Arize's default model
	// catalog. Optional; nil leaves the server default (false).
	IsDefaultModelsEnabled *bool
	// FunctionCallingEnabled, when non-nil, toggles function/tool calling.
	// Optional; nil leaves the server default (true).
	FunctionCallingEnabled *bool
}

// CreateVertexAIConfig is the VERTEX_AI create config. The provider
// discriminator is set by the SDK.
type CreateVertexAIConfig struct {
	// ProjectID is the GCP project ID Arize accesses Vertex through. Required.
	ProjectID string
	// Location is the GCP region (e.g. us-central1). Required.
	Location string
	// ProjectAccessLabel is the label used to verify Arize's access to the GCP
	// project. Required.
	ProjectAccessLabel string
}

// CreateNvidiaNIMConfig is the NVIDIA_NIM create config. The provider
// discriminator is set by the SDK.
type CreateNvidiaNIMConfig struct {
	// BaseURL is the self-hosted NIM endpoint URL (HTTPS). Optional; empty
	// leaves the provider default endpoint.
	BaseURL string
	// APIKey is the endpoint API key. Write-only (never returned). Optional;
	// empty is omitted.
	APIKey string
	// Headers is an optional cleartext header map sent to the endpoint.
	// Write-only; only header names are returned on read. When nil, no headers
	// are sent.
	Headers map[string]string
	// ModelNames, when non-nil, sets the custom model names to make available.
	// Optional; nil leaves the server default (none).
	ModelNames []string
	// IsDefaultModelsEnabled, when non-nil, toggles Arize's default model
	// catalog. Optional; nil leaves the server default (false).
	IsDefaultModelsEnabled *bool
	// FunctionCallingEnabled, when non-nil, toggles function/tool calling.
	// Optional; nil leaves the server default (true).
	FunctionCallingEnabled *bool
}

// CreateLiteLLMConfig is the LITELLM create config. The provider
// discriminator is set by the SDK.
type CreateLiteLLMConfig struct {
	// BaseURL is the LiteLLM endpoint URL requests are sent to (HTTPS).
	// Required: LiteLLM is self-hosted, so there is no default endpoint.
	BaseURL string
	// APIKey is the LiteLLM virtual key. Write-only (never returned).
	// Required: the key scopes the models Arize can resolve and call.
	APIKey string
	// Headers is an optional cleartext header map sent to the endpoint.
	// Write-only; only header names are returned on read. When nil, no headers
	// are sent.
	Headers map[string]string
	// ModelNames, when non-nil, sets the custom model names to make available.
	// Optional; nil leaves the server default (none).
	ModelNames []string
	// FunctionCallingEnabled, when non-nil, toggles function/tool calling.
	// Optional; nil leaves the server default (true).
	FunctionCallingEnabled *bool
}

// CreateFireworksConfig is the FIREWORKS create config. The provider
// discriminator is set by the SDK. Fireworks is a single hosted service, so
// there is no endpoint field and no custom request headers.
type CreateFireworksConfig struct {
	// APIKey is the Fireworks AI API key. Write-only (never returned; it
	// surfaces as HasApiKey on read). Required.
	APIKey string
	// ModelNames, when non-nil, sets the custom model names to make available.
	// Optional; nil leaves the server default (none). Arize also resolves the
	// models the key can reach from the Fireworks account, so an integration
	// created without any still has a selectable model list.
	ModelNames []string
	// IsDefaultModelsEnabled, when non-nil, toggles Arize's default model
	// catalog. Optional; nil leaves the server default (false).
	IsDefaultModelsEnabled *bool
	// FunctionCallingEnabled, when non-nil, toggles function/tool calling.
	// Optional; nil leaves the server default (true).
	FunctionCallingEnabled *bool
}

// CreateTogetherAiConfig is the TOGETHER_AI create config. The provider
// discriminator is set by the SDK. Together AI is a single hosted service, so
// there is no endpoint field and no custom request headers.
type CreateTogetherAiConfig struct {
	// APIKey is the Together AI API key. Write-only (never returned; it
	// surfaces as HasApiKey on read). Required.
	APIKey string
	// ModelNames, when non-nil, sets the custom model names to make available.
	// Optional; nil leaves the server default (none). Arize also resolves the
	// models the key can reach from the Together AI account, so an integration
	// created without any still has a selectable model list.
	ModelNames []string
	// IsDefaultModelsEnabled, when non-nil, toggles Arize's default model
	// catalog. Optional; nil leaves the server default (false).
	IsDefaultModelsEnabled *bool
	// FunctionCallingEnabled, when non-nil, toggles function/tool calling.
	// Optional; nil leaves the server default (true).
	FunctionCallingEnabled *bool
}

// CreateAWSBedrockConfig is the AWS_BEDROCK create config. The server requires
// at least one model source: enable IsDefaultModelsEnabled or provide at least
// one entry in ModelNames, otherwise the request is rejected with 422.
type CreateAWSBedrockConfig struct {
	// Auth is the provider-discriminated auth config. Exactly one of its modes
	// must be set. Required.
	Auth CreateAWSBedrockAuth
	// IsDefaultModelsEnabled, when non-nil, toggles Arize's default Bedrock
	// model catalog. Optional; nil leaves the server default (false).
	IsDefaultModelsEnabled *bool
	// ModelNames, when non-nil, sets the custom model names to make available.
	// Optional; nil leaves the server default (none).
	ModelNames []string
}

// CreateAWSBedrockAuth is the provider-discriminated AWS Bedrock auth config.
// Set exactly one of the mode fields; the SDK maps the set field onto the wire
// union and fills in the auth_type discriminator. Setting zero or more than
// one field is an error (ErrInvalidBedrockAuth).
type CreateAWSBedrockAuth struct {
	// Default selects DEFAULT (role-assumption) auth.
	Default *CreateAWSBedrockDefaultAuth
	// BearerToken selects BEARER_TOKEN auth.
	BearerToken *CreateAWSBedrockBearerTokenAuth
	// ProxyWithHeaders selects PROXY_WITH_HEADERS auth.
	ProxyWithHeaders *CreateAWSBedrockProxyWithHeadersAuth
}

// CreateAWSBedrockDefaultAuth is the DEFAULT (role-assumption) create auth. The
// auth_type discriminator is set by the SDK.
type CreateAWSBedrockDefaultAuth struct {
	// RoleARN is the AWS IAM role ARN Arize assumes for cross-account access.
	// Required.
	RoleARN string
	// ExternalID is the external ID on the assume-role policy. Optional; empty
	// is omitted.
	ExternalID string
	// BaseURL is a custom Bedrock endpoint URL. Optional; empty leaves the
	// provider default endpoint.
	BaseURL string
}

// CreateAWSBedrockBearerTokenAuth is the BEARER_TOKEN create auth. The
// auth_type discriminator is set by the SDK.
type CreateAWSBedrockBearerTokenAuth struct {
	// APIKey is the Bedrock bearer token. Write-only (never returned).
	// Required.
	APIKey string
	// BaseURL is a custom Bedrock endpoint URL. Optional; empty leaves the
	// provider default endpoint.
	BaseURL string
}

// CreateAWSBedrockProxyWithHeadersAuth is the PROXY_WITH_HEADERS create auth.
// The auth_type discriminator is set by the SDK.
type CreateAWSBedrockProxyWithHeadersAuth struct {
	// BaseURL is the proxy URL requests are forwarded to (HTTPS). Required.
	BaseURL string
	// Headers is an optional cleartext header map sent to the proxy.
	// Write-only; only header names are returned on read. When nil, no headers
	// are sent.
	Headers map[string]string
}

// UpdateAgentRequest is the request shape for Client.UpdateAgent. Only the
// non-nil patch fields are sent; omitted (nil) fields are left unchanged on
// the server. Collection fields are replace-on-provide.
//
// PATCH semantics for nullable fields (Description, Headers):
//   - nil          → preserve the existing value
//   - &"" / &empty → clear the field on the server
//   - &v           → set the field to v
type UpdateAgentRequest struct {
	// Integration is the target integration's name or ID. Required.
	Integration string
	// Space is an optional space name or ID used to narrow name resolution.
	// Ignored when Integration is an ID.
	Space string

	// Name, when non-nil, updates the integration name.
	Name *string
	// Description, when non-nil, updates the description. Pass &"" to clear.
	Description *string
	// Endpoint, when non-nil, updates the replay endpoint URL.
	Endpoint *string
	// InputSchema, when non-nil, replaces the request-payload JSON Schema.
	InputSchema *map[string]any
	// Headers, when non-nil, replaces the header map. Pass a pointer to an
	// empty map to clear all headers.
	Headers *map[string]string
	// RequestPresets, when non-nil, replaces the preset list (matched by name:
	// existing names update in place, new names insert, removed names delete).
	RequestPresets *[]AgentRequestPresetInput
	// Scopings, when non-nil, replaces the scoping rules. An empty slice
	// reverts to account-wide visibility.
	Scopings *[]IntegrationScoping
}

// UpdateLLMRequest is the request shape for Client.UpdateLLM. Only the non-nil
// patch fields are sent; omitted (nil) fields are left unchanged on the
// server. The provider is immutable and cannot be changed here.
//
// PATCH semantics for nullable fields (APIKey, BaseURL, Headers):
//   - nil          → preserve the existing value
//   - &"" / &empty → clear the field on the server
//   - &v           → set the field to v
//
// Field applicability is provider-specific and enforced server-side with 422
// (the SDK passes through whatever is set without re-validating): APIKey and
// FunctionCallingEnabled do not apply to AWS_BEDROCK or VERTEX_AI; Auth applies
// to AWS_BEDROCK only; BaseURL and Headers apply to CUSTOM, NVIDIA_NIM, and
// LITELLM only; IsDefaultModelsEnabled applies to AWS_BEDROCK, CUSTOM,
// NVIDIA_NIM, FIREWORKS, and TOGETHER_AI only; ModelNames applies to
// AWS_BEDROCK, CUSTOM, NVIDIA_NIM, LITELLM, FIREWORKS, and TOGETHER_AI only;
// ProjectID, Location, and ProjectAccessLabel apply to
// VERTEX_AI only.
type UpdateLLMRequest struct {
	// Integration is the target integration's name or ID. Required.
	Integration string
	// Space is an optional space name or ID used to narrow name resolution.
	// Ignored when Integration is an ID.
	Space string

	// Name, when non-nil, updates the integration name.
	Name *string
	// Scopings, when non-nil, replaces the scoping rules. An empty slice
	// reverts to account-wide visibility.
	Scopings *[]IntegrationScoping

	// APIKey, when non-nil, rotates the provider API key. Pass &"" to clear it.
	APIKey *string
	// FunctionCallingEnabled, when non-nil, updates the function-calling flag.
	FunctionCallingEnabled *bool
	// Auth, when non-nil, replaces the AWS Bedrock auth config wholesale
	// (auth_type may change). AWS_BEDROCK only. Exactly one auth mode must be
	// set on the value.
	Auth *CreateAWSBedrockAuth
	// BaseURL, when non-nil, updates the endpoint URL. Pass &"" to clear it
	// (NVIDIA_NIM then falls back to the provider default; CUSTOM and LITELLM
	// reject the clear with 422 because base_url is required).
	BaseURL *string
	// Headers, when non-nil, replaces the full custom-header set. Pass a
	// pointer to a nil map to clear all headers (sends JSON null); pass a
	// pointer to a populated map to replace. Preserve when nil.
	Headers *map[string]string
	// IsDefaultModelsEnabled, when non-nil, toggles Arize's default model
	// catalog.
	IsDefaultModelsEnabled *bool
	// ModelNames, when non-nil, replaces the custom model list.
	ModelNames *[]string
	// ProjectID, when non-nil, updates the Vertex AI GCP project ID.
	ProjectID *string
	// Location, when non-nil, updates the Vertex AI GCP region.
	Location *string
	// ProjectAccessLabel, when non-nil, updates the Vertex AI project-access
	// label.
	ProjectAccessLabel *string
}
