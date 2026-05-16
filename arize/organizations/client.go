package organizations

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
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
func (c *Client) List(ctx context.Context, params ListParams) (*OrganizationList, error) {
	prerelease.Warn("organizations.list", prerelease.Alpha)
	resp, err := c.gen.OrganizationsListWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single organization, resolving by name or ID.
func (c *Client) Get(ctx context.Context, nameOrID string) (*Organization, error) {
	prerelease.Warn("organizations.get", prerelease.Alpha)
	id, err := resolve.FindOrganizationID(ctx, c.gen, nameOrID)
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
func (c *Client) Create(ctx context.Context, req CreateRequest) (*Organization, error) {
	prerelease.Warn("organizations.create", prerelease.Alpha)
	resp, err := c.gen.OrganizationsCreateWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing organization, resolving by name or ID.
func (c *Client) Update(ctx context.Context, nameOrID string, req UpdateRequest) (*Organization, error) {
	prerelease.Warn("organizations.update", prerelease.Alpha)
	id, err := resolve.FindOrganizationID(ctx, c.gen, nameOrID)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.OrganizationsUpdateWithResponse(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
