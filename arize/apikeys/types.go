package apikeys

import (
	"time"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

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
	// Space identifies the target space. Required. Accepts either a space name
	// or ID.
	Space string
	// Description is an optional user-defined description. When empty, the key
	// is created without a description.
	Description string
	// ExpiresAt is the optional expiration timestamp. When zero, the key never
	// expires.
	ExpiresAt time.Time
	// SpaceRole is the optional space-level role for the bot user. When empty,
	// the server applies ApiKeySpaceRoleMember.
	SpaceRole ApiKeySpaceRole
	// OrgRole is the optional organization-level role for the bot user. When
	// empty, the server applies ApiKeyOrganizationRoleReadOnly.
	OrgRole ApiKeyOrganizationRole
	// AccountRole is the optional account-level role for the bot user. When
	// empty, the server applies ApiKeyAccountRoleMember.
	AccountRole ApiKeyAccountRole
}

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// KeyType is an optional filter on API key type. When empty, both user and
	// service keys are returned.
	KeyType ApiKeyType
	// Status is an optional filter on API key lifecycle status. When empty,
	// the server applies a default filter of ApiKeyStatusActive.
	Status ApiKeyStatus
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

// DeleteRequest is the request shape for Client.Delete.
type DeleteRequest struct {
	// ApiKeyID is the strict ID of the API key to delete.
	ApiKeyID string
}

// RefreshRequest is the request shape for Client.Refresh.
type RefreshRequest struct {
	// ApiKeyID is the strict ID of the API key to refresh.
	ApiKeyID string
	// ExpiresAt is the optional new expiration timestamp. When zero, the
	// refreshed key has no expiration (infinite lifetime).
	ExpiresAt time.Time
}

type (
	ApiKey        = generated.ApiKey
	ApiKeyList    = generated.ApiKeyList
	ApiKeyCreated = generated.ApiKeyCreated

	// ApiKeyType is the type of an API key (user or service). Shared across
	// responses, create requests, and list filters.
	ApiKeyType = generated.ApiKeyType
	// ApiKeyStatus is the lifecycle status of an API key.
	ApiKeyStatus = generated.ApiKeyStatus

	// ApiKeyRoles holds role assignments for the bot user created with a service key.
	ApiKeyRoles = generated.ApiKeyRoles
	// ApiKeyAccountRole is the account-level role for a service key's bot user.
	ApiKeyAccountRole = generated.ApiKeyAccountRole
	// ApiKeyOrganizationRole is the org-level role for a service key's bot user.
	ApiKeyOrganizationRole = generated.ApiKeyOrganizationRole
	// ApiKeySpaceRole is the space-level role for a service key's bot user.
	ApiKeySpaceRole = generated.ApiKeySpaceRole
)

const (
	ApiKeyTypeService ApiKeyType = generated.ApiKeyTypeService
	ApiKeyTypeUser    ApiKeyType = generated.ApiKeyTypeUser

	ApiKeyStatusActive  ApiKeyStatus = generated.ApiKeyStatusActive
	ApiKeyStatusDeleted ApiKeyStatus = generated.ApiKeyStatusDeleted

	ApiKeyAccountRoleAdmin  ApiKeyAccountRole = generated.ApiKeyAccountRoleAdmin
	ApiKeyAccountRoleMember ApiKeyAccountRole = generated.ApiKeyAccountRoleMember

	ApiKeyOrganizationRoleAdmin    ApiKeyOrganizationRole = generated.ApiKeyOrganizationRoleAdmin
	ApiKeyOrganizationRoleMember   ApiKeyOrganizationRole = generated.ApiKeyOrganizationRoleMember
	ApiKeyOrganizationRoleReadOnly ApiKeyOrganizationRole = generated.ApiKeyOrganizationRoleReadOnly

	ApiKeySpaceRoleAdmin    ApiKeySpaceRole = generated.ApiKeySpaceRoleAdmin
	ApiKeySpaceRoleMember   ApiKeySpaceRole = generated.ApiKeySpaceRoleMember
	ApiKeySpaceRoleReadOnly ApiKeySpaceRole = generated.ApiKeySpaceRoleReadOnly
)
