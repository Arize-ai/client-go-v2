package rolebindings

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
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

// Get returns a single role binding by ID.
func (c *Client) Get(ctx context.Context, bindingID string) (*RoleBindingResponse, error) {
	prerelease.Warn("rolebindings.get", prerelease.Alpha)
	resp, err := c.gen.RoleBindingsGetWithResponse(ctx, bindingID)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new role binding and returns it.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*RoleBindingResponse, error) {
	prerelease.Warn("rolebindings.create", prerelease.Alpha)
	resp, err := c.gen.RoleBindingsCreateWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing role binding and returns the updated binding.
func (c *Client) Update(ctx context.Context, bindingID string, req UpdateRequest) (*RoleBindingResponse, error) {
	prerelease.Warn("rolebindings.update", prerelease.Alpha)
	resp, err := c.gen.RoleBindingsUpdateWithResponse(ctx, bindingID, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes a role binding by ID.
func (c *Client) Delete(ctx context.Context, bindingID string) error {
	prerelease.Warn("rolebindings.delete", prerelease.Alpha)
	resp, err := c.gen.RoleBindingsDeleteWithResponse(ctx, bindingID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
