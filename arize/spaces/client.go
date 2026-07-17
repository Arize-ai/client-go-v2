package spaces

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Spaces API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of spaces. req.Organization, when non-empty,
// accepts an organization name or ID and restricts results to that
// organization.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*SpaceList, error) {
	prerelease.Warn("spaces.list", prerelease.Beta)
	var orgID string
	if req.Organization != "" {
		resolved, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
		if err != nil {
			return nil, err
		}
		orgID = resolved
	}
	params := generated.ListSpacesParams{
		OrgId:  optfields.PtrIfSet(orgID),
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListSpacesWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single space. req.Space accepts a name or ID.
func (c *Client) Get(
	ctx context.Context,
	req GetRequest,
) (*Space, error) {
	prerelease.Warn("spaces.get", prerelease.Beta)
	id, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetSpaceWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new space and returns it. req.Organization accepts a name
// or ID and identifies the parent organization.
func (c *Client) Create(
	ctx context.Context,
	req CreateRequest,
) (*Space, error) {
	prerelease.Warn("spaces.create", prerelease.Beta)
	orgID, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
	if err != nil {
		return nil, err
	}
	body := generated.CreateSpaceJSONRequestBody{
		Name:           req.Name,
		OrganizationId: orgID,
		Description:    optfields.PtrIfSet(req.Description),
	}
	resp, err := c.gen.CreateSpaceWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing space. req.Space accepts a name or ID. Leave a
// patch field nil to preserve its current value.
func (c *Client) Update(
	ctx context.Context,
	req UpdateRequest,
) (*Space, error) {
	prerelease.Warn("spaces.update", prerelease.Beta)
	id, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.UpdateSpaceJSONRequestBody{
		Name:        req.Name,
		Description: req.Description,
	}
	resp, err := c.gen.UpdateSpaceWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete irreversibly removes a space and cascades to all child resources
// (projects, datasets, monitors, custom metrics, etc.). req.Space accepts a
// name or ID.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("spaces.delete", prerelease.Beta)
	id, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.DeleteSpaceWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// AddUser adds a user to a space, or upserts their role if they are already a
// member. The user must already belong to the space's parent organization;
// auto-enrollment is not performed.
func (c *Client) AddUser(
	ctx context.Context,
	req AddUserRequest,
) (*SpaceMembership, error) {
	prerelease.Warn("spaces.add_user", prerelease.Beta)
	id, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AddSpaceUserJSONRequestBody{
		UserId: req.UserID,
		Role:   req.Role,
	}
	resp, err := c.gen.AddSpaceUserWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// RemoveUser removes a user from a space. Removes both the legacy
// SpaceMembers row and any RBAC role bindings for the user on this space.
func (c *Client) RemoveUser(
	ctx context.Context,
	req RemoveUserRequest,
) error {
	prerelease.Warn("spaces.remove_user", prerelease.Beta)
	id, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.RemoveSpaceUserWithResponse(ctx, id, req.UserID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
