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
	// the server applies APIKeySpaceRoleMember.
	SpaceRole APIKeySpaceRole
	// OrgRole is the optional organization-level role for the bot user. When
	// empty, the server applies APIKeyOrganizationRoleReadOnly.
	OrgRole APIKeyOrganizationRole
	// AccountRole is the optional account-level role for the bot user. When
	// empty, the server applies APIKeyAccountRoleMember.
	AccountRole APIKeyAccountRole
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

// DeleteRequest is the request shape for Client.Delete.
type DeleteRequest struct {
	// APIKeyID is the strict ID of the API key to delete.
	APIKeyID string
}

// RefreshRequest is the request shape for Client.Refresh.
type RefreshRequest struct {
	// APIKeyID is the strict ID of the API key to refresh.
	APIKeyID string
	// ExpiresAt is the optional new expiration timestamp. When zero, the
	// refreshed key has no expiration (infinite lifetime).
	ExpiresAt time.Time
}

type (
	APIKey        = generated.ApiKey
	APIKeyList    = generated.ApiKeyList
	APIKeyCreated = generated.ApiKeyCreated

	// APIKeyType is the type of an API key (user or service). Shared across
	// responses, create requests, and list filters.
	APIKeyType = generated.ApiKeyType
	// APIKeyStatus is the lifecycle status of an API key.
	APIKeyStatus = generated.ApiKeyStatus

	// APIKeyRoles holds role assignments for the bot user created with a service key.
	APIKeyRoles = generated.ApiKeyRoles
	// APIKeyAccountRole is the account-level role for a service key's bot user.
	APIKeyAccountRole = generated.ApiKeyAccountRole
	// APIKeyOrganizationRole is the org-level role for a service key's bot user.
	APIKeyOrganizationRole = generated.ApiKeyOrganizationRole
	// APIKeySpaceRole is the space-level role for a service key's bot user.
	APIKeySpaceRole = generated.ApiKeySpaceRole
)

const (
	APIKeyTypeService APIKeyType = generated.ApiKeyTypeService
	APIKeyTypeUser    APIKeyType = generated.ApiKeyTypeUser

	APIKeyStatusActive  APIKeyStatus = generated.ApiKeyStatusActive
	APIKeyStatusDeleted APIKeyStatus = generated.ApiKeyStatusDeleted

	APIKeyAccountRoleAdmin  APIKeyAccountRole = generated.ApiKeyAccountRoleAdmin
	APIKeyAccountRoleMember APIKeyAccountRole = generated.ApiKeyAccountRoleMember

	APIKeyOrganizationRoleAdmin    APIKeyOrganizationRole = generated.ApiKeyOrganizationRoleAdmin
	APIKeyOrganizationRoleMember   APIKeyOrganizationRole = generated.ApiKeyOrganizationRoleMember
	APIKeyOrganizationRoleReadOnly APIKeyOrganizationRole = generated.ApiKeyOrganizationRoleReadOnly

	APIKeySpaceRoleAdmin    APIKeySpaceRole = generated.ApiKeySpaceRoleAdmin
	APIKeySpaceRoleMember   APIKeySpaceRole = generated.ApiKeySpaceRoleMember
	APIKeySpaceRoleReadOnly APIKeySpaceRole = generated.ApiKeySpaceRoleReadOnly
)
