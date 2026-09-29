package rolebindings

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	RoleBinding = generated.RoleBinding

	// ListRoleBindings is the cursor-paginated list response shape.
	ListRoleBindings = generated.ListRoleBindingsResponse

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
	// empty, bindings are not filtered by user. For a service key, pass its bot
	// user's ID (BotUser.ID from apikeys.CreatedServiceApiKey) to list that key's
	// bindings.
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
	RoleID string
	// UserID is the ID of the user to bind the role to. For a service key,
	// this is the ID of the key's bot user — not the ID of the person who
	// created the key. It is returned as BotUser.ID from
	// apikeys.CreatedServiceApiKey.
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
