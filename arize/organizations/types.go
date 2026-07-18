package organizations

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, membership, and role types remain aliases to the generated
// wire shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	Organization               = generated.Organization
	ListOrganizations          = generated.ListOrganizationsResponse
	OrganizationMembership     = generated.OrganizationMembership
	OrganizationRole           = generated.OrganizationRole
	OrganizationRoleAssignment = generated.OrganizationRoleAssignment
	PredefinedOrgRole          = generated.OrganizationPredefinedRoleAssignment
	CustomOrgRole              = generated.OrganizationCustomRoleAssignment
	RoleAssignmentType         = generated.OrganizationRoleAssignmentType
)

// OrganizationRole enum values.
const (
	OrganizationRoleAdmin     = generated.OrganizationRoleADMIN
	OrganizationRoleMember    = generated.OrganizationRoleMEMBER
	OrganizationRoleReadOnly  = generated.OrganizationRoleREADONLY
	OrganizationRoleAnnotator = generated.OrganizationRoleANNOTATOR
)

// Role assignment discriminator values.
const (
	RoleAssignmentTypePredefined = generated.OrganizationRoleAssignmentTypePREDEFINED
	RoleAssignmentTypeCustom     = generated.OrganizationRoleAssignmentTypeCUSTOM
)

// GetRequest selects a single organization.
type GetRequest struct {
	// Organization accepts either an organization name or ID.
	Organization string
}

// CreateRequest is the request body for creating a new organization.
type CreateRequest struct {
	// Name of the organization (must be unique within the account).
	Name string
	// Description is an optional brief description of the organization's
	// purpose. When empty, the organization is created without a description.
	Description string
}

// UpdateRequest is the request body for updating an existing organization.
// At least one of Name or Description must be non-nil.
type UpdateRequest struct {
	// Organization accepts either an organization name or ID.
	Organization string
	// Name is optional. When non-nil, sets a new name for the organization;
	// when nil, the existing name is preserved.
	Name *string
	// Description is optional. When non-nil, sets a new description (pass a
	// pointer to an empty string to clear the existing description); when nil,
	// the existing description is preserved.
	Description *string
}

// DeleteRequest identifies an organization to delete. Deletion is irreversible
// and cascades to all child resources (spaces, projects, API keys, etc.).
type DeleteRequest struct {
	// Organization accepts either an organization name or ID.
	Organization string
}

// AddUserRequest adds a user to an organization, or upserts their role if they
// are already a member.
type AddUserRequest struct {
	// Organization accepts either an organization name or ID.
	Organization string
	// UserID is the unique identifier of the user to add.
	UserID string
	// Role is the predefined organization role to assign. Custom role
	// assignments are not yet supported by the server.
	Role PredefinedOrgRole
}

// RemoveUserRequest removes a user from an organization. Membership cascades
// to all child spaces.
type RemoveUserRequest struct {
	// Organization accepts either an organization name or ID.
	Organization string
	// UserID is the unique identifier of the user to remove.
	UserID string
}

// ListRequest holds optional filters for listing organizations.
type ListRequest struct {
	// Name is an optional case-insensitive substring filter on the
	// organization name. When empty, results are not filtered by name.
	Name string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}
