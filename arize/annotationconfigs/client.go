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
	prerelease.Warn("annotationconfigs.list", prerelease.Beta)
	params := &generated.ListAnnotationConfigsParams{
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
	resp, err := c.gen.ListAnnotationConfigsWithResponse(ctx, params)
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
	prerelease.Warn("annotationconfigs.get", prerelease.Beta)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetAnnotationConfigWithResponse(ctx, id)
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
	prerelease.Warn("annotationconfigs.create", prerelease.Beta)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.CreateAnnotationConfigJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeCONTINUOUS,
	}
	if err := body.FromCreateContinuousAnnotationConfigRequest(generated.CreateContinuousAnnotationConfigRequest{
		AnnotationConfigType:  generated.CreateContinuousAnnotationConfigRequestAnnotationConfigTypeCONTINUOUS,
		Name:                  req.Name,
		SpaceId:               spaceID,
		MinimumScore:          req.MinimumScore,
		MaximumScore:          req.MaximumScore,
		OptimizationDirection: optfields.PtrIfSet(req.OptimizationDirection),
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build continuous body: %w", err)
	}
	resp, err := c.gen.CreateAnnotationConfigWithResponse(ctx, body)
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
	prerelease.Warn("annotationconfigs.create", prerelease.Beta)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.CreateAnnotationConfigJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeCATEGORICAL,
	}
	if err := body.FromCreateCategoricalAnnotationConfigRequest(generated.CreateCategoricalAnnotationConfigRequest{
		AnnotationConfigType:  generated.CreateCategoricalAnnotationConfigRequestAnnotationConfigTypeCATEGORICAL,
		Name:                  req.Name,
		SpaceId:               spaceID,
		Values:                req.Values,
		OptimizationDirection: optfields.PtrIfSet(req.OptimizationDirection),
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build categorical body: %w", err)
	}
	resp, err := c.gen.CreateAnnotationConfigWithResponse(ctx, body)
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
	prerelease.Warn("annotationconfigs.create", prerelease.Beta)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.CreateAnnotationConfigJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeFREEFORM,
	}
	if err := body.FromCreateFreeformAnnotationConfigRequest(generated.CreateFreeformAnnotationConfigRequest{
		AnnotationConfigType: generated.CreateFreeformAnnotationConfigRequestAnnotationConfigTypeFREEFORM,
		Name:                 req.Name,
		SpaceId:              spaceID,
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build freeform body: %w", err)
	}
	resp, err := c.gen.CreateAnnotationConfigWithResponse(ctx, body)
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
	prerelease.Warn("annotationconfigs.update_categorical", prerelease.Beta)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.UpdateAnnotationConfigJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeCATEGORICAL,
	}
	if err := body.FromUpdateCategoricalAnnotationConfigRequest(generated.UpdateCategoricalAnnotationConfigRequest{
		AnnotationConfigType:  generated.UpdateCategoricalAnnotationConfigRequestAnnotationConfigTypeCATEGORICAL,
		Name:                  req.Name,
		OptimizationDirection: req.OptimizationDirection,
		Values:                req.Values,
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build categorical body: %w", err)
	}
	resp, err := c.gen.UpdateAnnotationConfigWithResponse(ctx, id, body)
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
	prerelease.Warn("annotationconfigs.update_continuous", prerelease.Beta)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.UpdateAnnotationConfigJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeCONTINUOUS,
	}
	if err := body.FromUpdateContinuousAnnotationConfigRequest(generated.UpdateContinuousAnnotationConfigRequest{
		AnnotationConfigType:  generated.UpdateContinuousAnnotationConfigRequestAnnotationConfigTypeCONTINUOUS,
		Name:                  req.Name,
		MinimumScore:          req.MinimumScore,
		MaximumScore:          req.MaximumScore,
		OptimizationDirection: req.OptimizationDirection,
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build continuous body: %w", err)
	}
	resp, err := c.gen.UpdateAnnotationConfigWithResponse(ctx, id, body)
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
	prerelease.Warn("annotationconfigs.update_freeform", prerelease.Beta)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.UpdateAnnotationConfigJSONRequestBody{
		AnnotationConfigType: generated.AnnotationConfigTypeFREEFORM,
	}
	if err := body.FromUpdateFreeformAnnotationConfigRequest(generated.UpdateFreeformAnnotationConfigRequest{
		AnnotationConfigType: generated.UpdateFreeformAnnotationConfigRequestAnnotationConfigTypeFREEFORM,
		Name:                 req.Name,
	}); err != nil {
		return nil, fmt.Errorf("annotationconfigs: build freeform body: %w", err)
	}
	resp, err := c.gen.UpdateAnnotationConfigWithResponse(ctx, id, body)
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
	prerelease.Warn("annotationconfigs.delete", prerelease.Beta)
	id, err := resolve.FindAnnotationConfigID(ctx, c.gen, req.AnnotationConfig, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.DeleteAnnotationConfigWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
