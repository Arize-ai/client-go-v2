package projects

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response and list types remain aliases to the generated wire shapes.
type (
	Project      = generated.Project
	ListProjects = generated.ListProjectsResponse
)

// GetRequest selects a single project.
type GetRequest struct {
	// Project accepts either a project name or ID.
	Project string
	// Space accepts either a space name or ID. Required when Project is a
	// name; ignored when Project is an ID.
	Space string
}

// CreateRequest is the request body for creating a new project.
type CreateRequest struct {
	// Name of the project (must be unique within the space).
	Name string
	// Space is the parent space for the new project. Accepts either a space
	// name or ID.
	Space string
}

// DeleteRequest selects a single project to delete.
type DeleteRequest struct {
	// Project accepts either a project name or ID.
	Project string
	// Space accepts either a space name or ID. Required when Project is a
	// name; ignored when Project is an ID.
	Space string
}

// ListRequest holds optional filters for listing projects.
type ListRequest struct {
	// Space, when non-empty, filters results by space. If the value is a
	// base64-encoded resource ID it is treated as a space ID (exact match);
	// otherwise it is used as a case-insensitive substring filter on the
	// space name (may match multiple spaces).
	Space string
	// Name is an optional case-insensitive substring filter on the project
	// name. When empty, results are not filtered by name.
	Name string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// UpdateRequest selects a project to rename.
type UpdateRequest struct {
	// Project accepts either a project name or ID.
	Project string
	// Space accepts either a space name or ID. Required when Project is a
	// name; ignored when Project is an ID.
	Space string
	// Name is the new name for the project. Must be unique within the space.
	Name string
}
