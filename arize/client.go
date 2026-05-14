package arize

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/Arize-ai/client-go-v2/arize/apikeys"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/resourcerestrictions"
	"github.com/Arize-ai/client-go-v2/arize/rolebindings"
	"github.com/Arize-ai/client-go-v2/arize/spans"
)

// Client is the top-level client for the Arize REST API.
// Access each resource domain through the typed subclient fields.
type Client struct {
	cfg Config
	gen *generated.ClientWithResponses

	APIKeys              *apikeys.Client
	ResourceRestrictions *resourcerestrictions.Client
	Spans                *spans.Client
	RoleBindings         *rolebindings.Client
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

	transport := http.DefaultTransport
	if resolved.InsecureSkipVerify {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
		transport = t
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
		gen: gen,

		APIKeys:              apikeys.New(gen, resolved),
		ResourceRestrictions: resourcerestrictions.New(gen, resolved),
		Spans:                spans.New(gen, resolved),
		RoleBindings:         rolebindings.New(gen, resolved),
	}, nil
}

// Config returns the resolved configuration used to construct this client.
// Useful for debugging, telemetry, or accessing endpoint-derived values like
// FlightHost/FlightPort that subclients themselves may not yet expose.
func (c *Client) Config() Config {
	return c.cfg
}
