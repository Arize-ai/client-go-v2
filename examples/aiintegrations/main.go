// Package main demonstrates how to use the aiintegrations subclient of the
// Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/aiintegrations
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/aiintegrations"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/get/update/delete.
	// space accepts either a space name or ID.
	const (
		space           = "U3BhY2U6MTox"
		integrationName = "example-anthropic"
	)

	listIntegrations(ctx, client)

	created := createIntegration(ctx, client, integrationName)
	getIntegration(ctx, client, created.Id)
	updateIntegration(ctx, client, created.Id)
	deleteIntegration(ctx, client, created.Id)

	// See createBedrockIntegration / createVertexIntegration below for how to
	// build provider-metadata payloads for AWS Bedrock and Vertex AI. Those
	// require real cloud-side configuration to actually work end-to-end.
	_ = space
}

// listIntegrations shows filtering by name (substring) and pagination cursor.
func listIntegrations(ctx context.Context, client *arize.Client) {
	resp, err := client.AIIntegrations.List(ctx, aiintegrations.ListRequest{
		Limit: 25,
	})
	if err != nil {
		log.Fatalf("list ai integrations: %v", err)
	}
	for _, ai := range resp.AiIntegrations {
		fmt.Printf("  %s\t%s\t%s\n", ai.Id, ai.Name, ai.Provider)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// createIntegration creates an Anthropic integration with a placeholder API
// key. Replace the key with a real one before running.
func createIntegration(ctx context.Context, client *arize.Client, name string) *aiintegrations.AIIntegration {
	created, err := client.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
		Name:     name,
		Provider: aiintegrations.AIIntegrationProviderAnthropic,
		APIKey:   "sk-placeholder",
	})
	if err != nil {
		log.Fatalf("create ai integration: %v", err)
	}
	fmt.Printf("created ai integration %s (%s)\n", created.Name, created.Id)
	return created
}

// getIntegration accepts a name or ID. Space is required when integration is
// a name; ignored when it is an ID.
func getIntegration(ctx context.Context, client *arize.Client, integration string) {
	ai, err := client.AIIntegrations.Get(ctx, aiintegrations.GetRequest{
		Integration: integration,
	})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("ai integration %q not found\n", integration)
			return
		}
		log.Fatalf("get ai integration: %v", err)
	}
	fmt.Printf("found ai integration %s (%s)\n", ai.Name, ai.Id)
}

// updateIntegration demonstrates two PATCH semantics:
//   - rotating the API key by passing a new non-empty value
//   - clearing the base URL by passing a pointer to the empty string (the
//     SDK emits JSON null on the wire — the OpenAPI "Pass null to remove"
//     signal)
//
// Fields left nil are preserved.
func updateIntegration(ctx context.Context, client *arize.Client, integrationID string) {
	newKey := "sk-rotated"
	clearBaseURL := "" // &"" → clears the existing base_url on the server
	updated, err := client.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
		Integration: integrationID,
		APIKey:      &newKey,
		BaseURL:     &clearBaseURL,
	})
	if err != nil {
		log.Fatalf("update ai integration: %v", err)
	}
	fmt.Printf("updated ai integration %s\n", updated.Id)
}

func deleteIntegration(ctx context.Context, client *arize.Client, integrationID string) {
	if err := client.AIIntegrations.Delete(ctx, aiintegrations.DeleteRequest{
		Integration: integrationID,
	}); err != nil {
		log.Fatalf("delete ai integration: %v", err)
	}
	fmt.Printf("deleted ai integration %s\n", integrationID)
}

// createBedrockIntegration shows how to build the provider-metadata payload
// for an AWS Bedrock integration. The role ARN must exist in your AWS
// account and grant Arize permission to assume it.
func createBedrockIntegration(ctx context.Context, client *arize.Client, name string) *aiintegrations.AIIntegration {
	created, err := client.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
		Name:     name,
		Provider: aiintegrations.AIIntegrationProviderAWSBedrock,
		ProviderMetadata: &aiintegrations.ProviderMetadata{
			AWS: &aiintegrations.AWSProviderMetadata{
				RoleArn: "arn:aws:iam::123456789012:role/ArizeBedrockRole",
			},
		},
	})
	if err != nil {
		log.Fatalf("create bedrock integration: %v", err)
	}
	return created
}

// createVertexIntegration shows how to build the provider-metadata payload
// for a Vertex AI (GCP) integration.
func createVertexIntegration(ctx context.Context, client *arize.Client, name string) *aiintegrations.AIIntegration {
	created, err := client.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
		Name:     name,
		Provider: aiintegrations.AIIntegrationProviderVertexAI,
		ProviderMetadata: &aiintegrations.ProviderMetadata{
			GCP: &aiintegrations.GCPProviderMetadata{
				ProjectId:          "my-gcp-project",
				Location:           "us-central1",
				ProjectAccessLabel: "production",
			},
		},
	})
	if err != nil {
		log.Fatalf("create vertex integration: %v", err)
	}
	return created
}

// keep createBedrockIntegration / createVertexIntegration referenced so the
// example compiles even when main() doesn't call them.
var _ = []func(context.Context, *arize.Client, string) *aiintegrations.AIIntegration{
	createBedrockIntegration,
	createVertexIntegration,
}
