// Package main demonstrates how to use the roles subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/roles
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/roles"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	const roleName = "example-role"

	listPredefinedRoles(ctx, client)
	listCustomRoles(ctx, client)

	role := createRole(ctx, client, roleName)
	getRole(ctx, client, roleName)
	updateRole(ctx, client, role.Id)
	deleteRole(ctx, client, role.Id)
}

// listPredefinedRoles shows the tri-state IsPredefined filter. Pass &true to
// return only system-defined roles, &false for custom roles, or leave it nil
// to return both.
func listPredefinedRoles(ctx context.Context, client *arize.Client) {
	predefined := true
	resp, err := client.Roles.List(ctx, roles.ListRequest{
		IsPredefined: &predefined,
		Limit:        25,
	})
	if err != nil {
		log.Fatalf("list predefined roles: %v", err)
	}
	fmt.Println("predefined roles:")
	for _, r := range resp.Roles {
		fmt.Printf("  %s\t%s\n", r.Id, r.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

func listCustomRoles(ctx context.Context, client *arize.Client) {
	custom := false
	resp, err := client.Roles.List(ctx, roles.ListRequest{
		IsPredefined: &custom,
		Limit:        25,
	})
	if err != nil {
		log.Fatalf("list custom roles: %v", err)
	}
	fmt.Println("custom roles:")
	for _, r := range resp.Roles {
		fmt.Printf("  %s\t%s\n", r.Id, r.Name)
	}
}

// getRole accepts either a role name or ID.
func getRole(ctx context.Context, client *arize.Client, nameOrID string) {
	role, err := client.Roles.Get(ctx, roles.GetRequest{Role: nameOrID})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("role %q not found\n", nameOrID)
			return
		}
		log.Fatalf("get role: %v", err)
	}
	fmt.Printf("found role %s (%s) with %d permission(s)\n", role.Name, role.Id, len(role.Permissions))
}

func createRole(ctx context.Context, client *arize.Client, name string) *roles.Role {
	role, err := client.Roles.Create(ctx, roles.CreateRequest{
		Name:        name,
		Description: "example custom role with read-only access to projects",
		Permissions: []roles.Permission{
			roles.Permissions.ProjectRead,
			roles.Permissions.ProjectSpanRead,
		},
	})
	if err != nil {
		log.Fatalf("create role: %v", err)
	}
	fmt.Printf("created role %s (%s)\n", role.Name, role.Id)
	return role
}

// updateRole demonstrates the PATCH semantics of UpdateRequest. Each field is
// a pointer: nil preserves the existing value; non-nil replaces it. Passing
// &"" for Description clears the description (Description is the only field
// that can be cleared — Name and Permissions cannot be set to empty).
func updateRole(ctx context.Context, client *arize.Client, roleID string) {
	newDescription := "" // clears the description on the server
	role, err := client.Roles.Update(ctx, roles.UpdateRequest{
		Role:        roleID,
		Description: &newDescription,
		Permissions: &[]roles.Permission{
			roles.Permissions.ProjectRead,
			roles.Permissions.ProjectSpanRead,
			roles.Permissions.DatasetRead,
		},
	})
	if err != nil {
		log.Fatalf("update role: %v", err)
	}
	fmt.Printf("updated role %s, now has %d permission(s)\n", role.Id, len(role.Permissions))
}

// deleteRole removes a custom role. Predefined (system) roles cannot be
// deleted — the server returns 400 if you try.
func deleteRole(ctx context.Context, client *arize.Client, roleID string) {
	if err := client.Roles.Delete(ctx, roles.DeleteRequest{Role: roleID}); err != nil {
		log.Fatalf("delete role: %v", err)
	}
	fmt.Printf("deleted role %s\n", roleID)
}
