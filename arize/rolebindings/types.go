package rolebindings

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	RoleBinding = generated.RoleBinding

	// RoleBindingList is the cursor-paginated list response shape.
	RoleBindingList = generated.RoleBindingList

	// RoleBindingResourceType is the resource type for a role binding (SPACE or PROJECT).
	RoleBindingResourceType = generated.RoleBindingResourceType
)

const (
	RoleBindingResourceTypePROJECT RoleBindingResourceType = generated.RoleBindingResourceTypePROJECT
	RoleBindingResourceTypeSPACE   RoleBindingResourceType = generated.RoleBindingResourceTypeSPACE
)

// ListRequest is the request for listing role bindings for the authenticated
// user's account, with cursor-based pagination.
type ListRequest struct {
	// ResourceType filters bindings by resource type
	// (RoleBindingResourceTypeSPACE or RoleBindingResourceTypePROJECT).
	// Required — the zero value is rejected by the server.
	ResourceType RoleBindingResourceType
	// UserID is an optional filter on the assigned user (global user ID). When
	// empty, bindings are not filtered by user.
	UserID string
	// Limit is the optional maximum number of items to return. When zero, the
	// SDK applies a default of 100. Server max is 100.
	Limit int
	// Cursor is the optional opaque pagination cursor returned from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

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
