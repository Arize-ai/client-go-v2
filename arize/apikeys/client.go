package apikeys

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
)

// Client provides access to the Arize API Keys API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of API keys.
func (c *Client) List(ctx context.Context, params ListParams) (*ApiKeyList, error) {
	prerelease.Warn("apikeys.list", prerelease.Alpha)
	resp, err := c.gen.ApiKeysListWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new API key and returns it.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*ApiKeyCreated, error) {
	prerelease.Warn("apikeys.create", prerelease.Alpha)
	resp, err := c.gen.ApiKeysCreateWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Delete removes an API key by ID.
func (c *Client) Delete(ctx context.Context, apiKeyID string) error {
	prerelease.Warn("apikeys.delete", prerelease.Alpha)
	resp, err := c.gen.ApiKeysDeleteWithResponse(ctx, apiKeyID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// Refresh rotates an API key and returns the new key.
func (c *Client) Refresh(ctx context.Context, apiKeyID string, req RefreshRequest) (*ApiKeyCreated, error) {
	prerelease.Warn("apikeys.refresh", prerelease.Alpha)
	resp, err := c.gen.ApiKeysRefreshWithResponse(ctx, apiKeyID, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
