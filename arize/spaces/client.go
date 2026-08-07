package spaces

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
	"github.com/Arize-ai/client-go-v2/arize/internal/roleconv"
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
) (*ListSpaces, error) {
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
	if req.IsPrivate {
		slog.Warn("spaces.create: private spaces restrict visibility to space members and admins; ensure members are added before the space becomes inaccessible to other users")
	}
	orgID, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
	if err != nil {
		return nil, err
	}
	body := generated.CreateSpaceJSONRequestBody{
		Name:           req.Name,
		OrganizationId: orgID,
		Description:    optfields.PtrIfSet(req.Description),
		IsPrivate:      optfields.PtrIfSet(req.IsPrivate),
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
	if req.IsPrivate != nil && *req.IsPrivate {
		slog.Warn("spaces.update: private spaces restrict visibility to space members and admins; ensure members are added before the space becomes inaccessible to other users")
	}
	id, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := map[string]any{}
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
	if req.IsPrivate != nil {
		body["is_private"] = *req.IsPrivate
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("spaces: marshal update body: %w", err)
	}
	resp, err := c.gen.UpdateSpaceWithBodyWithResponse(ctx, id, "application/json", bytes.NewReader(raw))
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
	if roleconv.IsZero(req.Role) {
		return nil, fmt.Errorf("spaces: Role is required; use AssignPredefinedRole or AssignCustomRole")
	}
	role, err := roleconv.SpaceRoleAssignmentRequest(req.Role)
	if err != nil {
		return nil, fmt.Errorf("spaces: build role request: %w", err)
	}
	body := generated.AddSpaceUserJSONRequestBody{
		UserId: req.UserID,
		Role:   role,
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
