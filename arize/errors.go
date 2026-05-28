package arize

import (
	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
	"github.com/Arize-ai/client-go-v2/arize/internal/sdkconfig"
)

// Config-validation errors are defined in internal/sdkconfig and re-exported
// here so callers can compare against arize.ErrMissingAPIKey directly.
var (
	ErrMissingAPIKey             = sdkconfig.ErrMissingAPIKey
	ErrMultipleEndpointOverrides = sdkconfig.ErrMultipleEndpointOverrides
)

// ResourceNotFoundError is returned when a resource cannot be resolved by name.
// Distinct from NotFoundError, which represents an HTTP 404 response from the
// server. Callers can check via errors.As(err, &arize.ResourceNotFoundError{}).
type ResourceNotFoundError = resolve.ResourceNotFoundError

// AmbiguousNameError is returned when a resource name matches more than one
// resource (e.g. a space name shared across different organizations). The
// caller should pass a resource ID instead of the name to disambiguate.
type AmbiguousNameError = resolve.AmbiguousNameError

// Re-export error types from internal/apierrors so callers only need to import
// the root arize package for both client construction and error handling.

type APIError = apierrors.APIError
type BadRequestError = apierrors.BadRequestError
type UnauthorizedError = apierrors.UnauthorizedError
type ForbiddenError = apierrors.ForbiddenError
type NotFoundError = apierrors.NotFoundError
type ConflictError = apierrors.ConflictError
type UnprocessableEntityError = apierrors.UnprocessableEntityError
type RateLimitError = apierrors.RateLimitError
type ServerError = apierrors.ServerError
