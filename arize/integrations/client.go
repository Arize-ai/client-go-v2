package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Integrations API (/v2/integrations),
// covering both LLM (model-provider) and agent (customer-hosted endpoint)
// integrations.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of integrations. req.Type is an optional
// filter: when set, only integrations of that type are returned; when
// omitted, the list spans every type, each item carrying its `type`
// discriminator. Defaults to a page size of 50.
func (c *Client) List(ctx context.Context, req ListRequest) (*ListIntegrations, error) {
	prerelease.Warn("integrations.list", prerelease.Alpha)
	params := &generated.ListIntegrationsParams{
		Type:   optfields.PtrIfSet(req.Type),
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	resp, err := c.gen.ListIntegrationsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single integration, resolving by name or ID. When req.Integration
// is a name, req.Type is required (names are unique per account and type); a
// name without a type, like an unresolvable name, returns
// arize.ResourceNotFoundError.
func (c *Client) Get(ctx context.Context, req GetRequest) (*Integration, error) {
	prerelease.Warn("integrations.get", prerelease.Alpha)
	id, err := resolve.FindIntegrationID(ctx, c.gen, req.Integration, req.Space, req.Type)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetIntegrationWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// CreateAgent creates a new agent (type=AGENT) integration and returns it.
func (c *Client) CreateAgent(ctx context.Context, req CreateAgentRequest) (*Integration, error) {
	prerelease.Warn("integrations.create_agent", prerelease.Alpha)

	cfg := generated.CreateAgentConfig{
		Endpoint:    req.Endpoint,
		InputSchema: req.InputSchema,
	}
	if req.Headers != nil {
		cfg.Headers = &req.Headers
	}
	if req.RequestPresets != nil {
		presets := createAgentPresets(req.RequestPresets)
		cfg.RequestPresets = &presets
	}

	agent := generated.CreateAgentIntegrationRequest{
		Type:        generated.CreateAgentIntegrationRequestTypeAGENT,
		Name:        req.Name,
		Config:      cfg,
		Description: optfields.PtrIfSet(req.Description),
	}
	if req.Scopings != nil {
		scopings := scopingRequests(req.Scopings)
		agent.Scopings = &scopings
	}

	var body generated.CreateIntegrationJSONRequestBody
	if err := body.FromCreateAgentIntegrationRequest(agent); err != nil {
		return nil, fmt.Errorf("integrations: build agent create body: %w", err)
	}
	return c.create(ctx, body)
}

// CreateLLM creates a new LLM (type=LLM) integration and returns it.
// req.Config selects the provider: set exactly one of its per-provider fields.
func (c *Client) CreateLLM(ctx context.Context, req CreateLLMRequest) (*Integration, error) {
	prerelease.Warn("integrations.create_llm", prerelease.Alpha)

	cfg, err := buildCreateLLMConfig(req.Config)
	if err != nil {
		return nil, err
	}

	llm := generated.CreateLlmIntegrationRequest{
		Type:   generated.CreateLlmIntegrationRequestTypeLLM,
		Name:   req.Name,
		Config: cfg,
	}
	if req.Scopings != nil {
		scopings := scopingRequests(req.Scopings)
		llm.Scopings = &scopings
	}

	var body generated.CreateIntegrationJSONRequestBody
	if err := body.FromCreateLlmIntegrationRequest(llm); err != nil {
		return nil, fmt.Errorf("integrations: build llm create body: %w", err)
	}
	return c.create(ctx, body)
}

// scopingRequests converts public scopings to the strict write-request wire
// form (identical fields; the schemas differ only in request strictness).
func scopingRequests(in []IntegrationScoping) []generated.IntegrationScopingRequest {
	out := make([]generated.IntegrationScopingRequest, len(in))
	for i, s := range in {
		out[i] = generated.IntegrationScopingRequest(s)
	}
	return out
}

// buildCreateLLMConfig maps the provider-discriminated CreateLLMConfig onto the
// generated wire union, setting each provider discriminator. Exactly one
// per-provider field must be set.
func buildCreateLLMConfig(in CreateLLMConfig) (generated.CreateLlmConfig, error) {
	var cfg generated.CreateLlmConfig

	set := 0
	for _, present := range []bool{
		in.OpenAI != nil, in.Anthropic != nil, in.Gemini != nil,
		in.AWSBedrock != nil, in.Custom != nil, in.VertexAI != nil, in.NvidiaNIM != nil,
		in.LiteLLM != nil, in.Fireworks != nil, in.TogetherAi != nil,
	} {
		if present {
			set++
		}
	}
	if set != 1 {
		return cfg, ErrInvalidProviderConfig
	}

	var err error
	switch {
	case in.OpenAI != nil:
		err = cfg.FromCreateOpenAiConfig(generated.CreateOpenAiConfig{
			Provider:                 generated.CreateOpenAiConfigProviderOPENAI,
			ApiKey:                   in.OpenAI.APIKey,
			IsFunctionCallingEnabled: in.OpenAI.FunctionCallingEnabled,
		})
	case in.Anthropic != nil:
		err = cfg.FromCreateAnthropicConfig(generated.CreateAnthropicConfig{
			Provider:                 generated.CreateAnthropicConfigProviderANTHROPIC,
			ApiKey:                   in.Anthropic.APIKey,
			IsFunctionCallingEnabled: in.Anthropic.FunctionCallingEnabled,
		})
	case in.Gemini != nil:
		err = cfg.FromCreateGeminiConfig(generated.CreateGeminiConfig{
			Provider:                 generated.CreateGeminiConfigProviderGEMINI,
			ApiKey:                   in.Gemini.APIKey,
			IsFunctionCallingEnabled: in.Gemini.FunctionCallingEnabled,
		})
	case in.AWSBedrock != nil:
		auth, authErr := buildCreateBedrockAuth(in.AWSBedrock.Auth)
		if authErr != nil {
			return cfg, authErr
		}
		err = cfg.FromCreateAwsBedrockConfig(generated.CreateAwsBedrockConfig{
			Provider:               generated.CreateAwsBedrockConfigProviderAWSBEDROCK,
			Auth:                   auth,
			IsDefaultModelsEnabled: in.AWSBedrock.IsDefaultModelsEnabled,
			ModelNames:             optfields.PtrSliceIfSet(in.AWSBedrock.ModelNames),
		})
	case in.Custom != nil:
		err = cfg.FromCreateCustomConfig(generated.CreateCustomConfig{
			Provider:                 generated.CreateCustomConfigProviderCUSTOM,
			BaseUrl:                  in.Custom.BaseURL,
			ApiKey:                   optfields.PtrIfSet(in.Custom.APIKey),
			Headers:                  optfields.PtrMapIfSet(in.Custom.Headers),
			ModelNames:               optfields.PtrSliceIfSet(in.Custom.ModelNames),
			IsDefaultModelsEnabled:   in.Custom.IsDefaultModelsEnabled,
			IsFunctionCallingEnabled: in.Custom.FunctionCallingEnabled,
		})
	case in.VertexAI != nil:
		err = cfg.FromCreateVertexAiConfig(generated.CreateVertexAiConfig{
			Provider:           generated.CreateVertexAiConfigProviderVERTEXAI,
			ProjectId:          in.VertexAI.ProjectID,
			Location:           in.VertexAI.Location,
			ProjectAccessLabel: in.VertexAI.ProjectAccessLabel,
		})
	case in.NvidiaNIM != nil:
		err = cfg.FromCreateNvidiaNimConfig(generated.CreateNvidiaNimConfig{
			Provider:                 generated.CreateNvidiaNimConfigProviderNVIDIANIM,
			BaseUrl:                  optfields.PtrIfSet(in.NvidiaNIM.BaseURL),
			ApiKey:                   optfields.PtrIfSet(in.NvidiaNIM.APIKey),
			Headers:                  optfields.PtrMapIfSet(in.NvidiaNIM.Headers),
			ModelNames:               optfields.PtrSliceIfSet(in.NvidiaNIM.ModelNames),
			IsDefaultModelsEnabled:   in.NvidiaNIM.IsDefaultModelsEnabled,
			IsFunctionCallingEnabled: in.NvidiaNIM.FunctionCallingEnabled,
		})
	case in.LiteLLM != nil:
		err = cfg.FromCreateLiteLlmConfig(generated.CreateLiteLlmConfig{
			Provider:                 generated.CreateLiteLlmConfigProviderLITELLM,
			BaseUrl:                  in.LiteLLM.BaseURL,
			ApiKey:                   in.LiteLLM.APIKey,
			Headers:                  optfields.PtrMapIfSet(in.LiteLLM.Headers),
			ModelNames:               optfields.PtrSliceIfSet(in.LiteLLM.ModelNames),
			IsFunctionCallingEnabled: in.LiteLLM.FunctionCallingEnabled,
		})
	case in.Fireworks != nil:
		err = cfg.FromCreateFireworksConfig(generated.CreateFireworksConfig{
			Provider:                 generated.CreateFireworksConfigProviderFIREWORKS,
			ApiKey:                   in.Fireworks.APIKey,
			ModelNames:               optfields.PtrSliceIfSet(in.Fireworks.ModelNames),
			IsDefaultModelsEnabled:   in.Fireworks.IsDefaultModelsEnabled,
			IsFunctionCallingEnabled: in.Fireworks.FunctionCallingEnabled,
		})
	case in.TogetherAi != nil:
		err = cfg.FromCreateTogetherAiConfig(generated.CreateTogetherAiConfig{
			Provider:                 generated.CreateTogetherAiConfigProviderTOGETHERAI,
			ApiKey:                   in.TogetherAi.APIKey,
			ModelNames:               optfields.PtrSliceIfSet(in.TogetherAi.ModelNames),
			IsDefaultModelsEnabled:   in.TogetherAi.IsDefaultModelsEnabled,
			IsFunctionCallingEnabled: in.TogetherAi.FunctionCallingEnabled,
		})
	}
	if err != nil {
		return cfg, fmt.Errorf("integrations: build llm config: %w", err)
	}
	return cfg, nil
}

// buildCreateBedrockAuth maps the discriminated CreateAWSBedrockAuth onto the
// generated wire union, setting the auth_type discriminator. Exactly one mode
// must be set. Shared by create and update (auth replaces wholesale on PATCH).
func buildCreateBedrockAuth(in CreateAWSBedrockAuth) (generated.CreateAwsBedrockAuth, error) {
	var auth generated.CreateAwsBedrockAuth

	set := 0
	for _, present := range []bool{
		in.Default != nil, in.BearerToken != nil, in.ProxyWithHeaders != nil,
	} {
		if present {
			set++
		}
	}
	if set != 1 {
		return auth, ErrInvalidBedrockAuth
	}

	var err error
	switch {
	case in.Default != nil:
		err = auth.FromCreateAwsBedrockDefaultAuth(generated.CreateAwsBedrockDefaultAuth{
			AuthType:   generated.CreateAwsBedrockDefaultAuthAuthTypeDEFAULT,
			RoleArn:    in.Default.RoleARN,
			ExternalId: optfields.PtrIfSet(in.Default.ExternalID),
			BaseUrl:    optfields.PtrIfSet(in.Default.BaseURL),
		})
	case in.BearerToken != nil:
		err = auth.FromCreateAwsBedrockBearerTokenAuth(generated.CreateAwsBedrockBearerTokenAuth{
			AuthType: generated.CreateAwsBedrockBearerTokenAuthAuthTypeBEARERTOKEN,
			ApiKey:   in.BearerToken.APIKey,
			BaseUrl:  optfields.PtrIfSet(in.BearerToken.BaseURL),
		})
	case in.ProxyWithHeaders != nil:
		err = auth.FromCreateAwsBedrockProxyWithHeadersAuth(generated.CreateAwsBedrockProxyWithHeadersAuth{
			AuthType: generated.CreateAwsBedrockProxyWithHeadersAuthAuthTypePROXYWITHHEADERS,
			BaseUrl:  in.ProxyWithHeaders.BaseURL,
			Headers:  optfields.PtrMapIfSet(in.ProxyWithHeaders.Headers),
		})
	}
	if err != nil {
		return auth, fmt.Errorf("integrations: build bedrock auth: %w", err)
	}
	return auth, nil
}

// create posts a prepared union body and returns the created integration.
func (c *Client) create(ctx context.Context, body generated.CreateIntegrationJSONRequestBody) (*Integration, error) {
	resp, err := c.gen.CreateIntegrationWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// UpdateAgent updates an existing agent integration, resolving by name or ID.
// Only non-nil patch fields are sent; omitted fields are left unchanged.
// Returns ErrNoUpdateFields without contacting the server when no patch fields
// are set. Nullable fields (Description, Headers) are cleared on the server
// when set to their empty value — see UpdateAgentRequest.
//
// The body is hand-marshaled as a map so nullable fields can send an explicit
// JSON null to clear: the generated typed body uses *T + omitempty, which
// cannot express the clear-via-null signal (matches the aiintegrations client).
func (c *Client) UpdateAgent(ctx context.Context, req UpdateAgentRequest) (*Integration, error) {
	prerelease.Warn("integrations.update_agent", prerelease.Alpha)

	config := map[string]any{}
	if req.Endpoint != nil {
		config["endpoint"] = *req.Endpoint
	}
	if req.InputSchema != nil {
		config["input_schema"] = *req.InputSchema
	}
	if req.Headers != nil {
		if len(*req.Headers) == 0 {
			config["headers"] = nil
		} else {
			config["headers"] = *req.Headers
		}
	}
	if req.RequestPresets != nil {
		config["request_presets"] = updateAgentPresets(*req.RequestPresets)
	}

	if req.Name == nil && req.Description == nil && req.Scopings == nil && len(config) == 0 {
		return nil, ErrNoUpdateFields
	}
	id, err := resolve.FindIntegrationID(ctx, c.gen, req.Integration, req.Space, IntegrationTypeAgent)
	if err != nil {
		return nil, err
	}

	body := map[string]any{"type": string(IntegrationTypeAgent)}
	if req.Name != nil {
		body["name"] = *req.Name
	}
	if req.Description != nil {
		if *req.Description == "" {
			body["description"] = nil
		} else {
			body["description"] = *req.Description
		}
	}
	if req.Scopings != nil {
		body["scopings"] = *req.Scopings
	}
	if len(config) > 0 {
		body["config"] = config
	}
	return c.update(ctx, id, body)
}

// UpdateLLM updates an existing LLM integration, resolving by name or ID. Only
// non-nil patch fields are sent; omitted fields are left unchanged. Returns
// ErrNoUpdateFields without contacting the server when no patch fields are set.
// Nullable fields (APIKey, BaseURL, Headers) are cleared on the server when set
// to their empty value — see UpdateLLMRequest.
//
// The body is hand-marshaled as a map so nullable fields can send an explicit
// JSON null to clear: the generated typed body uses *T + omitempty, which
// cannot express the clear-via-null signal (matches the aiintegrations client).
func (c *Client) UpdateLLM(ctx context.Context, req UpdateLLMRequest) (*Integration, error) {
	prerelease.Warn("integrations.update_llm", prerelease.Alpha)

	config := map[string]any{}
	if req.APIKey != nil {
		if *req.APIKey == "" {
			config["api_key"] = nil
		} else {
			config["api_key"] = *req.APIKey
		}
	}
	if req.FunctionCallingEnabled != nil {
		config["is_function_calling_enabled"] = *req.FunctionCallingEnabled
	}
	if req.Auth != nil {
		auth, err := buildCreateBedrockAuth(*req.Auth)
		if err != nil {
			return nil, err
		}
		config["auth"] = auth
	}
	if req.BaseURL != nil {
		if *req.BaseURL == "" {
			config["base_url"] = nil
		} else {
			config["base_url"] = *req.BaseURL
		}
	}
	if req.Headers != nil {
		if len(*req.Headers) == 0 {
			config["headers"] = nil
		} else {
			config["headers"] = *req.Headers
		}
	}
	if req.IsDefaultModelsEnabled != nil {
		config["is_default_models_enabled"] = *req.IsDefaultModelsEnabled
	}
	if req.ModelNames != nil {
		config["model_names"] = *req.ModelNames
	}
	if req.ProjectID != nil {
		config["project_id"] = *req.ProjectID
	}
	if req.Location != nil {
		config["location"] = *req.Location
	}
	if req.ProjectAccessLabel != nil {
		config["project_access_label"] = *req.ProjectAccessLabel
	}

	if req.Name == nil && req.Scopings == nil && len(config) == 0 {
		return nil, ErrNoUpdateFields
	}
	id, err := resolve.FindIntegrationID(ctx, c.gen, req.Integration, req.Space, IntegrationTypeLLM)
	if err != nil {
		return nil, err
	}

	body := map[string]any{"type": string(IntegrationTypeLLM)}
	if req.Name != nil {
		body["name"] = *req.Name
	}
	if req.Scopings != nil {
		body["scopings"] = *req.Scopings
	}
	if len(config) > 0 {
		body["config"] = config
	}
	return c.update(ctx, id, body)
}

// update marshals a prepared patch body — a map so nullable fields can send an
// explicit JSON null to clear — and returns the updated integration.
func (c *Client) update(ctx context.Context, id string, body map[string]any) (*Integration, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("integrations: marshal update body: %w", err)
	}
	resp, err := c.gen.UpdateIntegrationWithBodyWithResponse(ctx, id, "application/json", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes an integration, resolving by name or ID. When req.Integration
// is a name, req.Type is required; a name without a type, like an
// unresolvable name, returns arize.ResourceNotFoundError.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("integrations.delete", prerelease.Alpha)
	id, err := resolve.FindIntegrationID(ctx, c.gen, req.Integration, req.Space, req.Type)
	if err != nil {
		return err
	}
	resp, err := c.gen.DeleteIntegrationWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// createAgentPresets maps the hand-written preset inputs to the generated
// create shape.
func createAgentPresets(in []AgentRequestPresetInput) []generated.CreateAgentRequestPresetInput {
	out := make([]generated.CreateAgentRequestPresetInput, len(in))
	for i, p := range in {
		out[i] = generated.CreateAgentRequestPresetInput{
			Name:        p.Name,
			Config:      p.Config,
			Description: optfields.PtrIfSet(p.Description),
		}
	}
	return out
}

// updateAgentPresets maps the hand-written preset inputs to the generated
// update shape.
func updateAgentPresets(in []AgentRequestPresetInput) []generated.UpdateAgentRequestPresetInput {
	out := make([]generated.UpdateAgentRequestPresetInput, len(in))
	for i, p := range in {
		out[i] = generated.UpdateAgentRequestPresetInput{
			Name:        p.Name,
			Config:      p.Config,
			Description: optfields.PtrIfSet(p.Description),
		}
	}
	return out
}
