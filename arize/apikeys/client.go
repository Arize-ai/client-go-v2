package apikeys

import (
	"context"
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize API Keys API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of API keys. Defaults to a page size of 50.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*ListAPIKeys, error) {
	prerelease.Warn("apikeys.list", prerelease.Beta)
	params := &generated.ListApiKeysParams{
		KeyType: optfields.PtrIfSet(req.KeyType),
		Status:  optfields.PtrIfSet(req.Status),
		UserId:  optfields.PtrIfSet(req.UserID),
		Limit:   optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor:  optfields.PtrIfSet(req.Cursor),
	}
	if req.Space != "" {
		spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
		if err != nil {
			return nil, err
		}
		params.SpaceId = &spaceID
	}
	resp, err := c.gen.ListApiKeysWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new user API key and returns it.
func (c *Client) Create(
	ctx context.Context,
	req CreateRequest,
) (*APIKey, error) {
	prerelease.Warn("apikeys.create", prerelease.Beta)
	keyType := APIKeyTypeUser
	body := generated.CreateApiKeyRequest{
		Name:        req.Name,
		Description: optfields.PtrIfSet(req.Description),
		KeyType:     &keyType,
		ExpiresAt:   optfields.PtrIfSet(req.ExpiresAt),
	}
	resp, err := c.gen.CreateApiKeyWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// CreateServiceKey creates a new service API key and returns it.
func (c *Client) CreateServiceKey(
	ctx context.Context,
	req CreateServiceKeyRequest,
) (*APIKey, error) {
	prerelease.Warn("apikeys.create_service_key", prerelease.Beta)
	if req.Space == "" {
		return nil, fmt.Errorf("apikeys: space is required for service keys")
	}
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	keyType := APIKeyTypeService
	body := generated.CreateApiKeyRequest{
		Name:        req.Name,
		Description: optfields.PtrIfSet(req.Description),
		KeyType:     &keyType,
		ExpiresAt:   optfields.PtrIfSet(req.ExpiresAt),
		SpaceId:     &spaceID,
	}
	if req.SpaceRole != "" || req.OrgRole != "" || req.AccountRole != "" {
		body.Roles = &generated.ApiKeyRoles{
			SpaceRole:   optfields.PtrIfSet(req.SpaceRole),
			OrgRole:     optfields.PtrIfSet(req.OrgRole),
			AccountRole: optfields.PtrIfSet(req.AccountRole),
		}
	}
	resp, err := c.gen.CreateApiKeyWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Revoke sets an API key's status to revoked by ID. The key stops working
// immediately; revoking an already-revoked key is a no-op and still succeeds.
func (c *Client) Revoke(
	ctx context.Context,
	req RevokeRequest,
) error {
	prerelease.Warn("apikeys.revoke", prerelease.Beta)
	resp, err := c.gen.RevokeApiKeyWithResponse(ctx, req.APIKeyID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// Refresh rotates an API key and returns the new key.
func (c *Client) Refresh(
	ctx context.Context,
	req RefreshRequest,
) (*APIKey, error) {
	prerelease.Warn("apikeys.refresh", prerelease.Beta)
	body := generated.RefreshApiKeyRequest{
		ExpiresAt:          optfields.PtrIfSet(req.ExpiresAt),
		GracePeriodSeconds: optfields.PtrIfSet(req.GracePeriodSeconds),
	}
	resp, err := c.gen.RefreshApiKeyWithResponse(ctx, req.APIKeyID, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
