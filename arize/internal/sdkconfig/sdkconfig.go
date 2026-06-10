// Package sdkconfig holds the canonical Config type and resolution logic for
// the Arize Go SDK. It is referenced by every subclient via the internal/
// import path, while the public arize package re-exports the types via aliases.
// Splitting the type out of the public package avoids an import cycle between
// arize and its subclients.
package sdkconfig

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

// sdkPackageName is reported in the "sdk-package-name" header.
const sdkPackageName = "arize"

// allowedHTTPSchemes is the set of schemes accepted for APIScheme and OTLPScheme.
var allowedHTTPSchemes = map[string]struct{}{
	"http":  {},
	"https": {},
}

// parseBoolEnv reads a bool from the named env var. Truthy values
// {1, true, yes, on} (case-insensitive) return true; any other non-empty
// value returns false. The second return is false if the env var is
// empty/unset, so callers can preserve explicit defaults.
func parseBoolEnv(name string) (val, set bool) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return false, false
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true, true
	default:
		return false, true
	}
}

// envStr returns the env value (trimmed) if set, otherwise def.
func envStr(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return def
}

// envInt returns the env value parsed as int if set, otherwise def. Returns
// an error mentioning the env-var name if parsing fails.
func envInt(name string, def int) (int, error) {
	v := os.Getenv(name)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s is not a valid integer: %q", name, v)
	}
	return n, nil
}

// envFloat returns the env value parsed as float64 if set, otherwise def.
// Returns an error mentioning the env-var name if parsing fails.
func envFloat(name string, def float64) (float64, error) {
	v := os.Getenv(name)
	if v == "" {
		return def, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is not a valid float: %q", name, v)
	}
	return f, nil
}

// Region represents an Arize deployment region.
type Region string

const (
	RegionUSCentral Region = "us-central-1a"
	RegionEUWest    Region = "eu-west-1a"
	RegionCACentral Region = "ca-central-1a"
	RegionUSEast    Region = "us-east-1b"
)

const (
	defaultAPIHost      = "api.arize.com"
	defaultAPIScheme    = "https"
	defaultOTLPHost     = "otlp.arize.com"
	defaultOTLPScheme   = "https"
	defaultFlightHost   = "flight.arize.com"
	defaultFlightPort   = 443
	defaultFlightScheme = "grpc+tls"

	defaultMaxHTTPPayloadSizeMB = 8.0
	defaultArizeDirectory       = "~/.arize"
	defaultMaxPastYears         = 5
)

// RegionEndpoints holds the per-region endpoint defaults.
type RegionEndpoints struct {
	APIHost    string
	OTLPHost   string
	FlightHost string
	FlightPort int
}

// knownRegions is the canonical list of valid Arize regions. The endpoints map
// and IsValidRegion both derive from this slice, so adding a new region is a
// one-line change.
var knownRegions = []Region{
	RegionUSCentral,
	RegionEUWest,
	RegionCACentral,
	RegionUSEast,
}

// defaultRegionEndpoints maps each known region to its default endpoint
// configuration. Kept unexported so callers must go through RegionEndpointsFor
// and cannot mutate the global map.
var defaultRegionEndpoints = func() map[Region]RegionEndpoints {
	m := make(map[Region]RegionEndpoints, len(knownRegions))
	for _, r := range knownRegions {
		m[r] = regionEndpoints(r)
	}
	return m
}()

func regionEndpoints(r Region) RegionEndpoints {
	return RegionEndpoints{
		APIHost:    "api." + string(r) + ".arize.com",
		OTLPHost:   "otlp." + string(r) + ".arize.com",
		FlightHost: "flight." + string(r) + ".arize.com",
		FlightPort: defaultFlightPort,
	}
}

// IsValidRegion reports whether r is one of the known Arize regions.
func IsValidRegion(r Region) bool {
	return slices.Contains(knownRegions, r)
}

