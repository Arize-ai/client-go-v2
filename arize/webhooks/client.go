package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// ErrNoUpdateFields is returned by Update when every patch field is nil.
var ErrNoUpdateFields = errors.New("webhooks: update requires at least one of Name, Description, URL, AuthToken, TimeoutMs, or Headers")

// ErrUnpairedSourceFilter is returned by ListSubscriptions when only one of
// SourceType and SourceID is set.
var ErrUnpairedSourceFilter = errors.New("webhooks: SourceType and SourceID must be provided together")

// Client provides access to the Arize Webhooks API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of webhooks, most recently created first.
// req.Organization, when non-empty, accepts an organization name or ID and
// restricts results to that organization. Webhooks used as monitor
// notification channels are included.
func (c *Client) List(ctx context.Context, req ListRequest) (*ListWebhooks, error) {
	prerelease.Warn("webhooks.list", prerelease.Alpha)
	var orgID *string
	if req.Organization != "" {
		id, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
		if err != nil {
			return nil, err
		}
		orgID = &id
	}
	params := generated.ListWebhooksParams{
		OrgId:  orgID,
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListWebhooksWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single webhook, resolving by name or ID.
func (c *Client) Get(ctx context.Context, req GetRequest) (*Webhook, error) {
	prerelease.Warn("webhooks.get", prerelease.Alpha)
	id, err := resolve.FindWebhookID(ctx, c.gen, req.Webhook, req.Organization)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetWebhookWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new webhook, resolving the owning organization by name or
// ID. For WebhookAuthTypeHMACSHA256 webhooks the response carries
// SigningSecret; this is the only time the secret is returned, so store it.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*CreateWebhook, error) {
	prerelease.Warn("webhooks.create", prerelease.Alpha)
	orgID, err := resolve.FindOrganizationID(ctx, c.gen, req.Organization)
	if err != nil {
		return nil, err
	}
	body := generated.CreateWebhookJSONRequestBody{
		OrganizationId: orgID,
		Name:           req.Name,
		Url:            req.URL,
		Description:    optfields.PtrIfSet(req.Description),
		AuthType:       optfields.PtrIfSet(req.AuthType),
		AuthToken:      optfields.PtrIfSet(req.AuthToken),
		TimeoutMs:      optfields.PtrIfSet(req.TimeoutMs),
		Headers:        optfields.PtrMapIfSet(req.Headers),
	}
	resp, err := c.gen.CreateWebhookWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update patches a webhook, resolving by name or ID. Nil fields are left
// unchanged; ErrNoUpdateFields is returned when every patch field is nil.
func (c *Client) Update(ctx context.Context, req UpdateRequest) (*Webhook, error) {
	prerelease.Warn("webhooks.update", prerelease.Alpha)
	body := map[string]any{}
	if req.Name != nil {
		body["name"] = *req.Name
	}
	if req.Description != nil {
		if *req.Description == "" {
			body["description"] = nil
		} else {
			body["description"] = *req.Description
		}
	}
	if req.URL != nil {
		body["url"] = *req.URL
	}
	if req.AuthToken != nil {
		body["auth_token"] = *req.AuthToken
	}
	if req.TimeoutMs != nil {
		body["timeout_ms"] = *req.TimeoutMs
	}
	if req.Headers != nil {
		// headers is not nullable on the wire, so a nil map is sent as {}.
		headers := *req.Headers
		if headers == nil {
			headers = map[string]string{}
		}
		body["headers"] = headers
	}
	if len(body) == 0 {
		return nil, ErrNoUpdateFields
	}
	id, err := resolve.FindWebhookID(ctx, c.gen, req.Webhook, req.Organization)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("webhooks: marshal update body: %w", err)
	}
	resp, err := c.gen.UpdateWebhookWithBodyWithResponse(ctx, id, "application/json", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes a webhook, resolving by name or ID. The webhook stops
// receiving events and is detached from every prompt, evaluator, and monitor
// it was subscribed to.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("webhooks.delete", prerelease.Alpha)
	id, err := resolve.FindWebhookID(ctx, c.gen, req.Webhook, req.Organization)
	if err != nil {
		return err
	}
	resp, err := c.gen.DeleteWebhookWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// Test sends a test event to a webhook's endpoint, resolving the webhook by
// name or ID. A nil error means the test ran; inspect StatusCode and
// ErrorMessage on the response for the endpoint's outcome. Test deliveries
// are not supported for WebhookAuthTypeHMACSHA256 webhooks.
func (c *Client) Test(ctx context.Context, req TestRequest) (*TestWebhook, error) {
	prerelease.Warn("webhooks.test", prerelease.Alpha)
	id, err := resolve.FindWebhookID(ctx, c.gen, req.Webhook, req.Organization)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.TestWebhookWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// ListDeliveryAttempts returns a paginated list of a webhook's delivery
// attempts, most recent first, resolving the webhook by name or ID. Failed
// deliveries are retried, so one event may have several attempts.
func (c *Client) ListDeliveryAttempts(ctx context.Context, req ListDeliveryAttemptsRequest) (*ListWebhookDeliveryAttempts, error) {
	prerelease.Warn("webhooks.list_delivery_attempts", prerelease.Alpha)
	id, err := resolve.FindWebhookID(ctx, c.gen, req.Webhook, req.Organization)
	if err != nil {
		return nil, err
	}
	params := &generated.ListWebhookDeliveryAttemptsParams{
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListWebhookDeliveryAttemptsWithResponse(ctx, id, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// ListSubscriptions returns a paginated list of webhook subscriptions, most
// recently created first. SourceType and SourceID must be set together;
// ErrUnpairedSourceFilter is returned when only one is set. Subscriptions
// whose webhook has since been deleted are dropped after the page is read,
// so a page may hold fewer than Limit items while Pagination.HasMore is still
// true; keep paging until it is false.
func (c *Client) ListSubscriptions(ctx context.Context, req ListSubscriptionsRequest) (*ListWebhookSubscriptions, error) {
	prerelease.Warn("webhooks.list_subscriptions", prerelease.Alpha)
	if (req.SourceType == "") != (req.SourceID == "") {
		return nil, ErrUnpairedSourceFilter
	}
	params := &generated.ListWebhookSubscriptionsParams{
		SourceType: optfields.PtrIfSet(req.SourceType),
		SourceId:   optfields.PtrIfSet(req.SourceID),
		Limit:      optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor:     optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.ListWebhookSubscriptionsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// CreateSubscription subscribes a webhook (resolved by name or ID) to one
// event on one prompt or evaluator.
func (c *Client) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*WebhookSubscription, error) {
	prerelease.Warn("webhooks.create_subscription", prerelease.Alpha)
	webhookID, err := resolve.FindWebhookID(ctx, c.gen, req.Webhook, req.Organization)
	if err != nil {
		return nil, err
	}
	body := generated.CreateWebhookSubscriptionJSONRequestBody{
		WebhookId:  webhookID,
		SourceType: req.SourceType,
		SourceId:   req.SourceID,
		Event:      req.Event,
	}
	resp, err := c.gen.CreateWebhookSubscriptionWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// GetSubscription returns a single webhook subscription by strict ID. A 404
// is returned when the subscription does not exist, its source is not
// readable, or its webhook has since been deleted.
func (c *Client) GetSubscription(ctx context.Context, req GetSubscriptionRequest) (*WebhookSubscription, error) {
	prerelease.Warn("webhooks.get_subscription", prerelease.Alpha)
	resp, err := c.gen.GetWebhookSubscriptionWithResponse(ctx, req.SubscriptionID)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// DeleteSubscription removes a webhook subscription by strict ID. Other
// subscriptions on the source and the webhook itself are unaffected.
func (c *Client) DeleteSubscription(ctx context.Context, req DeleteSubscriptionRequest) error {
	prerelease.Warn("webhooks.delete_subscription", prerelease.Alpha)
	resp, err := c.gen.DeleteWebhookSubscriptionWithResponse(ctx, req.SubscriptionID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
