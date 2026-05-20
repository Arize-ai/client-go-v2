// Package main demonstrates how to use the apikeys subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/apikeys
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/apikeys"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	listAPIKeys(ctx, client)

	created := createAPIKey(ctx, client, "example-key")
	refreshAPIKey(ctx, client, created.Id)
	deleteAPIKey(ctx, client, created.Id)
}

// listAPIKeys shows filtering the list by key_type and status.
func listAPIKeys(ctx context.Context, client *arize.Client) {
	resp, err := client.APIKeys.List(ctx, apikeys.ListRequest{
		KeyType: apikeys.ApiKeyTypeUser,
		Status:  apikeys.ApiKeyStatusActive,
		Limit:   25,
	})
	if err != nil {
		log.Fatalf("list api keys: %v", err)
	}
	for _, k := range resp.ApiKeys {
		fmt.Printf("  %s\t%s\t%s\n", k.Id, k.Name, k.RedactedKey)
	}
}

// createAPIKey returns the only response that ever contains the plaintext
// Key — store it immediately, you cannot retrieve it later.
func createAPIKey(ctx context.Context, client *arize.Client, name string) *apikeys.ApiKeyCreated {
	created, err := client.APIKeys.Create(ctx, apikeys.CreateRequest{
		Name:      name,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		log.Fatalf("create api key: %v", err)
	}
	// The raw key value is only returned at creation. Store it securely —
	// it cannot be retrieved again.
	fmt.Printf("created api key %s — secret: %s\n", created.Id, created.Key)
	return created
}

func refreshAPIKey(ctx context.Context, client *arize.Client, keyID string) {
	rotated, err := client.APIKeys.Refresh(ctx, apikeys.RefreshRequest{
		ApiKeyID:  keyID,
		ExpiresAt: time.Now().Add(90 * 24 * time.Hour),
	})
	if err != nil {
		log.Fatalf("refresh api key: %v", err)
	}
	// The new raw key value is only returned at rotation. Store it securely —
	// it cannot be retrieved again.
	fmt.Printf("rotated api key %s — new secret: %s\n", rotated.Id, rotated.Key)
}

func deleteAPIKey(ctx context.Context, client *arize.Client, keyID string) {
	if err := client.APIKeys.Delete(ctx, apikeys.DeleteRequest{ApiKeyID: keyID}); err != nil {
		log.Fatalf("delete api key: %v", err)
	}
}
