package spaces

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, membership, and role types remain aliases to the generated
// wire shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	Space               = generated.Space
	SpaceList           = generated.SpaceList
	SpaceMembership     = generated.SpaceMembership
	SpaceRoleAssignment = generated.SpaceRoleAssignment
	PredefinedSpaceRole = generated.PredefinedRoleAssignment
	CustomSpaceRole     = generated.CustomRoleAssignment
	UserSpaceRole       = generated.UserSpaceRole
	RoleAssignmentType  = generated.SpaceRoleAssignmentType
)

// UserSpaceRole enum values.
const (
	UserSpaceRoleAdmin     = generated.UserSpaceRoleAdmin
	UserSpaceRoleMember    = generated.UserSpaceRoleMember
	UserSpaceRoleReadOnly  = generated.UserSpaceRoleReadOnly
	UserSpaceRoleAnnotator = generated.UserSpaceRoleAnnotator
)

// Role assignment discriminator values.
const (
	RoleAssignmentTypePredefined = generated.SpaceRoleAssignmentTypePredefined
	RoleAssignmentTypeCustom     = generated.SpaceRoleAssignmentTypeCustom
)

// AsPredefined returns the role's predefined assignment and true if the role's
// discriminator is "predefined". Returns the zero value and false otherwise —
// including when the discriminator says "custom" or when JSON parsing fails.
func AsPredefined(role SpaceRoleAssignment) (PredefinedSpaceRole, bool) {
	d, err := role.Discriminator()
	if err != nil || d != string(RoleAssignmentTypePredefined) {
		return PredefinedSpaceRole{}, false
	}
	pre, err := role.AsPredefinedRoleAssignment()
	if err != nil {
		return PredefinedSpaceRole{}, false
	}
	return pre, true
}

// AsCustom returns the role's custom assignment and true if the role's
// discriminator is "custom". Returns the zero value and false otherwise.
func AsCustom(role SpaceRoleAssignment) (CustomSpaceRole, bool) {
	d, err := role.Discriminator()
	if err != nil || d != string(RoleAssignmentTypeCustom) {
		return CustomSpaceRole{}, false
	}
	custom, err := role.AsCustomRoleAssignment()
	if err != nil {
		return CustomSpaceRole{}, false
	}
	return custom, true
}

// AssignPredefinedRole builds a role assignment naming one of the predefined
// UserSpaceRole values (e.g. UserSpaceRoleMember). Pass the result as
// AddUserRequest.Role.
func AssignPredefinedRole(name UserSpaceRole) SpaceRoleAssignment {
	var role SpaceRoleAssignment
	// From* only marshals a fixed two-field struct and sets the discriminator
	// internally; it cannot fail, so the error is safely discarded.
	_ = role.FromPredefinedRoleAssignment(PredefinedSpaceRole{Name: name})
	return role
}

// AssignCustomRole builds a role assignment referencing an existing custom RBAC
// role by its ID (see the roles package). Pass the result as
// AddUserRequest.Role.
func AssignCustomRole(roleID string) SpaceRoleAssignment {
	var role SpaceRoleAssignment
	_ = role.FromCustomRoleAssignment(CustomSpaceRole{Id: roleID})
	return role
}

// GetRequest selects a single space.
type GetRequest struct {
	// Space accepts either a space name or ID.
	Space string
}

// CreateRequest is the request body for creating a new space.
type CreateRequest struct {
	// Name of the space (must be unique within the organization).
	Name string
	// Organization accepts either an organization name or ID and identifies the
	// parent organization for the new space.
	Organization string
	// Description is an optional brief description of the space's purpose. When
	// empty, the space is created without a description.
	Description string
}

// UpdateRequest is the request body for updating an existing space. Leave a
// field nil to preserve its current value.
type UpdateRequest struct {
	// Space accepts either a space name or ID.
	Space string
	// Name is optional. When non-nil, sets a new name for the space; when nil,
	// the existing name is preserved.
	Name *string
	// Description is optional. When non-nil, sets a new description (pass a
	// pointer to an empty string to clear the existing description); when nil,
	// the existing description is preserved.
	Description *string
}

// DeleteRequest identifies a space to delete. Deletion is irreversible and
// cascades to all child resources (projects, datasets, monitors, custom
// metrics, etc.).
type DeleteRequest struct {
	// Space accepts either a space name or ID.
	Space string
}

// ListRequest holds optional filters for listing spaces.
type ListRequest struct {
	// Organization is an optional name-or-ID filter. When non-empty, only
	// spaces belonging to this organization are returned.
	Organization string
	// Name is an optional case-insensitive substring filter on the space name.
	// When empty, results are not filtered by name.
	Name string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// AddUserRequest adds a user to a space, or upserts their role if they are
// already a member.
type AddUserRequest struct {
	// Space accepts either a space name or ID.
	Space string
	// UserID is the unique identifier of the user to add.
	UserID string
	// Role is the role assignment for the user on this space. Build it with
	// AssignPredefinedRole (one of the UserSpaceRole values) or AssignCustomRole
	// (an existing custom role ID).
	Role SpaceRoleAssignment
}

// RemoveUserRequest removes a user from a space. Removes both the legacy
// SpaceMembers row and any RBAC role bindings for the user on this space.
type RemoveUserRequest struct {
	// Space accepts either a space name or ID.
	Space string
	// UserID is the unique identifier of the user to remove.
	UserID string
}
