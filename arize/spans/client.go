package spans

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
)

// Client provides access to the Arize Spans API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of spans matching the given filter.
//
// Unlike other List methods in this SDK, spans.List takes a body: the spans
// API uses POST because the filter DSL and projection list can be too large
// for a query string. Pagination stays in params.
func (c *Client) List(ctx context.Context, req ListRequest, params ListParams) (*SpanList, error) {
	prerelease.Warn("spans.list", prerelease.Alpha)
	resp, err := c.gen.SpansListWithResponse(ctx, &params, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes spans matching the given criteria. A nil return means all
// spans were fully deleted (HTTP 204). A non-nil SpanDeletePartial means the
// server returned HTTP 200: the request was partially processed and only the
// IDs listed in SpanDeletePartial.DeletedSpanIds were confirmed deleted —
// the caller should retry for the remainder.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) (*SpanDeletePartial, error) {
	prerelease.Warn("spans.delete", prerelease.Alpha)
	resp, err := c.gen.SpansDeleteWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
