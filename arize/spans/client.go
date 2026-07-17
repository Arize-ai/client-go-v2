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
	prerelease.Warn("spans.list", prerelease.Beta)
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
	params := generated.ListSpansParams{
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListSpansWithResponse(ctx, &params, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes spans by their IDs. The response always includes Completed
// (whether a retry is needed), DeletedSpanIds (confirmed deleted), and
// NotDeletedSpanIds (IDs not deleted — either not found within the supported
// lookback window, or not reached when Completed is false).
//
// When Completed is false, the operation could not fully complete. The caller
// should retry the original full request. The delete operation is idempotent,
// so re-submitting already-deleted IDs is safe.
//
// req.Project accepts a name or ID; req.Space is required when req.Project is
// a name.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) (*SpanDeleteResult, error) {
	prerelease.Warn("spans.delete", prerelease.Beta)
	projectID, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.DeleteSpansRequest{
		ProjectId: projectID,
		SpanIds:   req.SpanIDs,
	}
	resp, err := c.gen.DeleteSpansWithResponse(ctx, body)
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
	prerelease.Warn("spans.annotate", prerelease.Beta)
	projectID, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return err
	}
	body := generated.AnnotateSpansRequest{
		ProjectId:   projectID,
		Annotations: req.Annotations,
		StartTime:   optfields.PtrIfSet(req.Start),
		EndTime:     optfields.PtrIfSet(req.End),
	}
	resp, err := c.gen.AnnotateSpansWithResponse(ctx, body)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
