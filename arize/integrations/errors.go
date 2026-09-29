package integrations

import "errors"

// ErrNoUpdateFields is returned by UpdateAgent and UpdateLlm when no patch
// fields are set on the request. Compare with errors.Is.
var ErrNoUpdateFields = errors.New("integrations: at least one patch field must be provided")

// ErrInvalidProviderConfig is returned by CreateLLM when CreateLLMConfig does
// not have exactly one per-provider field set. Compare with errors.Is.
var ErrInvalidProviderConfig = errors.New("integrations: exactly one provider config must be set on CreateLLMConfig")

// ErrInvalidBedrockAuth is returned when a CreateAWSBedrockAuth does not have
// exactly one auth mode set. Compare with errors.Is.
var ErrInvalidBedrockAuth = errors.New("integrations: exactly one AWS Bedrock auth mode must be set")
