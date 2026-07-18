package projects

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
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

// List returns a paginated list of projects. When req.Space is a base64
// resource ID, it is sent as the space_id filter (exact match); otherwise it
// is sent as the space_name filter (case-insensitive substring match).
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*ListProjects, error) {
	prerelease.Warn("projects.list", prerelease.Beta)
	params := generated.ListProjectsParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	resp, err := c.gen.ListProjectsWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single project. req.Project accepts a name or ID; req.Space
// is required when req.Project is a name.
func (c *Client) Get(
	ctx context.Context,
	req GetRequest,
) (*Project, error) {
	prerelease.Warn("projects.get", prerelease.Beta)
	id, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetProjectWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new project. req.Space accepts a space name or ID.
func (c *Client) Create(
	ctx context.Context,
	req CreateRequest,
) (*Project, error) {
	prerelease.Warn("projects.create", prerelease.Beta)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.CreateProjectRequest{
		Name:    req.Name,
		SpaceId: spaceID,
	}
	resp, err := c.gen.CreateProjectWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Delete removes a project. req.Project accepts a name or ID; req.Space is
// required when req.Project is a name.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("projects.delete", prerelease.Beta)
	id, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.DeleteProjectWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// Update an existing project.
func (c *Client) Update(
	ctx context.Context,
	req UpdateRequest,
) (*Project, error) {
	prerelease.Warn("projects.update", prerelease.Beta)
	projectID, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.UpdateProjectRequest{
		Name: req.Name,
	}
	resp, err := c.gen.UpdateProjectWithResponse(ctx, projectID, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
