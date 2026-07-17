// Package main demonstrates how to use the spaces subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/spaces
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/spaces"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/update/delete.
	// Organization accepts either an organization name or ID.
	const (
		organization = "T3JnYW5pemF0aW9uOjE6MQ=="
		spaceName    = "example-space"
	)

	listSpaces(ctx, client)

	space := createSpace(ctx, client, spaceName, organization)
	getSpace(ctx, client, space.Name)
	updateSpace(ctx, client, space.Id, "example-space-renamed")

	// Membership management (these require a real user id).
	if userID := os.Getenv("ARIZE_EXAMPLE_USER_ID"); userID != "" {
		addSpaceUser(ctx, client, space.Id, userID)
		removeSpaceUser(ctx, client, space.Id, userID)
	} else {
		fmt.Println("set ARIZE_EXAMPLE_USER_ID to demo add/remove user")
	}

	deleteSpace(ctx, client, space.Id)
}

func listSpaces(ctx context.Context, client *arize.Client) {
	resp, err := client.Spaces.List(ctx, spaces.ListRequest{Limit: 25})
	if err != nil {
		log.Fatalf("list spaces: %v", err)
	}
	for _, s := range resp.Spaces {
		fmt.Printf("  %s\t%s\n", s.Id, s.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// getSpace accepts either a space name or ID.
func getSpace(ctx context.Context, client *arize.Client, nameOrID string) {
	space, err := client.Spaces.Get(ctx, spaces.GetRequest{Space: nameOrID})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("space %q not found\n", nameOrID)
			return
		}
		log.Fatalf("get space: %v", err)
	}
	fmt.Printf("found space %s (%s)\n", space.Name, space.Id)
}

// createSpace creates a space under the given organization. Organization
// accepts either an organization name or ID; it is resolved internally to an
// organization ID.
func createSpace(ctx context.Context, client *arize.Client, name, organization string) *spaces.Space {
	space, err := client.Spaces.Create(ctx, spaces.CreateRequest{
		Name:         name,
		Organization: organization,
		Description:  "created by the spaces example",
	})
	if err != nil {
		log.Fatalf("create space: %v", err)
	}
	fmt.Printf("created space %s (%s)\n", space.Name, space.Id)
	return space
}

// updateSpace accepts either a space name or ID. A nil patch field is left
// unchanged.
func updateSpace(ctx context.Context, client *arize.Client, spaceID, newName string) {
	space, err := client.Spaces.Update(ctx, spaces.UpdateRequest{
		Space: spaceID,
		Name:  &newName,
	})
	if err != nil {
		log.Fatalf("update space: %v", err)
	}
	fmt.Printf("renamed space %s to %s\n", space.Id, space.Name)
}

// addSpaceUser adds a user with the predefined "MEMBER" role. Build the role
// assignment with spaces.AssignPredefinedRole (one of the UserSpaceRole values)
// or spaces.AssignCustomRole (a custom RBAC role id).
func addSpaceUser(ctx context.Context, client *arize.Client, spaceID, userID string) {
	m, err := client.Spaces.AddUser(ctx, spaces.AddUserRequest{
		Space:  spaceID,
		UserID: userID,
		Role:   spaces.AssignPredefinedRole(spaces.UserSpaceRoleMember),
	})
	if err != nil {
		log.Fatalf("add user: %v", err)
	}
	fmt.Printf("added user %s to space %s (membership %s)\n", m.UserId, m.SpaceId, m.Id)

	// The server returns the role as a discriminated union. ValueByDiscriminator
	// reads the discriminator and returns the matching variant; type-switch to
	// cover both kinds.
	role, err := m.Role.ValueByDiscriminator()
	if err != nil {
		log.Fatalf("decode role: %v", err)
	}
	switch r := role.(type) {
	case spaces.PredefinedSpaceRole:
		fmt.Printf("  role: predefined %s\n", r.Name)
	case spaces.CustomSpaceRole:
		fmt.Printf("  role: custom %s\n", r.Id)
	}
}

func removeSpaceUser(ctx context.Context, client *arize.Client, spaceID, userID string) {
	if err := client.Spaces.RemoveUser(ctx, spaces.RemoveUserRequest{
		Space:  spaceID,
		UserID: userID,
	}); err != nil {
		log.Fatalf("remove user: %v", err)
	}
	fmt.Printf("removed user %s from space %s\n", userID, spaceID)
}

// deleteSpace irreversibly deletes the space and all child resources
// (projects, datasets, monitors, custom metrics, etc.). Use with care.
func deleteSpace(ctx context.Context, client *arize.Client, spaceID string) {
	if err := client.Spaces.Delete(ctx, spaces.DeleteRequest{Space: spaceID}); err != nil {
		log.Fatalf("delete space: %v", err)
	}
	fmt.Printf("deleted space %s\n", spaceID)
}
