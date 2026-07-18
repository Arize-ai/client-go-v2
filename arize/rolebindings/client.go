package rolebindings

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
)

// Client provides access to the Arize Role Bindings API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of role bindings for the authenticated user's
// account, filtered by resource type. Defaults to a page size of 50.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*ListRoleBindings, error) {
	prerelease.Warn("rolebindings.list", prerelease.Beta)
	params := &generated.ListRoleBindingsParams{
		ResourceType: req.ResourceType,
		UserId:       optfields.PtrIfSet(req.UserID),
		Limit:        optfields.PtrWithDefault(req.Limit, 50),
		Cursor:       optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListRoleBindingsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single role binding by ID.
func (c *Client) Get(
	ctx context.Context,
	req GetRequest,
) (*RoleBinding, error) {
	prerelease.Warn("rolebindings.get", prerelease.Beta)
	resp, err := c.gen.GetRoleBindingWithResponse(ctx, req.RoleBindingID)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new role binding and returns it.
func (c *Client) Create(
	ctx context.Context,
	req CreateRequest,
) (*RoleBinding, error) {
	prerelease.Warn("rolebindings.create", prerelease.Beta)
	body := generated.CreateRoleBindingRequest{
		ResourceId:   req.ResourceID,
		ResourceType: req.ResourceType,
		RoleId:       req.RoleID,
		UserId:       req.UserID,
	}
	resp, err := c.gen.CreateRoleBindingWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing role binding and returns the updated binding.
func (c *Client) Update(
	ctx context.Context,
	req UpdateRequest,
) (*RoleBinding, error) {
	prerelease.Warn("rolebindings.update", prerelease.Beta)
	body := generated.UpdateRoleBindingRequest{
		RoleId: req.RoleID,
	}
	resp, err := c.gen.UpdateRoleBindingWithResponse(ctx, req.RoleBindingID, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes a role binding by ID.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("rolebindings.delete", prerelease.Beta)
	resp, err := c.gen.DeleteRoleBindingWithResponse(ctx, req.RoleBindingID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
