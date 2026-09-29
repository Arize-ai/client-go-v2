// Package main demonstrates how to use the webhooks subclient of the Arize Go
// SDK v2.
//
// Webhooks are an Alpha feature; every call prints a one-time pre-release
// warning.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/webhooks
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/webhooks"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/update/delete.
	// Organization accepts either a name or ID; promptID is a strict ID.
	const (
		organization = "<organization-name-or-id>"
		webhookName  = "example-webhook"
		endpointURL  = "https://example.com/arize-events"
		promptID     = "<prompt-id>"
	)

	listWebhooks(ctx, client, organization)

	wh := createWebhook(ctx, client, organization, webhookName, endpointURL)
	getWebhook(ctx, client, webhookName, organization)
	updateWebhook(ctx, client, webhookName, organization)
	testWebhook(ctx, client, wh.Id)
	listDeliveryAttempts(ctx, client, wh.Id)

	sub := createSubscription(ctx, client, wh.Id, promptID)
	listSubscriptions(ctx, client, promptID)
	getSubscription(ctx, client, sub.Id)
	deleteSubscription(ctx, client, sub.Id)

	deleteWebhook(ctx, client, wh.Id)
}

// listWebhooks accepts an optional organization name or ID; leave it empty to
// list across every organization the key can read.
func listWebhooks(ctx context.Context, client *arize.Client, organization string) {
	resp, err := client.Webhooks.List(ctx, webhooks.ListRequest{Organization: organization, Limit: 25})
	if err != nil {
		log.Fatalf("list webhooks: %v", err)
	}
	for _, wh := range resp.Webhooks {
		fmt.Printf("  %s\t%s\t%s\n", wh.Id, wh.Name, wh.Url)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// createWebhook creates a BEARER webhook. For an HMAC_SHA256 webhook set
// AuthType to webhooks.WebhookAuthTypeHMACSHA256 and store SigningSecret from
// the response: it is returned only once.
func createWebhook(ctx context.Context, client *arize.Client, organization, name, url string) *webhooks.CreateWebhook {
	wh, err := client.Webhooks.Create(ctx, webhooks.CreateRequest{
		Organization: organization,
		Name:         name,
		URL:          url,
		Description:  "Example webhook created by the Go SDK",
		AuthType:     webhooks.WebhookAuthTypeBearer,
		AuthToken:    "Bearer example-token",
		TimeoutMs:    10000,
		Headers:      map[string]string{"X-Source": "arize-go-sdk"},
	})
	if err != nil {
		log.Fatalf("create webhook: %v", err)
	}
	fmt.Printf("created webhook %s (%s)\n", wh.Name, wh.Id)
	if wh.SigningSecret != nil {
		fmt.Println("signing secret returned; store it now, it cannot be fetched again")
	}
	return wh
}

// getWebhook accepts a webhook name or ID. Organization is required when the
// webhook is a name.
func getWebhook(ctx context.Context, client *arize.Client, webhook, organization string) {
	wh, err := client.Webhooks.Get(ctx, webhooks.GetRequest{Webhook: webhook, Organization: organization})
	if err != nil {
		var nfe *arize.ResourceNotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("webhook %q not found in organization %q\n", webhook, organization)
			return
		}
		log.Fatalf("get webhook: %v", err)
	}
	fmt.Printf("found webhook %s (%s) auth=%s timeout=%dms\n", wh.Name, wh.Id, wh.AuthType, wh.TimeoutMs)
}

// updateWebhook patches the description and timeout; nil fields are left
// unchanged. Pass a pointer to an empty string as Description to clear it.
func updateWebhook(ctx context.Context, client *arize.Client, webhook, organization string) {
	description := "Updated by the Go SDK example"
	timeout := 15000
	wh, err := client.Webhooks.Update(ctx, webhooks.UpdateRequest{
		Webhook:      webhook,
		Organization: organization,
		Description:  &description,
		TimeoutMs:    &timeout,
	})
	if err != nil {
		log.Fatalf("update webhook: %v", err)
	}
	fmt.Printf("updated webhook %s: %q, %dms\n", wh.Name, wh.Description, wh.TimeoutMs)
}

// testWebhook sends a test event. A nil error means the test ran; the response
// carries the endpoint's status code and error message.
func testWebhook(ctx context.Context, client *arize.Client, webhookID string) {
	resp, err := client.Webhooks.Test(ctx, webhooks.TestRequest{Webhook: webhookID})
	if err != nil {
		log.Fatalf("test webhook: %v", err)
	}
	if resp.ErrorMessage != nil {
		fmt.Printf("test delivery returned %d: %s\n", resp.StatusCode, *resp.ErrorMessage)
		return
	}
	fmt.Printf("test delivery returned %d\n", resp.StatusCode)
}

func listDeliveryAttempts(ctx context.Context, client *arize.Client, webhookID string) {
	resp, err := client.Webhooks.ListDeliveryAttempts(ctx, webhooks.ListDeliveryAttemptsRequest{Webhook: webhookID, Limit: 10})
	if err != nil {
		log.Fatalf("list delivery attempts: %v", err)
	}
	for _, a := range resp.DeliveryAttempts {
		status := "no response"
		if a.StatusCode != nil {
			status = fmt.Sprint(*a.StatusCode)
		}
		fmt.Printf("  event %s attempt %d -> %s\n", a.EventId, a.AttemptNumber, status)
	}
}

// createSubscription delivers one event from one prompt to the webhook. Create
// one subscription per event to deliver several events.
func createSubscription(ctx context.Context, client *arize.Client, webhookID, promptID string) *webhooks.WebhookSubscription {
	sub, err := client.Webhooks.CreateSubscription(ctx, webhooks.CreateSubscriptionRequest{
		Webhook:    webhookID,
		SourceType: webhooks.WebhookSourceTypePrompt,
		SourceID:   promptID,
		Event:      webhooks.WebhookEventTypePromptVersionCreated,
	})
	if err != nil {
		log.Fatalf("create subscription: %v", err)
	}
	fmt.Printf("created subscription %s (%s on %s)\n", sub.Id, sub.Event, sub.SourceId)
	return sub
}

// listSubscriptions filters to one source. SourceType and SourceID go
// together; leave both empty to list every readable subscription.
func listSubscriptions(ctx context.Context, client *arize.Client, promptID string) {
	resp, err := client.Webhooks.ListSubscriptions(ctx, webhooks.ListSubscriptionsRequest{
		SourceType: webhooks.WebhookSourceTypePrompt,
		SourceID:   promptID,
	})
	if err != nil {
		log.Fatalf("list subscriptions: %v", err)
	}
	for _, sub := range resp.Subscriptions {
		fmt.Printf("  %s\t%s\t%s\n", sub.Id, sub.Event, sub.WebhookId)
	}
}

func getSubscription(ctx context.Context, client *arize.Client, subscriptionID string) {
	sub, err := client.Webhooks.GetSubscription(ctx, webhooks.GetSubscriptionRequest{SubscriptionID: subscriptionID})
	if err != nil {
		log.Fatalf("get subscription: %v", err)
	}
	fmt.Printf("found subscription %s delivering %s\n", sub.Id, sub.Event)
}

func deleteSubscription(ctx context.Context, client *arize.Client, subscriptionID string) {
	if err := client.Webhooks.DeleteSubscription(ctx, webhooks.DeleteSubscriptionRequest{SubscriptionID: subscriptionID}); err != nil {
		log.Fatalf("delete subscription: %v", err)
	}
	fmt.Printf("deleted subscription %s\n", subscriptionID)
}

func deleteWebhook(ctx context.Context, client *arize.Client, webhookID string) {
	if err := client.Webhooks.Delete(ctx, webhooks.DeleteRequest{Webhook: webhookID}); err != nil {
		log.Fatalf("delete webhook: %v", err)
	}
	fmt.Printf("deleted webhook %s\n", webhookID)
}
