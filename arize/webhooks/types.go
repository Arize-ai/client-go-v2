package webhooks

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// Response, list, and nested types remain aliases to the generated wire
// shapes so callers can construct and assert on them without importing
// internal/generated.
type (
	// Webhook is an organization-level delivery destination: an HTTPS endpoint
	// plus the authentication used to call it. Credentials (auth token, header
	// values, signing secret) are never included.
	Webhook = generated.Webhook
	// CreateWebhook is the create response: the webhook plus SigningSecret for
	// HMAC_SHA256 webhooks. That field is returned only here and cannot be
	// fetched again.
	CreateWebhook = generated.CreateWebhookResponse
	ListWebhooks  = generated.ListWebhooksResponse
	// TestWebhook reports the endpoint's response to a test delivery.
	// StatusCode is 502 when no response was received.
	TestWebhook                 = generated.TestWebhookResponse
	WebhookDeliveryAttempt      = generated.WebhookDeliveryAttempt
	ListWebhookDeliveryAttempts = generated.ListWebhookDeliveryAttemptsResponse
	// WebhookSubscription delivers one event from one prompt or evaluator to
	// one webhook.
	WebhookSubscription      = generated.WebhookSubscription
	ListWebhookSubscriptions = generated.ListWebhookSubscriptionsResponse

	// WebhookAuthType is how deliveries from a webhook are authenticated. It
	// is fixed at creation.
	WebhookAuthType = generated.WebhookAuthType
	// WebhookEventType is an event a subscription delivers. Prompt events are
	// valid only for PROMPT sources, evaluator events only for EVALUATOR
	// sources.
	WebhookEventType = generated.WebhookEventType
	// WebhookSourceType is the kind of resource a subscription is attached to.
	WebhookSourceType = generated.WebhookSourceType
)

const (
	// WebhookAuthTypeBearer sends the stored auth token verbatim as the
	// Authorization header of each delivery.
	WebhookAuthTypeBearer WebhookAuthType = generated.WebhookAuthTypeBEARER
	// WebhookAuthTypeHMACSHA256 signs each delivery with a server-generated
	// secret returned once, in the create response.
	WebhookAuthTypeHMACSHA256 WebhookAuthType = generated.WebhookAuthTypeHMACSHA256

	WebhookEventTypePromptVersionCreated    WebhookEventType = generated.WebhookEventTypePROMPTVERSIONCREATED
	WebhookEventTypePromptVersionLabeled    WebhookEventType = generated.WebhookEventTypePROMPTVERSIONLABELED
	WebhookEventTypePromptVersionUnlabeled  WebhookEventType = generated.WebhookEventTypePROMPTVERSIONUNLABELED
	WebhookEventTypeEvaluatorVersionCreated WebhookEventType = generated.WebhookEventTypeEVALUATORVERSIONCREATED

	WebhookSourceTypePrompt    WebhookSourceType = generated.WebhookSourceTypePROMPT
	WebhookSourceTypeEvaluator WebhookSourceType = generated.WebhookSourceTypeEVALUATOR
)

// ListRequest holds optional filters for listing webhooks.
type ListRequest struct {
	// Organization is an optional name-or-ID filter. When non-empty, only
	// webhooks in that organization are returned; when empty, webhooks across
	// every organization the caller can read are returned.
	Organization string
	// Name is an optional case-insensitive substring filter on the webhook
	// name. When empty, results are not filtered by name.
	Name string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response's pagination.next_cursor. When empty, results start from the
	// first page.
	Cursor string
}

// GetRequest identifies the webhook to fetch.
type GetRequest struct {
	// Webhook accepts either a webhook name or ID.
	Webhook string
	// Organization accepts either an organization name or ID. Required when
	// Webhook is a name; ignored when Webhook is an ID.
	Organization string
}

// CreateRequest describes a new webhook.
type CreateRequest struct {
	// Organization accepts either an organization name or ID and identifies
	// the organization that owns the webhook.
	Organization string
	// Name is the webhook's name (unique within the organization, max 255
	// characters).
	Name string
	// URL is the HTTPS endpoint events are delivered to.
	URL string
	// Description is optional. When empty, the server stores an empty
	// description.
	Description string
	// AuthType is optional and fixed after creation. When empty, the server
	// applies its default (WebhookAuthTypeBearer). For
	// WebhookAuthTypeHMACSHA256 the server generates a signing secret and
	// returns it once, in the create response.
	AuthType WebhookAuthType
	// AuthToken is the optional complete Authorization header value sent with
	// each delivery, e.g. "Bearer my-token". It is sent verbatim, so include
	// the "Bearer " prefix if the endpoint expects one. Only valid when
	// AuthType is WebhookAuthTypeBearer. Never returned. When empty, no
	// Authorization header is configured.
	AuthToken string
	// TimeoutMs is the optional delivery timeout in milliseconds (1000 to
	// 60000). When zero, the server applies its default (30000).
	TimeoutMs int
	// Headers are optional custom HTTP headers sent with each delivery (at
	// most 20). Header values are never returned. When nil, no custom headers
	// are configured.
	Headers map[string]string
}

