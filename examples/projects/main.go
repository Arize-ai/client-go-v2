// Package main demonstrates how to use the projects subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/projects
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/projects"
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
		space       = "U3BhY2U6MTox"
		projectName = "example-project"
	)

	listProjects(ctx, client)

	proj := createProject(ctx, client, projectName, space)
	getProject(ctx, client, projectName, space)
	deleteProject(ctx, client, proj.Id)
}

func listProjects(ctx context.Context, client *arize.Client) {
	resp, err := client.Projects.List(ctx, projects.ListRequest{Limit: 25})
	if err != nil {
		log.Fatalf("list projects: %v", err)
	}
	for _, p := range resp.Projects {
		fmt.Printf("  %s\t%s\n", p.Id, p.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// getProject accepts a project name or ID. Space is required when project is
// a name.
func getProject(ctx context.Context, client *arize.Client, project, space string) {
	proj, err := client.Projects.Get(ctx, projects.GetRequest{Project: project, Space: space})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("project %q not found in space %q\n", project, space)
			return
		}
		log.Fatalf("get project: %v", err)
	}
	fmt.Printf("found project %s (%s)\n", proj.Name, proj.Id)
}

// createProject creates a project. Space accepts either a space name or ID;
// it is resolved internally to a space ID.
func createProject(ctx context.Context, client *arize.Client, name, space string) *projects.Project {
	proj, err := client.Projects.Create(ctx, projects.CreateRequest{
		Name:  name,
		Space: space,
	})
	if err != nil {
		log.Fatalf("create project: %v", err)
	}
	fmt.Printf("created project %s (%s)\n", proj.Name, proj.Id)
	return proj
}

func deleteProject(ctx context.Context, client *arize.Client, projectID string) {
	if err := client.Projects.Delete(ctx, projects.DeleteRequest{Project: projectID}); err != nil {
		log.Fatalf("delete project: %v", err)
	}
	fmt.Printf("deleted project %s\n", projectID)
}
