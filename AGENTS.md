# AGENTS.md — Arize Go SDK v2

Public Go SDK for the Arize AI platform. Module path: `github.com/Arize-ai/client-go-v2`.

## Layout

```
sdk/go/v2/
├── arize/                          # Public package — everything users import
│   ├── client.go                   # Client (top-level)
│   ├── config.go                   # Config (re-exports sdkconfig.Config)
│   ├── errors.go                   # Public typed-error aliases
│   ├── regions.go                  # Public Region API + RegionEndpointsFor lookup
│   ├── version.go                  # SDK version string
│   └── internal/
│       ├── apierrors/              # CheckResponse + typed HTTP errors
│       ├── generated/              # ⚠ AUTO-GENERATED OpenAPI client — never hand-edit
│       ├── prerelease/             # Pre-release warning helpers
│       ├── resolve/                # By-name resolvers (Dataset/Experiment/Project)
│       └── sdkconfig/              # Canonical Config + Resolve/Validate logic
```

The `internal/sdkconfig` split exists to break an import cycle: subclient packages need `Config` and the root `arize` package needs to alias its types — keeping the canonical type under `internal/` lets both sides import it without cycling. The public `arize` package re-exports types via Go type aliases.

## Conventions

### Errors

- **Compare errors with `errors.Is` / `errors.As`, never with `==` or `!=`.** Direct comparison breaks the moment any caller wraps with `fmt.Errorf("...: %w", err)`. Example:
  ```go
  // Good
  if errors.Is(err, arize.ErrMissingAPIKey) { ... }
  var nfe *arize.NotFoundError
  if errors.As(err, &nfe) { ... }

  // Bad — fragile, breaks under wrapping
  if err == arize.ErrMissingAPIKey { ... }
  ```
- Sentinel errors live in `internal/sdkconfig` (config) and `internal/apierrors` (HTTP), and are re-exported from the public `arize` package.
- HTTP errors are typed (`BadRequestError`, `NotFoundError`, …) and all embed `APIError`. Callers can match on a specific status class or unwrap to the base.

### Mirroring Python v8

This SDK tracks the semantics of `sdk/python/arize/v8`. When in doubt about defaults, env-var names, validation, or endpoint resolution, check the Python source first. Naming should follow Go idioms (e.g. `Client` not `ArizeClient` at struct-decl sites, `InsecureSkipVerify` matching `tls.Config` instead of inverting to `RequestVerify`), but env vars and behavior stay aligned.

**Source-of-truth files for config parity:**
- `sdk/python/arize/v8/src/arize/config.py` — `SDKConfiguration` dataclass, resolution, validation, repr/masking
- `sdk/python/arize/v8/src/arize/constants/config.py` — env-var names and default values

**Specific parity rules already baked into `internal/sdkconfig/sdkconfig.go`:**
- `Headers()` matches Python verbatim: `authorization` (lowercase, **raw API key, no `Bearer ` prefix**), `sdk-language`, `language-version` (`runtime.Version()`), `sdk-version`, `sdk-package-name` (=`"arize"`).
- Env-var booleans use `parseBoolEnv` (truthy set `{1, true, yes, on}` case-insensitive). Mirrors `_parse_bool` in Python.
- Inverted-bool convention: when Python defaults a bool to `True` (e.g. `request_verify`, `enable_caching`), the Go field is named for the negative side (`InsecureSkipVerify`, `DisableCaching`) so the Go zero value matches Python's default.
- Secret masking: `Config.String()` masks any field considered sensitive (today: `APIKey`). Add new sensitive fields to both the struct and `String()`.
- Multi-override validation error wraps the sentinel via `%w` and names the conflicting fields — keeps `errors.Is(err, ErrMultipleEndpointOverrides)` working while giving richer diagnostics (matches Python's error message style).

Before changing config behavior, re-read both Python files and update the parity tests in `sdk/go/v2/arize/config_test.go` (`TestConfig_Headers_*`, `TestConfig_Resolve_RequestVerifyEnv_*`, etc.).

### Public API surface

- The public `arize` package re-exports only what end users need. Helpers used solely by subclients (e.g. `apierrors.CheckResponse`) live under `internal/` and stay there.
- No screaming-snake-case identifiers (`REGION_ENDPOINTS` is wrong; `defaultRegionEndpoints` or a `RegionEndpointsFor` lookup is right).
- No mutable exported globals — expose lookup functions that return copies, not the underlying map.
- Don't keep dead/speculative fields. Add them when the feature lands.

### Tests

- Test package is `arize_test` (or `<pkg>_test` for internal packages) — black-box tests against the public API.
- Behavior tests for an internal helper belong in that internal package's `_test.go`, not in the public-package tests.

## Build & Test

```bash
cd sdk/go/v2
go build ./...
go test ./... -count=1
```

Bazel build files (`BUILD.bazel`) coexist with `go.mod`; both must stay in sync when files are added or moved.
