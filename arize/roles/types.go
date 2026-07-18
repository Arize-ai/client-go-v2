package roles

import (
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// Response, list, and enum types remain aliases to the generated wire shapes so
// callers can construct and assert on them without importing internal/generated.
type (
	Role      = generated.Role
	ListRoles = generated.ListRolesResponse

	// Permission is the enum of permission identifiers (e.g. PROJECT_READ).
	// Construct values via the Permissions namespace, e.g.
	// roles.Permissions.ProjectRead.
	Permission = generated.Permission
)

// Permissions namespaces every permission identifier so callers can reference
// them without bare string literals, e.g. roles.Permissions.ProjectRead. The
// values are generated from the proto enum; see
// internal/generated/permissions_gen.go.
var Permissions = generated.Permissions

// GetRequest selects a single role.
type GetRequest struct {
	// Role accepts either a role name or ID.
	Role string
}

// CreateRequest is the request body for creating a new role.
type CreateRequest struct {
	// Name of the role (must be unique within the account).
	Name string
	// Permissions to grant. At least one permission is required.
	Permissions []Permission
	// Description is an optional brief description of the role's purpose. When
	// empty, the role is created without a description.
	Description string
}

// UpdateRequest is the request body for updating an existing role. At least
// one of Name, Description, or Permissions must be non-nil.
type UpdateRequest struct {
	// Role accepts either a role name or ID.
	Role string
	// Name is optional. When non-nil, sets a new name for the role; when nil,
	// the existing name is preserved.
	Name *string
	// Description is optional. When non-nil, sets a new description (pass a
	// pointer to an empty string to clear the existing description); when nil,
	// the existing description is preserved.
	Description *string
	// Permissions is optional. When non-nil, replaces the existing permission
	// set with the provided slice (which must contain at least one
	// permission — the API rejects empty arrays); when nil, the existing
	// permissions are preserved. Passing a pointer to an empty slice causes
	// the server to return 400 — permissions cannot be cleared.
	Permissions *[]Permission
}

// DeleteRequest identifies a role to delete. Predefined (system) roles cannot
// be deleted.
type DeleteRequest struct {
	// Role accepts either a role name or ID.
	Role string
}

// ListRequest holds optional filters for listing roles.
type ListRequest struct {
	// IsPredefined is a tri-state filter on the role's predefined status. When
	// nil, both predefined and custom roles are returned; &true returns only
	// system-defined predefined roles; &false returns only custom
	// (account-defined) roles.
	//
	// Pointer (rather than a value type) because the filter has three states
	// that a bare bool cannot express — see the tri-state boolean carve-out
	// in sdk/go/v2/AGENTS.md. Mirrors Python's `bool | None`.
	IsPredefined *bool
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the SDK defaults to 100, matching Python.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}
