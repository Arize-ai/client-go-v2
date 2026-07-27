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

	toRevoke := createAPIKey(ctx, client, "example-key-to-revoke")
	revokeAPIKey(ctx, client, toRevoke.Id)

	// Provide real org and space IDs from your Arize account when running.
	createServiceKey(ctx, client, "example-service-key",
		"<org-hmac-id>", []string{"<space-hmac-id>"})
}

// listAPIKeys shows filtering the list by key_type and status.
func listAPIKeys(ctx context.Context, client *arize.Client) {
	resp, err := client.APIKeys.List(ctx, apikeys.ListRequest{
		KeyType: apikeys.APIKeyTypeUser,
		Status:  apikeys.APIKeyStatusActive,
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
func createAPIKey(ctx context.Context, client *arize.Client, name string) *apikeys.UserApiKeyCreated {
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

// refreshAPIKey rotates an API key and returns the new plaintext key —
// store it immediately, you cannot retrieve it later.
func refreshAPIKey(ctx context.Context, client *arize.Client, keyID string) {
	rotated, err := client.APIKeys.Refresh(ctx, apikeys.RefreshRequest{
		APIKeyID:  keyID,
		ExpiresAt: time.Now().Add(90 * 24 * time.Hour),
		// Allow 5 minutes for callers to switch to the new key before the
		// old one is invalidated. Omit or set to 0 for immediate revocation.
		GracePeriodSeconds: 300,
	})
	if err != nil {
		log.Fatalf("refresh api key: %v", err)
	}
	// The new raw key value is only returned at rotation. Store it securely —
	// it cannot be retrieved again.
	fmt.Printf("rotated api key %s — new secret: %s\n", rotated.Id, rotated.Key)
}

// revokeAPIKey sets the key's status to revoked, deactivating it immediately.
// Revoking is irreversible; revoking an already-revoked key is a no-op.
func revokeAPIKey(ctx context.Context, client *arize.Client, keyID string) {
	if err := client.APIKeys.Revoke(ctx, apikeys.RevokeRequest{APIKeyID: keyID}); err != nil {
		log.Fatalf("revoke api key: %v", err)
	}
}

// createServiceKey creates a bot service key scoped to a specific set of spaces.
// orgID and spaceIDs are HMAC-encoded IDs from the Arize UI or API.
//
// To assign a custom RBAC role to a space instead of a predefined one, set
// SpaceBinding.CustomRoleID to the role's HMAC-encoded ID and leave Role empty.
func createServiceKey(ctx context.Context, client *arize.Client, name, orgID string, spaceIDs []string) {
	spaces := make([]apikeys.SpaceBinding, 0, len(spaceIDs))
	for _, id := range spaceIDs {
		spaces = append(spaces, apikeys.SpaceBinding{
			Space: id,
			Role:  apikeys.AssignPredefinedSpaceRole(apikeys.SpaceRoleMember),
		})
	}

	created, err := client.APIKeys.CreateServiceKey(ctx, apikeys.CreateServiceKeyRequest{
		Name:        name,
		AccountRole: apikeys.AssignPredefinedAccountRole(apikeys.AccountRoleMember),
		Orgs: []apikeys.OrgBinding{
			{
				OrgID:  orgID,
				Role:   apikeys.AssignPredefinedOrgRole(apikeys.OrgRoleReadOnly),
				Spaces: spaces,
			},
		},
		ExpiresAt: time.Now().Add(90 * 24 * time.Hour),
	})
	if err != nil {
		log.Fatalf("create service key: %v", err)
	}
	// The raw key value is only returned at creation. Store it securely —
	// it cannot be retrieved again.
	fmt.Printf("created service key %s (bot user %s) — secret: %s\n",
		created.Id, created.BotUser.Id, created.Key)
}
