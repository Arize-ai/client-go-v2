package prompts

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Prompts API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of prompts. When req.Space is a base64
// resource ID, it is sent as the space_id filter (exact match); otherwise it
// is sent as the space_name filter (case-insensitive substring match).
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*PromptList, error) {
	prerelease.Warn("prompts.list", prerelease.Beta)
	params := generated.ListPromptsParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	resp, err := c.gen.ListPromptsWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single prompt, resolving by name or ID. If req.VersionID or
// req.Label is set, returns that specific version; otherwise returns the latest
// version. VersionID and Label are mutually exclusive.
func (c *Client) Get(
	ctx context.Context,
	req GetRequest,
) (*PromptWithVersion, error) {
	prerelease.Warn("prompts.get", prerelease.Beta)
	id, err := resolve.FindPromptID(ctx, c.gen, req.Prompt, req.Space)
	if err != nil {
		return nil, err
	}
	params := generated.GetPromptParams{
		VersionId: optfields.PtrIfSet(req.VersionID),
		Label:     optfields.PtrIfSet(req.Label),
	}
	resp, err := c.gen.GetPromptWithResponse(ctx, id, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new prompt (with its initial version), resolving the parent
// space by name or ID.
func (c *Client) Create(
	ctx context.Context,
	req CreateRequest,
) (*PromptWithVersion, error) {
	prerelease.Warn("prompts.create", prerelease.Beta)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.CreatePromptJSONRequestBody{
		Name:        req.Name,
		Description: optfields.PtrIfSet(req.Description),
		SpaceId:     spaceID,
		Version:     req.Version,
	}
	resp, err := c.gen.CreatePromptWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing prompt's metadata, resolving by name or ID. Leave
// req.Description nil to preserve its current value.
func (c *Client) Update(
	ctx context.Context,
	req UpdateRequest,
) (*Prompt, error) {
	prerelease.Warn("prompts.update", prerelease.Beta)
	id, err := resolve.FindPromptID(ctx, c.gen, req.Prompt, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.UpdatePromptJSONRequestBody{
		Description: req.Description,
	}
	resp, err := c.gen.UpdatePromptWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes a prompt, resolving by name or ID.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("prompts.delete", prerelease.Beta)
	id, err := resolve.FindPromptID(ctx, c.gen, req.Prompt, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.DeletePromptWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ListVersions returns a paginated list of versions for a prompt, resolving the
// prompt by name or ID.
func (c *Client) ListVersions(
	ctx context.Context,
	req ListVersionsRequest,
) (*PromptVersionList, error) {
	prerelease.Warn("prompts.list_versions", prerelease.Beta)
	id, err := resolve.FindPromptID(ctx, c.gen, req.Prompt, req.Space)
	if err != nil {
		return nil, err
	}
	params := generated.ListPromptVersionsParams{
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListPromptVersionsWithResponse(ctx, id, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// CreateVersion creates a new version for an existing prompt, resolving the
// prompt by name or ID.
func (c *Client) CreateVersion(
	ctx context.Context,
	req CreateVersionRequest,
) (*PromptVersion, error) {
	prerelease.Warn("prompts.create_version", prerelease.Beta)
	id, err := resolve.FindPromptID(ctx, c.gen, req.Prompt, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.CreatePromptVersionJSONRequestBody{
		CommitMessage:       req.CommitMessage,
		Provider:            req.Provider,
		Messages:            req.Messages,
		InputVariableFormat: optfields.PtrIfSet(req.InputVariableFormat),
		Model:               optfields.PtrIfSet(req.Model),
		InvocationParams:    req.InvocationParams,
		ProviderParams:      req.ProviderParams,
	}
	resp, err := c.gen.CreatePromptVersionWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// GetVersion returns a single prompt version by its ID. Version IDs are pure
// IDs with no name resolution.
func (c *Client) GetVersion(
	ctx context.Context,
	req GetVersionRequest,
) (*PromptVersion, error) {
	prerelease.Warn("prompts.get_version", prerelease.Beta)
	resp, err := c.gen.GetPromptVersionWithResponse(ctx, req.VersionID)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetVersionByLabel returns the prompt version pointed to by a label (e.g.
// "production"), resolving the prompt by name or ID.
func (c *Client) GetVersionByLabel(
	ctx context.Context,
	req GetVersionByLabelRequest,
) (*PromptVersion, error) {
	prerelease.Warn("prompts.get_version_by_label", prerelease.Beta)
	id, err := resolve.FindPromptID(ctx, c.gen, req.Prompt, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetPromptLabelWithResponse(ctx, id, req.LabelName)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SetVersionLabels assigns one or more labels to a specific prompt version,
// replacing all existing labels, and returns the updated prompt version.
// Version IDs are pure IDs with no name resolution.
func (c *Client) SetVersionLabels(
	ctx context.Context,
	req SetVersionLabelsRequest,
) (*PromptVersion, error) {
	prerelease.Warn("prompts.set_version_labels", prerelease.Beta)
	body := generated.SetPromptVersionLabelJSONRequestBody{
		Labels: req.Labels,
	}
	resp, err := c.gen.SetPromptVersionLabelWithResponse(ctx, req.VersionID, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// DeleteVersionLabel removes a label from a specific prompt version. Version
// IDs are pure IDs with no name resolution.
func (c *Client) DeleteVersionLabel(
	ctx context.Context,
	req DeleteVersionLabelRequest,
) error {
	prerelease.Warn("prompts.delete_version_label", prerelease.Beta)
	resp, err := c.gen.DeletePromptVersionLabelWithResponse(ctx, req.VersionID, req.LabelName)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
