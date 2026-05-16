package resourcerestrictions

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	ResourceRestrictionResponse = generated.ResourceRestrictionResponse
	CreateRequest               = generated.ResourceRestrictionCreate

	// ResourceRestriction is the nested restriction record inside ResourceRestrictionResponse.
	ResourceRestriction = generated.ResourceRestriction
	// ResourceRestrictionResourceType is the type of the restricted resource.
	ResourceRestrictionResourceType = generated.ResourceRestrictionResourceType
)

const (
	ResourceRestrictionResourceTypePROJECT ResourceRestrictionResourceType = generated.ResourceRestrictionResourceTypePROJECT
)