// RegionEndpointsFor returns the default endpoint configuration for r.
// The second return is false if r is not a known region. The returned struct
// is a value copy; mutating it does not affect future lookups.
func RegionEndpointsFor(r Region) (RegionEndpoints, bool) {
	e, ok := defaultRegionEndpoints[r]
	return e, ok
}

var (
	ErrMissingAPIKey             = errors.New("api_key is required; set Config.APIKey or ARIZE_API_KEY env var")
	ErrMultipleEndpointOverrides = errors.New("only one of Region, SingleHost, or BaseDomain may be set")
)

// Config holds all configuration for the Arize client.
//
// The override fields (Region, SingleHost/SinglePort, BaseDomain) are
// mutually exclusive and, when set, rewrite APIHost/OTLPHost/FlightHost/FlightPort
// during Resolve. Some fields below (OTLP*, Flight*) are not yet wired into the
// REST-only client surface but are reserved for future Flight/OTLP support.
type Config struct {
	APIKey    string
	APIHost   string
	APIScheme string

	OTLPHost   string
	OTLPScheme string

	FlightHost   string
	FlightPort   int
	FlightScheme string

	Region     Region
	SingleHost string
	SinglePort int
	BaseDomain string

	InsecureSkipVerify bool // default: false (TLS verified); env: ARIZE_REQUEST_VERIFY=false
	HTTPTimeout        time.Duration

	// MaxHTTPPayloadSizeMB caps outbound HTTP payload size in megabytes.
	// Default 8 MB; env: ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB.
	MaxHTTPPayloadSizeMB float64

	// ArizeDirectory is the local directory for SDK files (cache, logs, etc.).
	// Default "~/.arize"; env: ARIZE_DIRECTORY.
	ArizeDirectory string

	// DisableCaching turns off local caching. The zero value (false) means
	// caching is enabled — the field is named for the negative so the safe
	// default doesn't require an explicit setter. Env: ARIZE_ENABLE_CACHING.
	DisableCaching bool

	// MaxPastYears caps how far in the past prediction timestamps may be.
	// Default 5; env: ARIZE_MAX_PAST_YEARS. Resolve logs a warning if this
	// differs from the default.
	MaxPastYears int
}

