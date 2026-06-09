package evaluators

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	// Evaluator is a single evaluator's metadata (without a version body).
	Evaluator = generated.Evaluator

	// EvaluatorList is the cursor-paginated list response shape.
	EvaluatorList = generated.EvaluatorList

	// EvaluatorWithVersion is an evaluator together with one of its versions
	// (the latest, or a specific version when GetRequest.VersionID is set).
	EvaluatorWithVersion = generated.EvaluatorWithVersion

	// EvaluatorVersion is a versioned snapshot of an evaluator's
	// configuration. It is a oneOf: read the active variant with
	// ValueByDiscriminator and a type switch over EvaluatorVersionTemplate /
	// EvaluatorVersionCode.
	EvaluatorVersion = generated.EvaluatorVersion

	// EvaluatorVersionList is the cursor-paginated version list response shape.
	EvaluatorVersionList = generated.EvaluatorVersionList

	// EvaluatorVersionTemplate is the template (LLM-based) variant of an
	// EvaluatorVersion, returned by EvaluatorVersion.AsEvaluatorVersionTemplate.
	EvaluatorVersionTemplate = generated.EvaluatorVersionTemplate

	// EvaluatorVersionCode is the code (managed or custom) variant of an
	// EvaluatorVersion, returned by EvaluatorVersion.AsEvaluatorVersionCode.
	EvaluatorVersionCode = generated.EvaluatorVersionCode

	// EvaluatorType is the evaluator kind: template or code.
	EvaluatorType = generated.EvaluatorType

	// TemplateConfig is the configuration for a template (LLM-based)
	// evaluator version. Assign it to VersionConfig.Template.
	TemplateConfig = generated.TemplateConfig

	// EvaluatorLlmConfig is the LLM configuration nested inside a TemplateConfig.
	EvaluatorLlmConfig = generated.EvaluatorLlmConfig

	// InvocationParams holds LLM invocation parameters (temperature, etc.).
	InvocationParams = generated.InvocationParams

	// ProviderParams holds provider-specific LLM parameters.
	ProviderParams = generated.ProviderParams

	// DataGranularity is the granularity level an evaluator runs at.
	DataGranularity = generated.DataGranularity

	// OptimizationDirection is the score-optimization direction for a
	// classification evaluator.
	OptimizationDirection = generated.OptimizationDirection

	// ManagedCodeConfig is the configuration for a managed (built-in) code
	// evaluator version. Assign it to CodeConfig.Managed; the SDK sets its
	// type discriminator automatically.
	ManagedCodeConfig = generated.ManagedCodeConfig

	// CustomCodeConfig is the configuration for a custom (user Python) code
	// evaluator version. Assign it to CodeConfig.Custom; the SDK sets its
	// type discriminator automatically.
	CustomCodeConfig = generated.CustomCodeConfig

	// ManagedCodeEvaluator is the name of a built-in managed code evaluator.
	ManagedCodeEvaluator = generated.ManagedCodeEvaluator

	// StaticParam is a typed static parameter passed to a code evaluator.
	StaticParam = generated.StaticParam
)

const (
	// EvaluatorTypeTemplate is a template (LLM-based) evaluator.
	EvaluatorTypeTemplate EvaluatorType = generated.EvaluatorTypeTemplate
	// EvaluatorTypeCode is a code (managed built-in or custom Python) evaluator.
	EvaluatorTypeCode EvaluatorType = generated.EvaluatorTypeCode
)

// VersionConfig is the configuration payload for a new evaluator version,
// used by both Create (initial version) and CreateVersion. Set exactly one of
// Template or Code; the evaluator's type is derived from which is set.
type VersionConfig struct {
	// CommitMessage describes the change introduced by this version. Required.
	CommitMessage string
	// Template configures a template (LLM-based) version. Set exactly one of
	// Template or Code.
	Template *TemplateConfig
	// Code configures a code version. Set exactly one of Template or Code;
	// within Code set exactly one of Managed or Custom.
	Code *CodeConfig
}

