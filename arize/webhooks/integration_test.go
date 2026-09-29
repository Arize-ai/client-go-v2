//go:build integration

// Integration tests for the webhooks client against a live Arize API.
//
// The file is excluded from `go test ./...` and from the Bazel go_test by its
// build tag, and it skips itself when ARIZE_API_KEY is unset. Run it with:
//
//	ARIZE_API_KEY=<key> go test -tags integration ./arize/webhooks/ -run TestWebhooksIntegration -v
//
// Environment:
//   - ARIZE_API_KEY (required): the key under test.
//   - ARIZE_API_HOST / ARIZE_API_SCHEME (optional): the API host the SDK
//     already reads; defaults to production (api.arize.com, https).
//   - ARIZE_TEST_ORGANIZATION (optional): organization ID or name; defaults to
//     the first organization the key can read.
//
// Subtests run in file order and share state. Every resource created here is
// named `go-sdk-itest-<timestamp>...` and deleted in Cleanup whether or not
// the subtests pass.
package webhooks_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/organizations"
	"github.com/Arize-ai/client-go-v2/arize/prompts"
	"github.com/Arize-ai/client-go-v2/arize/spaces"
	"github.com/Arize-ai/client-go-v2/arize/webhooks"
)

const (
	itestTargetURL          = "https://example.com/hook"
	itestInitialDescription = "created by the go sdk integration tests"
	itestMaxPages           = 10
)

// isResourceID reports whether v decodes as a base64 "Type:..." resource ID.
func isResourceID(v string) bool {
	raw, err := base64.StdEncoding.DecodeString(v)
	return err == nil && strings.Contains(string(raw), ":")
}

// jsonKeys returns the top-level keys of v's JSON encoding.
func jsonKeys(t *testing.T, v any) map[string]bool {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	keys := make(map[string]bool, len(m))
	for k := range m {
		keys[k] = true
	}
	return keys
}

// assertNoCredentialKeys fails when the JSON encoding of v carries a
// credential field. allowSigningSecret is true only for the HMAC create
// result, the one response that legitimately returns the secret.
func assertNoCredentialKeys(t *testing.T, v any, allowSigningSecret bool) {
	t.Helper()
	keys := jsonKeys(t, v)
	for _, k := range []string{"auth_token", "headers"} {
		if keys[k] {
			t.Errorf("JSON has credential key %q", k)
		}
	}
	if keys["signing_secret"] && !allowSigningSecret {
		t.Errorf("JSON has key signing_secret")
	}
}

// offlineClient points at a closed port so a returned error proves the SDK
// failed before sending anything.
func offlineClient(t *testing.T) *arize.Client {
	t.Helper()
	c, err := arize.NewClient(arize.Config{APIKey: "unused", APIHost: "127.0.0.1:9", APIScheme: "http"})
	if err != nil {
		t.Fatalf("offline client: %v", err)
	}
	return c
}

// pickOrganization honours ARIZE_TEST_ORGANIZATION (ID or name) and otherwise
// returns the first organization the key can read.
func pickOrganization(ctx context.Context, t *testing.T, client *arize.Client) organizations.Organization {
	t.Helper()
	resp, err := client.Organizations.List(ctx, organizations.ListRequest{Limit: 100})
	if err != nil {
		t.Fatalf("list organizations: %v", err)
	}
	if override := os.Getenv("ARIZE_TEST_ORGANIZATION"); override != "" {
		for _, org := range resp.Organizations {
			if org.Id == override || org.Name == override {
				return org
			}
		}
		t.Fatalf("ARIZE_TEST_ORGANIZATION %q is not among the first 100 organizations this key can read", override)
	}
	if len(resp.Organizations) == 0 {
		t.Fatal("the API key cannot read any organization")
	}
	return resp.Organizations[0]
}