// Resolve returns a new Config with env vars applied to zero values, defaults
// set, and endpoint overrides (Region/SingleHost/SinglePort/BaseDomain) applied.
// Returns an error if a port-valued env var cannot be parsed as an integer.
//
// SinglePort only rewrites FlightPort, never APIHost. For an on-prem REST-only
// deployment served on a non-default port, set ARIZE_API_HOST to "host:port"
// or set Config.APIHost directly.
func (c Config) Resolve() (Config, error) {
	var err error
	if c.APIKey == "" {
		c.APIKey = os.Getenv("ARIZE_API_KEY")
	}
	if c.APIHost == "" {
		c.APIHost = envStr("ARIZE_API_HOST", defaultAPIHost)
	}
	if c.APIScheme == "" {
		c.APIScheme = envStr("ARIZE_API_SCHEME", defaultAPIScheme)
	}
	if c.OTLPHost == "" {
		c.OTLPHost = envStr("ARIZE_OTLP_HOST", defaultOTLPHost)
	}
	if c.OTLPScheme == "" {
		c.OTLPScheme = envStr("ARIZE_OTLP_SCHEME", defaultOTLPScheme)
	}
	if c.FlightHost == "" {
		c.FlightHost = envStr("ARIZE_FLIGHT_HOST", defaultFlightHost)
	}
	if c.FlightPort == 0 {
		if c.FlightPort, err = envInt("ARIZE_FLIGHT_PORT", defaultFlightPort); err != nil {
			return Config{}, err
		}
	}
	if c.FlightScheme == "" {
		c.FlightScheme = envStr("ARIZE_FLIGHT_SCHEME", defaultFlightScheme)
	}
	if c.Region == "" {
		c.Region = Region(os.Getenv("ARIZE_REGION"))
	}
	if c.SingleHost == "" {
		c.SingleHost = os.Getenv("ARIZE_SINGLE_HOST")
	}
	if c.SinglePort == 0 {
		if c.SinglePort, err = envInt("ARIZE_SINGLE_PORT", 0); err != nil {
			return Config{}, err
		}
	}
	if c.BaseDomain == "" {
		c.BaseDomain = os.Getenv("ARIZE_BASE_DOMAIN")
	}
	// ARIZE_REQUEST_VERIFY=true means verify TLS; InsecureSkipVerify is named
	// for the negative (matching tls.Config) so the env flag is inverted here.
	if !c.InsecureSkipVerify {
		if verify, ok := parseBoolEnv("ARIZE_REQUEST_VERIFY"); ok {
			c.InsecureSkipVerify = !verify
		}
	}
	if c.HTTPTimeout == 0 {
		c.HTTPTimeout = 30 * time.Second
	}
	if c.MaxHTTPPayloadSizeMB == 0 {
		if c.MaxHTTPPayloadSizeMB, err = envFloat("ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB", defaultMaxHTTPPayloadSizeMB); err != nil {
			return Config{}, err
		}
	}
	if c.ArizeDirectory == "" {
		c.ArizeDirectory = envStr("ARIZE_DIRECTORY", defaultArizeDirectory)
	}
	// ARIZE_ENABLE_CACHING=true means caching on; DisableCaching is named for
	// the negative (zero value = caching enabled) so the env flag is inverted.
	if !c.DisableCaching {
		if enable, ok := parseBoolEnv("ARIZE_ENABLE_CACHING"); ok {
			c.DisableCaching = !enable
		}
	}
	if c.MaxPastYears == 0 {
		if c.MaxPastYears, err = envInt("ARIZE_MAX_PAST_YEARS", defaultMaxPastYears); err != nil {
			return Config{}, err
		}
	}
	if c.MaxPastYears != defaultMaxPastYears {
		log.Printf("max_past_years is set to %d (default: %d). This setting allows timestamps older than the default limit. Please contact Arize support to enable custom timestamp limits for your account.",
			c.MaxPastYears, defaultMaxPastYears)
	}

	// Apply endpoint overrides. Later blocks intentionally overwrite earlier
	// ones when multiple are set, but Validate enforces that only one is set
	// at a time.
	if c.BaseDomain != "" {
		c.APIHost = "api." + c.BaseDomain
		c.OTLPHost = "otlp." + c.BaseDomain
		c.FlightHost = "flight." + c.BaseDomain
	}
	if c.SingleHost != "" {
		c.APIHost = c.SingleHost
		c.OTLPHost = c.SingleHost
		c.FlightHost = c.SingleHost
	}
	if c.SinglePort != 0 {
		c.FlightPort = c.SinglePort
	}
	if c.Region != "" {
		if endpoints, ok := defaultRegionEndpoints[c.Region]; ok {
			c.APIHost = endpoints.APIHost
			c.OTLPHost = endpoints.OTLPHost
			c.FlightHost = endpoints.FlightHost
			c.FlightPort = endpoints.FlightPort
		}
	}

	return c, nil
}

