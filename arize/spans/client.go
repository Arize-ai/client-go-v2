package spans

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
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
// Unlike other List methods in this SDK, spans.List uses POST because the
// filter DSL and projection list can be too large for a query string. Both
// the body fields (project, time range, filter) and the query params (limit,
// cursor) are flattened into ListRequest.
//
// req.Project accepts a name or ID; req.Space is required when req.Project is
// a name.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*SpanList, error) {
	prerelease.Warn("spans.list", prerelease.Alpha)
	projectID, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.ListSpansRequest{
		ProjectId: projectID,
		StartTime: optfields.PtrIfSet(req.Start),
		EndTime:   optfields.PtrIfSet(req.End),
		Filter:    optfields.PtrIfSet(req.Filter),
	}
	params := generated.SpansListParams{
		Limit:  optfields.PtrIfSet(req.Limit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.SpansListWithResponse(ctx, &params, body)
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
//
// req.Project accepts a name or ID; req.Space is required when req.Project is
// a name.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) (*SpanDeletePartial, error) {
	prerelease.Warn("spans.delete", prerelease.Alpha)
	projectID, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.DeleteSpansRequest{
		ProjectId: projectID,
		SpanIds:   req.SpanIDs,
	}
	resp, err := c.gen.SpansDeleteWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Annotate writes human annotations to a batch of spans. Annotations are
// upserted by annotation config name per span: resubmitting the same config
// name for the same span overwrites the previous value, so retries do not
// create duplicates.
//
// Up to 1000 spans may be annotated per request. Spans are looked up within
// the configured time window (defaulting to the last 31 days). If any span ID
// in the batch is not found within the window, the entire request is
// rejected.
//
// The write completes synchronously before the function returns (HTTP 202
// Accepted). Visibility in read queries may lag by a short interval.
//
// req.Project accepts a name or ID; req.Space is required when req.Project is
// a name.
func (c *Client) Annotate(
	ctx context.Context,
	req AnnotateRequest,
) error {
	prerelease.Warn("spans.annotate", prerelease.Alpha)
	projectID, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return err
	}
	body := generated.AnnotateSpansRequestBody{
		ProjectId:   projectID,
		Annotations: req.Annotations,
		StartTime:   optfields.PtrIfSet(req.Start),
		EndTime:     optfields.PtrIfSet(req.End),
	}
	resp, err := c.gen.SpansAnnotateWithResponse(ctx, body)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
