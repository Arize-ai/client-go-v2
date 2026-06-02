package evaluators

import (
	"context"
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Evaluators API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of evaluators. Defaults to a page size of 50.
func (c *Client) List(ctx context.Context, req ListRequest) (*EvaluatorList, error) {
	prerelease.Warn("evaluators.list", prerelease.Alpha)
	params := &generated.EvaluatorsListParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	resp, err := c.gen.EvaluatorsListWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single evaluator, resolving by name or ID. When
// req.VersionID is set it returns that specific version; otherwise it returns
// the latest version.
func (c *Client) Get(ctx context.Context, req GetRequest) (*EvaluatorWithVersion, error) {
	prerelease.Warn("evaluators.get", prerelease.Alpha)
	id, err := resolve.FindEvaluatorID(ctx, c.gen, req.Evaluator, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.EvaluatorsGetWithResponse(ctx, id, &generated.EvaluatorsGetParams{
		VersionId: optfields.PtrIfSet(req.VersionID),
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new evaluator with its initial version, resolving the
// parent space by name or ID. The evaluator's type is derived from
// req.Version (Template -> template, Code -> code).
func (c *Client) Create(ctx context.Context, req CreateRequest) (*EvaluatorWithVersion, error) {
	prerelease.Warn("evaluators.create", prerelease.Alpha)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	version, evalType, err := buildVersionCreate(req.Version)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.EvaluatorsCreateWithResponse(ctx, generated.EvaluatorsCreateJSONRequestBody{
		Name:        req.Name,
		Description: optfields.PtrIfSet(req.Description),
		SpaceId:     spaceID,
		Type:        evalType,
		Version:     version,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing evaluator's metadata, resolving by name or ID.
// Only non-nil patch fields are sent; nil fields are left unchanged. Returns
// ErrNoUpdateFields without contacting the server when no patch field is set.
func (c *Client) Update(ctx context.Context, req UpdateRequest) (*Evaluator, error) {
	prerelease.Warn("evaluators.update", prerelease.Alpha)
	if req.Name == nil && req.Description == nil {
		return nil, ErrNoUpdateFields
	}
	id, err := resolve.FindEvaluatorID(ctx, c.gen, req.Evaluator, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.EvaluatorsUpdateWithResponse(ctx, id, generated.EvaluatorsUpdateJSONRequestBody{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes an evaluator, resolving by name or ID.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("evaluators.delete", prerelease.Alpha)
	id, err := resolve.FindEvaluatorID(ctx, c.gen, req.Evaluator, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.EvaluatorsDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ListVersions returns a paginated list of versions for an evaluator,
// resolving the evaluator by name or ID. Defaults to a page size of 50.
func (c *Client) ListVersions(ctx context.Context, req ListVersionsRequest) (*EvaluatorVersionList, error) {
	prerelease.Warn("evaluators.list_versions", prerelease.Alpha)
	id, err := resolve.FindEvaluatorID(ctx, c.gen, req.Evaluator, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.EvaluatorVersionsListWithResponse(ctx, id, &generated.EvaluatorVersionsListParams{
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// CreateVersion appends a new version to an existing evaluator, resolving the
// evaluator by name or ID. The version's kind (set via req.Version) must match
// the parent evaluator's type.
func (c *Client) CreateVersion(ctx context.Context, req CreateVersionRequest) (*EvaluatorVersion, error) {
	prerelease.Warn("evaluators.create_version", prerelease.Alpha)
	id, err := resolve.FindEvaluatorID(ctx, c.gen, req.Evaluator, req.Space)
	if err != nil {
		return nil, err
	}
	version, _, err := buildVersionCreate(req.Version)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.EvaluatorVersionsCreateWithResponse(ctx, id, version)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// GetVersion returns a single evaluator version by its ID. Version IDs are
// pure IDs with no name resolution.
func (c *Client) GetVersion(ctx context.Context, req GetVersionRequest) (*EvaluatorVersion, error) {
	prerelease.Warn("evaluators.get_version", prerelease.Alpha)
	resp, err := c.gen.EvaluatorVersionsGetWithResponse(ctx, req.VersionID)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// buildVersionCreate translates a public VersionConfig into the generated
// version-create union and reports the derived evaluator type. Exactly one of
// Template or Code must be set.
func buildVersionCreate(v VersionConfig) (generated.EvaluatorVersionCreate, generated.EvaluatorType, error) {
	var out generated.EvaluatorVersionCreate
	switch {
	case v.Template != nil && v.Code != nil:
		return out, "", ErrConflictingVersionConfig
	case v.Template != nil:
		err := out.FromEvaluatorVersionTemplateCreate(generated.EvaluatorVersionTemplateCreate{
			CommitMessage:  v.CommitMessage,
			TemplateConfig: *v.Template,
		})
		return out, generated.EvaluatorTypeTemplate, err
	case v.Code != nil:
		code, err := buildCodeConfig(*v.Code)
		if err != nil {
			return out, "", err
		}
		err = out.FromEvaluatorVersionCodeCreate(generated.EvaluatorVersionCodeCreate{
			CommitMessage: v.CommitMessage,
			CodeConfig:    code,
		})
		return out, generated.EvaluatorTypeCode, err
	default:
		return out, "", fmt.Errorf("evaluators: VersionConfig requires exactly one of Template or Code")
	}
}

// buildCodeConfig translates a public CodeConfig oneOf into the generated
// code-config union, setting the inner type discriminator. Exactly one of
// Managed or Custom must be set.
func buildCodeConfig(c CodeConfig) (generated.CodeConfig, error) {
	var out generated.CodeConfig
	switch {
	case c.Managed != nil && c.Custom != nil:
		return out, ErrConflictingCodeConfig
	case c.Managed != nil:
		managed := *c.Managed
		managed.Type = generated.ManagedCodeConfigTypeManaged
		return out, out.FromManagedCodeConfig(managed)
	case c.Custom != nil:
		custom := *c.Custom
		custom.Type = generated.CustomCodeConfigTypeCustom
		return out, out.FromCustomCodeConfig(custom)
	default:
		return out, fmt.Errorf("evaluators: CodeConfig requires exactly one of Managed or Custom")
	}
}
