// Package main demonstrates how to use the integrations subclient of the
// Arize Go SDK v2: the polymorphic /v2/integrations surface covering both LLM
// (model-provider) and agent (customer-hosted endpoint) integrations.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/integrations
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/integrations"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/get/update/delete.
	const (
		llmName        = "example-openai"
		fireworksName  = "example-fireworks"
		togetherAiName = "example-together-ai"
		agentName      = "example-agent"
	)

	listIntegrations(ctx, client)

	llmID := createLLMIntegration(ctx, client, llmName)
	fireworksID := createFireworksIntegration(ctx, client, fireworksName)
	togetherAiID := createTogetherIntegration(ctx, client, togetherAiName)
	agentID := createAgentIntegration(ctx, client, agentName)

	getIntegration(ctx, client, llmID)
	updateLLMIntegration(ctx, client, llmID)
	updateAgentIntegration(ctx, client, agentID)

	deleteIntegration(ctx, client, llmID)
	deleteIntegration(ctx, client, fireworksID)
	deleteIntegration(ctx, client, togetherAiID)
	deleteIntegration(ctx, client, agentID)
}

// listIntegrations lists integrations two ways: filtered to a single type,
// and untyped — one merged list of every type, where each item is a
// type-tagged union discriminated by its `type` field.
func listIntegrations(ctx context.Context, client *arize.Client) {
	llms, err := client.Integrations.List(ctx, integrations.ListRequest{
		Type:  integrations.IntegrationTypeLLM,
		Limit: 25,
	})
	if err != nil {
		log.Fatalf("list llm integrations: %v", err)
	}
	fmt.Printf("llm integrations: %d\n", len(llms.Integrations))

	all, err := client.Integrations.List(ctx, integrations.ListRequest{Limit: 25})
	if err != nil {
		log.Fatalf("list integrations: %v", err)
	}
	for _, it := range all.Integrations {
		id, name, typ := unwrapIntegration(it)
		fmt.Printf("  %s\t%s\t%s\n", id, name, typ)
	}
	if all.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// createLLMIntegration creates an OpenAI integration with a placeholder API
// key. Set exactly one provider field on CreateLLMConfig; the SDK fills in
// the provider discriminator.
func createLLMIntegration(ctx context.Context, client *arize.Client, name string) string {
	created, err := client.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
		Name: name,
		Config: integrations.CreateLLMConfig{
			OpenAI: &integrations.CreateOpenAIConfig{APIKey: "sk-placeholder"},
		},
	})
	if err != nil {
		log.Fatalf("create llm integration: %v", err)
	}
	id, createdName, _ := unwrapIntegration(*created)
	fmt.Printf("created llm integration %s (%s)\n", createdName, id)
	return id
}

// createFireworksIntegration creates a Fireworks AI integration with only a
// placeholder API key. A hosted provider resolves its own model list, so Arize
// reads the models the key can reach from the Fireworks account and the
// integration has a selectable model list without any ModelNames.
func createFireworksIntegration(ctx context.Context, client *arize.Client, name string) string {
	created, err := client.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
		Name: name,
		Config: integrations.CreateLLMConfig{
			Fireworks: &integrations.CreateFireworksConfig{APIKey: "fw-placeholder"},
		},
	})
	if err != nil {
		log.Fatalf("create fireworks integration: %v", err)
	}
	id, createdName, _ := unwrapIntegration(*created)
	fmt.Printf("created fireworks integration %s (%s)\n", createdName, id)
	return id
}

// createTogetherIntegration creates a Together AI integration with only a
// placeholder API key. A hosted provider resolves its own model list, so Arize
// reads the models the key can reach from the Together AI account and the
// integration has a selectable model list without any ModelNames.
func createTogetherIntegration(ctx context.Context, client *arize.Client, name string) string {
	created, err := client.Integrations.CreateLLM(ctx, integrations.CreateLLMRequest{
		Name: name,
		Config: integrations.CreateLLMConfig{
			TogetherAi: &integrations.CreateTogetherAiConfig{APIKey: "together-placeholder"},
		},
	})
	if err != nil {
		log.Fatalf("create together ai integration: %v", err)
	}
	id, createdName, _ := unwrapIntegration(*created)
	fmt.Printf("created together ai integration %s (%s)\n", createdName, id)
	return id
}

