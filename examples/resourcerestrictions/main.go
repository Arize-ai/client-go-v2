// Package main demonstrates how to use the resourcerestrictions subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/resourcerestrictions
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/resourcerestrictions"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit this to match a real project in your account before running.
	const projectID = "UHJvamVjdDoxOmV4YW1wbGU"

	restrictResource(ctx, client, projectID)
	listRestrictions(ctx, client)
	unrestrictResource(ctx, client, projectID)
}

// listRestrictions prints all resource restrictions in the account, paging
// through results with the cursor returned in each response.
func listRestrictions(ctx context.Context, client *arize.Client) {
	cursor := ""
	for {
		page, err := client.ResourceRestrictions.List(ctx, resourcerestrictions.ListRequest{
			ResourceType: resourcerestrictions.ResourceRestrictionTypePROJECT,
			Limit:        50,
			Cursor:       cursor,
		})
		if err != nil {
			log.Fatalf("list restrictions: %v", err)
		}
		for _, rr := range page.ResourceRestrictions {
			fmt.Printf("restricted %s (type=%s, created=%s)\n", rr.ResourceId, rr.ResourceType, rr.CreatedAt)
		}
		if !page.Pagination.HasMore || page.Pagination.NextCursor == nil {
			break
		}
		cursor = *page.Pagination.NextCursor
	}
}

// restrictResource marks the given resource (here a project) as restricted.
// Restricting prevents roles bound at higher hierarchy levels (space, org,
// account) from granting access to it.
func restrictResource(ctx context.Context, client *arize.Client, resourceID string) {
	rr, err := client.ResourceRestrictions.Restrict(ctx, resourcerestrictions.RestrictRequest{
		ResourceID: resourceID,
	})
	if err != nil {
		log.Fatalf("restrict resource: %v", err)
	}
	fmt.Printf("restricted resource %s (type=%s)\n", rr.ResourceId, rr.ResourceType)
}

// unrestrictResource lifts the restriction from the resource identified by its
// ID — the ID of the *restricted resource* (e.g. a project ID), not the ID of
// the restriction record.
func unrestrictResource(ctx context.Context, client *arize.Client, resourceID string) {
	if err := client.ResourceRestrictions.Unrestrict(ctx, resourcerestrictions.UnrestrictRequest{
		ResourceID: resourceID,
	}); err != nil {
		log.Fatalf("unrestrict resource: %v", err)
	}
	fmt.Printf("lifted restriction from %s\n", resourceID)
}
