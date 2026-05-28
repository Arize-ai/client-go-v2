package aiintegrations

import "errors"

// ErrNoUpdateFields is returned by Update when no patch fields are set on
// UpdateRequest. Compare with errors.Is.
var ErrNoUpdateFields = errors.New("aiintegrations: at least one patch field must be provided")
