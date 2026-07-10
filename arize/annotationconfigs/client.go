package annotationconfigs

import (
	"context"
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Annotation Configs API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of annotation configs.
func (c *Client) List(ctx context.Context, req ListRequest) (*AnnotationConfigList, error) {
	prerelease.Warn("annotationconfigs.list", prerelease.Alpha)
	params := &generated.AnnotationConfigsListParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	if req.Space != "" {
		spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
		if err != nil {
			return nil, err
		}
		params.SpaceId = &spaceID
	}
	resp, err := c.gen.AnnotationConfigsListWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single annotation config, resolving by name or ID.
func (c *Client) Get(ctx context.Context, req GetRequest) (*AnnotationConfig, error) {
	prerelease.Warn("annotationconfigs.get", prerelease.Alpha)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AnnotationConfigsGetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// CreateContinuous creates a new continuous annotation config, resolving the
// parent space by name or ID.
func (c *Client) CreateContinuous(ctx context.Context, req CreateContinuousRequest) (*AnnotationConfig, error) {
	prerelease.Warn("annotationconfigs.create", prerelease.Alpha)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AnnotationConfigsCreateJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeContinuous,
	}
	if err := body.FromContinuousAnnotationConfigCreate(generated.ContinuousAnnotationConfigCreate{
		AnnotationConfigType:  generated.ContinuousAnnotationConfigCreateAnnotationConfigTypeContinuous,
		Name:                  req.Name,
		SpaceId:               spaceID,
		MinimumScore:          req.MinimumScore,
		MaximumScore:          req.MaximumScore,
		OptimizationDirection: optfields.PtrIfSet(req.OptimizationDirection),
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build continuous body: %w", err)
	}
	resp, err := c.gen.AnnotationConfigsCreateWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// CreateCategorical creates a new categorical annotation config, resolving
// the parent space by name or ID.
func (c *Client) CreateCategorical(ctx context.Context, req CreateCategoricalRequest) (*AnnotationConfig, error) {
	prerelease.Warn("annotationconfigs.create", prerelease.Alpha)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AnnotationConfigsCreateJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeCategorical,
	}
	if err := body.FromCategoricalAnnotationConfigCreate(generated.CategoricalAnnotationConfigCreate{
		AnnotationConfigType:  generated.CategoricalAnnotationConfigCreateAnnotationConfigTypeCategorical,
		Name:                  req.Name,
		SpaceId:               spaceID,
		Values:                req.Values,
		OptimizationDirection: optfields.PtrIfSet(req.OptimizationDirection),
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build categorical body: %w", err)
	}
	resp, err := c.gen.AnnotationConfigsCreateWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// CreateFreeform creates a new freeform annotation config, resolving the
// parent space by name or ID.
func (c *Client) CreateFreeform(ctx context.Context, req CreateFreeformRequest) (*AnnotationConfig, error) {
	prerelease.Warn("annotationconfigs.create", prerelease.Alpha)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AnnotationConfigsCreateJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeFreeform,
	}
	if err := body.FromFreeformAnnotationConfigCreate(generated.FreeformAnnotationConfigCreate{
		AnnotationConfigType: generated.FreeformAnnotationConfigCreateAnnotationConfigTypeFreeform,
		Name:                 req.Name,
		SpaceId:              spaceID,
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build freeform body: %w", err)
	}
	resp, err := c.gen.AnnotationConfigsCreateWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// UpdateCategorical modifies an existing categorical annotation config,
// resolving by name or ID. Fields left nil preserve their current
// values.
func (c *Client) UpdateCategorical(ctx context.Context, req UpdateCategoricalRequest) (*AnnotationConfig, error) {
	prerelease.Warn("annotationconfigs.update_categorical", prerelease.Alpha)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AnnotationConfigsUpdateJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeCategorical,
	}
	if err := body.FromCategoricalAnnotationConfigUpdate(generated.CategoricalAnnotationConfigUpdate{
		AnnotationConfigType:  generated.CategoricalAnnotationConfigUpdateAnnotationConfigTypeCategorical,
		Name:                  req.Name,
		OptimizationDirection: req.OptimizationDirection,
		Values:                req.Values,
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build categorical body: %w", err)
	}
	resp, err := c.gen.AnnotationConfigsUpdateWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// UpdateContinuous modifies an existing continuous annotation config,
// resolving by name or ID. Fields left nil preserve their current
// values.
func (c *Client) UpdateContinuous(ctx context.Context, req UpdateContinuousRequest) (*AnnotationConfig, error) {
	prerelease.Warn("annotationconfigs.update_continuous", prerelease.Alpha)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AnnotationConfigsUpdateJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeContinuous,
	}
	if err := body.FromContinuousAnnotationConfigUpdate(generated.ContinuousAnnotationConfigUpdate{
		AnnotationConfigType:  generated.ContinuousAnnotationConfigUpdateAnnotationConfigTypeContinuous,
		Name:                  req.Name,
		MinimumScore:          req.MinimumScore,
		MaximumScore:          req.MaximumScore,
		OptimizationDirection: req.OptimizationDirection,
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build continuous body: %w", err)
	}
	resp, err := c.gen.AnnotationConfigsUpdateWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// UpdateFreeform modifies an existing freeform annotation config, resolving
// by name or ID. Fields left nil preserve their current values.
func (c *Client) UpdateFreeform(ctx context.Context, req UpdateFreeformRequest) (*AnnotationConfig, error) {
	prerelease.Warn("annotationconfigs.update_freeform", prerelease.Alpha)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AnnotationConfigsUpdateJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeFreeform,
	}
	if err := body.FromFreeformAnnotationConfigUpdate(generated.FreeformAnnotationConfigUpdate{
		AnnotationConfigType: generated.FreeformAnnotationConfigUpdateAnnotationConfigTypeFreeform,
		Name:                 req.Name,
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build freeform body: %w", err)
	}
	resp, err := c.gen.AnnotationConfigsUpdateWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes an annotation config, resolving by name or ID.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("annotationconfigs.delete", prerelease.Alpha)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.AnnotationConfigsDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
