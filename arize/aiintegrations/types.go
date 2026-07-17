package aiintegrations

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	// AIIntegration is a single AI integration configuration for an external
	// LLM provider.
	AIIntegration = generated.AiIntegration

	// AIIntegrationList is the cursor-paginated list response shape.
	AIIntegrationList = generated.ListAiIntegrationsResponse

	// AIIntegrationProvider is the LLM provider for an integration.
	AIIntegrationProvider = generated.AiIntegrationProvider

	// AIIntegrationAuthType is the authentication method an integration uses
	// when calling the upstream provider.
	AIIntegrationAuthType = generated.AiIntegrationAuthType

	// AIIntegrationScoping is a single visibility rule (account, organization,
	// or space) applied to an integration.
	AIIntegrationScoping = generated.AiIntegrationScoping

	// AWSProviderMetadata holds AWS Bedrock provider metadata. Wrap in a
	// ProviderMetadata and assign to CreateRequest.ProviderMetadata or
	// UpdateRequest.ProviderMetadata. The kind discriminator is set
	// automatically by ProviderMetadata.MarshalJSON.
	AWSProviderMetadata = generated.AwsProviderMetadata

	// GCPProviderMetadata holds Vertex AI (GCP) provider metadata. Wrap in a
	// ProviderMetadata and assign to CreateRequest.ProviderMetadata or
	// UpdateRequest.ProviderMetadata. The kind discriminator is set
	// automatically by ProviderMetadata.MarshalJSON.
	GCPProviderMetadata = generated.GcpProviderMetadata
)

const (
	AIIntegrationProviderOpenAI      AIIntegrationProvider = generated.AiIntegrationProviderOPENAI
	AIIntegrationProviderAzureOpenAI AIIntegrationProvider = generated.AiIntegrationProviderAZUREOPENAI
	AIIntegrationProviderAWSBedrock  AIIntegrationProvider = generated.AiIntegrationProviderAWSBEDROCK
	AIIntegrationProviderVertexAI    AIIntegrationProvider = generated.AiIntegrationProviderVERTEXAI
	AIIntegrationProviderAnthropic   AIIntegrationProvider = generated.AiIntegrationProviderANTHROPIC
	AIIntegrationProviderCustom      AIIntegrationProvider = generated.AiIntegrationProviderCUSTOM
	AIIntegrationProviderNvidiaNim   AIIntegrationProvider = generated.AiIntegrationProviderNVIDIANIM
	AIIntegrationProviderGemini      AIIntegrationProvider = generated.AiIntegrationProviderGEMINI

	AIIntegrationAuthTypeDefault          AIIntegrationAuthType = generated.AiIntegrationAuthTypeDEFAULT
	AIIntegrationAuthTypeProxyWithHeaders AIIntegrationAuthType = generated.AiIntegrationAuthTypePROXYWITHHEADERS
	AIIntegrationAuthTypeBearerToken      AIIntegrationAuthType = generated.AiIntegrationAuthTypeBEARERTOKEN
)

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// Name, when non-empty, filters integrations to those whose name contains
	// the given case-insensitive substring. When empty, no name filtering is
	// applied.
	Name string
	// Space, when non-empty, filters results by space. If the value is a
	// base64-encoded resource ID it is treated as a space ID (exact match);
	// otherwise it is used as a case-insensitive substring filter on the
	// space name (may match multiple spaces).
	Space string
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
	// Space is the parent space's name or ID. Required when Integration is a
	// name; ignored when Integration is an ID.
	Space string
}

