// Package main demonstrates how to use the users subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/users
//
// The read-only paths (List, Get) run unconditionally. The mutating paths
// (Create, Update, ResendInvitation, ResetPassword, Delete, BulkDelete) send
// real invitation/password emails and change account membership, so they are
// gated behind environment variables:
//
//	ARIZE_EXAMPLE_USER         user ID or email to demo Get
//	ARIZE_EXAMPLE_INVITE_EMAIL email to invite, then update + delete
//	ARIZE_EXAMPLE_BULK_EMAILS  comma-separated emails/IDs to demo BulkDelete
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/users"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	listUsers(ctx, client)

	if userOrEmail := os.Getenv("ARIZE_EXAMPLE_USER"); userOrEmail != "" {
		getUser(ctx, client, userOrEmail)
	} else {
		fmt.Println("set ARIZE_EXAMPLE_USER (id or email) to demo Get")
	}

	// Create → Update → invitation/password flows → Delete. These send real
	// emails and mutate membership, so gate them behind an env var.
	if email := os.Getenv("ARIZE_EXAMPLE_INVITE_EMAIL"); email != "" {
		user := createUser(ctx, client, "Example User", email)
		updateUser(ctx, client, user.Id)
		resendInvitation(ctx, client, user.Id)
		resetPassword(ctx, client, user.Id)
		deleteUser(ctx, client, user.Id)
	} else {
		fmt.Println("set ARIZE_EXAMPLE_INVITE_EMAIL to demo create/update/invite/delete")
	}

	if bulk := os.Getenv("ARIZE_EXAMPLE_BULK_EMAILS"); bulk != "" {
		bulkDeleteUsers(ctx, client, strings.Split(bulk, ","))
	} else {
		fmt.Println("set ARIZE_EXAMPLE_BULK_EMAILS (comma-separated emails/ids) to demo BulkDelete")
	}
}

// listUsers shows filtering by email substring and account status.
func listUsers(ctx context.Context, client *arize.Client) {
	resp, err := client.Users.List(ctx, users.ListRequest{
		Status: []users.UserStatus{users.UserStatusActive, users.UserStatusInvited},
		Limit:  25,
	})
	if err != nil {
		log.Fatalf("list users: %v", err)
	}
	for _, u := range resp.Users {
		fmt.Printf("  %s\t%s\t%s\n", u.Id, u.Email, u.Status)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// getUser accepts either a user ID or an email address. A non-matching email
// yields a *ResourceNotFoundError.
func getUser(ctx context.Context, client *arize.Client, userOrEmail string) {
	user, err := client.Users.Get(ctx, users.GetRequest{User: userOrEmail})
	if err != nil {
		var nfe *arize.ResourceNotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("user %q not found\n", userOrEmail)
			return
		}
		log.Fatalf("get user: %v", err)
	}
	fmt.Printf("found user %s (%s)\n", user.Name, user.Id)

	// The user's account role is a discriminated union; ValueByDiscriminator
	// returns the matching variant — read the predefined arm.
	role, err := user.Role.ValueByDiscriminator()
	if err != nil {
		log.Fatalf("decode role: %v", err)
	}
	switch pre := role.(type) {
	case users.PredefinedUserRole:
		fmt.Printf("  role: predefined %s\n", pre.Name)
	}
}

// createUser invites a new account user with the predefined "member" role.
func createUser(ctx context.Context, client *arize.Client, name, email string) *users.User {
	user, err := client.Users.Create(ctx, users.CreateRequest{
		Name:       name,
		Email:      email,
		Role:       users.AssignPredefinedRole(users.UserRoleMember),
		InviteMode: users.InviteModeEmailLink,
	})
	if err != nil {
		log.Fatalf("create user: %v", err)
	}
	fmt.Printf("created user %s (%s) — status %s\n", user.Email, user.Id, user.Status)
	return user
}

// updateUser renames the user and grants developer permissions. Leave a field
// nil to preserve its current value.
func updateUser(ctx context.Context, client *arize.Client, userID string) {
	newName := "Example User (renamed)"
	isDeveloper := true
	user, err := client.Users.Update(ctx, users.UpdateRequest{
		UserID:      userID,
		Name:        &newName,
		IsDeveloper: &isDeveloper,
	})
	if err != nil {
		log.Fatalf("update user: %v", err)
	}
	fmt.Printf("updated user %s → name %q, developer=%t\n", user.Id, user.Name, user.IsDeveloper)
}

// resendInvitation resends the invite email for a still-pending user.
func resendInvitation(ctx context.Context, client *arize.Client, userID string) {
	if err := client.Users.ResendInvitation(ctx, users.ResendInvitationRequest{UserID: userID}); err != nil {
		log.Fatalf("resend invitation: %v", err)
	}
	fmt.Printf("resent invitation to user %s\n", userID)
}

// resetPassword triggers a password-reset email (password-auth users only).
func resetPassword(ctx context.Context, client *arize.Client, userID string) {
	if err := client.Users.ResetPassword(ctx, users.ResetPasswordRequest{UserID: userID}); err != nil {
		log.Fatalf("reset password: %v", err)
	}
	fmt.Printf("triggered password reset for user %s\n", userID)
}

// deleteUser soft-deletes the user (cascades to memberships, API keys, role
// bindings). Idempotent.
func deleteUser(ctx context.Context, client *arize.Client, userID string) {
	if err := client.Users.Delete(ctx, users.DeleteRequest{UserID: userID}); err != nil {
		log.Fatalf("delete user: %v", err)
	}
	fmt.Printf("deleted user %s\n", userID)
}

// bulkDeleteUsers deletes users by email and/or ID. Per-user outcomes are
// returned rather than aborting the batch: an unresolved email is NotFound, a
// failed delete is Failed.
func bulkDeleteUsers(ctx context.Context, client *arize.Client, emailsOrIDs []string) {
	results, err := client.Users.BulkDelete(ctx, users.BulkDeleteRequest{Emails: emailsOrIDs})
	if err != nil {
		log.Fatalf("bulk delete users: %v", err)
	}
	for _, r := range results {
		switch r.Status {
		case users.DeletionStatusDeleted:
			fmt.Printf("  deleted %s\n", r.UserID)
		case users.DeletionStatusNotFound:
			fmt.Printf("  not found: %s\n", r.Email)
		default:
			fmt.Printf("  %s %s: %s\n", r.Status, r.UserID, r.Error)
		}
	}
}
