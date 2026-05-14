package spans

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/sdkconfig"
)

// Client provides access to the Arize Spans API.
type Client struct {
	gen *generated.ClientWithResponses
	cfg sdkconfig.Config
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses, cfg sdkconfig.Config) *Client {
	return &Client{gen: gen, cfg: cfg}
}

// List returns a paginated list of spans matching the given filter.
func (c *Client) List(ctx context.Context, req ListSpansRequest, params ListParams) (*SpanList, error) {
	prerelease.Warn("spans.list", prerelease.Alpha)
	reqBody := generated.SpansListJSONRequestBody{
		ProjectId: req.ProjectId,
		EndTime:   req.EndTime,
		Filter:    req.Filter,
		StartTime: req.StartTime,
	}
	resp, err := c.gen.SpansListWithResponse(ctx, &params, reqBody)
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
func (c *Client) Delete(ctx context.Context, req DeleteSpansRequest) (*SpanDeletePartial, error) {
	prerelease.Warn("spans.delete", prerelease.Alpha)
	reqBody := generated.SpansDeleteJSONRequestBody{
		ProjectId: req.ProjectId,
		SpanIds:   req.SpanIds,
	}
	resp, err := c.gen.SpansDeleteWithResponse(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
