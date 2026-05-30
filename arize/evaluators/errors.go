package evaluators

import "errors"

// ErrNoUpdateFields is returned by Update when no patch fields are set on
// UpdateRequest. Compare with errors.Is.
var ErrNoUpdateFields = errors.New("evaluators: at least one patch field must be provided")

// ErrConflictingVersionConfig is returned by Create and CreateVersion when both
// VersionConfig.Template and VersionConfig.Code are set; exactly one is
// required. Compare with errors.Is.
var ErrConflictingVersionConfig = errors.New("evaluators: VersionConfig.Template and VersionConfig.Code are mutually exclusive; set exactly one")

// ErrConflictingCodeConfig is returned by Create and CreateVersion when both
// CodeConfig.Managed and CodeConfig.Custom are set; exactly one is required.
// Compare with errors.Is.
var ErrConflictingCodeConfig = errors.New("evaluators: CodeConfig.Managed and CodeConfig.Custom are mutually exclusive; set exactly one")
