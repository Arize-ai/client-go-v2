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
		Limit:  optfields.PtrIfSet(req.Limit),
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

// Create creates a new annotation config, resolving the parent space by name
// or ID. The annotation config body is a discriminated union; this method
// builds the right variant based on req.Type.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*AnnotationConfig, error) {
	prerelease.Warn("annotationconfigs.create", prerelease.Alpha)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AnnotationConfigsCreateJSONRequestBody{
		AnnotationConfigType: req.Type,
	}
	switch req.Type {
	case generated.AnnotationConfigTypeCategorical:
		if err := body.FromCategoricalAnnotationConfigCreate(generated.CategoricalAnnotationConfigCreate{
			AnnotationConfigType:  generated.CategoricalAnnotationConfigCreateAnnotationConfigTypeCategorical,
			Name:                  req.Name,
			SpaceId:               spaceID,
			OptimizationDirection: optfields.PtrIfSet(req.OptimizationDirection),
			Values:                req.Values,
		}); err != nil {
			return nil, fmt.Errorf("annotationconfigs: build categorical body: %w", err)
		}
	case generated.AnnotationConfigTypeContinuous:
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
	case generated.AnnotationConfigTypeFreeform:
		if err := body.FromFreeformAnnotationConfigCreate(generated.FreeformAnnotationConfigCreate{
			AnnotationConfigType: generated.FreeformAnnotationConfigCreateAnnotationConfigTypeFreeform,
			Name:                 req.Name,
			SpaceId:              spaceID,
		}); err != nil {
			return nil, fmt.Errorf("annotationconfigs: build freeform body: %w", err)
		}
	default:
		return nil, fmt.Errorf("annotationconfigs: unknown Type %q (must be categorical, continuous, or freeform)", req.Type)
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