// createAgentIntegration creates an agent integration pointing at a
// customer-hosted HTTPS endpoint, with a JSON Schema describing the request
// payload and a reusable request preset.
func createAgentIntegration(ctx context.Context, client *arize.Client, name string) string {
	created, err := client.Integrations.CreateAgent(ctx, integrations.CreateAgentRequest{
		Name:     name,
		Endpoint: "https://agent.example.com/invoke",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"input": map[string]any{"type": "string"}},
		},
		RequestPresets: []integrations.AgentRequestPresetInput{
			{Name: "default", Config: map[string]any{"input": "hello"}},
		},
	})
	if err != nil {
		log.Fatalf("create agent integration: %v", err)
	}
	id, createdName, _ := unwrapIntegration(*created)
	fmt.Printf("created agent integration %s (%s)\n", createdName, id)
	return id
}

// getIntegration accepts a name or ID. Type is required when integration is a
// name (names are unique per account and type); ignored when it is an ID.
func getIntegration(ctx context.Context, client *arize.Client, integration string) {
	it, err := client.Integrations.Get(ctx, integrations.GetRequest{
		Integration: integration,
	})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("integration %q not found\n", integration)
			return
		}
		log.Fatalf("get integration: %v", err)
	}
	id, name, typ := unwrapIntegration(*it)
	fmt.Printf("found %s integration %s (%s)\n", typ, name, id)
}

// updateLLMIntegration demonstrates two PATCH semantics:
//   - rotating the API key by passing a new non-empty value
//   - fields left nil are preserved
func updateLLMIntegration(ctx context.Context, client *arize.Client, integrationID string) {
	newKey := "sk-rotated"
	updated, err := client.Integrations.UpdateLLM(ctx, integrations.UpdateLLMRequest{
		Integration: integrationID,
		APIKey:      &newKey,
	})
	if err != nil {
		log.Fatalf("update llm integration: %v", err)
	}
	id, _, _ := unwrapIntegration(*updated)
	fmt.Printf("updated llm integration %s\n", id)
}

// updateAgentIntegration renames the agent and clears its description by
// passing a pointer to the empty string (the SDK emits JSON null on the
// wire — the OpenAPI "Pass null to remove" signal).
func updateAgentIntegration(ctx context.Context, client *arize.Client, integrationID string) {
	newName := "example-agent-renamed"
	clearDescription := "" // &"" → clears the existing description on the server
	updated, err := client.Integrations.UpdateAgent(ctx, integrations.UpdateAgentRequest{
		Integration: integrationID,
		Name:        &newName,
		Description: &clearDescription,
	})
	if err != nil {
		log.Fatalf("update agent integration: %v", err)
	}
	id, _, _ := unwrapIntegration(*updated)
	fmt.Printf("updated agent integration %s\n", id)
}

func deleteIntegration(ctx context.Context, client *arize.Client, integrationID string) {
	if err := client.Integrations.Delete(ctx, integrations.DeleteRequest{
		Integration: integrationID,
	}); err != nil {
		log.Fatalf("delete integration: %v", err)
	}
	fmt.Printf("deleted integration %s\n", integrationID)
}

// unwrapIntegration reads the active variant of the type-tagged Integration
// union and returns its shared id/name fields plus the discriminator.
func unwrapIntegration(it integrations.Integration) (id, name, typ string) {
	d, err := it.Discriminator()
	if err != nil {
		log.Fatalf("integration discriminator: %v", err)
	}
	switch d {
	case string(integrations.IntegrationTypeLLM):
		llm, err := it.AsLlmIntegration()
		if err != nil {
			log.Fatalf("unwrap llm integration: %v", err)
		}
		return llm.Id, llm.Name, d
	case string(integrations.IntegrationTypeAgent):
		agent, err := it.AsAgentIntegration()
		if err != nil {
			log.Fatalf("unwrap agent integration: %v", err)
		}
		return agent.Id, agent.Name, d
	default:
		log.Fatalf("unknown integration type %q", d)
		return "", "", ""
	}
}
