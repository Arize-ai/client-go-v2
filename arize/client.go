package arize

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/Arize-ai/client-go-v2/arize/aiintegrations"
	"github.com/Arize-ai/client-go-v2/arize/annotationconfigs"
	"github.com/Arize-ai/client-go-v2/arize/annotationqueues"
	"github.com/Arize-ai/client-go-v2/arize/apikeys"
	"github.com/Arize-ai/client-go-v2/arize/datasets"
	"github.com/Arize-ai/client-go-v2/arize/evaluators"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/organizations"
	"github.com/Arize-ai/client-go-v2/arize/projects"
	"github.com/Arize-ai/client-go-v2/arize/prompts"
	"github.com/Arize-ai/client-go-v2/arize/resourcerestrictions"
	"github.com/Arize-ai/client-go-v2/arize/rolebindings"
	"github.com/Arize-ai/client-go-v2/arize/roles"
	"github.com/Arize-ai/client-go-v2/arize/spaces"
	"github.com/Arize-ai/client-go-v2/arize/spans"
)

// Client is the top-level client for the Arize REST API.
// Access each resource domain through the typed subclient fields.
type Client struct {
	cfg Config

	AIIntegrations       *aiintegrations.Client
	APIKeys              *apikeys.Client
	AnnotationConfigs    *annotationconfigs.Client
	AnnotationQueues     *annotationqueues.Client
	Datasets             *datasets.Client
	Evaluators           *evaluators.Client
	Organizations        *organizations.Client
	Projects             *projects.Client
	Prompts              *prompts.Client
	ResourceRestrictions *resourcerestrictions.Client
	RoleBindings         *rolebindings.Client
	Roles                *roles.Client
	Spaces               *spaces.Client
	Spans                *spans.Client
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

		AIIntegrations:       aiintegrations.New(gen),
		APIKeys:              apikeys.New(gen),
		AnnotationConfigs:    annotationconfigs.New(gen),
		AnnotationQueues:     annotationqueues.New(gen),
		Datasets:             datasets.New(gen),
		Evaluators:           evaluators.New(gen),
		Organizations:        organizations.New(gen),
		Projects:             projects.New(gen),
		Prompts:              prompts.New(gen),
		ResourceRestrictions: resourcerestrictions.New(gen),
		RoleBindings:         rolebindings.New(gen),
		Roles:                roles.New(gen),
		Spaces:               spaces.New(gen),
		Spans:                spans.New(gen),
	}, nil
}

// Config returns the resolved configuration used to construct this client.
// Useful for debugging, telemetry, or accessing endpoint-derived values like
// FlightHost/FlightPort that subclients themselves may not yet expose.
func (c *Client) Config() Config {
	return c.cfg
}
