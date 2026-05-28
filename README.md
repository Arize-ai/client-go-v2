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
  - [Operations on Projects](#operations-on-projects)
    - [List Projects](#list-projects)
    - [Create a Project](#create-a-project)
    - [Get a Project](#get-a-project)
    - [Update a Project](#update-a-project)
    - [Delete a Project](#delete-a-project)
  - [Operations on Resource Restrictions](#operations-on-resource-restrictions)
    - [Create a Resource Restriction](#create-a-resource-restriction)
    - [Delete a Resource Restriction](#delete-a-resource-restriction)
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
- **Projects** subclient — list, create, get, update, and delete projects.
- **Resource Restrictions** subclient — create and delete restrictions on Arize resources.
- **Typed HTTP errors** matchable via `errors.As` (`BadRequestError`, `UnauthorizedError`, `NotFoundError`, …).

Additional resource domains (spans, datasets, experiments, prompts, evaluators, tasks, organizations, spaces, users) will be added incrementally.

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

err := client.ResourceRestrictions.Delete(ctx, resourcerestrictions.DeleteRequest{ResourceRestrictionID: "nonexistent"})

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

## Operations on Projects

Use `client.Projects` to manage projects, which are namespaces for organizing tracing data.

### List Projects

```go
resp, err := client.Projects.List(ctx, projects.ListRequest{
    Space: "<space-id-or-name>", // optional
    Name:  "prod",               // optional substring filter
    Limit: 50,                   // optional
})
```

### Create a Project

```go
proj, err := client.Projects.Create(ctx, projects.CreateRequest{
    Name:  "my-project", // must be unique within the space
    Space: "<space-id-or-name>",
})
```

### Get a Project

```go
proj, err := client.Projects.Get(ctx, projects.GetRequest{
    Project: "<project-id-or-name>",
    Space:   "<space-id-or-name>", // required when Project is a name
})
```

### Update a Project

```go
proj, err := client.Projects.Update(ctx, projects.UpdateRequest{
    Project: "<project-id-or-name>",
    Space:   "<space-id-or-name>", // required when Project is a name
    Name:    "renamed-project",    // must be unique within the space
})
```

### Delete a Project

```go
err := client.Projects.Delete(ctx, projects.DeleteRequest{
    Project: "<project-id-or-name>",
    Space:   "<space-id-or-name>", // required when Project is a name
})
```

## Operations on Resource Restrictions

Use `client.ResourceRestrictions` to restrict access to resources (e.g. projects) for users without the relevant permissions.

### Create a Resource Restriction

```go
import (
    "context"

    "github.com/Arize-ai/client-go-v2/arize"
    "github.com/Arize-ai/client-go-v2/arize/resourcerestrictions"
)

ctx := context.Background()
resp, err := client.ResourceRestrictions.Create(ctx, resourcerestrictions.CreateRequest{
    ResourceID: "<resource-id>",
})
if err != nil {
    // handle error
}
_ = resp.ResourceRestriction
```

### Delete a Resource Restriction

```go
if err := client.ResourceRestrictions.Delete(ctx, resourcerestrictions.DeleteRequest{ResourceRestrictionID: "<resource-id>"}); err != nil {
    var nfe *arize.NotFoundError
    if errors.As(err, &nfe) {
        // resource restriction already gone
    }
}
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

Copyright 2025 Arize AI, Inc. All Rights Reserved.
