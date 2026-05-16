package apikeys

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	ApiKey         = generated.ApiKey
	ApiKeyList     = generated.ApiKeyList
	ApiKeyCreated  = generated.ApiKeyCreated
	CreateRequest  = generated.CreateApiKeyRequestBody
	RefreshRequest = generated.RefreshApiKeyRequestBody
	ListParams     = generated.ApiKeysListParams

	// ApiKeyKeyType is the key type field on an ApiKey response.
	ApiKeyKeyType = generated.ApiKeyKeyType
	// ApiKeyStatus is the lifecycle status of an API key.
	ApiKeyStatus = generated.ApiKeyStatus
	// ApiKeyCreatedKeyType is the key type field on an ApiKeyCreated response.
	ApiKeyCreatedKeyType = generated.ApiKeyCreatedKeyType
	// ApiKeyCreateKeyType is the key_type field on CreateApiKeyRequest.
	ApiKeyCreateKeyType = generated.ApiKeyCreateKeyType
	// ListParamsKeyType is the key_type filter field on ListParams.
	ListParamsKeyType = generated.ApiKeysListParamsKeyType

	// ApiKeyRoles holds role assignments for the bot user created with a service key.
	ApiKeyRoles = generated.ApiKeyRoles
	// ApiKeyRolesAccountRole is the account-level role for a service key's bot user.
	ApiKeyRolesAccountRole = generated.ApiKeyRolesAccountRole
	// ApiKeyRolesOrgRole is the org-level role for a service key's bot user.
	ApiKeyRolesOrgRole = generated.ApiKeyRolesOrgRole
	// ApiKeyRolesSpaceRole is the space-level role for a service key's bot user.
	ApiKeyRolesSpaceRole = generated.ApiKeyRolesSpaceRole
)

const (
	ApiKeyKeyTypeService ApiKeyKeyType = generated.ApiKeyKeyTypeService
	ApiKeyKeyTypeUser    ApiKeyKeyType = generated.ApiKeyKeyTypeUser

	ApiKeyStatusActive  ApiKeyStatus = generated.ApiKeyStatusActive
	ApiKeyStatusDeleted ApiKeyStatus = generated.ApiKeyStatusDeleted

	ApiKeyCreatedKeyTypeService ApiKeyCreatedKeyType = generated.ApiKeyCreatedKeyTypeService
	ApiKeyCreatedKeyTypeUser    ApiKeyCreatedKeyType = generated.ApiKeyCreatedKeyTypeUser

	ApiKeyCreateKeyTypeService ApiKeyCreateKeyType = generated.ApiKeyCreateKeyTypeService
	ApiKeyCreateKeyTypeUser    ApiKeyCreateKeyType = generated.ApiKeyCreateKeyTypeUser

	ListParamsKeyTypeService ListParamsKeyType = generated.ApiKeysListParamsKeyTypeService
	ListParamsKeyTypeUser    ListParamsKeyType = generated.ApiKeysListParamsKeyTypeUser

	ApiKeyRolesAccountRoleAdmin  ApiKeyRolesAccountRole = generated.ApiKeyRolesAccountRoleAdmin
	ApiKeyRolesAccountRoleMember ApiKeyRolesAccountRole = generated.ApiKeyRolesAccountRoleMember

	ApiKeyRolesOrgRoleAdmin    ApiKeyRolesOrgRole = generated.ApiKeyRolesOrgRoleAdmin
	ApiKeyRolesOrgRoleMember   ApiKeyRolesOrgRole = generated.ApiKeyRolesOrgRoleMember
	ApiKeyRolesOrgRoleReadOnly ApiKeyRolesOrgRole = generated.ApiKeyRolesOrgRoleReadOnly

	ApiKeyRolesSpaceRoleAdmin    ApiKeyRolesSpaceRole = generated.ApiKeyRolesSpaceRoleAdmin
	ApiKeyRolesSpaceRoleMember   ApiKeyRolesSpaceRole = generated.ApiKeyRolesSpaceRoleMember
	ApiKeyRolesSpaceRoleReadOnly ApiKeyRolesSpaceRole = generated.ApiKeyRolesSpaceRoleReadOnly
)
