package aiintegrations

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

// Client provides access to the Arize AI Integrations API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of AI integrations. Defaults to a page size of 50.
func (c *Client) List(ctx context.Context, req ListRequest) (*AIIntegrationList, error) {
	prerelease.Warn("aiintegrations.list", prerelease.Alpha)
	params := &generated.AiIntegrationsListParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	resp, err := c.gen.AiIntegrationsListWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single AI integration, resolving by name or ID.
func (c *Client) Get(ctx context.Context, req GetRequest) (*AIIntegration, error) {
	prerelease.Warn("aiintegrations.get", prerelease.Alpha)
	id, err := resolve.FindAIIntegrationID(ctx, c.gen, req.Integration, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AiIntegrationsGetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new AI integration and returns it.
//
// The body is hand-marshaled so that ProviderMetadata (a hand-written
// discriminator wrapper) can be serialized via its own MarshalJSON.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*AIIntegration, error) {
	prerelease.Warn("aiintegrations.create", prerelease.Alpha)
	body := map[string]any{
		"name":     req.Name,
		"provider": req.Provider,
	}
	if req.APIKey != "" {
		body["api_key"] = req.APIKey
	}
	if req.BaseURL != "" {
		body["base_url"] = req.BaseURL
	}
	if req.AuthType != "" {
		body["auth_type"] = req.AuthType
	}
	// Always send so the SDK doesn't silently inherit a future change to the
	// server-side default.
	body["enable_default_models"] = req.EnableDefaultModels
	if req.DisableFunctionCalling {
		// Inverted: caller asked to disable; send function_calling_enabled=false.
		body["function_calling_enabled"] = false
	}
	if req.ModelNames != nil {
		body["model_names"] = req.ModelNames
	}
	if req.Headers != nil {
		body["headers"] = req.Headers
	}
	if req.Scopings != nil {
		body["scopings"] = req.Scopings
	}
	if req.ProviderMetadata != nil {
		body["provider_metadata"] = req.ProviderMetadata
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("aiintegrations: marshal create body: %w", err)
	}
	resp, err := c.gen.AiIntegrationsCreateWithBodyWithResponse(ctx, "application/json", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing AI integration, resolving by name or ID. Only
// non-nil patch fields are sent on the wire; omitted fields are left
// unchanged on the server. Returns ErrNoUpdateFields without contacting the
// server when no patch fields are set.
//
// For the four nullable fields (APIKey, BaseURL, Headers, ProviderMetadata),
// a non-nil pointer to the zero value is sent on the wire as JSON null so the
// server clears the field — matching the OpenAPI "Pass null to remove"
// contract. The body is hand-marshaled here because the generated typed body
// uses *T + omitempty, which cannot express the clear-via-null signal.
func (c *Client) Update(ctx context.Context, req UpdateRequest) (*AIIntegration, error) {
	prerelease.Warn("aiintegrations.update", prerelease.Alpha)
	if req.Name == nil && req.Provider == nil && req.APIKey == nil &&
		req.BaseURL == nil && req.ModelNames == nil && req.Headers == nil &&
		req.EnableDefaultModels == nil && req.FunctionCallingEnabled == nil &&
		req.AuthType == nil && req.ProviderMetadata == nil && req.Scopings == nil {
		return nil, ErrNoUpdateFields
	}
	id, err := resolve.FindAIIntegrationID(ctx, c.gen, req.Integration, req.Space)
	if err != nil {
		return nil, err
	}
	body := make(map[string]any)
	if req.Name != nil {
		body["name"] = *req.Name
	}
	if req.Provider != nil {
		body["provider"] = *req.Provider
	}
	if req.APIKey != nil {
		if *req.APIKey == "" {
			body["api_key"] = nil
		} else {
			body["api_key"] = *req.APIKey
		}
	}
	if req.BaseURL != nil {
		if *req.BaseURL == "" {
			body["base_url"] = nil
		} else {
			body["base_url"] = *req.BaseURL
		}
	}
	if req.ModelNames != nil {
		body["model_names"] = *req.ModelNames
	}
	if req.Headers != nil {
		if len(*req.Headers) == 0 {
			body["headers"] = nil
		} else {
			body["headers"] = *req.Headers
		}
	}
	if req.EnableDefaultModels != nil {
		body["enable_default_models"] = *req.EnableDefaultModels
	}
	if req.FunctionCallingEnabled != nil {
		body["function_calling_enabled"] = *req.FunctionCallingEnabled
	}
	if req.AuthType != nil {
		body["auth_type"] = *req.AuthType
	}
	if req.ProviderMetadata != nil {
		body["provider_metadata"] = req.ProviderMetadata
	}
	if req.Scopings != nil {
		body["scopings"] = *req.Scopings
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("aiintegrations: marshal update body: %w", err)
	}
	resp, err := c.gen.AiIntegrationsUpdateWithBodyWithResponse(ctx, id, "application/json", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes an AI integration, resolving by name or ID.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("aiintegrations.delete", prerelease.Alpha)
	id, err := resolve.FindAIIntegrationID(ctx, c.gen, req.Integration, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.AiIntegrationsDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
