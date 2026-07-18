package annotationconfigs

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	AnnotationConfig            = generated.AnnotationConfig
	ListAnnotationConfigs       = generated.ListAnnotationConfigsResponse
	CategoricalAnnotationConfig = generated.CategoricalAnnotationConfig
	ContinuousAnnotationConfig  = generated.ContinuousAnnotationConfig
	FreeformAnnotationConfig    = generated.FreeformAnnotationConfig
	CategoricalAnnotationValue  = generated.CategoricalAnnotationValue
	OptimizationDirection       = generated.OptimizationDirection
	AnnotationConfigType        = generated.AnnotationConfigType
)

const (
	AnnotationConfigTypeCategorical AnnotationConfigType = generated.AnnotationConfigTypeCATEGORICAL
	AnnotationConfigTypeContinuous  AnnotationConfigType = generated.AnnotationConfigTypeCONTINUOUS
	AnnotationConfigTypeFreeform    AnnotationConfigType = generated.AnnotationConfigTypeFREEFORM

	OptimizationDirectionMaximize OptimizationDirection = generated.OptimizationDirectionMAXIMIZE
	OptimizationDirectionMinimize OptimizationDirection = generated.OptimizationDirectionMINIMIZE
	OptimizationDirectionNone     OptimizationDirection = generated.OptimizationDirectionNONE
)

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// Space, when non-empty, filters annotation configs to a single space.
	// Accepts either a space name or ID.
	Space string
	// Name, when non-empty, applies a case-insensitive substring filter on the
	// annotation config name.
	Name string
	// Limit is the optional maximum number of items to return. When zero, the
	// SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor returned from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

// GetRequest identifies the annotation config to fetch.
type GetRequest struct {
	// AnnotationConfig is the config's name or ID. Required.
	AnnotationConfig string
	// Space is the parent space's name or ID. Required when AnnotationConfig
	// is a name; ignored when AnnotationConfig is an ID.
	Space string
}

// UpdateCategoricalRequest describes a patch to an existing categorical
// annotation config, resolved by name or ID. All patch fields are optional:
// leave a field nil to preserve its current value.
type UpdateCategoricalRequest struct {
	// AnnotationConfig is the config's name or ID. Required.
	AnnotationConfig string
	// Space is the parent space's name or ID. Required when AnnotationConfig
	// is a name; ignored when AnnotationConfig is an ID.
	Space string
	// Name is optional. When non-nil, sets a new name (must be unique within
	// the space); when nil, the existing name is preserved.
	Name *string
	// OptimizationDirection is optional. When non-nil, sets a new
	// optimization direction; when nil, the existing value is preserved.
	OptimizationDirection *OptimizationDirection
	// Values is optional. When non-nil, replaces the full set of allowed
	// labels (2–100 items); when nil, the existing set is preserved.
	Values *[]CategoricalAnnotationValue
}

// UpdateContinuousRequest describes a patch to an existing continuous
// annotation config, resolved by name or ID. All patch fields are optional:
// leave a field nil to preserve its current value.
type UpdateContinuousRequest struct {
	// AnnotationConfig is the config's name or ID. Required.
	AnnotationConfig string
	// Space is the parent space's name or ID. Required when AnnotationConfig
	// is a name; ignored when AnnotationConfig is an ID.
	Space string
	// Name is optional. When non-nil, sets a new name (must be unique within
	// the space); when nil, the existing name is preserved.
	Name *string
	// MinimumScore is optional. When non-nil, sets a new lower bound of the
	// score range; when nil, the existing value is preserved.
	MinimumScore *float64
	// MaximumScore is optional. When non-nil, sets a new upper bound of the
	// score range; when nil, the existing value is preserved.
	MaximumScore *float64
	// OptimizationDirection is optional. When non-nil, sets a new
	// optimization direction; when nil, the existing value is preserved.
	OptimizationDirection *OptimizationDirection
}

// UpdateFreeformRequest describes a patch to an existing freeform annotation
// config, resolved by name or ID. All patch fields are optional: leave a
// field nil to preserve its current value.
type UpdateFreeformRequest struct {
	// AnnotationConfig is the config's name or ID. Required.
	AnnotationConfig string
	// Space is the parent space's name or ID. Required when AnnotationConfig
	// is a name; ignored when AnnotationConfig is an ID.
	Space string
	// Name is optional. When non-nil, sets a new name (must be unique within
	// the space); when nil, the existing name is preserved.
	Name *string
}

// DeleteRequest identifies the annotation config to delete.
type DeleteRequest struct {
	// AnnotationConfig is the config's name or ID. Required.
	AnnotationConfig string
	// Space is the parent space's name or ID. Required when AnnotationConfig
	// is a name; ignored when AnnotationConfig is an ID.
	Space string
}

// CreateContinuousRequest describes a new continuous annotation config, i.e.
// a numeric score within a fixed range.
type CreateContinuousRequest struct {
	// Space is the parent space's name or ID. Required.
	Space string
	// Name is the annotation config's name. Required.
	Name string
	// MinimumScore is the lower bound of the score range. Required.
	MinimumScore float64
	// MaximumScore is the upper bound of the score range. Required.
	MaximumScore float64
	// OptimizationDirection indicates which end of the score range is
	// considered better. When the zero value, the server applies its
	// default (`none`).
	OptimizationDirection OptimizationDirection
}

// CreateCategoricalRequest describes a new categorical annotation config,
// i.e. a fixed set of labeled values a scorer can choose from.
type CreateCategoricalRequest struct {
	// Space is the parent space's name or ID. Required.
	Space string
	// Name is the annotation config's name. Required.
	Name string
	// Values is the set of allowed labels. Required.
	Values []CategoricalAnnotationValue
	// OptimizationDirection indicates which values are considered better.
	// When the zero value, the server applies its default (`none`).
	OptimizationDirection OptimizationDirection
}

// CreateFreeformRequest describes a new freeform annotation config, i.e.
// open-ended text feedback with no predefined scale or set of values.
type CreateFreeformRequest struct {
	// Space is the parent space's name or ID. Required.
	Space string
	// Name is the annotation config's name. Required.
	Name string
}
