package projects

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

type (
	Project      = generated.Project
	ProjectList  = generated.ProjectList
	CreateRequest = generated.ProjectCreate

	// ListParams holds optional filters for listing projects.
	// Fields are pointers; nil means "unset".
	ListParams = generated.ProjectsListParams
)

