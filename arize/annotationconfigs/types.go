package annotationconfigs

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	AnnotationConfig            = generated.AnnotationConfig
	AnnotationConfigList        = generated.AnnotationConfigList
	CategoricalAnnotationConfig = generated.CategoricalAnnotationConfig
	ContinuousAnnotationConfig  = generated.ContinuousAnnotationConfig
	FreeformAnnotationConfig    = generated.FreeformAnnotationConfig
	CategoricalAnnotationValue  = generated.CategoricalAnnotationValue
	OptimizationDirection       = generated.OptimizationDirection
	AnnotationConfigType        = generated.AnnotationConfigType
)

const (
	AnnotationConfigTypeCategorical AnnotationConfigType = generated.AnnotationConfigTypeCategorical
	AnnotationConfigTypeContinuous  AnnotationConfigType = generated.AnnotationConfigTypeContinuous
	AnnotationConfigTypeFreeform    AnnotationConfigType = generated.AnnotationConfigTypeFreeform

	OptimizationDirectionMaximize OptimizationDirection = generated.OptimizationDirectionMaximize
	OptimizationDirectionMinimize OptimizationDirection = generated.OptimizationDirectionMinimize
	OptimizationDirectionNone     OptimizationDirection = generated.OptimizationDirectionNone
)

// AsCategorical returns the config's categorical variant and true if the
// discriminator is "categorical". Returns the zero value and false otherwise —
// including when the discriminator names another variant or JSON parsing fails.
func AsCategorical(ac AnnotationConfig) (CategoricalAnnotationConfig, bool) {
	d, err := ac.Discriminator()
	if err != nil || d != string(AnnotationConfigTypeCategorical) {
		return CategoricalAnnotationConfig{}, false
	}
	v, err := ac.AsCategoricalAnnotationConfig()
	if err != nil {
		return CategoricalAnnotationConfig{}, false
	}
	return v, true
}

// AsContinuous returns the config's continuous variant and true if the
// discriminator is "continuous". Returns the zero value and false otherwise.
func AsContinuous(ac AnnotationConfig) (ContinuousAnnotationConfig, bool) {
	d, err := ac.Discriminator()
	if err != nil || d != string(AnnotationConfigTypeContinuous) {
		return ContinuousAnnotationConfig{}, false
	}
	v, err := ac.AsContinuousAnnotationConfig()
	if err != nil {
		return ContinuousAnnotationConfig{}, false
	}
	return v, true
}

// AsFreeform returns the config's freeform variant and true if the
// discriminator is "freeform". Returns the zero value and false otherwise.
func AsFreeform(ac AnnotationConfig) (FreeformAnnotationConfig, bool) {
	d, err := ac.Discriminator()
	if err != nil || d != string(AnnotationConfigTypeFreeform) {
		return FreeformAnnotationConfig{}, false
	}
	v, err := ac.AsFreeformAnnotationConfig()
	if err != nil {
		return FreeformAnnotationConfig{}, false
	}
	return v, true
}

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// Space, when non-empty, filters annotation configs to a single space.
	// Accepts either a space name or ID.
	Space string
	// Name, when non-empty, applies a case-insensitive substring filter on the
	// annotation config name.
	Name string
	// Limit is the optional maximum number of items to return. When zero, the
	// server applies its default page size.
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

// DeleteRequest identifies the annotation config to delete.
type DeleteRequest struct {
	// AnnotationConfig is the config's name or ID. Required.
	AnnotationConfig string
	// Space is the parent space's name or ID. Required when AnnotationConfig
	// is a name; ignored when AnnotationConfig is an ID.
	Space string
}

// CreateRequest describes a new annotation config. Type-specific fields:
//   - Categorical: Values is required
//   - Continuous: MinimumScore and MaximumScore are required
//   - Freeform: no extra fields
type CreateRequest struct {
	// Space is the parent space's name or ID. Required.
	Space string
	// Name is the annotation config's name. Required.
	Name string
	// Type selects the annotation config variant. Required.
	Type AnnotationConfigType
	// OptimizationDirection (categorical/continuous only) indicates which way
	// is "better". When the zero value, the server applies its default
	// (`none`).
	OptimizationDirection OptimizationDirection
	// MinimumScore (continuous only) is the lower bound of the score range.
	// Required for continuous.
	MinimumScore float64
	// MaximumScore (continuous only) is the upper bound of the score range.
	// Required for continuous.
	MaximumScore float64
	// Values (categorical only) is the set of allowed labels. Required for
	// categorical.
	Values []CategoricalAnnotationValue
}
