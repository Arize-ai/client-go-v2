package auditlogs

import (
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// StartTime, when non-zero, filters results to entries created at or after
	// this time (inclusive). When zero, the server defaults to 30 days before
	// EndTime.
	StartTime time.Time
	// EndTime, when non-zero, filters results to entries created at or before
	// this time (inclusive). When zero, the server defaults to the current time.
	EndTime time.Time
	// UserID, when non-empty, filters results to entries for this user
	// (base64-encoded resource ID). When empty, no user filtering is applied.
	UserID string
	// OperationType, when non-empty, filters results to entries with this
	// operation type. When empty, no operation-type filtering is applied.
	OperationType AuditLogOperationType
	// Limit is the optional maximum number of items to return. When zero, the
	// SDK applies a default of 50. Server max is 100.
	Limit int
	// Cursor is the optional opaque pagination cursor returned from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

type (
	// AuditLog is a single audit log entry recording an authenticated user action.
	AuditLog = generated.AuditLog

	// AuditLogList is the cursor-paginated list response shape.
	AuditLogList = generated.ListAuditLogsResponse

	// AuditLogOperationType is the type of audited operation.
	AuditLogOperationType = generated.AuditLogOperationType
)

const (
	// AuditLogOperationTypeQUERY is a read-only query operation.
	AuditLogOperationTypeQUERY AuditLogOperationType = generated.AuditLogOperationTypeQUERY
	// AuditLogOperationTypeMUTATION is a write operation (GraphQL mutation or REST API call).
	AuditLogOperationTypeMUTATION AuditLogOperationType = generated.AuditLogOperationTypeMUTATION
	// AuditLogOperationTypeSUBSCRIPTION is a real-time subscription operation.
	AuditLogOperationTypeSUBSCRIPTION AuditLogOperationType = generated.AuditLogOperationTypeSUBSCRIPTION
)
