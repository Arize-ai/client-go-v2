package resourcerestrictions

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// ResourceRestriction is the restriction record returned by Restrict.
type ResourceRestriction = generated.ResourceRestriction

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