// CodeConfig is the oneOf input for a code evaluator version. Set exactly one
// of Managed or Custom; the SDK sets the inner type discriminator
// automatically.
type CodeConfig struct {
	// Managed configures a built-in managed code evaluator.
	Managed *ManagedCodeConfig
	// Custom configures a custom (user-supplied Python) code evaluator.
	Custom *CustomCodeConfig
}

// ListRequest is the request shape for Client.List.
type ListRequest struct {
	// Space, when non-empty, filters results by space. If the value is a
	// base64-encoded resource ID it is treated as a space ID (exact match);
	// otherwise it is used as a case-insensitive substring filter on the
	// space name (may match multiple spaces). When empty, no space filtering
	// is applied.
	Space string
	// Name, when non-empty, filters evaluators to those whose name contains
	// the given case-insensitive substring. When empty, no name filtering is
	// applied.
	Name string
	// Limit is the optional maximum number of items to return. When zero, the
	// SDK applies a default of 50. Server max is 100.
	Limit int
	// Cursor is the optional opaque pagination cursor returned from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

// GetRequest is the request shape for Client.Get.
type GetRequest struct {
	// Evaluator is the evaluator's name or ID. Required.
	Evaluator string
	// Space is the parent space's name or ID. Required when Evaluator is a
	// name; ignored when Evaluator is an ID.
	Space string
	// VersionID, when non-empty, returns that specific version. When empty,
	// the latest version is returned.
	VersionID string
}

// CreateRequest is the request shape for Client.Create. It creates an
// evaluator together with its initial version. The evaluator's type is
// derived from Version (Template -> template, Code -> code).
type CreateRequest struct {
	// Space is the parent space's name or ID. Required.
	Space string
	// Name is the evaluator's name; must be unique within the space. Required.
	Name string
	// Description is an optional description. When empty, no description is set.
	Description string
	// Version is the initial version's configuration. Required; set exactly
	// one of Version.Template or Version.Code.
	Version VersionConfig
}

// UpdateRequest is the request shape for Client.Update. Only non-nil patch
// fields are sent; nil fields are left unchanged on the server. Update returns
// ErrNoUpdateFields without contacting the server when no patch field is set.
type UpdateRequest struct {
	// Evaluator is the target evaluator's name or ID. Required.
	Evaluator string
	// Space is the parent space's name or ID. Required when Evaluator is a
	// name; ignored when Evaluator is an ID.
	Space string
	// Name, when non-nil, updates the evaluator's name.
	Name *string
	// Description, when non-nil, updates the evaluator's description.
	Description *string
}

// DeleteRequest is the request shape for Client.Delete.
type DeleteRequest struct {
	// Evaluator is the evaluator's name or ID. Required.
	Evaluator string
	// Space is the parent space's name or ID. Required when Evaluator is a
	// name; ignored when Evaluator is an ID.
	Space string
}

// ListVersionsRequest is the request shape for Client.ListVersions.
type ListVersionsRequest struct {
	// Evaluator is the evaluator's name or ID. Required.
	Evaluator string
	// Space is the parent space's name or ID. Required when Evaluator is a
	// name; ignored when Evaluator is an ID.
	Space string
	// Limit is the optional maximum number of items to return. When zero, the
	// SDK applies a default of 50. Server max is 100.
	Limit int
	// Cursor is the optional opaque pagination cursor returned from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

// CreateVersionRequest is the request shape for Client.CreateVersion. The new
// version's kind must match the parent evaluator's type.
type CreateVersionRequest struct {
	// Evaluator is the evaluator's name or ID. Required.
	Evaluator string
	// Space is the parent space's name or ID. Required when Evaluator is a
	// name; ignored when Evaluator is an ID.
	Space string
	// Version is the new version's configuration. Required; set exactly one of
	// Version.Template or Version.Code.
	Version VersionConfig
}

// GetVersionRequest is the request shape for Client.GetVersion.
type GetVersionRequest struct {
	// VersionID is the evaluator version's ID. Required. Version IDs are pure
	// IDs with no name resolution.
	VersionID string
}
