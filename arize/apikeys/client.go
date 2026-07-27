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
) (*ListApiKeysResponse, error) {
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
) (*UserApiKeyCreated, error) {
	prerelease.Warn("apikeys.create", prerelease.Beta)
	var body generated.CreateApiKeyRequest
	if err := body.FromCreateUserApiKeyRequest(generated.CreateUserApiKeyRequest{
		KeyType:     generated.CreateUserApiKeyRequestKeyTypeUSER,
		Name:        req.Name,
		Description: optfields.PtrIfSet(req.Description),
		ExpiresAt:   optfields.PtrIfSet(req.ExpiresAt),
	}); err != nil {
		return nil, fmt.Errorf("apikeys: build request body: %w", err)
	}
	resp, err := c.gen.CreateApiKeyWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	created, err := resp.JSON201.AsUserApiKeyCreated()
	if err != nil {
		return nil, fmt.Errorf("apikeys: decode user key response: %w", err)
	}
	return &created, nil
}

// CreateServiceKey creates a new service API key and returns it.
func (c *Client) CreateServiceKey(
	ctx context.Context,
	req CreateServiceKeyRequest,
) (*ServiceApiKeyCreated, error) {
	prerelease.Warn("apikeys.create_service_key", prerelease.Beta)
	if len(req.Orgs) == 0 {
		return nil, fmt.Errorf("apikeys: at least one organization binding with at least one space is required")
	}

	orgBindings := make([]generated.ServiceKeyOrgAssignment, 0, len(req.Orgs))
	for j, orgBinding := range req.Orgs {
		if orgBinding.OrgID == "" {
			return nil, fmt.Errorf("apikeys: orgs[%d]: OrgID is required", j)
		}
		if len(orgBinding.Spaces) == 0 {
			return nil, fmt.Errorf("apikeys: orgs[%d]: at least one space binding is required", j)
		}

		var orgRole *generated.OrganizationRoleAssignment
		if !isZeroRole(orgBinding.Role) {
			r := orgBinding.Role
			orgRole = &r
		}

		// Translate SDK space bindings because callers may pass a space name,
		// while the API request body requires the resolved space ID.
		spaceBindings := make([]generated.ServiceKeySpaceAssignment, 0, len(orgBinding.Spaces))
		for k, sb := range orgBinding.Spaces {
			if sb.Space == "" {
				return nil, fmt.Errorf("apikeys: orgs[%d].spaces[%d]: space is required", j, k)
			}
			spaceID, err := resolve.FindSpaceID(ctx, c.gen, sb.Space)
			if err != nil {
				return nil, err
			}

			var spaceRole *generated.SpaceRoleAssignment
			if !isZeroRole(sb.Role) {
				r := sb.Role
				spaceRole = &r
			}

			spaceBindings = append(spaceBindings, generated.ServiceKeySpaceAssignment{
				SpaceId: spaceID,
				Role:    spaceRole,
			})
		}

		orgBindings = append(orgBindings, generated.ServiceKeyOrgAssignment{
			OrgId:  orgBinding.OrgID,
			Role:   orgRole,
			Spaces: spaceBindings,
		})
	}

	// Build the optional account-level role; nil when zero (server applies default).
	var accountRole *generated.UserRoleAssignment
	if !isZeroRole(req.AccountRole) {
		r := req.AccountRole
		accountRole = &r
	}

	svc := generated.CreateServiceApiKeyRequest{
		KeyType:       generated.CreateServiceApiKeyRequestKeyTypeSERVICE,
		Name:          req.Name,
		Description:   optfields.PtrIfSet(req.Description),
		ExpiresAt:     optfields.PtrIfSet(req.ExpiresAt),
		AccountRole:   accountRole,
		Organizations: orgBindings,
	}

	var body generated.CreateApiKeyRequest
	if err := body.FromCreateServiceApiKeyRequest(svc); err != nil {
		return nil, fmt.Errorf("apikeys: build request body: %w", err)
	}

	resp, err := c.gen.CreateApiKeyWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	created, err := resp.JSON201.AsServiceApiKeyCreated()
	if err != nil {
		return nil, fmt.Errorf("apikeys: decode service key response: %w", err)
	}
	return &created, nil
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

// Refresh rotates an API key and returns the replacement in the legacy refresh shape.
func (c *Client) Refresh(
	ctx context.Context,
	req RefreshRequest,
) (*RefreshApiKeyResponse, error) {
	prerelease.Warn("apikeys.refresh", prerelease.Beta)
	body := generated.RefreshApiKeyRequestBody{
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

// isZeroRole reports whether a union role value is unset (zero). The union types
// store their data as json.RawMessage; when nil the MarshalJSON returns "null".
func isZeroRole(r interface{ MarshalJSON() ([]byte, error) }) bool {
	b, err := r.MarshalJSON()
	return err == nil && string(b) == "null"
}
