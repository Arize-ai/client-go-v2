// Package main demonstrates how to use the prompts subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/prompts
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/prompts"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/get/delete.
	// Space accepts either a space name or ID.
	const (
		space      = "U3BhY2U6MTox"
		promptName = "example-prompt"
	)

	listPrompts(ctx, client)

	p := createPrompt(ctx, client, promptName, space)
	getPrompt(ctx, client, promptName, space)
	v := addVersion(ctx, client, promptName, space)
	promoteVersion(ctx, client, v.Id)
	deletePrompt(ctx, client, p.Id)
}

func listPrompts(ctx context.Context, client *arize.Client) {
	resp, err := client.Prompts.List(ctx, prompts.ListRequest{Limit: 25})
	if err != nil {
		log.Fatalf("list prompts: %v", err)
	}
	for _, p := range resp.Prompts {
		fmt.Printf("  %s\t%s\n", p.Id, p.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// getPrompt accepts a prompt name or ID. Space is required when prompt is a
// name.
func getPrompt(ctx context.Context, client *arize.Client, prompt, space string) {
	p, err := client.Prompts.Get(ctx, prompts.GetRequest{Prompt: prompt, Space: space})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("prompt %q not found in space %q\n", prompt, space)
			return
		}
		log.Fatalf("get prompt: %v", err)
	}
	fmt.Printf("found prompt %s (%s)\n", p.Name, p.Id)
}

// createPrompt creates a prompt with its initial version. Space accepts either
// a space name or ID; it is resolved internally to a space ID.
func createPrompt(ctx context.Context, client *arize.Client, name, space string) *prompts.PromptWithVersion {
	content := "You are a helpful assistant. Answer the user's question: {question}"
	p, err := client.Prompts.Create(ctx, prompts.CreateRequest{
		Name:  name,
		Space: space,
		Version: prompts.PromptVersionCreate{
			CommitMessage: "initial version",
			Provider:      prompts.LlmProviderOpenAi,
			Messages: []prompts.LLMMessage{
				{Role: prompts.MessageRoleSystem, Content: &content},
			},
		},
	})
	if err != nil {
		log.Fatalf("create prompt: %v", err)
	}
	fmt.Printf("created prompt %s (%s)\n", p.Name, p.Id)
	return p
}

// addVersion appends a new version to an existing prompt, resolved by name.
func addVersion(ctx context.Context, client *arize.Client, prompt, space string) *prompts.PromptVersion {
	content := "You are a concise assistant. Answer briefly: {question}"
	v, err := client.Prompts.CreateVersion(ctx, prompts.CreateVersionRequest{
		Prompt:        prompt,
		Space:         space,
		CommitMessage: "make answers concise",
		Provider:      prompts.LlmProviderOpenAi,
		Messages: []prompts.LLMMessage{
			{Role: prompts.MessageRoleSystem, Content: &content},
		},
	})
	if err != nil {
		log.Fatalf("create version: %v", err)
	}
	fmt.Printf("created version %s\n", v.Id)
	return v
}

// promoteVersion assigns the "production" label to a specific version.
func promoteVersion(ctx context.Context, client *arize.Client, versionID string) {
	labels, err := client.Prompts.SetVersionLabels(ctx, prompts.SetVersionLabelsRequest{
		VersionID: versionID,
		Labels:    []string{"production"},
	})
	if err != nil {
		log.Fatalf("set version labels: %v", err)
	}
	fmt.Printf("labeled version %s; labels now %v\n", versionID, labels.Labels)
}

func deletePrompt(ctx context.Context, client *arize.Client, promptID string) {
	if err := client.Prompts.Delete(ctx, prompts.DeleteRequest{Prompt: promptID}); err != nil {
		log.Fatalf("delete prompt: %v", err)
	}
	fmt.Printf("deleted prompt %s\n", promptID)
}
