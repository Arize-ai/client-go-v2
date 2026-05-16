package resourcerestrictions

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
)

// Client provides access to the Arize Resource Restrictions API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// Create creates a new resource restriction and returns it.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*ResourceRestrictionResponse, error) {
	prerelease.Warn("resourcerestrictions.create", prerelease.Alpha)
	resp, err := c.gen.ResourceRestrictionsCreateWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes a resource restriction by resource ID.
func (c *Client) Delete(ctx context.Context, resourceID string) error {
	prerelease.Warn("resourcerestrictions.delete", prerelease.Alpha)
	resp, err := c.gen.ResourceRestrictionsDeleteWithResponse(ctx, resourceID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
