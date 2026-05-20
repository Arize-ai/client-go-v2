package organizations

import "errors"

// ErrNoUpdateFields is returned by Update when neither Name nor Description is
// provided. Compare with errors.Is.
var ErrNoUpdateFields = errors.New("organizations: at least one of name or description must be provided")
