package users

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, enum, and role-assignment types are aliases to the generated
// wire shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	User                   = generated.User
	ListUsers              = generated.ListUsersResponse
	UserStatus             = generated.UserStatus
	UserRole               = generated.UserRole
	InviteMode             = generated.InviteMode
	UserRoleAssignment     = generated.UserRoleAssignment
	PredefinedUserRole     = generated.PredefinedUserRoleAssignment
	CustomUserRole         = generated.CustomUserRoleAssignment
	UserRoleAssignmentType = generated.UserRoleAssignmentType
)

// User status values.
const (
	UserStatusActive  = generated.UserStatusACTIVE
	UserStatusInvited = generated.UserStatusINVITED
	UserStatusExpired = generated.UserStatusEXPIRED
)

// Predefined account-level role names.
const (
	UserRoleAdmin     = generated.UserRoleADMIN
	UserRoleMember    = generated.UserRoleMEMBER
	UserRoleAnnotator = generated.UserRoleANNOTATOR
)

// Invite modes for Client.Create.
const (
	InviteModeNone              = generated.InviteModeNONE
	InviteModeEmailLink         = generated.InviteModeEMAILLINK
	InviteModeTemporaryPassword = generated.InviteModeTEMPORARYPASSWORD
)

// Role assignment discriminator values.
const (
	RoleAssignmentTypePredefined = generated.UserRoleAssignmentTypePREDEFINED
	RoleAssignmentTypeCustom     = generated.UserRoleAssignmentTypeCUSTOM
)

// AssignPredefinedRole builds an account-level role assignment naming one of
// the predefined UserRole values (UserRoleAdmin, UserRoleMember,
// UserRoleAnnotator). Pass the result as CreateRequest.Role.
func AssignPredefinedRole(name UserRole) UserRoleAssignment {
	var role UserRoleAssignment
	// From* only marshals a fixed two-field struct and sets the discriminator
	// internally; it cannot fail, so the error is safely discarded.
	_ = role.FromPredefinedUserRoleAssignment(PredefinedUserRole{Name: name, Type: RoleAssignmentTypePredefined})
	return role
}

// AssignCustomRole builds an account-level role assignment referencing an
// existing custom RBAC role by its ID (see the roles package). Pass the result
// as CreateRequest.Role.
//
// Note: custom account-role assignments are reserved for future use server-side.
func AssignCustomRole(roleID string) UserRoleAssignment {
	var role UserRoleAssignment
	_ = role.FromCustomUserRoleAssignment(CustomUserRole{Id: roleID, Type: RoleAssignmentTypeCustom})
	return role
}

// DeletionStatus is the per-user outcome of a BulkDelete attempt.
type DeletionStatus string

const (
	DeletionStatusDeleted  DeletionStatus = "deleted"
	DeletionStatusFailed   DeletionStatus = "failed"
	DeletionStatusNotFound DeletionStatus = "not_found"
)

// BulkUserDeletionResult records the outcome of one deletion attempt in
// Client.BulkDelete.
type BulkUserDeletionResult struct {
	// UserID is the ID of the user targeted for deletion. It is empty when an
	// email could not be resolved to a user (Status NotFound).
	UserID string
	// Email is the address the deletion was requested for, set only when the
	// user was specified by email. Empty when the user was specified by ID.
	Email string
	// Status is the outcome: Deleted, Failed, or NotFound.
	Status DeletionStatus
	// Error is a human-readable message, populated for Failed and NotFound.
	Error string
}

// ListRequest holds optional filters for Client.List.
type ListRequest struct {
	// Email is an optional case-insensitive substring filter on user email
	// (max 255 chars). When empty, results are not filtered by email.
	Email string
	// Status is an optional filter on account status. When nil or empty, the
	// server returns active, invited, and expired users.
	Status []UserStatus
	// Limit is the optional maximum number of items to return (server max 100).
	// When zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous response's
	// pagination.next_cursor. When empty, results start from the first page.
	Cursor string
}

// GetRequest selects a single user.
type GetRequest struct {
	// User accepts either a user ID or an email address. An email is resolved
	// to an ID via the users list endpoint (case-insensitive exact match); a
	// non-matching email yields a *ResourceNotFoundError.
	User string
}

// CreateRequest is the request body for Client.Create.
type CreateRequest struct {
	// Name is the full display name of the new user (1–255 chars).
	Name string
	// Email is the email address of the user to invite. It also serves as the
	// idempotency key when InviteMode is not InviteModeNone.
	Email string
	// Role is the account-level role assignment. Build it with
	// AssignPredefinedRole (one of the UserRole values) or AssignCustomRole.
	Role UserRoleAssignment
	// InviteMode controls whether and how an invitation is sent
	// (InviteModeNone, InviteModeEmailLink, or InviteModeTemporaryPassword).
	InviteMode InviteMode
	// IsDeveloper optionally sets developer permissions. When nil, the server
	// applies its role-based default (true for admin/member, false for
	// annotator). Non-nil sets the value explicitly.
	IsDeveloper *bool
}

// UpdateRequest is the request body for Client.Update. At least one of Name or
// IsDeveloper must be non-nil. Leave a field nil to preserve its current value.
type UpdateRequest struct {
	// UserID is the strict ID of the user to update.
	UserID string
	// Name is optional. When non-nil, sets a new display name (1–255 chars;
	// whitespace-only is rejected by the server). When nil, the existing name
	// is preserved. The name cannot be cleared.
	Name *string
	// IsDeveloper is optional. When non-nil, grants (true) or revokes (false)
	// developer permissions. When nil, the current value is preserved.
	IsDeveloper *bool
}

// DeleteRequest identifies a user to delete (soft-delete; cascades to org/space
// memberships, API keys, and role bindings).
type DeleteRequest struct {
	// UserID is the strict ID of the user to delete.
	UserID string
}

// ResendInvitationRequest identifies a pending (invited) user to resend an
// invitation email to.
type ResendInvitationRequest struct {
	// UserID is the strict ID of the user.
	UserID string
}

// ResetPasswordRequest identifies a user to send a password-reset email to.
type ResetPasswordRequest struct {
	// UserID is the strict ID of the user.
	UserID string
}

// BulkDeleteRequest selects users to delete by ID and/or email. At least one of
// UserIDs or Emails must be non-empty.
type BulkDeleteRequest struct {
	// UserIDs are user IDs to delete directly.
	UserIDs []string
	// Emails are addresses to resolve (case-insensitive exact match) and then
	// delete. Unresolved emails are recorded as NotFound, not errors.
	Emails []string
}
