package prompts

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, version, and nested types remain aliases to the generated
// wire shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	Prompt            = generated.Prompt
	PromptWithVersion = generated.PromptWithVersion
	PromptList        = generated.PromptList
	PromptVersion     = generated.PromptVersion
	PromptVersionList = generated.PromptVersionList

	// PromptVersionLabels is the set of labels currently on a prompt version.
	PromptVersionLabels = generated.PromptVersionLabelsResponse

	// PromptVersionCreate is the initial/new version configuration supplied when
	// creating a prompt.
	PromptVersionCreate = generated.PromptVersionCreateRequest

	// LLMMessage is a single message in a prompt template.
	LLMMessage = generated.LLMMessage
	// MessageRole is the role of a prompt message's author.
	MessageRole = generated.MessageRole
	// InvocationParams configures the LLM invocation (temperature, etc.).
	InvocationParams = generated.InvocationParams
	// ProviderParams configures provider-specific parameters.
	ProviderParams = generated.ProviderParams

	// InputVariableFormat declares how prompt variables are interpolated.
	InputVariableFormat = generated.InputVariableFormat
	// LlmProvider is the LLM provider (e.g. "openai").
	LlmProvider = generated.LlmProvider
)

// MessageRole enum values.
const (
	MessageRoleSystem    = generated.MessageRoleSystem
	MessageRoleUser      = generated.MessageRoleUser
	MessageRoleAssistant = generated.MessageRoleAssistant
	MessageRoleTool      = generated.MessageRoleTool
)

// InputVariableFormat enum values.
const (
	InputVariableFormatFString  = generated.InputVariableFormatFString
	InputVariableFormatMustache = generated.InputVariableFormatMustache
	InputVariableFormatNone     = generated.InputVariableFormatNone
)

// LlmProvider enum values.
const (
	LlmProviderAnthropic   = generated.LlmProviderAnthropic
	LlmProviderAwsBedrock  = generated.LlmProviderAwsBedrock
	LlmProviderAzureOpenAi = generated.LlmProviderAzureOpenAi
	LlmProviderCustom      = generated.LlmProviderCustom
	LlmProviderOpenAi      = generated.LlmProviderOpenAi
	LlmProviderVertexAi    = generated.LlmProviderVertexAi
)

// ListRequest holds optional filters for listing prompts.
type ListRequest struct {
	// Space, when non-empty, filters results by space. If the value is a
	// base64-encoded resource ID it is treated as a space ID (exact match);
	// otherwise it is used as a case-insensitive substring filter on the
	// space name (may match multiple spaces).
	Space string
	// Name is an optional case-insensitive substring filter on the prompt name.
	// When empty, results are not filtered by name.
	Name string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the server applies its default page size.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous response's
	// pagination.next_cursor. When empty, results start from the first page.
	Cursor string
}

// GetRequest identifies the prompt to fetch.
type GetRequest struct {
	// Prompt accepts either a prompt name or ID. Required.
	Prompt string
	// Space accepts either a space name or ID. Required when Prompt is a name;
	// ignored when Prompt is an ID.
	Space string
	// VersionID is optional. When non-empty, returns this specific prompt
	// version; when empty, returns the latest version. Mutually exclusive with
	// Label.
	VersionID string
	// Label is optional. When non-empty, returns the version pointed to by this
	// label (e.g. "production"); when empty, returns the latest version.
	// Mutually exclusive with VersionID.
	Label string
}

// CreateRequest describes a new prompt with its initial version.
type CreateRequest struct {
	// Space accepts either a space name or ID and identifies the parent space.
	// Required.
	Space string
	// Name is the prompt's name (must be unique within the space). Required.
	Name string
	// Description is optional. When empty, the prompt is created without a
	// description.
	Description string
	// Version is the initial prompt version (commit message, messages, provider,
	// etc.). Required.
	Version PromptVersionCreate
}

// UpdateRequest identifies the prompt to update and the fields to patch.
type UpdateRequest struct {
	// Prompt accepts either a prompt name or ID. Required.
	Prompt string
	// Space accepts either a space name or ID. Required when Prompt is a name;
	// ignored when Prompt is an ID.
	Space string
	// Description is optional. When non-nil, sets a new description (pass a
	// pointer to an empty string to clear the existing description); when nil,
	// the existing description is preserved.
	Description *string
}

// DeleteRequest identifies the prompt to delete.
type DeleteRequest struct {
	// Prompt accepts either a prompt name or ID. Required.
	Prompt string
	// Space accepts either a space name or ID. Required when Prompt is a name;
	// ignored when Prompt is an ID.
	Space string
}

// ListVersionsRequest identifies the prompt (resolved by name or ID) and
// pagination options for listing its versions.
type ListVersionsRequest struct {
	// Prompt accepts either a prompt name or ID. Required.
	Prompt string
	// Space accepts either a space name or ID. Required when Prompt is a name;
	// ignored when Prompt is an ID.
	Space string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the server applies its default page size.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous response's
	// pagination.next_cursor. When empty, results start from the first page.
	Cursor string
}

// CreateVersionRequest identifies the prompt and describes the new version.
type CreateVersionRequest struct {
	// Prompt accepts either a prompt name or ID. Required.
	Prompt string
	// Space accepts either a space name or ID. Required when Prompt is a name;
	// ignored when Prompt is an ID.
	Space string
	// CommitMessage is the version's commit message. Required.
	CommitMessage string
	// Provider is the LLM provider (e.g. "openai"). Required.
	Provider LlmProvider
	// Messages is the prompt messages list. Required.
	Messages []LLMMessage
	// InputVariableFormat is optional and declares how prompt variables are
	// interpolated. When empty, the server defaults to f_string.
	InputVariableFormat InputVariableFormat
	// Model is optional and selects a specific provider model. When empty, no
	// default model is set on the version.
	Model string
	// InvocationParams is optional and configures invocation params
	// (temperature, etc.). When nil, no invocation params are set.
	InvocationParams *InvocationParams
	// ProviderParams is optional and configures provider-specific params. When
	// nil, no provider params are set.
	ProviderParams *ProviderParams
}

// GetVersionRequest identifies a single prompt version by its ID.
type GetVersionRequest struct {
	// VersionID is the strict ID of the prompt version. No name resolution is
	// performed.
	VersionID string
}

// GetVersionByLabelRequest identifies the prompt and the label name to look up.
type GetVersionByLabelRequest struct {
	// Prompt accepts either a prompt name or ID. Required.
	Prompt string
	// Space accepts either a space name or ID. Required when Prompt is a name;
	// ignored when Prompt is an ID.
	Space string
	// LabelName is the label to resolve (e.g. "production"). Required.
	LabelName string
}

// SetVersionLabelsRequest assigns labels to a specific prompt version.
type SetVersionLabelsRequest struct {
	// VersionID is the strict ID of the prompt version. No name resolution is
	// performed.
	VersionID string
	// Labels is the set of label names to assign to the version. Replaces all
	// existing labels; pass an empty slice to remove all labels.
	Labels []string
}

// DeleteVersionLabelRequest removes a label from a specific prompt version.
type DeleteVersionLabelRequest struct {
	// VersionID is the strict ID of the prompt version. No name resolution is
	// performed.
	VersionID string
	// LabelName is the label to remove. Required.
	LabelName string
}
