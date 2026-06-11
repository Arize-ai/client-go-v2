<p align="center">
  <a href="https://arize.com/ax">
    <img src="https://storage.googleapis.com/arize-assets/arize-logo-white.jpg" width="600" />
  </a>
  <br/>
  <a target="_blank" href="https://pkg.go.dev/github.com/Arize-ai/client-go-v2">
    <img src="https://pkg.go.dev/badge/github.com/Arize-ai/client-go-v2.svg" alt="Go Reference">
  </a>
  <a target="_blank" href="https://pkg.go.dev/github.com/Arize-ai/client-go-v2">
    <img src="https://img.shields.io/github/go-mod/go-version/Arize-ai/client-go-v2?label=go">
  </a>
  <a target="_blank" href="https://arize-ai.slack.com/join/shared_invite/zt-2w57bhem8-hq24MB6u7yE_ZF_ilOYSBw#/shared-invite/email">
    <img src="https://img.shields.io/badge/slack-@arize-blue.svg?logo=slack">
  </a>
</p>

---

# Table of Contents <!-- omit in toc -->

- [Overview](#overview)
- [Status](#status)
- [Installation](#installation)
  - [Migrating from the Legacy Go Client](#migrating-from-the-legacy-go-client)
- [Usage](#usage)
  - [Constructing a Client](#constructing-a-client)
  - [Regions](#regions)
  - [Endpoint Overrides](#endpoint-overrides)
  - [Error Handling](#error-handling)
  - [Operations on Spaces](#operations-on-spaces)
    - [List Spaces](#list-spaces)
    - [Get a Space](#get-a-space)
    - [Create a Space](#create-a-space)
    - [Update a Space](#update-a-space)
    - [Delete a Space](#delete-a-space)
    - [Add a Space Member](#add-a-space-member)
    - [Remove a Space Member](#remove-a-space-member)
  - [Operations on Projects](#operations-on-projects)
    - [List Projects](#list-projects)
    - [Get a Project](#get-a-project)
    - [Create a Project](#create-a-project)
    - [Update a Project](#update-a-project)
    - [Delete a Project](#delete-a-project)
  - [Operations on Spans](#operations-on-spans)
    - [List Spans](#list-spans)
    - [Delete Spans](#delete-spans)
    - [Annotate Spans](#annotate-spans)
  - [Operations on Datasets](#operations-on-datasets)
    - [List Datasets](#list-datasets)
    - [Get a Dataset](#get-a-dataset)
    - [Create a Dataset](#create-a-dataset)
    - [Update a Dataset](#update-a-dataset)
    - [Delete a Dataset](#delete-a-dataset)
    - [List Examples](#list-examples)
    - [Append Examples](#append-examples)
    - [Annotate Examples](#annotate-examples)
  - [Operations on Experiments](#operations-on-experiments)
    - [List Experiments](#list-experiments)
    - [Get an Experiment](#get-an-experiment)
    - [Create an Experiment](#create-an-experiment)
    - [Delete an Experiment](#delete-an-experiment)
    - [List Runs](#list-runs)
    - [Append Runs](#append-runs)
  - [Operations on Prompts](#operations-on-prompts)
    - [List Prompts](#list-prompts)
    - [Get a Prompt](#get-a-prompt)
    - [Create a Prompt](#create-a-prompt)
    - [Update a Prompt](#update-a-prompt)
    - [Delete a Prompt](#delete-a-prompt)
    - [List Versions](#list-versions)
    - [Create a Version](#create-a-version)
    - [Get a Version](#get-a-version)
    - [Get a Version by Label](#get-a-version-by-label)
    - [Set Version Labels](#set-version-labels)
    - [Delete a Version Label](#delete-a-version-label)
  - [Operations on Evaluators](#operations-on-evaluators)
    - [List Evaluators](#list-evaluators)
    - [Get an Evaluator](#get-an-evaluator)
    - [Create an Evaluator](#create-an-evaluator)
    - [Update an Evaluator](#update-an-evaluator)
    - [Delete an Evaluator](#delete-an-evaluator)
    - [List Versions](#list-versions-1)
    - [Create a Version](#create-a-version-1)
    - [Get a Version](#get-a-version-1)
  - [Operations on Annotation Configs](#operations-on-annotation-configs)
    - [List Annotation Configs](#list-annotation-configs)
    - [Get an Annotation Config](#get-an-annotation-config)
    - [Create an Annotation Config](#create-an-annotation-config)
    - [Delete an Annotation Config](#delete-an-annotation-config)
  - [Operations on AI Integrations](#operations-on-ai-integrations)
    - [List AI Integrations](#list-ai-integrations)
    - [Get an AI Integration](#get-an-ai-integration)
    - [Create an AI Integration](#create-an-ai-integration)
    - [Update an AI Integration](#update-an-ai-integration)
    - [Delete an AI Integration](#delete-an-ai-integration)
  - [Operations on Organizations](#operations-on-organizations)
    - [List Organizations](#list-organizations)
    - [Get an Organization](#get-an-organization)
    - [Create an Organization](#create-an-organization)
    - [Update an Organization](#update-an-organization)
    - [Delete an Organization](#delete-an-organization)
    - [Add a User](#add-a-user)
    - [Remove a User](#remove-a-user)
  - [Operations on Roles](#operations-on-roles)
    - [List Roles](#list-roles)
    - [Get a Role](#get-a-role)
    - [Create a Role](#create-a-role)
    - [Update a Role](#update-a-role)
    - [Delete a Role](#delete-a-role)
  - [Operations on Role Bindings](#operations-on-role-bindings)
    - [Get a Role Binding](#get-a-role-binding)
    - [Create a Role Binding](#create-a-role-binding)
    - [Update a Role Binding](#update-a-role-binding)
    - [Delete a Role Binding](#delete-a-role-binding)
  - [Operations on API Keys](#operations-on-api-keys)
    - [List API Keys](#list-api-keys)
    - [Create an API Key](#create-an-api-key)
    - [Create a Service Key](#create-a-service-key)
    - [Refresh an API Key](#refresh-an-api-key)
    - [Revoke an API Key](#revoke-an-api-key)
  - [Operations on Resource Restrictions](#operations-on-resource-restrictions)
    - [Restrict a Resource](#restrict-a-resource)
    - [Unrestrict a Resource](#unrestrict-a-resource)
  - [Operations on Annotation Queues](#operations-on-annotation-queues)
    - [List Annotation Queues](#list-annotation-queues)
    - [Create an Annotation Queue](#create-an-annotation-queue)
    - [Get an Annotation Queue](#get-an-annotation-queue)
    - [Add Records to a Queue](#add-records-to-a-queue)
    - [Annotate a Record](#annotate-a-record)
    - [Update an Annotation Queue](#update-an-annotation-queue)
    - [Delete an Annotation Queue](#delete-an-annotation-queue)
  - [Operations on Users](#operations-on-users)
    - [List Users](#list-users)
    - [Get a User](#get-a-user)
    - [Create a User](#create-a-user)
    - [Update a User](#update-a-user)
    - [Resend an Invitation](#resend-an-invitation)
    - [Reset a Password](#reset-a-password)
    - [Delete a User](#delete-a-user)
    - [Bulk Delete Users](#bulk-delete-users)
  - [Operations on Tasks](#operations-on-tasks)
    - [List Tasks](#list-tasks)
    - [Get a Task](#get-a-task)
    - [Create an Evaluation Task](#create-an-evaluation-task)
    - [Create a Run-Experiment Task](#create-a-run-experiment-task)
    - [Update a Task](#update-a-task)
    - [Delete a Task](#delete-a-task)
    - [Trigger a Run](#trigger-a-run)
    - [List Task Runs](#list-task-runs)
    - [Get a Run](#get-a-run)
    - [Cancel a Run](#cancel-a-run)
    - [Wait for a Run](#wait-for-a-run)
- [SDK Configuration](#sdk-configuration)
  - [Environment Variables](#environment-variables)
  - [TLS Verification](#tls-verification)
  - [HTTP Timeout](#http-timeout)
  - [Local Directory and Caching](#local-directory-and-caching)
- [Community](#community)

# Overview

A helper package to interact with Arize AI APIs from Go.

Arize is an AI engineering platform. It helps engineers develop, evaluate, and observe AI applications and agents.

Arize has both Enterprise and OSS products to support this goal:

- [Arize AX](https://arize.com/) — an enterprise AI engineering platform from development to production, with an embedded AI Copilot
- [Phoenix](https://github.com/Arize-ai/phoenix) — a lightweight, open-source project for tracing, prompt engineering, and evaluation
- [OpenInference](https://github.com/Arize-ai/openinference) — an open-source instrumentation package to trace LLM applications across models and frameworks

We log over 1 trillion inferences and spans, 10 million evaluation runs, and 2 million OSS downloads every month.

# Status

> **Pre-release (`v0.x.y`).** The public API surface is unstable and may change without notice until the first `v1.0.0` release. Each call site logs a one-time pre-release warning.

The Go SDK v2 currently exposes the following surface area:

- **Client construction** with environment-aware configuration, region resolution, and on-prem endpoint overrides.
- **Typed HTTP errors** matchable via `errors.As` (`BadRequestError`, `UnauthorizedError`, `NotFoundError`, …).
- Resource subclients on `*arize.Client`:
  - **Spaces** — list, get, create, update, delete, and manage memberships.
  - **Projects** — list, get, create, update, delete.
  - **Spans** — list, delete, and annotate spans.
  - **Datasets** — list, get, create, update, delete, and manage examples.
  - **Experiments** — list, get, create, delete, and list their runs.
  - **Prompts** — list, get, create, update, delete, and manage versions and labels.
  - **Evaluators** — list, get, create, update, delete, and manage versions.
  - **Annotation Configs** — list, get, create, delete.
  - **AI Integrations** — list, get, create, update, delete.
  - **Organizations** — list, get, create, update, delete, and manage memberships.
  - **Roles** & **Role Bindings** — manage RBAC roles and their bindings.
  - **API Keys** — list, create, create service keys, refresh, revoke.
  - **Resource Restrictions** — restrict and unrestrict access to Arize resources.
  - **Annotation Queues** — list, get, create, update, delete, add records, and annotate.
  - **Users** — list, get, create, update, delete, bulk delete, resend invitations, and reset passwords.
  - **Tasks** — list, get, create, update, delete, and trigger, list, poll, and cancel their runs.

Additional resource domains will be added incrementally.

Runnable, end-to-end programs for every subclient live in [`examples/`](./examples).

# Installation

```bash
go get github.com/Arize-ai/client-go-v2@latest
```

Module path:

```
github.com/Arize-ai/client-go-v2
```

Package documentation is available at [pkg.go.dev/github.com/Arize-ai/client-go-v2](https://pkg.go.dev/github.com/Arize-ai/client-go-v2).

## Migrating from the Legacy Go Client

The legacy `client_golang` package lives at [Arize-ai/client-go-v1](https://github.com/Arize-ai/client-go-v1). It is in maintenance mode — new feature work targets `v2`. The v2 surface is REST-based and intentionally diverges from v1; treat the move as a port rather than a drop-in upgrade.

# Usage

## Constructing a Client

The client reads its API key from `Config.APIKey` or the `ARIZE_API_KEY` environment variable. All other fields fall back to environment variables and then to documented defaults during `Resolve()`.

```go
package main

import (
    "context"
    "log"

    "github.com/Arize-ai/client-go-v2/arize"
)

func main() {
    client, err := arize.NewClient(arize.Config{
        APIKey: "<your-api-key>", // or set ARIZE_API_KEY
    })
    if err != nil {
        log.Fatal(err)
    }

    _ = context.Background()
    _ = client
}
```

`NewClient` resolves the config (applies env vars and defaults), validates it, and returns an error if anything is missing or inconsistent. Common sentinel errors:

```go
if errors.Is(err, arize.ErrMissingAPIKey)             { /* APIKey unset */ }
if errors.Is(err, arize.ErrMultipleEndpointOverrides) { /* Region + SingleHost + BaseDomain conflict */ }
```

## Regions

Set `Config.Region` (or the `ARIZE_REGION` env var) to route the client at a specific Arize deployment region. Known regions:

| Constant            | Value             |
| ------------------- | ----------------- |
| `RegionUSCentral`   | `us-central-1a`   |
| `RegionEUWest`      | `eu-west-1a`      |
| `RegionCACentral`   | `ca-central-1a`   |
| `RegionUSEast`      | `us-east-1b`      |

```go
client, err := arize.NewClient(arize.Config{
    APIKey: "<your-api-key>",
    Region: arize.RegionEUWest,
})
```

To inspect the endpoints a region resolves to without constructing a client:

```go
endpoints, ok := arize.RegionEndpointsFor(arize.RegionEUWest)
// endpoints.APIHost == "api.eu-west-1a.arize.com"
```

## Endpoint Overrides

For on-prem deployments, use **one** of the following override fields. Setting more than one returns `arize.ErrMultipleEndpointOverrides` from `Validate()`.

```go
// 1. Base domain — derives api.<domain>, otlp.<domain>, flight.<domain>
client, _ := arize.NewClient(arize.Config{
    APIKey:     "<your-api-key>",
    BaseDomain: "arize.example.com",
})

// 2. Single host — points API, OTLP, and Flight at the same host
client, _ = arize.NewClient(arize.Config{
    APIKey:     "<your-api-key>",
    SingleHost: "arize.internal",
    SinglePort: 8443, // optional; rewrites FlightPort
})

// 3. Explicit per-component host/scheme
client, _ = arize.NewClient(arize.Config{
    APIKey:    "<your-api-key>",
    APIHost:   "arize.internal:8080",
    APIScheme: "http",
})
```

## Error Handling

All HTTP errors implement `error` and embed `arize.APIError`. Match on a specific status class with `errors.As`:

```go
import "errors"

err := client.ResourceRestrictions.Unrestrict(ctx, resourcerestrictions.UnrestrictRequest{ResourceID: "nonexistent"})

var nfe *arize.NotFoundError
if errors.As(err, &nfe) {
    // HTTP 404 — handle the missing-resource case
}

var apiErr *arize.APIError
if errors.As(err, &apiErr) {
    // Any HTTP error — read apiErr.StatusCode, apiErr.Body, etc.
}
```

Available typed errors: `BadRequestError`, `UnauthorizedError`, `ForbiddenError`, `NotFoundError`, `ConflictError`, `RateLimitError`, `ServerError`. Compare with `errors.Is` / `errors.As`, never with `==` — wrapping with `fmt.Errorf("...: %w", err)` breaks direct comparison.

Every method takes `(ctx context.Context, req XRequest)` — path identifiers, body fields, and query
params all live on `req`. Fields named for a bare resource (`Space`, `Project`, `Organization`, …) accept
**either a name or an ID**; the SDK resolves them internally. ID-only fields use the `<Resource>ID` suffix.
The snippets below assume `ctx` and `client` are in scope and the matching subclient package is imported
(e.g. `import "github.com/Arize-ai/client-go-v2/arize/spaces"`). Each subclient has a full runnable
program in [`examples/`](./examples).

## Operations on Spaces

Use `client.Spaces` to manage spaces (containers for projects, datasets, …) and their
memberships. `Organization` accepts a name or ID.

A full runnable example lives in [`examples/spaces`](./examples/spaces).

### List Spaces

```go
resp, err := client.Spaces.List(ctx, spaces.ListRequest{Limit: 25})
```

### Get a Space

```go
space, err := client.Spaces.Get(ctx, spaces.GetRequest{Space: "<space-id-or-name>"})
```

### Create a Space

```go
space, err := client.Spaces.Create(ctx, spaces.CreateRequest{
    Name:         "my-space",
    Organization: "<org-id-or-name>", // accepts a name or ID
    Description:  "optional",
})
```

### Update a Space

Patch semantics: nil fields are preserved.

```go
newName := "renamed-space"
space, err := client.Spaces.Update(ctx, spaces.UpdateRequest{Space: "<space-id-or-name>", Name: &newName})
```

### Delete a Space

Irreversible — removes all child resources.

```go
err := client.Spaces.Delete(ctx, spaces.DeleteRequest{Space: "<space-id>"})
```

### Add a Space Member

Build the role with `AssignPredefinedRole` or `AssignCustomRole`.

```go
m, err := client.Spaces.AddUser(ctx, spaces.AddUserRequest{
    Space: "<space-id>", UserID: "<user-id>",
    Role:  spaces.AssignPredefinedRole(spaces.UserSpaceRoleMember),
})
```

### Remove a Space Member

```go
err := client.Spaces.RemoveUser(ctx, spaces.RemoveUserRequest{Space: "<space-id>", UserID: "<user-id>"})
```

## Operations on Projects

Use `client.Projects` to manage projects, which are namespaces for organizing tracing
data. `Space` is required when `Project` is a name.

A full runnable example lives in [`examples/projects`](./examples/projects).

### List Projects

```go
resp, err := client.Projects.List(ctx, projects.ListRequest{
    Space: "<space-id-or-name>", // optional filter
    Name:  "prod",               // optional substring filter
    Limit: 50,
})
```

### Get a Project

```go
proj, err := client.Projects.Get(ctx, projects.GetRequest{
    Project: "<project-id-or-name>",
    Space:   "<space-id-or-name>", // required when Project is a name
})
```

### Create a Project

```go
proj, err := client.Projects.Create(ctx, projects.CreateRequest{
    Name:  "my-project", // must be unique within the space
    Space: "<space-id-or-name>",
})
```

### Update a Project

```go
proj, err := client.Projects.Update(ctx, projects.UpdateRequest{
    Project: "<project-id-or-name>",
    Space:   "<space-id-or-name>",
    Name:    "renamed-project",
})
```

### Delete a Project

```go
err := client.Projects.Delete(ctx, projects.DeleteRequest{Project: "<project-id-or-name>", Space: "<space-id-or-name>"})
```

## Operations on Spans

Use `client.Spans` to list, delete, and annotate spans. `List` is a POST under the hood
(the filter DSL can be too large for a query string), so its body fields (project,
time range, filter) and query params (limit, cursor) are flattened into one
`ListRequest`.

A full runnable example lives in [`examples/spans`](./examples/spans).

### List Spans

Name-or-ID `Project`, optional time range and filter DSL.

```go
resp, err := client.Spans.List(ctx, spans.ListRequest{
    Project: "<project-id-or-name>",
    Space:   "<space-id-or-name>", // required when Project is a name
    End:     time.Now(),
    Filter:  "status_code = 'ERROR'",
    Limit:   50,
})
```

### Delete Spans

A partial success returns a non-nil `*SpanDeletePartial` (HTTP 200); a full delete returns nil (HTTP 204).

```go
partial, err := client.Spans.Delete(ctx, spans.DeleteRequest{
    Project: "<project-id-or-name>",
    Space:   "<space-id-or-name>",
    SpanIDs: []string{"span-1", "span-2"},
})
```

### Annotate Spans

Up to 1000 spans per call; re-submitting the same config name overwrites (no duplicates).

```go
err := client.Spans.Annotate(ctx, spans.AnnotateRequest{
    Project:     "<project-id-or-name>",
    Space:       "<space-id-or-name>",
    Annotations: []spans.AnnotateRecordInput{ /* RecordId + AnnotationInput values */ },
})
```

## Operations on Datasets

Use `client.Datasets` to manage datasets and their examples. `Space` is required when
`Dataset` is a name. Each example is an arbitrary set of user-defined fields.

A full runnable example lives in [`examples/datasets`](./examples/datasets).

### List Datasets

```go
resp, err := client.Datasets.List(ctx, datasets.ListRequest{Space: "<space-id-or-name>", Limit: 25})
```

### Get a Dataset

```go
ds, err := client.Datasets.Get(ctx, datasets.GetRequest{Dataset: "<dataset-id-or-name>", Space: "<space-id-or-name>"})
```

### Create a Dataset

At least one example is required (empty returns `datasets.ErrNoExamples`).

```go
ds, err := client.Datasets.Create(ctx, datasets.CreateRequest{
    Name:  "my-dataset",
    Space: "<space-id-or-name>",
    Examples: []datasets.DatasetExampleCreate{
        {"input": "What is Arize?", "output": "An AI observability platform."},
    },
})
```

### Update a Dataset

```go
ds, err := client.Datasets.Update(ctx, datasets.UpdateRequest{Dataset: "<dataset-id-or-name>", Space: "<space-id-or-name>", Name: "renamed"})
```

### Delete a Dataset

```go
err := client.Datasets.Delete(ctx, datasets.DeleteRequest{Dataset: "<dataset-id>"})
```

### List Examples

```go
ex, err := client.Datasets.ListExamples(ctx, datasets.ListExamplesRequest{Dataset: "<dataset-id-or-name>", Space: "<space-id-or-name>", Limit: 50})
```

### Append Examples

Appends to the latest dataset version.

```go
ins, err := client.Datasets.AppendExamples(ctx, datasets.AppendExamplesRequest{
    Dataset: "<dataset-id-or-name>", Space: "<space-id-or-name>",
    Examples: []datasets.DatasetExampleCreate{{"input": "q", "output": "a"}},
})
```

### Annotate Examples

```go
err := client.Datasets.AnnotateExamples(ctx, datasets.AnnotateExamplesRequest{
    Dataset: "<dataset-id-or-name>", Space: "<space-id-or-name>",
    Annotations: []datasets.AnnotateRecordInput{ /* RecordId + AnnotationInput values */ },
})
```

## Operations on Experiments

Use `client.Experiments` to manage experiments — a named set of task runs over a dataset,
optionally carrying evaluator results. `Dataset` is required to create or resolve an
experiment by name; `Space` is required when `Dataset` is itself a name. On `Create`,
each run is a row of user-named columns: `TaskFields` names the example-ID and output
columns, and `EvaluatorColumns` remaps evaluator result columns to the wire format.

A full runnable example lives in [`examples/experiments`](./examples/experiments).

### List Experiments

```go
resp, err := client.Experiments.List(ctx, experiments.ListRequest{
    Dataset: "<dataset-id-or-name>", // optional filter
    Space:   "<space-id-or-name>",   // required when Dataset is a name
    Limit:   25,
})
```

### Get an Experiment

```go
exp, err := client.Experiments.Get(ctx, experiments.GetRequest{
    Experiment: "<experiment-id-or-name>",
    Dataset:    "<dataset-id-or-name>", // required when Experiment is a name
    Space:      "<space-id-or-name>",   // required when Dataset is also a name
})
```

### Create an Experiment

Each run maps your column names to the wire format via `TaskFields` (required) and
`EvaluatorColumns` (optional). Non-string outputs are JSON-encoded automatically.

```go
exp, err := client.Experiments.Create(ctx, experiments.CreateRequest{
    Dataset: "<dataset-id-or-name>",
    Space:   "<space-id-or-name>", // required when Dataset is a name
    Name:    "my-experiment",
    Runs: []map[string]any{
        {"example_id": "ex-1", "answer": "Paris", "relevance_score": 0.9, "relevance_label": "relevant"},
    },
    TaskFields: experiments.TaskFields{ExampleID: "example_id", Output: "answer"},
    EvaluatorColumns: map[string]experiments.EvaluatorFields{
        "relevance": {Score: "relevance_score", Label: "relevance_label"},
    },
})
```

### Delete an Experiment

```go
err := client.Experiments.Delete(ctx, experiments.DeleteRequest{Experiment: "<experiment-id>"})
```

### List Runs

```go
runs, err := client.Experiments.ListRuns(ctx, experiments.ListRunsRequest{
    Experiment: "<experiment-id-or-name>",
    Dataset:    "<dataset-id-or-name>", // required when Experiment is a name
    Space:      "<space-id-or-name>",   // required when Dataset is also a name
    Limit:      50,
})
```

### Append Runs

Append between 1 and 1000 new runs to an existing experiment. Each run must include `ExampleId` (the ID of an example from the experiment's dataset) and `Output`; additional user-defined fields go in `AdditionalProperties`. The response includes the updated experiment and the generated run IDs in input order.

```go
result, err := client.Experiments.AppendRuns(ctx, experiments.AppendRunsRequest{
    ExperimentID: "<experiment-id>",
    ExperimentRuns: []experiments.ExperimentRunCreate{
        {ExampleId: "example-1", Output: "An AI observability platform."},
        {ExampleId: "example-2", Output: "A unit of work in a trace."},
    },
})
```

## Operations on Prompts

Use `client.Prompts` to manage prompts, their versions, and version labels. `Space` is
required when `Prompt` is a name; version-by-ID and label operations take a strict
`VersionID`.

A full runnable example lives in [`examples/prompts`](./examples/prompts).

### List Prompts

```go
resp, err := client.Prompts.List(ctx, prompts.ListRequest{Limit: 25})
```

### Get a Prompt

`Get` accepts an optional `VersionID` or `Label` to pin a version.

```go
p, err := client.Prompts.Get(ctx, prompts.GetRequest{Prompt: "<prompt-id-or-name>", Space: "<space-id-or-name>"})
```

### Create a Prompt

Created with the initial version.

```go
content := "You are a helpful assistant. Answer: {question}"
p, err := client.Prompts.Create(ctx, prompts.CreateRequest{
    Name:  "my-prompt",
    Space: "<space-id-or-name>",
    Version: prompts.PromptVersionCreate{
        CommitMessage: "initial version",
        Provider:      prompts.LlmProviderOpenAi,
        Messages:      []prompts.LLMMessage{{Role: prompts.MessageRoleSystem, Content: &content}},
    },
})
```

### Update a Prompt

`Description` is a PATCH pointer.

```go
desc := "new description"
p, err := client.Prompts.Update(ctx, prompts.UpdateRequest{Prompt: "<prompt-id-or-name>", Space: "<space-id-or-name>", Description: &desc})
```

### Delete a Prompt

```go
err := client.Prompts.Delete(ctx, prompts.DeleteRequest{Prompt: "<prompt-id>"})
```

### List Versions

```go
versions, err := client.Prompts.ListVersions(ctx, prompts.ListVersionsRequest{Prompt: "<prompt-id-or-name>", Space: "<space-id-or-name>", Limit: 25})
```

### Create a Version

```go
content := "You are a helpful assistant. Answer: {question}"
v, err := client.Prompts.CreateVersion(ctx, prompts.CreateVersionRequest{
    Prompt: "<prompt-id-or-name>", Space: "<space-id-or-name>",
    CommitMessage: "tweak", Provider: prompts.LlmProviderOpenAi,
    Messages: []prompts.LLMMessage{{Role: prompts.MessageRoleSystem, Content: &content}},
})
```

### Get a Version

```go
v, err := client.Prompts.GetVersion(ctx, prompts.GetVersionRequest{VersionID: "<version-id>"})
```

### Get a Version by Label

```go
v, err := client.Prompts.GetVersionByLabel(ctx, prompts.GetVersionByLabelRequest{Prompt: "<prompt-id-or-name>", Space: "<space-id-or-name>", LabelName: "production"})
```

### Set Version Labels

Replaces all labels on a version.

```go
labels, err := client.Prompts.SetVersionLabels(ctx, prompts.SetVersionLabelsRequest{VersionID: "<version-id>", Labels: []string{"production"}})
```

### Delete a Version Label

Removes one label.

```go
err := client.Prompts.DeleteVersionLabel(ctx, prompts.DeleteVersionLabelRequest{VersionID: "<version-id>", LabelName: "production"})
```

## Operations on Evaluators

Use `client.Evaluators` to manage evaluators and their versions. An evaluator is either a
`template` (LLM-based) or `code` (managed built-in or custom Python) evaluator — the
type is derived from `Version`. `Space` is required when `Evaluator` is a name.

A full runnable example lives in [`examples/evaluators`](./examples/evaluators).

### List Evaluators

```go
resp, err := client.Evaluators.List(ctx, evaluators.ListRequest{Limit: 25})
```

### Get an Evaluator

The returned version is a oneOf — read the active variant via `ValueByDiscriminator` and a type switch.

```go
ev, err := client.Evaluators.Get(ctx, evaluators.GetRequest{Evaluator: "<evaluator-id-or-name>", Space: "<space-id-or-name>"})
if v, err := ev.Version.ValueByDiscriminator(); err == nil {
    if tmpl, ok := v.(evaluators.EvaluatorVersionTemplate); ok {
        _ = tmpl.TemplateConfig.Template
    }
}
```

### Create an Evaluator

A template evaluator with its initial version.

```go
ev, err := client.Evaluators.Create(ctx, evaluators.CreateRequest{
    Name:  "relevance",
    Space: "<space-id-or-name>",
    Version: evaluators.VersionConfig{
        CommitMessage: "initial version",
        Template: &evaluators.TemplateConfig{
            Name:      "relevance",
            Template:  "Is the answer relevant?\n{{input}}",
            LlmConfig: evaluators.EvaluatorLlmConfig{AiIntegrationId: "<ai-integration-id>", ModelName: "gpt-4o"},
        },
    },
})
```

### Update an Evaluator

`Name`/`Description` are PATCH pointers — at least one is required, else `evaluators.ErrNoUpdateFields`.

```go
newName := "renamed-evaluator"
md, err := client.Evaluators.Update(ctx, evaluators.UpdateRequest{Evaluator: "<evaluator-id-or-name>", Space: "<space-id-or-name>", Name: &newName})
```

### Delete an Evaluator

```go
err := client.Evaluators.Delete(ctx, evaluators.DeleteRequest{Evaluator: "<evaluator-id>"})
```

### List Versions

```go
vers, err := client.Evaluators.ListVersions(ctx, evaluators.ListVersionsRequest{Evaluator: "<evaluator-id-or-name>", Space: "<space-id-or-name>", Limit: 25})
```

### Create a Version

The new version's kind must match the parent evaluator's type.

```go
ver, err := client.Evaluators.CreateVersion(ctx, evaluators.CreateVersionRequest{
    Evaluator: "<evaluator-id-or-name>", Space: "<space-id-or-name>",
    Version: evaluators.VersionConfig{CommitMessage: "tighten rubric", Template: &evaluators.TemplateConfig{ /* ... */ }},
})
```

### Get a Version

```go
ver, err := client.Evaluators.GetVersion(ctx, evaluators.GetVersionRequest{VersionID: "<version-id>"})
```

## Operations on Annotation Configs

Use `client.AnnotationConfigs` to manage annotation configs. The config type
(`Categorical`, `Continuous`, `Freeform`) selects which fields are required. The
returned `AnnotationConfig` is a discriminated union — read it via
`ValueByDiscriminator` and a type switch over the variant.

A full runnable example lives in [`examples/annotationconfigs`](./examples/annotationconfigs).

### List Annotation Configs

```go
resp, err := client.AnnotationConfigs.List(ctx, annotationconfigs.ListRequest{Space: "<space-id-or-name>", Limit: 25})
```

### Get an Annotation Config

```go
ac, err := client.AnnotationConfigs.Get(ctx, annotationconfigs.GetRequest{AnnotationConfig: "<config-id-or-name>", Space: "<space-id-or-name>"})
if v, err := ac.ValueByDiscriminator(); err == nil {
    if cfg, ok := v.(annotationconfigs.CategoricalAnnotationConfig); ok {
        _ = cfg.Values
    }
}
```

### Create an Annotation Config

Categorical (`Values` required), continuous (`Min`/`MaxScore` required), or freeform (no extra fields).

```go
score1, score0 := 1.0, 0.0
ac, err := client.AnnotationConfigs.Create(ctx, annotationconfigs.CreateRequest{
    Space: "<space-id-or-name>",
    Name:  "quality",
    Type:  annotationconfigs.AnnotationConfigTypeCategorical,
    Values: []annotationconfigs.CategoricalAnnotationValue{
        {Label: "good", Score: &score1},
        {Label: "bad", Score: &score0},
    },
})
```

### Delete an Annotation Config

```go
err := client.AnnotationConfigs.Delete(ctx, annotationconfigs.DeleteRequest{AnnotationConfig: "<config-id>"})
```

## Operations on AI Integrations

Use `client.AIIntegrations` to manage connections to LLM providers (OpenAI, Anthropic,
AWS Bedrock, Vertex AI, …). On `Update`, nil patch fields are preserved and
pointer-to-empty-string clears a clearable field.

A full runnable example lives in [`examples/aiintegrations`](./examples/aiintegrations).

### List AI Integrations

```go
resp, err := client.AIIntegrations.List(ctx, aiintegrations.ListRequest{Limit: 25})
```

### Get an AI Integration

```go
ai, err := client.AIIntegrations.Get(ctx, aiintegrations.GetRequest{Integration: "<integration-id-or-name>"})
```

### Create an AI Integration

Provider key inline, or `ProviderMetadata` for AWS Bedrock / Vertex AI.

```go
ai, err := client.AIIntegrations.Create(ctx, aiintegrations.CreateRequest{
    Name:     "my-anthropic",
    Provider: aiintegrations.AIIntegrationProviderAnthropic,
    APIKey:   "<provider-api-key>",
})
```

### Update an AI Integration

Rotate the key, clear the base URL (`&""` emits JSON null), preserve everything else.

```go
newKey, clearBaseURL := "<new-key>", ""
ai, err := client.AIIntegrations.Update(ctx, aiintegrations.UpdateRequest{
    Integration: "<integration-id>", APIKey: &newKey, BaseURL: &clearBaseURL,
})
```

### Delete an AI Integration

```go
err := client.AIIntegrations.Delete(ctx, aiintegrations.DeleteRequest{Integration: "<integration-id>"})
```

## Operations on Organizations

Use `client.Organizations` to manage organizations and their memberships. `Organization`
accepts a name or ID.

A full runnable example lives in [`examples/organizations`](./examples/organizations).

### List Organizations

```go
resp, err := client.Organizations.List(ctx, organizations.ListRequest{Limit: 25})
```

### Get an Organization

```go
org, err := client.Organizations.Get(ctx, organizations.GetRequest{Organization: "<org-id-or-name>"})
```

### Create an Organization

```go
org, err := client.Organizations.Create(ctx, organizations.CreateRequest{Name: "acme"})
```

### Update an Organization

`Name`/`Description` are PATCH pointers.

```go
newName := "acme-renamed"
org, err := client.Organizations.Update(ctx, organizations.UpdateRequest{Organization: "<org-id-or-name>", Name: &newName})
```

### Delete an Organization

Irreversible.

```go
err := client.Organizations.Delete(ctx, organizations.DeleteRequest{Organization: "<org-id>"})
```

### Add a User

The returned membership's `Role` is a discriminated union — read it via `ValueByDiscriminator` and a type switch.

```go
m, err := client.Organizations.AddUser(ctx, organizations.AddUserRequest{
    Organization: "<org-id>", UserID: "<user-id>",
    Role: organizations.PredefinedOrgRole{Name: organizations.OrganizationRoleMember},
})
```

### Remove a User

```go
err := client.Organizations.RemoveUser(ctx, organizations.RemoveUserRequest{Organization: "<org-id>", UserID: "<user-id>"})
```

## Operations on Roles

Use `client.Roles` to manage RBAC roles. Permissions are referenced through the typed
`roles.Permissions` namespace. `Role` accepts a name or ID.

A full runnable example lives in [`examples/roles`](./examples/roles).

### List Roles

`IsPredefined` is a tri-state filter: `&true` (system roles), `&false` (custom), nil (both).

```go
predefined := true
resp, err := client.Roles.List(ctx, roles.ListRequest{IsPredefined: &predefined, Limit: 25})
```

### Get a Role

```go
role, err := client.Roles.Get(ctx, roles.GetRequest{Role: "<role-id-or-name>"})
```

### Create a Role

```go
role, err := client.Roles.Create(ctx, roles.CreateRequest{
    Name:        "read-only",
    Description: "read-only access to projects",
    Permissions: []roles.Permission{roles.Permissions.ProjectRead, roles.Permissions.ProjectSpanRead},
})
```

### Update a Role

PATCH pointers: nil preserves, non-nil replaces; `&""` clears `Description` (`Name`/`Permissions` cannot be emptied).

```go
desc := ""
role, err := client.Roles.Update(ctx, roles.UpdateRequest{
    Role:        "<role-id>",
    Description: &desc,
    Permissions: &[]roles.Permission{roles.Permissions.ProjectRead, roles.Permissions.DatasetRead},
})
```

### Delete a Role

Predefined roles cannot be deleted (server returns 400).

```go
err := client.Roles.Delete(ctx, roles.DeleteRequest{Role: "<role-id>"})
```

## Operations on Role Bindings

Use `client.RoleBindings` to bind a role to a user on a resource. These take strict IDs
(no name resolution).

A full runnable example lives in [`examples/rolebindings`](./examples/rolebindings).

### Get a Role Binding

```go
rb, err := client.RoleBindings.Get(ctx, rolebindings.GetRequest{RoleBindingID: "<binding-id>"})
```

### Create a Role Binding

```go
rb, err := client.RoleBindings.Create(ctx, rolebindings.CreateRequest{
    ResourceID:   "<space-id>",
    ResourceType: rolebindings.RoleBindingResourceTypeSPACE,
    RoleID:       "<role-id>",
    UserID:       "<user-id>",
})
```

### Update a Role Binding

```go
rb, err := client.RoleBindings.Update(ctx, rolebindings.UpdateRequest{RoleBindingID: "<binding-id>", RoleID: "<new-role-id>"})
```

### Delete a Role Binding

```go
err := client.RoleBindings.Delete(ctx, rolebindings.DeleteRequest{RoleBindingID: "<binding-id>"})
```

## Operations on API Keys

Use `client.APIKeys` to manage API keys. The plaintext key is returned **only** at creation
or refresh — store it immediately, it cannot be retrieved later. `CreateServiceKey`
provisions a bot user with space/org/account roles.

A full runnable example lives in [`examples/apikeys`](./examples/apikeys).

### List API Keys

Filter by key type and status.

```go
resp, err := client.APIKeys.List(ctx, apikeys.ListRequest{
    KeyType: apikeys.APIKeyTypeUser,
    Status:  apikeys.APIKeyStatusActive,
    Limit:   25,
})
```

### Create an API Key

`created.Key` holds the only copy of the secret.

```go
created, err := client.APIKeys.Create(ctx, apikeys.CreateRequest{
    Name:      "example-key",
    ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // zero = never expires
})
```

### Create a Service Key

A key bound to a bot user with roles in a space.

```go
svc, err := client.APIKeys.CreateServiceKey(ctx, apikeys.CreateServiceKeyRequest{
    Name:  "ci-bot",
    Space: "<space-id-or-name>",
})
```

### Refresh an API Key

Refresh rotates an existing API key, invalidating the old one and returning a new key value. Use `GracePeriodSeconds` to keep the old key valid for a short overlap period while rolling over clients.

```go
newKey, err := client.APIKeys.Refresh(ctx, apikeys.RefreshRequest{
    APIKeyID:           "<api-key-id>",
    ExpiresAt:          time.Now().Add(90 * 24 * time.Hour), // optional; zero = never expires
    GracePeriodSeconds: 3600, // optional; if not specified, old key is deleted immediately
})
// newKey.ApiKeyValue — the new raw key value (only returned at refresh time)

```

### Revoke an API Key

Revoke sets the key's status to `revoked`, deactivating it immediately. Revoking is
irreversible; revoking an already-revoked key is a no-op and still succeeds.

```go
err := client.APIKeys.Revoke(ctx, apikeys.RevokeRequest{APIKeyID: "<key-id>"})
```

## Operations on Resource Restrictions

Use `client.ResourceRestrictions` to restrict a resource (e.g. a project), preventing
roles bound at higher levels (space, org, account) from granting access. Both
calls take the restricted **resource** ID (e.g. a project ID), not a
restriction-record ID. Only `PROJECT` resources are supported today.

A full runnable example lives in [`examples/resourcerestrictions`](./examples/resourcerestrictions).

### Restrict a Resource

`Restrict` is idempotent — restricting an already-restricted resource returns the existing restriction without error.

```go
rr, err := client.ResourceRestrictions.Restrict(ctx, resourcerestrictions.RestrictRequest{
    ResourceID: "<resource-id>",
})
_ = rr // *ResourceRestriction
```

### Unrestrict a Resource

```go
if err := client.ResourceRestrictions.Unrestrict(ctx, resourcerestrictions.UnrestrictRequest{ResourceID: "<resource-id>"}); err != nil {
    var nfe *arize.NotFoundError
    if errors.As(err, &nfe) {
        // resource restriction already gone
    }
}
```

## Operations on Annotation Queues

Use `client.AnnotationQueues` to manage annotation queues — collections of records
(spans or dataset examples) routed to annotators for human labeling.

> **Alpha:** annotation queues are a pre-release feature. Every method emits a
> one-time pre-release warning and the API may change in a backward-incompatible way.

A full runnable example lives in [`examples/annotationqueues`](./examples/annotationqueues).

### List Annotation Queues

```go
resp, err := client.AnnotationQueues.List(ctx, annotationqueues.ListRequest{
    Space: "<space-id-or-name>", // optional
    Name:  "eval",               // optional substring filter
    Limit: 50,                   // optional
})
```

### Get an Annotation Queue

```go
queue, err := client.AnnotationQueues.Get(ctx, annotationqueues.GetRequest{
    AnnotationQueue: "<queue-id-or-name>",
    Space:           "<space-id-or-name>", // required when AnnotationQueue is a name
})
```

### Create an Annotation Queue

```go
queue, err := client.AnnotationQueues.Create(ctx, annotationqueues.CreateRequest{
    Space:            "<space-id-or-name>",
    Name:             "my-queue",
    Instructions:     "Rate each response for helpfulness.",      // optional
    AnnotatorEmails:  []annotationqueues.Email{"annotator@example.com"}, // optional
    AssignmentMethod: annotationqueues.AssignmentMethodAll,        // optional
})
```

### Add Records to a Queue

Build a record source with `NewSpanRecordSource` (spans) or `NewExampleRecordSource`
(dataset examples), then add up to two sources per request.

```go
src, err := annotationqueues.NewSpanRecordSource(annotationqueues.AnnotationQueueSpanRecordInput{
    ProjectId: "<project-id>",
    StartTime: time.Now().Add(-24 * time.Hour),
    EndTime:   time.Now(),
})
if err != nil {
    // handle error
}
resp, err := client.AnnotationQueues.AddRecords(ctx, annotationqueues.AddRecordsRequest{
    AnnotationQueue: "<queue-id-or-name>",
    Space:           "<space-id-or-name>",
    RecordSources:   []annotationqueues.AnnotationQueueRecordInput{src},
})
```

### Annotate a Record

`RecordID` is a strict ID (no name resolution) — read it from `ListRecords`.

```go
score := 0.9
result, err := client.AnnotationQueues.Annotate(ctx, annotationqueues.AnnotateRequest{
    AnnotationQueue: "<queue-id-or-name>",
    Space:           "<space-id-or-name>",
    RecordID:        "<record-id>",
    Annotations:     []annotationqueues.AnnotationInput{{Name: "helpfulness", Score: &score}},
})
```

### Update an Annotation Queue

Patch semantics: a `nil` field is left unchanged, a non-nil field replaces the value.

```go
instructions := "Updated instructions."
queue, err := client.AnnotationQueues.Update(ctx, annotationqueues.UpdateRequest{
    AnnotationQueue: "<queue-id-or-name>",
    Space:           "<space-id-or-name>", // required when AnnotationQueue is a name
    Instructions:    &instructions,
})
```

### Delete an Annotation Queue

```go
err := client.AnnotationQueues.Delete(ctx, annotationqueues.DeleteRequest{
    AnnotationQueue: "<queue-id-or-name>",
    Space:           "<space-id-or-name>", // required when AnnotationQueue is a name
})
```

## Operations on Users

Use `client.Users` to manage account users. These take strict user IDs, except `Get`,
whose `User` field accepts a user ID **or** an email address.

> **Alpha:** users are a pre-release feature. Every method emits a one-time
> pre-release warning and the API may change in a backward-incompatible way.

A full runnable example lives in [`examples/users`](./examples/users).

### List Users

`Email` is a case-insensitive substring filter; `Status` filters by account state
(active, invited, expired).

```go
resp, err := client.Users.List(ctx, users.ListRequest{
    Status: []users.UserStatus{users.UserStatusActive, users.UserStatusInvited},
    Limit:  25,
})
```

### Get a User

`User` accepts a user ID or an email address (resolved by case-insensitive exact
match; a non-matching email yields a `*ResourceNotFoundError`).

```go
user, err := client.Users.Get(ctx, users.GetRequest{User: "user@example.com"})
```

### Create a User

Build the account-level role with `AssignPredefinedRole` (or `AssignCustomRole`).
`InviteMode` controls whether and how an invitation is sent.

```go
user, err := client.Users.Create(ctx, users.CreateRequest{
    Name:       "Ada Lovelace",
    Email:      "user@example.com",
    Role:       users.AssignPredefinedRole(users.UserRoleMember),
    InviteMode: users.InviteModeEmailLink,
})
```

### Update a User

`Name`/`IsDeveloper` are PATCH pointers: nil preserves the current value, non-nil
sets it. At least one must be non-nil.

```go
newName := "Ada Lovelace"
isDeveloper := true
user, err := client.Users.Update(ctx, users.UpdateRequest{
    UserID:      "<user-id>",
    Name:        &newName,
    IsDeveloper: &isDeveloper,
})
```

### Delete a User

Soft-delete; cascades to org/space memberships, API keys, and role bindings.
Idempotent.

```go
err := client.Users.Delete(ctx, users.DeleteRequest{UserID: "<user-id>"})
```

### Bulk Delete Users

Delete by ID and/or email. Per-user outcomes are returned rather than aborting
the batch: an unresolved email is `DeletionStatusNotFound`, a failed delete is
`DeletionStatusFailed`.

```go
results, err := client.Users.BulkDelete(ctx, users.BulkDeleteRequest{
    UserIDs: []string{"<user-id>"},
    Emails:  []string{"user@example.com"},
})
for _, r := range results {
    // UserID is set for resolved users; Email is set when the user was
    // specified by email (and is the only identifier for not_found).
    fmt.Printf("user=%q email=%q: %s\n", r.UserID, r.Email, r.Status)
}
```

### Resend an Invitation

The target user must still be in the invited state.

```go
err := client.Users.ResendInvitation(ctx, users.ResendInvitationRequest{UserID: "<user-id>"})
```

### Reset a Password

Password-auth users only (not SSO/SAML); the account must be verified.

```go
err := client.Users.ResetPassword(ctx, users.ResetPasswordRequest{UserID: "<user-id>"})
```

## Operations on Tasks

Use `client.Tasks` to manage tasks — automated jobs that evaluate data
(`template_evaluation` / `code_evaluation`) or run experiments
(`run_experiment`) — and their async runs.

> **Alpha:** tasks are a pre-release feature. Every method emits a one-time
> pre-release warning and the API may change in a backward-incompatible way.

A full runnable example lives in [`examples/tasks`](./examples/tasks).

### List Tasks

`Space`, `Project`, and `Dataset` accept a name or ID and filter results;
`Type` filters by task type.

```go
resp, err := client.Tasks.List(ctx, tasks.ListRequest{
    Space: "demo",
    Type:  tasks.TaskTypeTemplateEvaluation,
    Limit: 25,
})
```

### Get a Task

`Task` accepts a task name or ID; `Space` is required when it is a name.

```go
task, err := client.Tasks.Get(ctx, tasks.GetRequest{Task: "my-task", Space: "demo"})
```

### Create an Evaluation Task

Creates a `template_evaluation` or `code_evaluation` task (set `Type`).
Exactly one of `Project` or `Dataset` must be set; dataset-based tasks
require at least one entry in `ExperimentIDs`, and `SamplingRate` /
`IsContinuous` apply only to project-based tasks.

```go
task, err := client.Tasks.CreateEvaluationTask(ctx, tasks.CreateEvaluationTaskRequest{
    Name:    "relevance-eval",
    Type:    tasks.TaskTypeTemplateEvaluation,
    Project: "my-project",
    Space:   "demo",
    Evaluators: []tasks.EvaluatorInput{
        {EvaluatorID: "<evaluator-id>"},
    },
    SamplingRate: 0.5,
})
```

### Create a Run-Experiment Task

`RunConfiguration` is a oneOf — populate exactly one variant via
`FromLlmGenerationRunConfig` or `FromTemplateEvaluationRunConfig`.

```go
var rc tasks.RunConfiguration
err := rc.FromTemplateEvaluationRunConfig(tasks.TemplateEvaluationRunConfig{
    AiIntegrationId:    "<ai-integration-id>",
    Template:           "Is the answer relevant?\n{{input}}",
    ProvideExplanation: true,
})

task, err := client.Tasks.CreateRunExperimentTask(ctx, tasks.CreateRunExperimentTaskRequest{
    Name:             "nightly-experiment",
    Dataset:          "my-dataset",
    Space:            "demo",
    RunConfiguration: rc,
})
```

### Update a Task

Patch fields are pointers: nil preserves the current value, non-nil sets it
(a pointer to `""` clears `QueryFilter`). The SDK fetches the task first to
validate the fields against its type — `SamplingRate`, `IsContinuous`,
`QueryFilter`, and `Evaluators` apply only to evaluation tasks;
`RunConfiguration` only to run_experiment tasks. An empty patch returns
`tasks.ErrNoUpdateFields`.

```go
rate := float32(0.25)
task, err := client.Tasks.Update(ctx, tasks.UpdateRequest{
    Task:         "relevance-eval",
    Space:        "demo",
    SamplingRate: &rate,
})
```

### Delete a Task

Irreversible; cascades to the task's runs and configurations.

```go
err := client.Tasks.Delete(ctx, tasks.DeleteRequest{Task: "relevance-eval", Space: "demo"})
```

### Trigger a Run

Starts an async run (returned in `pending` status). The SDK fetches the task
first to validate the fields against its type: evaluation tasks take
`DataStartTime` / `DataEndTime` / `MaxSpans` / `OverrideEvaluations` /
`ExperimentIDs`; run_experiment tasks take `ExperimentName` (required),
`DatasetVersionID`, `ExampleIDs` / `MaxExamples` (mutually exclusive),
`TracingMetadata`, and `EvaluationTaskIDs`.

```go
run, err := client.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{
    Task:          "relevance-eval",
    Space:         "demo",
    DataStartTime: time.Now().Add(-time.Hour),
    MaxSpans:      1000,
})
```

### List Task Runs

`Status` filters by run state (pending, running, completed, failed, cancelled).

```go
resp, err := client.Tasks.ListRuns(ctx, tasks.ListRunsRequest{
    Task:   "relevance-eval",
    Space:  "demo",
    Status: tasks.TaskRunStatusCompleted,
})
```

### Get a Run

Takes a strict run ID; use it to poll a run triggered by `TriggerRun`.

```go
run, err := client.Tasks.GetRun(ctx, tasks.GetRunRequest{RunID: "<run-id>"})
```

### Cancel a Run

Only valid while the run is pending or running.

```go
run, err := client.Tasks.CancelRun(ctx, tasks.CancelRunRequest{RunID: "<run-id>"})
```

### Wait for a Run

Polls until the run reaches a terminal state (completed, failed, or
cancelled). Defaults: 5s poll interval, 10m timeout; on expiry the error
wraps `tasks.ErrWaitTimeout`.

```go
run, err := client.Tasks.WaitForRun(ctx, tasks.WaitForRunRequest{
    RunID:        "<run-id>",
    PollInterval: 5 * time.Second,
    Timeout:      5 * time.Minute,
})
```

# SDK Configuration

## Environment Variables

Most `Config` fields fall back to an environment variable when unset. Highlights:

| Env var                          | Config field            | Default                |
| -------------------------------- | ----------------------- | ---------------------- |
| `ARIZE_API_KEY`                  | `APIKey`                | _(required)_           |
| `ARIZE_API_HOST`                 | `APIHost`               | `api.arize.com`        |
| `ARIZE_API_SCHEME`               | `APIScheme`             | `https`                |
| `ARIZE_REGION`                   | `Region`                | _(unset)_              |
| `ARIZE_BASE_DOMAIN`              | `BaseDomain`            | _(unset)_              |
| `ARIZE_SINGLE_HOST`              | `SingleHost`            | _(unset)_              |
| `ARIZE_SINGLE_PORT`              | `SinglePort`            | _(unset)_              |
| `ARIZE_REQUEST_VERIFY`           | `InsecureSkipVerify`    | `false` (verified)     |
| `ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB` | `MaxHTTPPayloadSizeMB`  | `8`                    |
| `ARIZE_DIRECTORY`                | `ArizeDirectory`        | `~/.arize`             |
| `ARIZE_ENABLE_CACHING`           | `DisableCaching`        | `true` (caching on)    |
| `ARIZE_MAX_PAST_YEARS`           | `MaxPastYears`          | `5`                    |

Boolean env vars accept `1`, `true`, `yes`, `on` (case-insensitive) as truthy; any other non-empty value is treated as false.

> **Note:** `InsecureSkipVerify` and `DisableCaching` are named for the negative so their Go zero values give the safe default (TLS verification on, caching on). The corresponding env vars (`ARIZE_REQUEST_VERIFY`, `ARIZE_ENABLE_CACHING`) keep the positive name.

## TLS Verification

`InsecureSkipVerify` defaults to `false` — TLS certificates are verified. To opt out (e.g. for testing against a local server with a self-signed cert):

```go
client, _ := arize.NewClient(arize.Config{
    APIKey:             "<your-api-key>",
    InsecureSkipVerify: true, // or set ARIZE_REQUEST_VERIFY=false
})
```

## HTTP Timeout

```go
client, _ := arize.NewClient(arize.Config{
    APIKey:      "<your-api-key>",
    HTTPTimeout: 60 * time.Second, // defaults to 30s
})
```

## Local Directory and Caching

`ArizeDirectory` is reserved for SDK-managed local files (cache, logs). `DisableCaching` is wired through `Config` but caching itself is not yet active in v2 — these knobs exist now to keep call sites stable as features land.

```go
client, _ := arize.NewClient(arize.Config{
    APIKey:         "<your-api-key>",
    ArizeDirectory: "/var/lib/arize",
    DisableCaching: true,
})
```

# Community

Join our community to connect with thousands of AI builders.

- 🌍 Join our [Slack community](https://arize-ai.slack.com/join/shared_invite/zt-11t1vbu4x-xkBIHmOREQnYnYDH1GDfCg?__hstc=259489365.a667dfafcfa0169c8aee4178d115dc81.1733501603539.1733501603539.1733501603539.1&__hssc=259489365.1.1733501603539&__hsfp=3822854628&submissionGuid=381a0676-8f38-437b-96f2-fc10875658df#/shared-invite/email).
- 📚 Read our [documentation](https://docs.arize.com/arize).
- 💡 Ask questions and provide feedback in the _#arize-support_ channel.
- 𝕏 Follow us on [𝕏](https://twitter.com/ArizeAI).
- 🧑‍🏫 Deep dive into everything [Agents](http://arize.com/ai-agents/) and [LLM Evaluations](https://arize.com/llm-evaluation) on Arize's Learning Hubs.

Copyright 2026 Arize AI, Inc. All Rights Reserved.
