package webhooks_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/webhooks"
)

func webhookID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Webhook:1:" + suffix))
}

func orgID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Organization:1:" + suffix))
}

func subscriptionID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("WebhookSubscription:1:" + suffix))
}

func promptID(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Prompt:1:" + suffix))
}

func newTestClient(t *testing.T, h http.HandlerFunc) *arize.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := arize.NewClient(arize.Config{APIKey: "key", APIHost: srv.Listener.Addr().String(), APIScheme: "http"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// wireCreate mirrors the JSON shape of the create request body so tests can
// assert what the SDK sent without importing internal/generated.
type wireCreate struct {
	OrganizationID string            `json:"organization_id"`
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	Description    *string           `json:"description"`
	AuthType       *string           `json:"auth_type"`
	AuthToken      *string           `json:"auth_token"`
	TimeoutMs      *int              `json:"timeout_ms"`
	Headers        map[string]string `json:"headers"`
}

type wireCreateSubscription struct {
	WebhookID  string `json:"webhook_id"`
	SourceType string `json:"source_type"`
	SourceID   string `json:"source_id"`
	Event      string `json:"event"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// listOneWebhook serves a webhooks list page holding exactly one webhook, used
// by the name-resolution cases.
func listOneWebhook(t *testing.T, w http.ResponseWriter, r *http.Request, wantOrg, name, id string) {
	t.Helper()
	if got := r.URL.Query().Get("org_id"); got != wantOrg {
		t.Errorf("org_id: want %q, got %q", wantOrg, got)
	}
	if got := r.URL.Query().Get("name"); got != name {
		t.Errorf("name: want %q, got %q", name, got)
	}
	writeJSON(w, 200, webhooks.ListWebhooks{
		Webhooks:   []webhooks.Webhook{{Id: id, Name: name}},
		Pagination: arize.PaginationMetadata{HasMore: false},
	})
}

func TestWebhooks(t *testing.T) {
	updatedName := "hook-renamed"
	emptyString := ""
	newURL := "https://example.com/new"
	newToken := "Bearer new-token"
	newTimeout := 5000
	newHeaders := map[string]string{"X-Env": "prod"}

	tests := []struct {
		name    string
		handler http.HandlerFunc
		invoke  func(ctx context.Context, c *arize.Client) (any, error)
		check   func(t *testing.T, got any, err error)
	}{
		{
			name: "List",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v2/webhooks" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				for _, k := range []string{"org_id", "name", "cursor"} {
					if r.URL.Query().Has(k) {
						t.Errorf("%s should be omitted when unset: %v", k, r.URL.Query())
					}
				}
				if got := r.URL.Query().Get("limit"); got != "50" {
					t.Errorf("limit should default to 50, got %q", got)
				}
				writeJSON(w, 200, webhooks.ListWebhooks{
					Webhooks:   []webhooks.Webhook{{Id: "wh-1", Name: "hook-1"}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.List(ctx, webhooks.ListRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*webhooks.ListWebhooks)
				if len(resp.Webhooks) != 1 || resp.Webhooks[0].Id != "wh-1" {
					t.Errorf("unexpected webhooks: %+v", resp.Webhooks)
				}
			},
		},
		{
			name: "List_Filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if q.Get("org_id") != orgID("o1") || q.Get("name") != "prod" || q.Get("limit") != "25" || q.Get("cursor") != "c1" {
					t.Errorf("query: %v", q)
				}
				writeJSON(w, 200, webhooks.ListWebhooks{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.List(ctx, webhooks.ListRequest{Organization: orgID("o1"), Name: "prod", Limit: 25, Cursor: "c1"})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "List_OrganizationByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/organizations":
					if got := r.URL.Query().Get("name"); got != "acme" {
						t.Errorf("organization name filter: %q", got)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"organizations":[{"id":"` + orgID("acme") + `","name":"acme"}],"pagination":{"has_more":false}}`))
				case "/v2/webhooks":
					if got := r.URL.Query().Get("org_id"); got != orgID("acme") {
						t.Errorf("org_id: want %q, got %q", orgID("acme"), got)
					}
					writeJSON(w, 200, webhooks.ListWebhooks{Pagination: arize.PaginationMetadata{HasMore: false}})
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(500)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.List(ctx, webhooks.ListRequest{Organization: "acme"})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Get",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v2/webhooks/"+webhookID("wh-1") {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				writeJSON(w, 200, webhooks.Webhook{Id: webhookID("wh-1"), Name: "hook-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: webhookID("wh-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if wh := got.(*webhooks.Webhook); wh.Name != "hook-1" {
					t.Errorf("unexpected name: %s", wh.Name)
				}
			},
		},
		{
			name: "Get_ByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/v2/webhooks":
					listOneWebhook(t, w, r, orgID("o1"), "hook-1", webhookID("wh-1"))
				case r.URL.Path == "/v2/webhooks/"+webhookID("wh-1"):
					writeJSON(w, 200, webhooks.Webhook{Id: webhookID("wh-1"), Name: "hook-1"})
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(500)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: "hook-1", Organization: orgID("o1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if wh := got.(*webhooks.Webhook); wh.Id != webhookID("wh-1") {
					t.Errorf("unexpected id: %s", wh.Id)
				}
			},
		},
		{
			name: "Get_NameWithoutOrganization",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("server should not be called: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(500)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: "hook-1"})
			},
			check: func(t *testing.T, _ any, err error) {
				var rnfe *arize.ResourceNotFoundError
				if !errors.As(err, &rnfe) {
					t.Fatalf("want *arize.ResourceNotFoundError, got %T: %v", err, err)
				}
				if !strings.Contains(rnfe.Hint, "organization") {
					t.Errorf("hint should mention organization: %q", rnfe.Hint)
				}
			},
		},
		{
			name: "Get_NameNotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, 200, webhooks.ListWebhooks{
					Webhooks:   []webhooks.Webhook{{Id: webhookID("wh-2"), Name: "hook-1-staging"}},
					Pagination: arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: "hook-1", Organization: orgID("o1")})
			},
			check: func(t *testing.T, _ any, err error) {
				var rnfe *arize.ResourceNotFoundError
				if !errors.As(err, &rnfe) {
					t.Fatalf("want *arize.ResourceNotFoundError, got %T: %v", err, err)
				}
				if len(rnfe.Available) != 1 || rnfe.Available[0] != "hook-1-staging" {
					t.Errorf("Available: want [hook-1-staging], got %v", rnfe.Available)
				}
			},
		},
		{
			name: "Create",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v2/webhooks" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				var body wireCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode create body: %v", err)
				}
				if body.OrganizationID != orgID("o1") || body.Name != "hook-new" || body.URL != "https://example.com/hook" {
					t.Errorf("required fields: %+v", body)
				}
				if body.Description == nil || *body.Description != "ci events" {
					t.Errorf("description: %v", body.Description)
				}
				if body.AuthType == nil || *body.AuthType != "BEARER" {
					t.Errorf("auth_type: %v", body.AuthType)
				}
				if body.AuthToken == nil || *body.AuthToken != "Bearer tok" {
					t.Errorf("auth_token: %v", body.AuthToken)
				}
				if body.TimeoutMs == nil || *body.TimeoutMs != 10000 {
					t.Errorf("timeout_ms: %v", body.TimeoutMs)
				}
				if body.Headers["X-Env"] != "prod" {
					t.Errorf("headers: %v", body.Headers)
				}
				writeJSON(w, 201, webhooks.CreateWebhook{Id: webhookID("wh-new"), Name: "hook-new", AuthType: webhooks.WebhookAuthTypeBearer})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Create(ctx, webhooks.CreateRequest{
					Organization: orgID("o1"),
					Name:         "hook-new",
					URL:          "https://example.com/hook",
					Description:  "ci events",
					AuthType:     webhooks.WebhookAuthTypeBearer,
					AuthToken:    "Bearer tok",
					TimeoutMs:    10000,
					Headers:      map[string]string{"X-Env": "prod"},
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*webhooks.CreateWebhook)
				if resp.Id != webhookID("wh-new") {
					t.Errorf("unexpected id: %s", resp.Id)
				}
				if resp.SigningSecret != nil {
					t.Errorf("signing_secret should be absent for BEARER: %q", *resp.SigningSecret)
				}
			},
		},
		{
			name: "Create_MinimalOmitsOptionalFields",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode create body: %v", err)
				}
				for _, k := range []string{"description", "auth_type", "auth_token", "timeout_ms", "headers"} {
					if _, ok := body[k]; ok {
						t.Errorf("%s should be omitted when unset, got %s", k, body[k])
					}
				}
				writeJSON(w, 201, webhooks.CreateWebhook{Id: webhookID("wh-new"), Name: "hook-new"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Create(ctx, webhooks.CreateRequest{
					Organization: orgID("o1"), Name: "hook-new", URL: "https://example.com/hook",
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Create_OrganizationByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/organizations":
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"organizations":[{"id":"` + orgID("acme") + `","name":"acme"}],"pagination":{"has_more":false}}`))
				case "/v2/webhooks":
					var body wireCreate
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("decode create body: %v", err)
					}
					if body.OrganizationID != orgID("acme") {
						t.Errorf("organization_id: want %q, got %q", orgID("acme"), body.OrganizationID)
					}
					writeJSON(w, 201, webhooks.CreateWebhook{Id: webhookID("wh-new"), Name: "hook-new"})
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(500)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Create(ctx, webhooks.CreateRequest{Organization: "acme", Name: "hook-new", URL: "https://example.com/hook"})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Create_HMACReturnsSigningSecret",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body wireCreate
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode create body: %v", err)
				}
				if body.AuthType == nil || *body.AuthType != "HMAC_SHA256" {
					t.Errorf("auth_type: %v", body.AuthType)
				}
				secret := "whsec_secret"
				hint := "whsec_…cret"
				writeJSON(w, 201, webhooks.CreateWebhook{
					Id: webhookID("wh-hmac"), Name: "hook-hmac",
					AuthType:          webhooks.WebhookAuthTypeHMACSHA256,
					SigningSecret:     &secret,
					SigningSecretHint: &hint,
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Create(ctx, webhooks.CreateRequest{
					Organization: orgID("o1"), Name: "hook-hmac", URL: "https://example.com/hook",
					AuthType: webhooks.WebhookAuthTypeHMACSHA256,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*webhooks.CreateWebhook)
				if resp.SigningSecret == nil || *resp.SigningSecret != "whsec_secret" {
					t.Errorf("signing_secret: %v", resp.SigningSecret)
				}
				if resp.AuthType != webhooks.WebhookAuthTypeHMACSHA256 {
					t.Errorf("auth_type: %s", resp.AuthType)
				}
			},
		},
		{
			name: "Update",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch || r.URL.Path != "/v2/webhooks/"+webhookID("wh-1") {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got := string(body["name"]); got != `"hook-renamed"` {
					t.Errorf("name: got %s", got)
				}
				for _, k := range []string{"description", "url", "auth_token", "timeout_ms", "headers"} {
					if _, ok := body[k]; ok {
						t.Errorf("%s should be omitted when nil, got %s", k, body[k])
					}
				}
				writeJSON(w, 200, webhooks.Webhook{Id: webhookID("wh-1"), Name: updatedName})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Update(ctx, webhooks.UpdateRequest{Webhook: webhookID("wh-1"), Name: &updatedName})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if wh := got.(*webhooks.Webhook); wh.Name != updatedName {
					t.Errorf("unexpected name: %s", wh.Name)
				}
			},
		},
		{
			name: "Update_AllFields",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Name        string            `json:"name"`
					Description string            `json:"description"`
					URL         string            `json:"url"`
					AuthToken   string            `json:"auth_token"`
					TimeoutMs   int               `json:"timeout_ms"`
					Headers     map[string]string `json:"headers"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Name != updatedName || body.Description != "new desc" || body.URL != newURL ||
					body.AuthToken != newToken || body.TimeoutMs != newTimeout || body.Headers["X-Env"] != "prod" {
					t.Errorf("body: %+v", body)
				}
				writeJSON(w, 200, webhooks.Webhook{Id: webhookID("wh-1"), Name: updatedName})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				desc := "new desc"
				return c.Webhooks.Update(ctx, webhooks.UpdateRequest{
					Webhook:     webhookID("wh-1"),
					Name:        &updatedName,
					Description: &desc,
					URL:         &newURL,
					AuthToken:   &newToken,
					TimeoutMs:   &newTimeout,
					Headers:     &newHeaders,
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Update_ClearDescriptionAndHeaders",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got, ok := body["description"]; !ok || string(got) != "null" {
					t.Errorf("description: want JSON null, got %s", got)
				}
				if got, ok := body["headers"]; !ok || string(got) != "{}" {
					t.Errorf("headers: want {}, got %s", got)
				}
				writeJSON(w, 200, webhooks.Webhook{Id: webhookID("wh-1"), Name: "hook-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				empty := map[string]string{}
				return c.Webhooks.Update(ctx, webhooks.UpdateRequest{Webhook: webhookID("wh-1"), Description: &emptyString, Headers: &empty})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Update_NilHeadersMapSendsEmptyObject",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got, ok := body["headers"]; !ok || string(got) != "{}" {
					t.Errorf("headers: want {}, got %s", got)
				}
				writeJSON(w, 200, webhooks.Webhook{Id: webhookID("wh-1"), Name: "hook-1"})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				var nilHeaders map[string]string
				return c.Webhooks.Update(ctx, webhooks.UpdateRequest{Webhook: webhookID("wh-1"), Headers: &nilHeaders})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Update_ByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/v2/webhooks":
					listOneWebhook(t, w, r, orgID("o1"), "hook-1", webhookID("wh-1"))
				case r.Method == http.MethodPatch && r.URL.Path == "/v2/webhooks/"+webhookID("wh-1"):
					writeJSON(w, 200, webhooks.Webhook{Id: webhookID("wh-1"), Name: updatedName})
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Update(ctx, webhooks.UpdateRequest{Webhook: "hook-1", Organization: orgID("o1"), Name: &updatedName})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Update_NoFields",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("server should not be called: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(500)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Update(ctx, webhooks.UpdateRequest{Webhook: webhookID("wh-1")})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, webhooks.ErrNoUpdateFields) {
					t.Fatalf("want ErrNoUpdateFields, got %v", err)
				}
			},
		},
		{
			name: "Delete",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete || r.URL.Path != "/v2/webhooks/"+webhookID("wh-1") {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Webhooks.Delete(ctx, webhooks.DeleteRequest{Webhook: webhookID("wh-1")})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Delete_ByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/v2/webhooks":
					listOneWebhook(t, w, r, orgID("o1"), "hook-1", webhookID("wh-1"))
				case r.Method == http.MethodDelete && r.URL.Path == "/v2/webhooks/"+webhookID("wh-1"):
					w.WriteHeader(204)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Webhooks.Delete(ctx, webhooks.DeleteRequest{Webhook: "hook-1", Organization: orgID("o1")})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "Test",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v2/webhooks/"+webhookID("wh-1")+"/test" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				msg := "connection refused"
				writeJSON(w, 200, webhooks.TestWebhook{StatusCode: 502, ErrorMessage: &msg})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Test(ctx, webhooks.TestRequest{Webhook: webhookID("wh-1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*webhooks.TestWebhook)
				if resp.StatusCode != 502 || resp.ErrorMessage == nil || *resp.ErrorMessage != "connection refused" {
					t.Errorf("unexpected response: %+v", resp)
				}
			},
		},
		{
			name: "Test_ByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/v2/webhooks":
					listOneWebhook(t, w, r, orgID("o1"), "hook-1", webhookID("wh-1"))
				case r.Method == http.MethodPost && r.URL.Path == "/v2/webhooks/"+webhookID("wh-1")+"/test":
					writeJSON(w, 200, webhooks.TestWebhook{StatusCode: 200})
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Test(ctx, webhooks.TestRequest{Webhook: "hook-1", Organization: orgID("o1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if resp := got.(*webhooks.TestWebhook); resp.StatusCode != 200 {
					t.Errorf("status code: want 200, got %d", resp.StatusCode)
				}
			},
		},
		{
			name: "ListDeliveryAttempts",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v2/webhooks/"+webhookID("wh-1")+"/delivery-attempts" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				if q := r.URL.Query(); q.Get("limit") != "500" || q.Get("cursor") != "c2" {
					t.Errorf("query: %v", q)
				}
				status := 200
				writeJSON(w, 200, webhooks.ListWebhookDeliveryAttempts{
					DeliveryAttempts: []webhooks.WebhookDeliveryAttempt{{EventId: "evt-1", AttemptNumber: 1, StatusCode: &status}},
					Pagination:       arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.ListDeliveryAttempts(ctx, webhooks.ListDeliveryAttemptsRequest{Webhook: webhookID("wh-1"), Limit: 500, Cursor: "c2"})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*webhooks.ListWebhookDeliveryAttempts)
				if len(resp.DeliveryAttempts) != 1 || resp.DeliveryAttempts[0].EventId != "evt-1" {
					t.Errorf("unexpected attempts: %+v", resp.DeliveryAttempts)
				}
			},
		},
		{
			name: "ListDeliveryAttempts_ByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/webhooks":
					listOneWebhook(t, w, r, orgID("o1"), "hook-1", webhookID("wh-1"))
				case "/v2/webhooks/" + webhookID("wh-1") + "/delivery-attempts":
					writeJSON(w, 200, webhooks.ListWebhookDeliveryAttempts{Pagination: arize.PaginationMetadata{HasMore: false}})
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(500)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.ListDeliveryAttempts(ctx, webhooks.ListDeliveryAttemptsRequest{Webhook: "hook-1", Organization: orgID("o1")})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "ListSubscriptions",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v2/webhook-subscriptions" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				for _, k := range []string{"source_type", "source_id", "cursor"} {
					if r.URL.Query().Has(k) {
						t.Errorf("%s should be omitted when unset: %v", k, r.URL.Query())
					}
				}
				if got := r.URL.Query().Get("limit"); got != "50" {
					t.Errorf("limit should default to 50, got %q", got)
				}
				writeJSON(w, 200, webhooks.ListWebhookSubscriptions{
					Subscriptions: []webhooks.WebhookSubscription{{Id: subscriptionID("s1"), Event: webhooks.WebhookEventTypePromptVersionCreated}},
					Pagination:    arize.PaginationMetadata{HasMore: false},
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.ListSubscriptions(ctx, webhooks.ListSubscriptionsRequest{})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				resp := got.(*webhooks.ListWebhookSubscriptions)
				if len(resp.Subscriptions) != 1 || resp.Subscriptions[0].Id != subscriptionID("s1") {
					t.Errorf("unexpected subscriptions: %+v", resp.Subscriptions)
				}
			},
		},
		{
			name: "ListSubscriptions_Filters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if q.Get("source_type") != "PROMPT" || q.Get("source_id") != promptID("p1") || q.Get("limit") != "10" || q.Get("cursor") != "c3" {
					t.Errorf("query: %v", q)
				}
				writeJSON(w, 200, webhooks.ListWebhookSubscriptions{Pagination: arize.PaginationMetadata{HasMore: false}})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.ListSubscriptions(ctx, webhooks.ListSubscriptionsRequest{
					SourceType: webhooks.WebhookSourceTypePrompt, SourceID: promptID("p1"), Limit: 10, Cursor: "c3",
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "ListSubscriptions_SourceTypeWithoutID",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("server should not be called: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(500)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.ListSubscriptions(ctx, webhooks.ListSubscriptionsRequest{SourceType: webhooks.WebhookSourceTypeEvaluator})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, webhooks.ErrUnpairedSourceFilter) {
					t.Fatalf("want ErrUnpairedSourceFilter, got %v", err)
				}
			},
		},
		{
			name: "ListSubscriptions_SourceIDWithoutType",
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("server should not be called: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(500)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.ListSubscriptions(ctx, webhooks.ListSubscriptionsRequest{SourceID: promptID("p1")})
			},
			check: func(t *testing.T, _ any, err error) {
				if !errors.Is(err, webhooks.ErrUnpairedSourceFilter) {
					t.Fatalf("want ErrUnpairedSourceFilter, got %v", err)
				}
			},
		},
		{
			name: "CreateSubscription",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v2/webhook-subscriptions" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				var body wireCreateSubscription
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.WebhookID != webhookID("wh-1") || body.SourceType != "PROMPT" || body.SourceID != promptID("p1") || body.Event != "PROMPT_VERSION_LABELED" {
					t.Errorf("body: %+v", body)
				}
				writeJSON(w, 201, webhooks.WebhookSubscription{
					Id: subscriptionID("s1"), WebhookId: webhookID("wh-1"),
					SourceType: webhooks.WebhookSourceTypePrompt, SourceId: promptID("p1"),
					Event: webhooks.WebhookEventTypePromptVersionLabeled,
				})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.CreateSubscription(ctx, webhooks.CreateSubscriptionRequest{
					Webhook:    webhookID("wh-1"),
					SourceType: webhooks.WebhookSourceTypePrompt,
					SourceID:   promptID("p1"),
					Event:      webhooks.WebhookEventTypePromptVersionLabeled,
				})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				sub := got.(*webhooks.WebhookSubscription)
				if sub.Id != subscriptionID("s1") || sub.Event != webhooks.WebhookEventTypePromptVersionLabeled {
					t.Errorf("unexpected subscription: %+v", sub)
				}
			},
		},
		{
			name: "CreateSubscription_WebhookByName",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/webhooks":
					listOneWebhook(t, w, r, orgID("o1"), "hook-1", webhookID("wh-1"))
				case "/v2/webhook-subscriptions":
					var body wireCreateSubscription
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					if body.WebhookID != webhookID("wh-1") {
						t.Errorf("webhook_id: want %q, got %q", webhookID("wh-1"), body.WebhookID)
					}
					writeJSON(w, 201, webhooks.WebhookSubscription{Id: subscriptionID("s1")})
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(500)
				}
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.CreateSubscription(ctx, webhooks.CreateSubscriptionRequest{
					Webhook: "hook-1", Organization: orgID("o1"),
					SourceType: webhooks.WebhookSourceTypeEvaluator, SourceID: promptID("e1"),
					Event: webhooks.WebhookEventTypeEvaluatorVersionCreated,
				})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "GetSubscription",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v2/webhook-subscriptions/"+subscriptionID("s1") {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				writeJSON(w, 200, webhooks.WebhookSubscription{Id: subscriptionID("s1"), WebhookId: webhookID("wh-1")})
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.GetSubscription(ctx, webhooks.GetSubscriptionRequest{SubscriptionID: subscriptionID("s1")})
			},
			check: func(t *testing.T, got any, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if sub := got.(*webhooks.WebhookSubscription); sub.WebhookId != webhookID("wh-1") {
					t.Errorf("unexpected webhook id: %s", sub.WebhookId)
				}
			},
		},
		{
			name: "DeleteSubscription",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete || r.URL.Path != "/v2/webhook-subscriptions/"+subscriptionID("s1") {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(204)
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Webhooks.DeleteSubscription(ctx, webhooks.DeleteSubscriptionRequest{SubscriptionID: subscriptionID("s1")})
			},
			check: func(t *testing.T, _ any, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "GetSubscription_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(404)
				_, _ = w.Write([]byte(`{"title":"Not Found","detail":"subscription not found","status":404}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.GetSubscription(ctx, webhooks.GetSubscriptionRequest{SubscriptionID: subscriptionID("s-missing")})
			},
			check: func(t *testing.T, _ any, err error) {
				var apiErr *arize.NotFoundError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *arize.NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "DeleteSubscription_NotFound",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(404)
				_, _ = w.Write([]byte(`{"title":"Not Found","detail":"subscription not found","status":404}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return nil, c.Webhooks.DeleteSubscription(ctx, webhooks.DeleteSubscriptionRequest{SubscriptionID: subscriptionID("s-missing")})
			},
			check: func(t *testing.T, _ any, err error) {
				var apiErr *arize.NotFoundError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *arize.NotFoundError, got %T: %v", err, err)
				}
			},
		},
		{
			name: "Error_404",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(404)
				_, _ = w.Write([]byte(`{"title":"Not Found","detail":"webhook not found","status":404}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: webhookID("wh-missing")})
			},
			check: func(t *testing.T, _ any, err error) {
				var apiErr *arize.NotFoundError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *arize.NotFoundError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != 404 {
					t.Errorf("status: %d", apiErr.StatusCode)
				}
			},
		},
		{
			name: "Error_409",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(409)
				_, _ = w.Write([]byte(`{"title":"Conflict","detail":"duplicate subscription","status":409}`))
			},
			invoke: func(ctx context.Context, c *arize.Client) (any, error) {
				return c.Webhooks.CreateSubscription(ctx, webhooks.CreateSubscriptionRequest{
					Webhook: webhookID("wh-1"), SourceType: webhooks.WebhookSourceTypePrompt,
					SourceID: promptID("p1"), Event: webhooks.WebhookEventTypePromptVersionCreated,
				})
			},
			check: func(t *testing.T, _ any, err error) {
				var apiErr *arize.ConflictError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *arize.ConflictError, got %T: %v", err, err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, tt.handler)
			got, err := tt.invoke(context.Background(), client)
			tt.check(t, got, err)
		})
	}
}