// UpdateRequest identifies the webhook to update and the fields to patch. At
// least one patch field must be non-nil. AuthType cannot be changed after
// creation, and the signing secret of an HMAC_SHA256 webhook cannot be
// rotated; create a new webhook instead.
type UpdateRequest struct {
	// Webhook accepts either a webhook name or ID.
	Webhook string
	// Organization accepts either an organization name or ID. Required when
	// Webhook is a name; ignored when Webhook is an ID.
	Organization string
	// Name, when non-nil, sets a new name (unique within the organization);
	// when nil, the existing name is preserved.
	Name *string
	// Description, when non-nil, sets a new description. JSON null clears it;
	// because nil preserves the existing description, pass a pointer to an
	// empty string to have the SDK send JSON null.
	Description *string
	// URL, when non-nil, sets a new HTTPS endpoint; when nil, the existing URL
	// is preserved.
	URL *string
	// AuthToken, when non-nil, replaces the Authorization header value sent
	// with each delivery. Only valid when the webhook's AuthType is
	// WebhookAuthTypeBearer. A pointer to an empty string removes the token,
	// so deliveries carry no Authorization header. When nil, the existing
	// token is preserved.
	AuthToken *string
	// TimeoutMs, when non-nil, sets a new delivery timeout in milliseconds
	// (1000 to 60000); when nil, the existing timeout is preserved.
	TimeoutMs *int
	// Headers, when non-nil, replaces the whole custom header map; headers not
	// included are removed, and an empty map removes them all. When nil, the
	// existing headers are preserved.
	Headers *map[string]string
}

// DeleteRequest identifies the webhook to delete.
type DeleteRequest struct {
	// Webhook accepts either a webhook name or ID.
	Webhook string
	// Organization accepts either an organization name or ID. Required when
	// Webhook is a name; ignored when Webhook is an ID.
	Organization string
}

// TestRequest identifies the webhook to send a test delivery to.
type TestRequest struct {
	// Webhook accepts either a webhook name or ID.
	Webhook string
	// Organization accepts either an organization name or ID. Required when
	// Webhook is a name; ignored when Webhook is an ID.
	Organization string
}

// ListDeliveryAttemptsRequest identifies the webhook (resolved by name or ID)
// and pagination options for listing its delivery attempts.
type ListDeliveryAttemptsRequest struct {
	// Webhook accepts either a webhook name or ID.
	Webhook string
	// Organization accepts either an organization name or ID. Required when
	// Webhook is a name; ignored when Webhook is an ID.
	Organization string
	// Limit is the optional maximum number of items to return (max 500). When
	// zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

// ListSubscriptionsRequest holds optional filters for listing webhook
// subscriptions. SourceType and SourceID must be set together or not at all.
type ListSubscriptionsRequest struct {
	// SourceType, when non-empty, restricts results to one kind of source.
	// Must be paired with SourceID. When both are empty, subscriptions from
	// every readable source are returned.
	SourceType WebhookSourceType
	// SourceID, when non-empty, restricts results to one prompt or evaluator
	// by strict ID. Must be paired with SourceType.
	SourceID string
	// Limit is the optional maximum number of items to return (max 100). When
	// zero, the SDK applies a default of 50.
	Limit int
	// Cursor is the optional opaque pagination cursor from a previous
	// response. When empty, results start from the first page.
	Cursor string
}

// CreateSubscriptionRequest subscribes a webhook to one event on one prompt
// or evaluator. To deliver several events to the same webhook, create one
// subscription per event.
type CreateSubscriptionRequest struct {
	// Webhook accepts either a webhook name or ID. The webhook must belong to
	// the source's organization.
	Webhook string
	// Organization accepts either an organization name or ID. Required when
	// Webhook is a name; ignored when Webhook is an ID.
	Organization string
	// SourceType is the kind of resource to attach the webhook to.
	SourceType WebhookSourceType
	// SourceID is the strict ID of the prompt or evaluator.
	SourceID string
	// Event is the event to deliver. It must belong to SourceType.
	Event WebhookEventType
}

// GetSubscriptionRequest identifies the subscription to fetch.
type GetSubscriptionRequest struct {
	// SubscriptionID is the strict ID of the subscription (no name resolution).
	SubscriptionID string
}

// DeleteSubscriptionRequest identifies the subscription to delete.
type DeleteSubscriptionRequest struct {
	// SubscriptionID is the strict ID of the subscription (no name resolution).
	SubscriptionID string
}
