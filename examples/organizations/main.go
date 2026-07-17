// Package main demonstrates how to use the organizations subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/organizations
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/organizations"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	listOrganizations(ctx, client)

	org := createOrganization(ctx, client, "example-org")
	getOrganization(ctx, client, org.Name)
	updateOrganization(ctx, client, org.Id, "example-org-renamed")

	// Membership management (these require a real user id).
	if userID := os.Getenv("ARIZE_EXAMPLE_USER_ID"); userID != "" {
		addOrganizationUser(ctx, client, org.Id, userID)
		removeOrganizationUser(ctx, client, org.Id, userID)
	} else {
		fmt.Println("set ARIZE_EXAMPLE_USER_ID to demo add/remove user")
	}

	deleteOrganization(ctx, client, org.Id)
}

func listOrganizations(ctx context.Context, client *arize.Client) {
	resp, err := client.Organizations.List(ctx, organizations.ListRequest{Limit: 25})
	if err != nil {
		log.Fatalf("list organizations: %v", err)
	}
	for _, o := range resp.Organizations {
		fmt.Printf("  %s\t%s\n", o.Id, o.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// getOrganization accepts either an organization name or ID.
func getOrganization(ctx context.Context, client *arize.Client, nameOrID string) {
	org, err := client.Organizations.Get(ctx, organizations.GetRequest{Organization: nameOrID})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("organization %q not found\n", nameOrID)
			return
		}
		log.Fatalf("get organization: %v", err)
	}
	fmt.Printf("found organization %s (%s)\n", org.Name, org.Id)
}

func createOrganization(ctx context.Context, client *arize.Client, name string) *organizations.Organization {
	org, err := client.Organizations.Create(ctx, organizations.CreateRequest{Name: name})
	if err != nil {
		log.Fatalf("create organization: %v", err)
	}
	fmt.Printf("created organization %s (%s)\n", org.Name, org.Id)
	return org
}

// updateOrganization accepts either an organization name or ID.
func updateOrganization(ctx context.Context, client *arize.Client, orgID, newName string) {
	org, err := client.Organizations.Update(ctx, organizations.UpdateRequest{
		Organization: orgID,
		Name:         &newName,
	})
	if err != nil {
		log.Fatalf("update organization: %v", err)
	}
	fmt.Printf("renamed organization %s to %s\n", org.Id, org.Name)
}

// addOrganizationUser adds a user with the predefined "MEMBER" role. Custom
// role assignments are not yet supported by the server.
func addOrganizationUser(ctx context.Context, client *arize.Client, orgID, userID string) {
	m, err := client.Organizations.AddUser(ctx, organizations.AddUserRequest{
		Organization: orgID,
		UserID:       userID,
		Role:         organizations.PredefinedOrgRole{Name: organizations.OrganizationRoleMember},
	})
	if err != nil {
		log.Fatalf("add user: %v", err)
	}
	fmt.Printf("added user %s to organization %s (membership %s)\n", m.UserId, m.OrganizationId, m.Id)

	// The server returns the role as a discriminated union. ValueByDiscriminator
	// reads the discriminator and returns the matching variant; type-switch to
	// handle every variant.
	role, err := m.Role.ValueByDiscriminator()
	if err != nil {
		log.Printf("decode role: %v", err)
		return
	}
	switch r := role.(type) {
	case organizations.PredefinedOrgRole:
		fmt.Printf("  role: predefined %s\n", r.Name)
	case organizations.CustomOrgRole:
		fmt.Printf("  role: custom %s\n", r.Id)
	}
}

func removeOrganizationUser(ctx context.Context, client *arize.Client, orgID, userID string) {
	if err := client.Organizations.RemoveUser(ctx, organizations.RemoveUserRequest{
		Organization: orgID,
		UserID:       userID,
	}); err != nil {
		log.Fatalf("remove user: %v", err)
	}
	fmt.Printf("removed user %s from organization %s\n", userID, orgID)
}

// deleteOrganization irreversibly deletes the organization and all child
// resources. Use with care.
func deleteOrganization(ctx context.Context, client *arize.Client, orgID string) {
	if err := client.Organizations.Delete(ctx, organizations.DeleteRequest{Organization: orgID}); err != nil {
		log.Fatalf("delete organization: %v", err)
	}
	fmt.Printf("deleted organization %s\n", orgID)
}
