package apikeys

import (
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// SpaceBinding declares one space that the service key's bot user should access.
//
// The binding (Space) answers "which space?". The role assignment (Role) answers
// "with what role?" — either a named predefined role or a custom RBAC role by ID.
// Build Role with [AssignPredefinedSpaceRole] or [AssignCustomSpaceRole].
type SpaceBinding struct {
	// Space identifies the target space. Required. Accepts either a space name or ID.
	Space string
	// Role is the role to assign the bot user within this space. Build it with
	// AssignPredefinedSpaceRole (e.g. SpaceRoleMember) for a predefined role, or
	// AssignCustomSpaceRole (an existing custom RBAC role ID) for a custom role.
	// When zero, the server applies the predefined member role.
	Role SpaceRoleAssignment
}

// OrgBinding declares one organization that the service key's bot user should access,
// together with the spaces within that organization.
//
// The binding (OrgID) answers "which organization?". The role assignment (Role) answers
// "with what org-level role?". Spaces carries the per-space bindings nested under this org.
// Build Role with [AssignPredefinedOrgRole] or [AssignCustomOrgRole].
type OrgBinding struct {
	// OrgID is the HMAC-encoded ID of the organization. Required.
	OrgID string
	// Role is the role to assign the bot user within this organization. Build it with
	// AssignPredefinedOrgRole (e.g. OrgRoleReadOnly) for a predefined role, or
	// AssignCustomOrgRole (an existing custom RBAC role ID) for a custom role.
	// When zero, the server applies the predefined read-only role.
	Role OrgRoleAssignment
	// Spaces is the list of space bindings within this organization. At least one
	// space is required per organization. All spaces must belong to this organization.
	Spaces []SpaceBinding
}

// CreateRequest is the request shape for Client.Create (user keys).
type CreateRequest struct {
	// Name is the user-defined name for the API key.
	Name string
	// Description is an optional user-defined description. When empty, the key
	// is created without a description.
	Description string
	// ExpiresAt is the optional expiration timestamp. When zero, the key never
	// expires.
	ExpiresAt time.Time
}

// CreateServiceKeyRequest is the request shape for Client.CreateServiceKey.
type CreateServiceKeyRequest struct {
	// Name is the user-defined name for the API key.
	Name string
	// AccountRole is the optional account-level role assignment for the bot user.
	// Build it with AssignPredefinedAccountRole (one of the AccountRole values) or
	// AssignCustomAccountRole (an existing custom role ID). When zero, the server
	// applies the predefined member role.
	AccountRole AccountRoleAssignment
	// Orgs is the list of organization bindings for the bot user. At least one
	// organization with at least one space is required.
	Orgs []OrgBinding
	// Description is an optional user-defined description. When empty, the key
	// is created without a description.
	Description string
	// ExpiresAt is the optional expiration timestamp. When zero, the key never
	// expires.
	ExpiresAt time.Time
}

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// KeyType is an optional filter on API key type. When empty, both user and
	// service keys are returned.
	KeyType APIKeyType
	// Status is an optional filter on API key lifecycle status. When empty,
	// the server applies a default filter of APIKeyStatusActive.
	Status APIKeyStatus
	// Space, when non-empty, filters API keys to a single space. Accepts
	// either a space name or ID.
	Space string
	// UserID is an optional filter on the user who created the key (base64
	// unique identifier). When empty, results are not filtered by user.
	UserID string
	// Limit is the optional maximum number of items to return. When zero, the
	// SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor returned from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

// RevokeRequest is the request shape for Client.Revoke.
type RevokeRequest struct {
	// APIKeyID is the strict ID of the API key to revoke.
	APIKeyID string
}

// RefreshRequest is the request shape for Client.Refresh.
type RefreshRequest struct {
	// APIKeyID is the strict ID of the API key to refresh.
	APIKeyID string
	// ExpiresAt is the optional new expiration timestamp. When zero, the
	// refreshed key has no expiration (infinite lifetime).
	ExpiresAt time.Time
	// GracePeriodSeconds is the optional number of seconds the old key remains
	// valid after the refresh. When zero, the old key is invalidated immediately.
	GracePeriodSeconds int
}

type (
	APIKey                = generated.ApiKey
	ListApiKeysResponse   = generated.ListApiKeysResponse
	CreateApiKeyResponse  = generated.CreateApiKeyResponse
	RefreshApiKeyResponse = generated.RefreshApiKeyResponse

	// UserApiKeyCreated is the response body when a user API key is created.
	UserApiKeyCreated = generated.UserApiKeyCreated
	// ServiceApiKeyCreated is the response body when a service API key is created.
	ServiceApiKeyCreated = generated.ServiceApiKeyCreated

	// APIKeyType is the type of an API key (user or service). Shared across
	// responses, create requests, and list filters.
	APIKeyType = generated.ApiKeyType
	// APIKeyStatus is the lifecycle status of an API key.
	APIKeyStatus = generated.ApiKeyStatus

	// SpaceRoleAssignment is the role assignment for a service key's bot user within
	// a space. Build with AssignPredefinedSpaceRole or AssignCustomSpaceRole.
	SpaceRoleAssignment = generated.SpaceRoleAssignment
	// OrgRoleAssignment is the role assignment for a service key's bot user within
	// an organization. Build with AssignPredefinedOrgRole or AssignCustomOrgRole.
	OrgRoleAssignment = generated.OrganizationRoleAssignment
	// AccountRoleAssignment is the role assignment for a service key's bot user at
	// the account level. Build with AssignPredefinedAccountRole or AssignCustomAccountRole.
	AccountRoleAssignment = generated.UserRoleAssignment

	// PredefinedSpaceRole is the predefined variant of SpaceRoleAssignment.
	PredefinedSpaceRole = generated.PredefinedRoleAssignment
	// CustomSpaceRole is the custom variant of SpaceRoleAssignment.
	CustomSpaceRole = generated.CustomRoleAssignment
	// PredefinedOrgRole is the predefined variant of OrgRoleAssignment.
	PredefinedOrgRole = generated.OrganizationPredefinedRoleAssignment
	// CustomOrgRole is the custom variant of OrgRoleAssignment.
	CustomOrgRole = generated.OrganizationCustomRoleAssignment
	// PredefinedAccountRole is the predefined variant of AccountRoleAssignment.
	PredefinedAccountRole = generated.PredefinedUserRoleAssignment
	// CustomAccountRole is the custom variant of AccountRoleAssignment.
	CustomAccountRole = generated.CustomUserRoleAssignment

	// SpaceRole is the predefined space-level role for a service key's bot user.
	SpaceRole = generated.UserSpaceRole
	// OrgRole is the predefined organization-level role for a service key's bot user.
	OrgRole = generated.OrganizationRole
	// AccountRole is the predefined account-level role for a service key's bot user.
	AccountRole = generated.UserRole
)

const (
	APIKeyTypeService APIKeyType = generated.ApiKeyTypeSERVICE
	APIKeyTypeUser    APIKeyType = generated.ApiKeyTypeUSER

	APIKeyStatusActive  APIKeyStatus = generated.ApiKeyStatusACTIVE
	APIKeyStatusRevoked APIKeyStatus = generated.ApiKeyStatusREVOKED

	SpaceRoleAdmin    SpaceRole = generated.UserSpaceRoleADMIN
	SpaceRoleMember   SpaceRole = generated.UserSpaceRoleMEMBER
	SpaceRoleReadOnly SpaceRole = generated.UserSpaceRoleREADONLY

	OrgRoleAdmin    OrgRole = generated.OrganizationRoleADMIN
	OrgRoleMember   OrgRole = generated.OrganizationRoleMEMBER
	OrgRoleReadOnly OrgRole = generated.OrganizationRoleREADONLY

	AccountRoleAdmin  AccountRole = generated.UserRoleADMIN
	AccountRoleMember AccountRole = generated.UserRoleMEMBER
)

// AssignPredefinedSpaceRole builds a space role assignment for one of the predefined
// SpaceRole values (e.g. SpaceRoleMember). Pass the result as SpaceBinding.Role.
func AssignPredefinedSpaceRole(name SpaceRole) SpaceRoleAssignment {
	var r SpaceRoleAssignment
	// From* only marshals a fixed struct and sets the discriminator internally;
	// it cannot fail, so the error is safely discarded.
	_ = r.FromPredefinedRoleAssignment(PredefinedSpaceRole{Name: name})
	return r
}

// AssignCustomSpaceRole builds a space role assignment referencing an existing custom
// RBAC role by its ID (see the roles package). Pass the result as SpaceBinding.Role.
func AssignCustomSpaceRole(roleID string) SpaceRoleAssignment {
	var r SpaceRoleAssignment
	_ = r.FromCustomRoleAssignment(CustomSpaceRole{Id: roleID})
	return r
}

// AssignPredefinedOrgRole builds an org role assignment for one of the predefined
// OrgRole values (e.g. OrgRoleReadOnly). Pass the result as OrgBinding.Role.
func AssignPredefinedOrgRole(name OrgRole) OrgRoleAssignment {
	var r OrgRoleAssignment
	_ = r.FromOrganizationPredefinedRoleAssignment(PredefinedOrgRole{Name: name})
	return r
}

// AssignCustomOrgRole builds an org role assignment referencing an existing custom
// RBAC role by its ID. Pass the result as OrgBinding.Role.
func AssignCustomOrgRole(roleID string) OrgRoleAssignment {
	var r OrgRoleAssignment
	_ = r.FromOrganizationCustomRoleAssignment(CustomOrgRole{Id: roleID})
	return r
}

// AssignPredefinedAccountRole builds an account role assignment for one of the
// predefined AccountRole values (e.g. AccountRoleMember). Pass the result as
// CreateServiceKeyRequest.AccountRole.
func AssignPredefinedAccountRole(name AccountRole) AccountRoleAssignment {
	var r AccountRoleAssignment
	_ = r.FromPredefinedUserRoleAssignment(PredefinedAccountRole{Name: name})
	return r
}

// AssignCustomAccountRole builds an account role assignment referencing an existing
// custom RBAC role by its ID. Pass the result as CreateServiceKeyRequest.AccountRole.
func AssignCustomAccountRole(roleID string) AccountRoleAssignment {
	var r AccountRoleAssignment
	_ = r.FromCustomUserRoleAssignment(CustomAccountRole{Id: roleID})
	return r
}
