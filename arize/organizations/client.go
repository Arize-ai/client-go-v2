package organizations

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Organizations API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of organizations.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*OrganizationList, error) {
	prerelease.Warn("organizations.list", prerelease.Beta)
	params := generated.OrganizationsListParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.OrganizationsListWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single organization. req.Organization accepts a name or ID.
func (c *Client) Get(
	ctx context.Context,
	req GetRequest,
) (*Organization, error) {
	prerelease.Warn("organizations.get", prerelease.Beta)
	id, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.OrganizationsGetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new organization and returns it.
func (c *Client) Create(
	ctx context.Context,
	req CreateRequest,
) (*Organization, error) {
	prerelease.Warn("organizations.create", prerelease.Beta)
	body := generated.OrganizationCreate{
		Name:        req.Name,
		Description: optfields.PtrIfSet(req.Description),
	}
	resp, err := c.gen.OrganizationsCreateWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing organization. req.Organization accepts a name or
// ID. Leave a patch field nil to preserve its current value.
func (c *Client) Update(
	ctx context.Context,
	req UpdateRequest,
) (*Organization, error) {
	prerelease.Warn("organizations.update", prerelease.Beta)
	id, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
	if err != nil {
		return nil, err
	}
	body := generated.OrganizationUpdate{
		Name:        req.Name,
		Description: req.Description,
	}
	resp, err := c.gen.OrganizationsUpdateWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete irreversibly removes an organization and cascades to all child
// resources (spaces, projects, API keys, datasets, monitors, etc.).
// req.Organization accepts a name or ID.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("organizations.delete", prerelease.Beta)
	id, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
	if err != nil {
		return err
	}
	resp, err := c.gen.OrganizationsDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// AddUser adds a user to an organization, or upserts their role if they are
// already a member. Custom role assignments are not yet supported for
// organizations; pass a PredefinedOrgRole built from one of the
// OrganizationRole* constants.
func (c *Client) AddUser(
	ctx context.Context,
	req AddUserRequest,
) (*OrganizationMembership, error) {
	prerelease.Warn("organizations.add_user", prerelease.Alpha)
	id, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
	if err != nil {
		return nil, err
	}
	role := req.Role
	role.Type = RoleAssignmentTypePredefined
	var assignment generated.OrganizationRoleAssignment
	if err := assignment.FromOrganizationPredefinedRoleAssignment(role); err != nil {
		return nil, err
	}
	body := generated.OrganizationMembershipInput{
		UserId: req.UserID,
		Role:   assignment,
	}
	resp, err := c.gen.OrganizationsAddUserWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// RemoveUser removes a user from an organization. Membership cascades to all
// child spaces.
func (c *Client) RemoveUser(
	ctx context.Context,
	req RemoveUserRequest,
) error {
	prerelease.Warn("organizations.remove_user", prerelease.Alpha)
	id, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
	if err != nil {
		return err
	}
	resp, err := c.gen.OrganizationsRemoveUserWithResponse(ctx, id, req.UserID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
