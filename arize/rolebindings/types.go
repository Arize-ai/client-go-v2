package rolebindings

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	RoleBinding         = generated.RoleBinding
	RoleBindingResponse = generated.RoleBindingResponse
	CreateRequest       = generated.CreateRoleBindingRequestBody
	UpdateRequest       = generated.UpdateRoleBindingRequestBody

	// RoleBindingResourceType is the resource type for a role binding (SPACE or PROJECT).
	RoleBindingResourceType = generated.RoleBindingResourceType
)

const (
	RoleBindingResourceTypePROJECT RoleBindingResourceType = generated.RoleBindingResourceTypePROJECT
	RoleBindingResourceTypeSPACE   RoleBindingResourceType = generated.RoleBindingResourceTypeSPACE
)