// Validate returns an error if the resolved Config is invalid.
func (c Config) Validate() error {
	if c.APIKey == "" {
		return ErrMissingAPIKey
	}
	overrides := 0
	if c.Region != "" {
		overrides++
	}
	if c.SingleHost != "" || c.SinglePort != 0 {
		overrides++
	}
	if c.BaseDomain != "" {
		overrides++
	}
	if overrides > 1 {
		var conflicts []string
		if c.Region != "" {
			conflicts = append(conflicts, fmt.Sprintf("Region=%q", c.Region))
		}
		if c.SingleHost != "" {
			conflicts = append(conflicts, fmt.Sprintf("SingleHost=%q", c.SingleHost))
		}
		if c.SinglePort != 0 {
			conflicts = append(conflicts, fmt.Sprintf("SinglePort=%d", c.SinglePort))
		}
		if c.BaseDomain != "" {
			conflicts = append(conflicts, fmt.Sprintf("BaseDomain=%q", c.BaseDomain))
		}
		return fmt.Errorf("%w: got %s", ErrMultipleEndpointOverrides, strings.Join(conflicts, ", "))
	}
	if c.SinglePort != 0 && (c.SinglePort < 1 || c.SinglePort > 65535) {
		return fmt.Errorf("single_port must be 1-65535, got %d", c.SinglePort)
	}
	if c.FlightPort < 1 || c.FlightPort > 65535 {
		return fmt.Errorf("flight_port must be 1-65535, got %d", c.FlightPort)
	}
	if _, ok := allowedHTTPSchemes[strings.ToLower(c.APIScheme)]; !ok {
		return fmt.Errorf("api_scheme must be one of [http https], got %q", c.APIScheme)
	}
	if _, ok := allowedHTTPSchemes[strings.ToLower(c.OTLPScheme)]; !ok {
		return fmt.Errorf("otlp_scheme must be one of [http https], got %q", c.OTLPScheme)
	}
	if c.Region != "" && !IsValidRegion(c.Region) {
		return fmt.Errorf("region %q is not a known region", c.Region)
	}
	if c.MaxHTTPPayloadSizeMB < 1 {
		return fmt.Errorf("max_http_payload_size_mb must be >= 1, got %v", c.MaxHTTPPayloadSizeMB)
	}
	if c.MaxPastYears < 1 {
		return fmt.Errorf("max_past_years must be >= 1, got %d", c.MaxPastYears)
	}
	return nil
}

// maskSecret returns the first 6 chars of s followed by "***", "***" if the
// secret is too short to keep any prefix, or empty if s itself is empty.
func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 6 {
		return "***"
	}
	return s[:6] + "***"
}

// String implements fmt.Stringer and masks sensitive fields (APIKey) so
// fmt.Println / %v / %+v / log lines do not leak the API key.
func (c Config) String() string {
	return fmt.Sprintf(
		"Config{APIKey:%s, APIHost:%q, APIScheme:%q, OTLPHost:%q, OTLPScheme:%q, "+
			"FlightHost:%q, FlightPort:%d, FlightScheme:%q, "+
			"Region:%q, SingleHost:%q, SinglePort:%d, BaseDomain:%q, "+
			"InsecureSkipVerify:%t, HTTPTimeout:%s, "+
			"MaxHTTPPayloadSizeMB:%v, ArizeDirectory:%q, DisableCaching:%t, MaxPastYears:%d}",
		maskSecret(c.APIKey), c.APIHost, c.APIScheme, c.OTLPHost, c.OTLPScheme,
		c.FlightHost, c.FlightPort, c.FlightScheme,
		c.Region, c.SingleHost, c.SinglePort, c.BaseDomain,
		c.InsecureSkipVerify, c.HTTPTimeout,
		c.MaxHTTPPayloadSizeMB, c.ArizeDirectory, c.DisableCaching, c.MaxPastYears,
	)
}

// APIURL returns the base URL for REST API calls.
func (c Config) APIURL() string {
	return fmt.Sprintf("%s://%s", c.APIScheme, c.APIHost)
}

// Headers returns HTTP headers to attach to every request. The authorization
// header carries the API key with a "Bearer " prefix, as required by the Arize
// API gateway (apiproxy AuthMiddleware), which rejects any authorization header
// lacking the prefix.
func (c Config) Headers() map[string]string {
	return map[string]string{
		"authorization":    "Bearer " + c.APIKey,
		"sdk-language":     "go",
		"language-version": runtime.Version(),
		"sdk-version":      SDKVersion,
		"sdk-package-name": sdkPackageName,
	}
}
