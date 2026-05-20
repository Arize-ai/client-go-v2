// Package main demonstrates how to use the rolebindings subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/rolebindings
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/rolebindings"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account.
	const (
		spaceID = "U3BhY2U6MTox"
		roleID  = "Um9sZToxOmFkbWlu"
		userID  = "VXNlcjoxOjEyMw"
	)

	binding := createRoleBinding(ctx, client, spaceID, roleID, userID)
	getRoleBinding(ctx, client, binding.Id)
	updateRoleBinding(ctx, client, binding.Id, roleID)
	deleteRoleBinding(ctx, client, binding.Id)
}

func getRoleBinding(ctx context.Context, client *arize.Client, bindingID string) {
	rb, err := client.RoleBindings.Get(ctx, rolebindings.GetRequest{RoleBindingID: bindingID})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("role binding %q not found\n", bindingID)
			return
		}
		log.Fatalf("get role binding: %v", err)
	}
	fmt.Printf("found role binding %s (user=%s role=%s)\n", rb.Id, rb.UserId, rb.RoleId)
}

func createRoleBinding(ctx context.Context, client *arize.Client, spaceID, roleID, userID string) *rolebindings.RoleBinding {
	rb, err := client.RoleBindings.Create(ctx, rolebindings.CreateRequest{
		ResourceID:   spaceID,
		ResourceType: rolebindings.RoleBindingResourceTypeSPACE,
		RoleID:       roleID,
		UserID:       userID,
	})
	if err != nil {
		log.Fatalf("create role binding: %v", err)
	}
	fmt.Printf("created role binding %s\n", rb.Id)
	return rb
}

func updateRoleBinding(ctx context.Context, client *arize.Client, bindingID, newRoleID string) {
	rb, err := client.RoleBindings.Update(ctx, rolebindings.UpdateRequest{
		RoleBindingID: bindingID,
		RoleID:        newRoleID,
	})
	if err != nil {
		log.Fatalf("update role binding: %v", err)
	}
	fmt.Printf("updated role binding %s — now role=%s\n", rb.Id, rb.RoleId)
}

func deleteRoleBinding(ctx context.Context, client *arize.Client, bindingID string) {
	if err := client.RoleBindings.Delete(ctx, rolebindings.DeleteRequest{RoleBindingID: bindingID}); err != nil {
		log.Fatalf("delete role binding: %v", err)
	}
	fmt.Printf("deleted role binding %s\n", bindingID)
}
