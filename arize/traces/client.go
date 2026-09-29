package traces

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Traces API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of traces matching the given filter.
//
// Like spans.List, traces.List uses POST because the filter DSL can be too
// large for a query string. Both the body fields (project, time range, filter)
// and the query params (limit, cursor) are flattened into ListRequest.
//
// A trace is returned when any of its spans matches the filter. Each trace
// carries a flat list of spans; reconstruct the trace tree client-side using
// each span's parent_id.
//
// req.Project accepts a name or ID; req.Space is required when req.Project is
// a name.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*ListTraces, error) {
	prerelease.Warn("traces.list", prerelease.Beta)
	projectID, err := resolve.FindProjectID(ctx, c.gen, req.Project, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.ListTracesRequest{
		ProjectId: projectID,
		StartTime: optfields.PtrIfSet(req.Start),
		EndTime:   optfields.PtrIfSet(req.End),
		Filter:    optfields.PtrIfSet(req.Filter),
	}
	params := generated.ListTracesParams{
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListTracesWithResponse(ctx, &params, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
