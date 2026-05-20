// Package resourcerestrictions provides access to the Arize resource
// restrictions API.
//
// Resource restrictions prevent roles bound at higher hierarchy levels (space,
// org, account) from granting access to the restricted resource. Only space
// admins or users with the PROJECT_RESTRICT permission can restrict or
// unrestrict a resource.
//
// Currently only PROJECT resources are supported.
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

// Restrict marks a resource as restricted.
//
// Restricting a resource prevents roles bound at higher hierarchy levels
// (space, org, account) from granting access. Only space admins or users with
// the PROJECT_RESTRICT permission can perform this action.
//
// This operation is idempotent — restricting an already-restricted resource
// returns the existing restriction without error.
//
// Currently only PROJECT resources are supported.
func (c *Client) Restrict(
	ctx context.Context,
	req RestrictRequest,
) (*ResourceRestriction, error) {
	prerelease.Warn("resourcerestrictions.restrict", prerelease.Alpha)
	resp, err := c.gen.ResourceRestrictionsCreateWithResponse(ctx, generated.ResourceRestrictionCreate{
		ResourceId: req.ResourceID,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return &resp.JSON200.ResourceRestriction, nil
}

// Unrestrict removes a restriction from a resource.
//
// Removing a restriction means that roles bound at other levels of the
// hierarchy (space, org, account) can once again grant access to the resource.
//
// ResourceID is the ID of the restricted resource (e.g. a project ID), not the
// ID of the restriction record itself — the API dereferences from the
// restricted resource to the restriction it holds.
func (c *Client) Unrestrict(
	ctx context.Context,
	req UnrestrictRequest,
) error {
	prerelease.Warn("resourcerestrictions.unrestrict", prerelease.Alpha)
	resp, err := c.gen.ResourceRestrictionsDeleteWithResponse(ctx, req.ResourceID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