// CreateRequest is the request shape for Client.Create. Integration names
// must be unique within the account.
type CreateRequest struct {
	// Name is the user-defined integration name. Required.
	Name string
	// Provider is the AI provider this integration calls. Required.
	//
	// When Provider is AIIntegrationProviderAWSBedrock or
	// AIIntegrationProviderVertexAI, ProviderMetadata must also be set. The
	// SDK does not pre-validate this; the server returns
	// *arize.UnprocessableEntityError (HTTP 422) if ProviderMetadata is
	// omitted or does not match the provider.
	Provider AIIntegrationProvider
	// APIKey is the optional provider API key. Write-only; never returned by
	// the server. When empty, no API key is configured.
	APIKey string
	// BaseURL is the optional custom base URL for the provider. When empty,
	// the provider's default endpoint is used.
	BaseURL string
	// ModelNames is the optional list of supported model names. When nil,
	// the field is omitted from the request and no allowlist is configured.
	// Pass a non-nil empty slice to send an explicit empty allowlist.
	ModelNames []string
	// Headers is the optional set of custom headers included in requests to
	// the provider. When nil, the field is omitted from the request and no
	// custom headers are sent. Pass a non-nil empty map to send an explicit
	// empty headers object.
	Headers map[string]string
	// EnableDefaultModels enables the provider's default model list. Optional;
	// the SDK always sends this field on the wire so it does not inherit a
	// future change to the server-side default. Go's zero value (false)
	// matches the current server default.
	EnableDefaultModels bool
	// DisableFunctionCalling, when true, disables function/tool calling for
	// this integration. Optional; zero (false) leaves the server default of
	// true (function calling enabled) in place. Inverted relative to the wire
	// field `function_calling_enabled` so Go's zero value matches the server
	// default.
	DisableFunctionCalling bool
	// AuthType is the optional authentication method. When zero, the server
	// applies its default (AIIntegrationAuthTypeDefault).
	AuthType AIIntegrationAuthType
	// ProviderMetadata is the optional provider-specific configuration.
	// Required when Provider is AIIntegrationProviderAWSBedrock (set
	// ProviderMetadata.AWS) or AIIntegrationProviderVertexAI (set
	// ProviderMetadata.GCP). Not pre-validated by the SDK — the server
	// returns *arize.UnprocessableEntityError (HTTP 422) on mismatch.
	ProviderMetadata *ProviderMetadata
	// Scopings is the optional list of visibility rules. When nil, the field
	// is omitted from the request and the server applies its default of
	// account-wide visibility. Pass a non-nil empty slice to send an explicit
	// empty scopings list; the server-side interpretation of an empty list is
	// implementation-defined.
	Scopings []AIIntegrationScoping
}

// UpdateRequest is the request shape for Client.Update. Only the patch fields
// that are non-nil are sent on the wire; omitted (nil) fields are left
// unchanged on the server.
//
// PATCH semantics for nullable fields (e.g. APIKey, BaseURL, Headers,
// ProviderMetadata):
//   - nil          → preserve the existing value
//   - &"" / &zero  → clear the field on the server
//   - &v           → set the field to v
type UpdateRequest struct {
	// Integration is the target integration's name or ID. Required.
	Integration string
	// Space is the parent space's name or ID. Required when Integration is a
	// name; ignored when Integration is an ID.
	Space string

	// Name, when non-nil, updates the integration's name.
	Name *string
	// Provider, when non-nil, updates the AI provider. Changing provider on a
	// live integration is unusual — prefer creating a new integration. When
	// patching Provider to AIIntegrationProviderAWSBedrock or
	// AIIntegrationProviderVertexAI, ProviderMetadata must be patched in the
	// same call. Not pre-validated by the SDK — the server returns
	// *arize.UnprocessableEntityError (HTTP 422) on mismatch.
	Provider *AIIntegrationProvider
	// APIKey, when non-nil, updates the provider API key. Pass &"" to clear.
	APIKey *string
	// BaseURL, when non-nil, updates the custom base URL. Pass &"" to clear.
	BaseURL *string
	// ModelNames, when non-nil, replaces the entire supported-models list.
	ModelNames *[]string
	// Headers, when non-nil, replaces the entire custom-headers map. Pass a
	// pointer to an empty map to clear all headers.
	Headers *map[string]string
	// EnableDefaultModels, when non-nil, updates the enable-default-models
	// flag.
	EnableDefaultModels *bool
	// FunctionCallingEnabled, when non-nil, updates the function-calling
	// flag. Unlike CreateRequest.DisableFunctionCalling this is non-inverted
	// because PATCH already distinguishes nil (preserve) from &false (set
	// to false).
	FunctionCallingEnabled *bool
	// AuthType, when non-nil, updates the authentication method.
	AuthType *AIIntegrationAuthType
	// ProviderMetadata, when non-nil, updates the provider-specific
	// configuration. Set ProviderMetadata.AWS or ProviderMetadata.GCP to
	// update; pass an empty &ProviderMetadata{} to clear.
	ProviderMetadata *ProviderMetadata
	// Scopings, when non-nil, replaces the entire scopings list.
	Scopings *[]AIIntegrationScoping
}

// DeleteRequest is the request shape for Client.Delete.
type DeleteRequest struct {
	// Integration is the integration's name or ID. Required.
	Integration string
	// Space is the parent space's name or ID. Required when Integration is a
	// name; ignored when Integration is an ID.
	Space string
}