// findOrCreatePrompt scans the organization's spaces for any prompt and
// returns its ID. When none exists it creates one named promptName in the
// first space and returns created=true so Cleanup can remove it.
func findOrCreatePrompt(ctx context.Context, t *testing.T, client *arize.Client, orgID, promptName string) (id string, created bool) {
	t.Helper()
	spaceList, err := client.Spaces.List(ctx, spaces.ListRequest{Organization: orgID, Limit: 100})
	if err != nil {
		t.Fatalf("list spaces: %v", err)
	}
	if len(spaceList.Spaces) == 0 {
		t.Fatalf("organization %s has no space to hold the test prompt", orgID)
	}
	for _, space := range spaceList.Spaces {
		promptList, err := client.Prompts.List(ctx, prompts.ListRequest{Space: space.Id, Limit: 1})
		if err != nil {
			t.Fatalf("list prompts in space %s: %v", space.Id, err)
		}
		if len(promptList.Prompts) > 0 {
			return promptList.Prompts[0].Id, false
		}
	}
	content := "Hello {name}"
	model := "gpt-4o-mini"
	format := prompts.InputVariableFormatFString
	p, err := client.Prompts.Create(ctx, prompts.CreateRequest{
		Space: spaceList.Spaces[0].Id,
		Name:  promptName,
		Version: prompts.PromptVersionCreate{
			CommitMessage:       "initial version",
			InputVariableFormat: &format,
			Provider:            prompts.LlmProviderOpenAi,
			Model:               &model,
			Messages:            []prompts.LLMMessageRequest{{Role: prompts.MessageRoleUser, Content: &content}},
		},
	})
	if err != nil {
		t.Fatalf("create prompt: %v", err)
	}
	return p.Id, true
}

