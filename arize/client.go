package arize

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Arize-ai/client-go-v2/arize/aiintegrations"
	"github.com/Arize-ai/client-go-v2/arize/annotationconfigs"
	"github.com/Arize-ai/client-go-v2/arize/auditlogs"
	"github.com/Arize-ai/client-go-v2/arize/annotationqueues"
	"github.com/Arize-ai/client-go-v2/arize/apikeys"
	"github.com/Arize-ai/client-go-v2/arize/datasets"
	"github.com/Arize-ai/client-go-v2/arize/evaluators"
	"github.com/Arize-ai/client-go-v2/arize/experiments"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/organizations"
	"github.com/Arize-ai/client-go-v2/arize/projects"
	"github.com/Arize-ai/client-go-v2/arize/prompts"
	"github.com/Arize-ai/client-go-v2/arize/resourcerestrictions"
	"github.com/Arize-ai/client-go-v2/arize/rolebindings"
	"github.com/Arize-ai/client-go-v2/arize/roles"
	"github.com/Arize-ai/client-go-v2/arize/spaces"
	"github.com/Arize-ai/client-go-v2/arize/spans"
	"github.com/Arize-ai/client-go-v2/arize/tasks"
	"github.com/Arize-ai/client-go-v2/arize/users"
)

// Client is the top-level client for the Arize REST API.
// Access each resource domain through the typed subclient fields.
type Client struct {
	cfg Config

	AIIntegrations       *aiintegrations.Client
	APIKeys              *apikeys.Client
	AuditLogs            *auditlogs.Client
	AnnotationConfigs    *annotationconfigs.Client
	AnnotationQueues     *annotationqueues.Client
	Datasets             *datasets.Client
	Evaluators           *evaluators.Client
	Experiments          *experiments.Client
	Organizations        *organizations.Client
	Projects             *projects.Client
	Prompts              *prompts.Client
	ResourceRestrictions *resourcerestrictions.Client
	RoleBindings         *rolebindings.Client
	Roles                *roles.Client
	Spaces               *spaces.Client
	Spans                *spans.Client
	Tasks                *tasks.Client
	Users                *users.Client
}

// NewClient constructs a Client from the provided Config.
// Config is resolved (env vars applied, defaults set) and validated before use.
func NewClient(cfg Config) (*Client, error) {
	resolved, err := cfg.Resolve()
	if err != nil {
		return nil, err
	}
	if err := resolved.Validate(); err != nil {
		return nil, err
	}

	transport, err := buildTransport(resolved)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Timeout:   resolved.HTTPTimeout,
		Transport: transport,
	}

	headersCopy := resolved.Headers()
	gen, err := generated.NewClientWithResponses(
		resolved.APIURL(),
		generated.WithHTTPClient(httpClient),
		generated.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			for k, v := range headersCopy {
				req.Header.Set(k, v)
			}
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		cfg: resolved,

		AIIntegrations:       aiintegrations.New(gen),
		APIKeys:              apikeys.New(gen),
		AuditLogs:            auditlogs.New(gen),
		AnnotationConfigs:    annotationconfigs.New(gen),
		AnnotationQueues:     annotationqueues.New(gen),
		Datasets:             datasets.New(gen),
		Evaluators:           evaluators.New(gen),
		Experiments:          experiments.New(gen),
		Organizations:        organizations.New(gen),
		Projects:             projects.New(gen),
		Prompts:              prompts.New(gen),
		ResourceRestrictions: resourcerestrictions.New(gen),
		RoleBindings:         rolebindings.New(gen),
		Roles:                roles.New(gen),
		Spaces:               spaces.New(gen),
		Spans:                spans.New(gen),
		Tasks:                tasks.New(gen),
		Users:                users.New(gen),
	}, nil
}

// Config returns the resolved configuration used to construct this client.
// Useful for debugging, telemetry, or accessing endpoint-derived values like
// FlightHost/FlightPort that subclients themselves may not yet expose.
func (c *Client) Config() Config {
	return c.cfg
}

// buildTransport constructs an http.RoundTripper from the resolved Config.
// Precedence:
//   - InsecureSkipVerify → skip TLS verification entirely (validated to be
//     mutually exclusive with SSLCACert).
//   - SSLCACert set → load the file into a custom cert pool.
//   - ProxyURL set → override the proxy for this transport.
//
// When none of the above apply, http.DefaultTransport is returned unchanged so
// that the default behaviour (system CAs, env-var proxy) is preserved.
func buildTransport(cfg Config) (http.RoundTripper, error) {
	needsCustomTransport := cfg.InsecureSkipVerify || cfg.SSLCACert != "" || cfg.ProxyURL != ""
	if !needsCustomTransport {
		return http.DefaultTransport, nil
	}

	t := http.DefaultTransport.(*http.Transport).Clone()

	if cfg.InsecureSkipVerify {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	} else if cfg.SSLCACert != "" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			pool = x509.NewCertPool()
		}
		pem, err := os.ReadFile(cfg.SSLCACert)
		if err != nil {
			return nil, fmt.Errorf("ssl_ca_cert: reading %q: %w", cfg.SSLCACert, err)
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("ssl_ca_cert: no valid PEM certificates found in %q", cfg.SSLCACert)
		}
		t.TLSClientConfig = &tls.Config{RootCAs: pool}
	}

	if cfg.ProxyURL != "" {
		u, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("proxy_url: invalid URL %q: %w", cfg.ProxyURL, err)
		}
		t.Proxy = http.ProxyURL(u)
	}

	return t, nil
}
