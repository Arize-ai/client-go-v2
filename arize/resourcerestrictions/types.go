package resourcerestrictions

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, and enum types remain aliases to the generated wire shapes so
// callers can construct and assert on them without importing internal/generated.
type (
	// ResourceRestriction is the restriction record returned by Restrict and
	// listed by List.
	ResourceRestriction = generated.ResourceRestriction

	// ResourceRestrictionList is the paginated envelope returned by List. It
	// carries the ResourceRestrictions slice and Pagination metadata.
	ResourceRestrictionList = generated.ListResourceRestrictionsResponse

	// ResourceRestrictionType is the type of a restricted resource. Currently
	// only PROJECT is supported.
	ResourceRestrictionType = generated.ResourceRestrictionType
)

// ResourceRestrictionTypePROJECT restricts a project within a space. It is the
// only supported resource type today.
const ResourceRestrictionTypePROJECT ResourceRestrictionType = generated.ResourceRestrictionTypePROJECT

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// ResourceType is an optional filter on the restricted resource type. When
	// empty, restrictions of all supported resource types are returned
	// (currently only PROJECT).
	ResourceType ResourceRestrictionType
	// Limit is the optional maximum number of items to return. When zero, the
	// SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor returned from a previous
	// response (Pagination.NextCursor). When empty, results start from the
	// first page.
	Cursor string
}

// RestrictRequest is the request for marking a resource as restricted.
type RestrictRequest struct {
	// ResourceID is the unique identifier of the resource to restrict (e.g. a
	// project ID). Strict ID only — no name resolution. Must encode a project
	// resource ID; currently only PROJECT resources are supported.
	ResourceID string
}

// UnrestrictRequest is the request for removing a restriction from a resource.
type UnrestrictRequest struct {
	// ResourceID is the unique identifier of the restricted resource (e.g. a
	// project ID), not the ID of the restriction record itself — the API
	// dereferences from the restricted resource to the restriction it holds.
	ResourceID string
}
