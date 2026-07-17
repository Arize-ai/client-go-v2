package auditlogs

import (
	"context"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
)

// Client provides access to the Arize Audit Logs API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of audit log entries for the account, ordered
// newest first. The caller must be an account admin and the account must have
// audit logging enabled.
func (c *Client) List(ctx context.Context, req ListRequest) (*AuditLogList, error) {
	prerelease.Warn("audit_logs.list", prerelease.Beta)
	params := &generated.ListAuditLogsParams{
		StartTime:     optfields.PtrIfSet(req.StartTime),
		EndTime:       optfields.PtrIfSet(req.EndTime),
		UserId:        optfields.PtrIfSet(req.UserID),
		OperationType: optfields.PtrIfSet(req.OperationType),
		Limit:         optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor:        optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListAuditLogsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