func TestWebhooksIntegration(t *testing.T) {
	if os.Getenv("ARIZE_API_KEY") == "" {
		t.Skip("ARIZE_API_KEY is not set")
	}
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	t.Cleanup(cancel)

	prefix := fmt.Sprintf("go-sdk-itest-%d", time.Now().UnixMilli())
	bearerName := prefix + "-bearer"
	hmacName := prefix + "-hmac"
	renamedBearerName := prefix + "-bearer-renamed"
	promptName := prefix + "-prompt"

	org := pickOrganization(ctx, t, client)
	promptID, promptCreated := findOrCreatePrompt(ctx, t, client, org.Id, promptName)

	var (
		bearer       *webhooks.CreateWebhook
		hmac         *webhooks.CreateWebhook
		subscription *webhooks.WebhookSubscription
	)

	// Cleanup runs after every subtest, including on failure. Anything already
	// gone is fine; every other error is a failure.
	t.Cleanup(func() {
		var nfe *arize.NotFoundError
		if subscription != nil {
			err := client.Webhooks.DeleteSubscription(ctx, webhooks.DeleteSubscriptionRequest{SubscriptionID: subscription.Id})
			if err != nil && !errors.As(err, &nfe) {
				t.Errorf("cleanup: delete subscription %s: %v", subscription.Id, err)
			}
		}
		leftovers, err := client.Webhooks.List(ctx, webhooks.ListRequest{Organization: org.Id, Name: prefix, Limit: 100})
		if err != nil {
			t.Errorf("cleanup: list webhooks: %v", err)
		} else {
			for _, wh := range leftovers.Webhooks {
				err := client.Webhooks.Delete(ctx, webhooks.DeleteRequest{Webhook: wh.Id})
				if err != nil && !errors.As(err, &nfe) {
					t.Errorf("cleanup: delete webhook %s: %v", wh.Id, err)
				}
			}
		}
		if promptCreated {
			err := client.Prompts.Delete(ctx, prompts.DeleteRequest{Prompt: promptID})
			if err != nil && !errors.As(err, &nfe) {
				t.Errorf("cleanup: delete prompt %s: %v", promptID, err)
			}
		}
	})

	// needs fails the subtest when an earlier subtest it depends on did not
	// produce its resource.
	needs := func(t *testing.T, name string, ok bool) {
		t.Helper()
		if !ok {
			t.Fatalf("depends on %q, which did not complete", name)
		}
	}

	// ── Create ──────────────────────────────────────────────────────────

	t.Run("create BEARER webhook without returning credentials", func(t *testing.T) {
		wh, err := client.Webhooks.Create(ctx, webhooks.CreateRequest{
			Organization: org.Id,
			Name:         bearerName,
			URL:          itestTargetURL,
			Description:  itestInitialDescription,
			AuthType:     webhooks.WebhookAuthTypeBearer,
			AuthToken:    "Bearer go-sdk-itest-token",
			Headers:      map[string]string{"X-Test-Run": prefix},
			TimeoutMs:    5000,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		bearer = wh
		if !isResourceID(wh.Id) {
			t.Errorf("Id = %q, want a resource ID", wh.Id)
		}
		if wh.OrganizationId != org.Id {
			t.Errorf("OrganizationId = %q, want %q", wh.OrganizationId, org.Id)
		}
		if wh.Name != bearerName {
			t.Errorf("Name = %q, want %q", wh.Name, bearerName)
		}
		if wh.Url != itestTargetURL {
			t.Errorf("Url = %q, want %q", wh.Url, itestTargetURL)
		}
		if wh.Description != itestInitialDescription {
			t.Errorf("Description = %q, want %q", wh.Description, itestInitialDescription)
		}
		if wh.AuthType != webhooks.WebhookAuthTypeBearer {
			t.Errorf("AuthType = %q, want BEARER", wh.AuthType)
		}
		if wh.TimeoutMs != 5000 {
			t.Errorf("TimeoutMs = %d, want 5000", wh.TimeoutMs)
		}
		if wh.CreatedAt.IsZero() || wh.UpdatedAt.IsZero() {
			t.Errorf("CreatedAt/UpdatedAt zero: %v / %v", wh.CreatedAt, wh.UpdatedAt)
		}
		if wh.SigningSecret != nil {
			t.Errorf("SigningSecret = %q, want nil on a BEARER webhook", *wh.SigningSecret)
		}
		if wh.SigningSecretHint != nil {
			t.Errorf("SigningSecretHint = %q, want nil on a BEARER webhook", *wh.SigningSecretHint)
		}
		assertNoCredentialKeys(t, wh, false)
	})

	t.Run("create HMAC_SHA256 webhook returns the signing secret once", func(t *testing.T) {
		wh, err := client.Webhooks.Create(ctx, webhooks.CreateRequest{
			Organization: org.Id,
			Name:         hmacName,
			URL:          itestTargetURL,
			AuthType:     webhooks.WebhookAuthTypeHMACSHA256,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		hmac = wh
		if !isResourceID(wh.Id) {
			t.Errorf("Id = %q, want a resource ID", wh.Id)
		}
		if wh.AuthType != webhooks.WebhookAuthTypeHMACSHA256 {
			t.Errorf("AuthType = %q, want HMAC_SHA256", wh.AuthType)
		}
		if wh.Description != "" {
			t.Errorf("Description = %q, want empty", wh.Description)
		}
		if wh.TimeoutMs != 30000 {
			t.Errorf("TimeoutMs = %d, want server default 30000", wh.TimeoutMs)
		}
		if wh.SigningSecret == nil || *wh.SigningSecret == "" {
			t.Errorf("SigningSecret = %v, want a non-empty string", wh.SigningSecret)
		}
		if wh.SigningSecretHint == nil || *wh.SigningSecretHint == "" {
			t.Errorf("SigningSecretHint = %v, want a non-empty string", wh.SigningSecretHint)
		}
		assertNoCredentialKeys(t, wh, true)
		if !jsonKeys(t, wh)["signing_secret"] {
			t.Errorf("JSON lacks signing_secret on the HMAC create result")
		}
	})

	// ── Get ─────────────────────────────────────────────────────────────

	t.Run("get by ID returns no credentials", func(t *testing.T) {
		needs(t, "create BEARER", bearer != nil)
		got, err := client.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: bearer.Id})
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Id != bearer.Id {
			t.Errorf("Id = %q, want %q", got.Id, bearer.Id)
		}
		if got.Name != bearerName {
			t.Errorf("Name = %q, want %q", got.Name, bearerName)
		}
		if got.OrganizationId != org.Id {
			t.Errorf("OrganizationId = %q, want %q", got.OrganizationId, org.Id)
		}
		if got.SigningSecretHint != nil {
			t.Errorf("SigningSecretHint = %q, want nil on a BEARER webhook", *got.SigningSecretHint)
		}
		assertNoCredentialKeys(t, got, false)
	})

	t.Run("get HMAC_SHA256 by ID returns only the signing secret hint", func(t *testing.T) {
		needs(t, "create HMAC", hmac != nil)
		got, err := client.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: hmac.Id})
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Id != hmac.Id {
			t.Errorf("Id = %q, want %q", got.Id, hmac.Id)
		}
		if got.SigningSecretHint == nil || hmac.SigningSecretHint == nil || *got.SigningSecretHint != *hmac.SigningSecretHint {
			t.Errorf("SigningSecretHint = %v, want %v", got.SigningSecretHint, hmac.SigningSecretHint)
		}
		assertNoCredentialKeys(t, got, false)
	})

	t.Run("get by name resolves the organization", func(t *testing.T) {
		needs(t, "create BEARER and HMAC", bearer != nil && hmac != nil)
		tests := []struct {
			name         string
			webhook      string
			organization string
			wantID       string
		}{
			{name: "organization name", webhook: bearerName, organization: org.Name, wantID: bearer.Id},
			{name: "organization ID", webhook: hmacName, organization: org.Id, wantID: hmac.Id},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := client.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: tt.webhook, Organization: tt.organization})
				if err != nil {
					t.Fatalf("get: %v", err)
				}
				if got.Id != tt.wantID {
					t.Errorf("Id = %q, want %q", got.Id, tt.wantID)
				}
			})
		}
	})

	t.Run("get unknown name returns ResourceNotFoundError with near misses", func(t *testing.T) {
		needs(t, "create BEARER and HMAC", bearer != nil && hmac != nil)
		_, err := client.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: prefix, Organization: org.Id})
		var nfe *arize.ResourceNotFoundError
		if !errors.As(err, &nfe) {
			t.Fatalf("err = %v, want *arize.ResourceNotFoundError", err)
		}
		if nfe.ResourceType != "webhook" || nfe.Name != prefix {
			t.Errorf("ResourceType/Name = %q/%q, want webhook/%q", nfe.ResourceType, nfe.Name, prefix)
		}
		available := strings.Join(nfe.Available, ",")
		for _, want := range []string{bearerName, hmacName} {
			if !strings.Contains(available, want) {
				t.Errorf("Available = %v, want it to include %q", nfe.Available, want)
			}
		}
	})

	t.Run("get by name without organization fails before any request", func(t *testing.T) {
		_, err := offlineClient(t).Webhooks.Get(ctx, webhooks.GetRequest{Webhook: bearerName})
		var nfe *arize.ResourceNotFoundError
		if !errors.As(err, &nfe) {
			t.Fatalf("err = %v, want *arize.ResourceNotFoundError", err)
		}
		if nfe.Hint == "" {
			t.Errorf("Hint is empty, want a hint naming the organization field")
		}
	})

	// ── List ────────────────────────────────────────────────────────────

	t.Run("list filters by name substring", func(t *testing.T) {
		needs(t, "create BEARER and HMAC", bearer != nil && hmac != nil)
		resp, err := client.Webhooks.List(ctx, webhooks.ListRequest{Organization: org.Id, Name: prefix})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		ids := map[string]bool{}
		for _, wh := range resp.Webhooks {
			ids[wh.Id] = true
			if !strings.HasPrefix(wh.Name, prefix) {
				t.Errorf("Name = %q, want prefix %q", wh.Name, prefix)
			}
			assertNoCredentialKeys(t, wh, false)
		}
		if !ids[bearer.Id] || !ids[hmac.Id] {
			t.Errorf("listed ids = %v, want both %q and %q", ids, bearer.Id, hmac.Id)
		}
	})

	t.Run("list pages with Limit 1 until HasMore is false", func(t *testing.T) {
		needs(t, "create BEARER and HMAC", bearer != nil && hmac != nil)
		var ids []string
		seen := map[string]bool{}
		cursor := ""
		pages := 0
		for {
			page, err := client.Webhooks.List(ctx, webhooks.ListRequest{Organization: org.Id, Name: prefix, Limit: 1, Cursor: cursor})
			if err != nil {
				t.Fatalf("list page %d: %v", pages+1, err)
			}
			pages++
			if len(page.Webhooks) > 1 {
				t.Errorf("page %d has %d items, want at most 1", pages, len(page.Webhooks))
			}
			for _, wh := range page.Webhooks {
				if seen[wh.Id] {
					t.Errorf("duplicate id %q across pages", wh.Id)
				}
				seen[wh.Id] = true
				ids = append(ids, wh.Id)
			}
			if !page.Pagination.HasMore {
				break
			}
			if page.Pagination.NextCursor == nil || *page.Pagination.NextCursor == "" {
				t.Fatalf("page %d: HasMore is true but NextCursor is empty", pages)
			}
			cursor = *page.Pagination.NextCursor
			if pages >= itestMaxPages {
				t.Fatalf("pagination did not finish after %d pages", itestMaxPages)
			}
		}
		if pages < 2 {
			t.Errorf("pages = %d, want at least 2", pages)
		}
		if !seen[bearer.Id] || !seen[hmac.Id] {
			t.Errorf("paged ids = %v, want both %q and %q", ids, bearer.Id, hmac.Id)
		}
	})

	// ── Update ──────────────────────────────────────────────────────────

	t.Run("update Name and TimeoutMs leaves other fields alone", func(t *testing.T) {
		needs(t, "create BEARER", bearer != nil)
		timeout := 7000
		name := renamedBearerName
		updated, err := client.Webhooks.Update(ctx, webhooks.UpdateRequest{Webhook: bearer.Id, Name: &name, TimeoutMs: &timeout})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Id != bearer.Id {
			t.Errorf("Id = %q, want %q", updated.Id, bearer.Id)
		}
		if updated.Name != renamedBearerName {
			t.Errorf("Name = %q, want %q", updated.Name, renamedBearerName)
		}
		if updated.TimeoutMs != 7000 {
			t.Errorf("TimeoutMs = %d, want 7000", updated.TimeoutMs)
		}
		if updated.Url != bearer.Url {
			t.Errorf("Url = %q, want unchanged %q", updated.Url, bearer.Url)
		}
		if updated.Description != itestInitialDescription {
			t.Errorf("Description = %q, want unchanged %q", updated.Description, itestInitialDescription)
		}
		if updated.AuthType != webhooks.WebhookAuthTypeBearer {
			t.Errorf("AuthType = %q, want unchanged BEARER", updated.AuthType)
		}
		if updated.OrganizationId != org.Id {
			t.Errorf("OrganizationId = %q, want unchanged %q", updated.OrganizationId, org.Id)
		}
		if updated.UpdatedAt.Before(bearer.UpdatedAt) {
			t.Errorf("UpdatedAt = %v, want >= %v", updated.UpdatedAt, bearer.UpdatedAt)
		}
		got, err := client.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: bearer.Id})
		if err != nil {
			t.Fatalf("get after update: %v", err)
		}
		if got.Name != renamedBearerName || got.TimeoutMs != 7000 {
			t.Errorf("get after update: Name/TimeoutMs = %q/%d, want %q/7000", got.Name, got.TimeoutMs, renamedBearerName)
		}
	})

	t.Run("update with empty Description clears it", func(t *testing.T) {
		needs(t, "create BEARER", bearer != nil)
		empty := ""
		updated, err := client.Webhooks.Update(ctx, webhooks.UpdateRequest{Webhook: bearer.Id, Description: &empty})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Description != "" {
			t.Errorf("Description = %q, want empty", updated.Description)
		}
		if updated.Name != renamedBearerName {
			t.Errorf("Name = %q, want unchanged %q", updated.Name, renamedBearerName)
		}
	})

	t.Run("update with no fields fails before any request", func(t *testing.T) {
		needs(t, "create BEARER", bearer != nil)
		_, err := offlineClient(t).Webhooks.Update(ctx, webhooks.UpdateRequest{Webhook: bearer.Id})
		if !errors.Is(err, webhooks.ErrNoUpdateFields) {
			t.Fatalf("err = %v, want ErrNoUpdateFields", err)
		}
	})

	// ── Test delivery and delivery attempts ─────────────────────────────

	t.Run("test BEARER webhook returns the endpoint's response", func(t *testing.T) {
		needs(t, "create BEARER", bearer != nil)
		result, err := client.Webhooks.Test(ctx, webhooks.TestRequest{Webhook: bearer.Id})
		if err != nil {
			t.Fatalf("test: %v", err)
		}
		if result.StatusCode < 100 || result.StatusCode > 599 {
			t.Errorf("StatusCode = %d, want an HTTP status", result.StatusCode)
		}
		if result.ErrorMessage != nil {
			t.Logf("endpoint answered %d: %s", result.StatusCode, *result.ErrorMessage)
		} else {
			t.Logf("endpoint answered %d", result.StatusCode)
		}
	})

	t.Run("test HMAC_SHA256 webhook returns BadRequestError", func(t *testing.T) {
		needs(t, "create HMAC", hmac != nil)
		_, err := client.Webhooks.Test(ctx, webhooks.TestRequest{Webhook: hmac.Id})
		var bre *arize.BadRequestError
		if !errors.As(err, &bre) {
			t.Fatalf("err = %v, want *arize.BadRequestError", err)
		}
		if bre.StatusCode != 400 {
			t.Errorf("StatusCode = %d, want 400", bre.StatusCode)
		}
	})

	t.Run("list delivery attempts returns an empty page", func(t *testing.T) {
		needs(t, "create BEARER", bearer != nil)
		resp, err := client.Webhooks.ListDeliveryAttempts(ctx, webhooks.ListDeliveryAttemptsRequest{Webhook: bearer.Id})
		if err != nil {
			t.Fatalf("list delivery attempts: %v", err)
		}
		if len(resp.DeliveryAttempts) != 0 {
			t.Errorf("DeliveryAttempts has %d items, want 0 on a webhook with no subscriptions", len(resp.DeliveryAttempts))
		}
		if resp.Pagination.HasMore {
			t.Errorf("HasMore = true, want false")
		}
	})

	// ── Subscriptions ───────────────────────────────────────────────────

	t.Run("create PROMPT_VERSION_CREATED subscription on a prompt", func(t *testing.T) {
		needs(t, "create BEARER", bearer != nil)
		sub, err := client.Webhooks.CreateSubscription(ctx, webhooks.CreateSubscriptionRequest{
			Webhook:    bearer.Id,
			SourceType: webhooks.WebhookSourceTypePrompt,
			SourceID:   promptID,
			Event:      webhooks.WebhookEventTypePromptVersionCreated,
		})
		if err != nil {
			t.Fatalf("create subscription: %v", err)
		}
		subscription = sub
		if !isResourceID(sub.Id) {
			t.Errorf("Id = %q, want a resource ID", sub.Id)
		}
		if sub.WebhookId != bearer.Id {
			t.Errorf("WebhookId = %q, want %q", sub.WebhookId, bearer.Id)
		}
		if sub.SourceType != webhooks.WebhookSourceTypePrompt {
			t.Errorf("SourceType = %q, want PROMPT", sub.SourceType)
		}
		if sub.SourceId != promptID {
			t.Errorf("SourceId = %q, want %q", sub.SourceId, promptID)
		}
		if sub.Event != webhooks.WebhookEventTypePromptVersionCreated {
			t.Errorf("Event = %q, want PROMPT_VERSION_CREATED", sub.Event)
		}
		if sub.CreatedAt.IsZero() {
			t.Errorf("CreatedAt is zero")
		}
	})

	t.Run("create subscription rejects invalid input", func(t *testing.T) {
		needs(t, "create BEARER", bearer != nil)
		tests := []struct {
			name  string
			event webhooks.WebhookEventType
			check func(t *testing.T, err error)
		}{
			{
				name:  "duplicate returns ConflictError",
				event: webhooks.WebhookEventTypePromptVersionCreated,
				check: func(t *testing.T, err error) {
					var ce *arize.ConflictError
					if !errors.As(err, &ce) {
						t.Fatalf("err = %v, want *arize.ConflictError", err)
					}
					if ce.StatusCode != 409 {
						t.Errorf("StatusCode = %d, want 409", ce.StatusCode)
					}
				},
			},
			{
				name:  "evaluator event on a prompt returns UnprocessableEntityError",
				event: webhooks.WebhookEventTypeEvaluatorVersionCreated,
				check: func(t *testing.T, err error) {
					var uee *arize.UnprocessableEntityError
					if !errors.As(err, &uee) {
						t.Fatalf("err = %v, want *arize.UnprocessableEntityError", err)
					}
					if uee.StatusCode != 422 {
						t.Errorf("StatusCode = %d, want 422", uee.StatusCode)
					}
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := client.Webhooks.CreateSubscription(ctx, webhooks.CreateSubscriptionRequest{
					Webhook:    bearer.Id,
					SourceType: webhooks.WebhookSourceTypePrompt,
					SourceID:   promptID,
					Event:      tt.event,
				})
				tt.check(t, err)
			})
		}
	})

	t.Run("list subscriptions with only SourceType fails before any request", func(t *testing.T) {
		_, err := offlineClient(t).Webhooks.ListSubscriptions(ctx, webhooks.ListSubscriptionsRequest{SourceType: webhooks.WebhookSourceTypePrompt})
		if !errors.Is(err, webhooks.ErrUnpairedSourceFilter) {
			t.Fatalf("err = %v, want ErrUnpairedSourceFilter", err)
		}
	})

	t.Run("list subscriptions by source contains the subscription", func(t *testing.T) {
		needs(t, "create subscription", subscription != nil)
		resp, err := client.Webhooks.ListSubscriptions(ctx, webhooks.ListSubscriptionsRequest{
			SourceType: webhooks.WebhookSourceTypePrompt,
			SourceID:   promptID,
		})
		if err != nil {
			t.Fatalf("list subscriptions: %v", err)
		}
		found := false
		for _, s := range resp.Subscriptions {
			if s.Id == subscription.Id {
				found = true
			}
		}
		if !found {
			t.Errorf("subscription %q not in list of %d items", subscription.Id, len(resp.Subscriptions))
		}
	})

	t.Run("get subscription by ID", func(t *testing.T) {
		needs(t, "create subscription", subscription != nil)
		got, err := client.Webhooks.GetSubscription(ctx, webhooks.GetSubscriptionRequest{SubscriptionID: subscription.Id})
		if err != nil {
			t.Fatalf("get subscription: %v", err)
		}
		if got.Id != subscription.Id {
			t.Errorf("Id = %q, want %q", got.Id, subscription.Id)
		}
		if got.WebhookId != bearer.Id {
			t.Errorf("WebhookId = %q, want %q", got.WebhookId, bearer.Id)
		}
		if got.Event != webhooks.WebhookEventTypePromptVersionCreated {
			t.Errorf("Event = %q, want PROMPT_VERSION_CREATED", got.Event)
		}
	})

	t.Run("delete subscription then get returns NotFoundError", func(t *testing.T) {
		needs(t, "create subscription", subscription != nil)
		id := subscription.Id
		if err := client.Webhooks.DeleteSubscription(ctx, webhooks.DeleteSubscriptionRequest{SubscriptionID: id}); err != nil {
			t.Fatalf("delete subscription: %v", err)
		}
		subscription = nil
		_, err := client.Webhooks.GetSubscription(ctx, webhooks.GetSubscriptionRequest{SubscriptionID: id})
		var nfe *arize.NotFoundError
		if !errors.As(err, &nfe) {
			t.Fatalf("err = %v, want *arize.NotFoundError", err)
		}
	})

	// ── Delete ──────────────────────────────────────────────────────────

	t.Run("delete webhooks then get returns NotFoundError", func(t *testing.T) {
		needs(t, "create BEARER and HMAC", bearer != nil && hmac != nil)
		tests := []struct {
			name string
			req  webhooks.DeleteRequest
			id   string
		}{
			{name: "by ID", req: webhooks.DeleteRequest{Webhook: bearer.Id}, id: bearer.Id},
			{name: "by name", req: webhooks.DeleteRequest{Webhook: hmacName, Organization: org.Id}, id: hmac.Id},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := client.Webhooks.Delete(ctx, tt.req); err != nil {
					t.Fatalf("delete: %v", err)
				}
				_, err := client.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: tt.id})
				var nfe *arize.NotFoundError
				if !errors.As(err, &nfe) {
					t.Fatalf("get after delete: err = %v, want *arize.NotFoundError", err)
				}
			})
		}
	})
}
