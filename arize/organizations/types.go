package organizations

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	Organization     = generated.Organization
	OrganizationList = generated.OrganizationList
	CreateRequest    = generated.CreateOrganizationRequestBody
	UpdateRequest    = generated.UpdateOrganizationRequestBody

	// ListParams holds optional filters for listing organizations.
	ListParams = generated.OrganizationsListParams
)

