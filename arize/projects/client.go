package projects

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Projects API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of projects.
func (c *Client) List(ctx context.Context, params ListParams) (*ProjectList, error) {
	prerelease.Warn("projects.list", prerelease.Alpha)
	resp, err := c.gen.ProjectsListWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single project, resolving by name or ID.
// space is required when project is a name rather than an ID.
func (c *Client) Get(ctx context.Context, project, space string) (*Project, error) {
	prerelease.Warn("projects.get", prerelease.Alpha)
	id, err := resolve.FindProjectID(ctx, c.gen, project, space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.ProjectsGetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new project, resolving the parent space by name or ID.
func (c *Client) Create(ctx context.Context, space, name string) (*Project, error) {
	prerelease.Warn("projects.create", prerelease.Alpha)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.ProjectsCreateWithResponse(ctx, CreateRequest{
		Name:    name,
		SpaceId: spaceID,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Delete removes a project, resolving by name or ID.
// space is required when project is a name rather than an ID.
func (c *Client) Delete(ctx context.Context, project, space string) error {
	prerelease.Warn("projects.delete", prerelease.Alpha)
	id, err := resolve.FindProjectID(ctx, c.gen, project, space)
	if err != nil {
		return err
	}
	resp, err := c.gen.ProjectsDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
