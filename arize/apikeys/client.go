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
) (*APIKeyList, error) {
	prerelease.Warn("apikeys.list", prerelease.Alpha)
	params := &generated.ApiKeysListParams{
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
	resp, err := c.gen.ApiKeysListWithResponse(ctx, params)
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
) (*APIKeyCreated, error) {
	prerelease.Warn("apikeys.create", prerelease.Alpha)
	keyType := APIKeyTypeUser
	body := generated.ApiKeyCreate{
		Name:        req.Name,
		Description: optfields.PtrIfSet(req.Description),
		KeyType:     &keyType,
		ExpiresAt:   optfields.PtrIfSet(req.ExpiresAt),
	}
	resp, err := c.gen.ApiKeysCreateWithResponse(ctx, body)
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
) (*APIKeyCreated, error) {
	prerelease.Warn("apikeys.create_service_key", prerelease.Alpha)
	if req.Space == "" {
		return nil, fmt.Errorf("apikeys: space is required for service keys")
	}
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	keyType := APIKeyTypeService
	body := generated.ApiKeyCreate{
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
	resp, err := c.gen.ApiKeysCreateWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Delete removes an API key by ID.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("apikeys.delete", prerelease.Alpha)
	resp, err := c.gen.ApiKeysDeleteWithResponse(ctx, req.APIKeyID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// Refresh rotates an API key and returns the new key.
func (c *Client) Refresh(
	ctx context.Context,
	req RefreshRequest,
) (*APIKeyCreated, error) {
	prerelease.Warn("apikeys.refresh", prerelease.Alpha)
	body := generated.RefreshApiKeyRequestBody{
		ExpiresAt:          optfields.PtrIfSet(req.ExpiresAt),
		GracePeriodSeconds: optfields.PtrIfSet(req.GracePeriodSeconds),
	}
	resp, err := c.gen.ApiKeysRefreshWithResponse(ctx, req.APIKeyID, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
