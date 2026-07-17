package roles

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Roles API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of roles. Defaults to a page size of 100.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*RoleList, error) {
	prerelease.Warn("roles.list", prerelease.Beta)
	params := &generated.ListRolesParams{
		IsPredefined: req.IsPredefined,
		Limit:        optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor:       optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListRolesWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single role. req.Role accepts a name or ID.
func (c *Client) Get(
	ctx context.Context,
	req GetRequest,
) (*Role, error) {
	prerelease.Warn("roles.get", prerelease.Beta)
	id, err := resolve.FindRoleID(ctx, c.gen, req.Role)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetRoleWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new role.
func (c *Client) Create(
	ctx context.Context,
	req CreateRequest,
) (*Role, error) {
	prerelease.Warn("roles.create", prerelease.Beta)
	body := generated.CreateRoleRequest{
		Name:        req.Name,
		Permissions: req.Permissions,
		Description: optfields.PtrIfSet(req.Description),
	}
	resp, err := c.gen.CreateRoleWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update patches an existing role. req.Role accepts a name or ID. Leave a
// patch field nil to preserve its current value.
func (c *Client) Update(
	ctx context.Context,
	req UpdateRequest,
) (*Role, error) {
	prerelease.Warn("roles.update", prerelease.Beta)
	id, err := resolve.FindRoleID(ctx, c.gen, req.Role)
	if err != nil {
		return nil, err
	}
	body := generated.UpdateRoleRequest{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	}
	resp, err := c.gen.UpdateRoleWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes a role. req.Role accepts a name or ID. Predefined (system)
// roles cannot be deleted.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("roles.delete", prerelease.Beta)
	id, err := resolve.FindRoleID(ctx, c.gen, req.Role)
	if err != nil {
		return err
	}
	resp, err := c.gen.DeleteRoleWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
