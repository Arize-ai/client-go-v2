package rolebindings

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	RoleBinding = generated.RoleBinding

	// RoleBindingResourceType is the resource type for a role binding (SPACE or PROJECT).
	RoleBindingResourceType = generated.RoleBindingResourceType
)

const (
	RoleBindingResourceTypePROJECT RoleBindingResourceType = generated.RoleBindingResourceTypePROJECT
	RoleBindingResourceTypeSPACE   RoleBindingResourceType = generated.RoleBindingResourceTypeSPACE
)

// GetRequest is the request for retrieving a single role binding.
type GetRequest struct {
	RoleBindingID string
}

// CreateRequest is the request for creating a role binding.
// All ID fields are strict IDs — name resolution is not performed.
type CreateRequest struct {
	RoleID       string
	UserID       string
	ResourceType RoleBindingResourceType
	ResourceID   string
}

// UpdateRequest is the request for updating an existing role binding.
type UpdateRequest struct {
	RoleBindingID string
	RoleID        string
}

// DeleteRequest is the request for deleting a role binding.
type DeleteRequest struct {
	RoleBindingID string
}
