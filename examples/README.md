# Arize Go SDK v2 — Examples

Runnable examples for every subclient on `*arize.Client`. Each subdirectory is a self-contained `package main` showing how to construct the client and call each verb on one subclient.

## Layout

```
examples/
├── aiintegrations/       client.AIIntegrations.{List, Get, Create, Update, Delete}
├── annotationconfigs/    client.AnnotationConfigs.{List, Get, Create, Delete}
├── annotationqueues/     client.AnnotationQueues.{List, Get, Create, Update, Delete, ListRecords, AddRecords, DeleteRecords, Annotate, Assign}
├── apikeys/              client.APIKeys.{List, Create, CreateServiceKey, Refresh, Delete}
├── datasets/             client.Datasets.{List, Get, Create, Update, Delete, ListExamples, AppendExamples, AnnotateExamples}
├── evaluators/           client.Evaluators.{List, Get, Create, Update, Delete, ListVersions, CreateVersion, GetVersion}
├── organizations/        client.Organizations.{List, Get, Create, Update, Delete, AddUser, RemoveUser}
├── projects/             client.Projects.{List, Get, Create, Delete}
├── prompts/              client.Prompts.{List, Get, Create, Update, Delete, ListVersions, CreateVersion, GetVersion, GetVersionByLabel, SetVersionLabels, DeleteVersionLabel}
├── resourcerestrictions/ client.ResourceRestrictions.{Create, Delete}
├── rolebindings/         client.RoleBindings.{Get, Create, Update, Delete}
├── roles/                client.Roles.{List, Get, Create, Update, Delete}
├── spaces/               client.Spaces.{List, Get, Create, Update, Delete, AddUser, RemoveUser}
└── spans/                client.Spans.{List, Delete, Annotate}
```

## Running

Each example expects `ARIZE_API_KEY` in the environment. From `sdk/go/v2/`:

```bash
ARIZE_API_KEY=<your-key> go run ./examples/projects
ARIZE_API_KEY=<your-key> go run ./examples/apikeys
# ...etc
```

Other env vars accepted by `arize.Config` (see `arize/config.go`):

| Var | Default | Purpose |
|---|---|---|
| `ARIZE_API_HOST` | `api.arize.com` | Override API host (on-prem, staging, …) |
| `ARIZE_API_SCHEME` | `https` | Set to `http` for local servers |

## Editing before running

The examples include placeholder IDs (space, project, role, user) that you must replace with values from your own account before running the create/update/delete paths. Look for `const (...)` blocks at the top of each `main()`.

The destructive operations (Create → Update → Delete) are wired together in `main()` so the examples leave no leftover state, but you can still trip up against rate limits or RBAC if the key doesn't have permission — comment out blocks you don't want.

## Conventions on display

Every method on every subclient has the signature `Method(ctx context.Context, req XRequest)`. Path identifiers, body fields, and query params all live on `req`. Where the field name is a bare resource (`Project`, `Space`, `Organization`), it accepts either a name or an ID — the SDK resolves it internally. ID-only fields use the `<Resource>ID` suffix.

```go
// Get / Delete — name-or-ID
client.Projects.Get(ctx, projects.GetRequest{Project: "my-project", Space: "demo"})
client.Organizations.Get(ctx, organizations.GetRequest{Organization: "my-org"})

// Get / Delete — strict ID
client.APIKeys.Delete(ctx, apikeys.DeleteRequest{ApiKeyID: keyID})
client.RoleBindings.Get(ctx, rolebindings.GetRequest{RoleBindingID: bindingID})

// Create
client.Organizations.Create(ctx, organizations.CreateRequest{Name: "acme"})
client.Projects.Create(ctx, projects.CreateRequest{Name: "new", Space: "demo"})

// Update — request carries the target + the patch
client.Organizations.Update(ctx, organizations.UpdateRequest{
    Organization: "my-org",
    Name:         &newName,
})

// Refresh — request carries the target ID + the body
client.APIKeys.Refresh(ctx, apikeys.RefreshRequest{
    ApiKeyID:  keyID,
    ExpiresAt: &expiresAt,
})

// List
client.APIKeys.List(ctx, apikeys.ListRequest{Limit: &limit})
client.Projects.List(ctx, projects.ListRequest{Space: "demo", Limit: &limit})

// List with POST body + query (spans.List) — body and query are flattened
client.Spans.List(ctx, spans.ListRequest{
    Project: "my-project",
    Space:   "demo",
    Filter:  strPtr("status_code = 'ERROR'"),
    Limit:   &limit,
})
```

See `sdk/go/v2/AGENTS.md` (§ "Public types and request shapes") for the full convention.
