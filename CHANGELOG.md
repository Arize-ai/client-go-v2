# Changelog

## [0.8.0](https://github.com/Arize-ai/arize/compare/arize-go-sdk-v2/v0.7.0...arize-go-sdk-v2/v0.8.0) (2026-05-29)


### 🎁 New Features

* Add Evaluators support ([#73226](https://github.com/Arize-ai/arize/issues/73226)) ([1b02ac8](https://github.com/Arize-ai/arize/commit/1b02ac8bd754ffb02cb07ba3439df2a66e48c4cc))
* Add Annotation Queues support ([#71421](https://github.com/Arize-ai/arize/issues/71421)) ([39a971e](https://github.com/Arize-ai/arize/commit/39a971ebec5a3235d85ea76a5084dd2530e5c79f))

## [0.7.0](https://github.com/Arize-ai/arize/compare/arize-go-sdk-v2/v0.6.0...arize-go-sdk-v2/v0.7.0) (2026-05-28)


### 🎁 New Features

* Add AI Integrations support ([#71415](https://github.com/Arize-ai/arize/issues/71415)) ([b1ccab6](https://github.com/Arize-ai/arize/commit/b1ccab60f7aaf9da296d39f2d470f0fdba07b0e9))
* Add datasets support ([#72996](https://github.com/Arize-ai/arize/issues/72996)) ([561eef6](https://github.com/Arize-ai/arize/commit/561eef62bb4e8790c46c215d3ce405e7449ab7f0))
* Add annotationconfigs support ([#71417](https://github.com/Arize-ai/arize/issues/71417)) ([91fef1f](https://github.com/Arize-ai/arize/commit/91fef1f67dcb9801c909edfd66136e302d3cfa0d))
* Support for project updates ([#72941](https://github.com/Arize-ai/arize/issues/72941)) ([6f3fb05](https://github.com/Arize-ai/arize/commit/6f3fb0525ae386fe57ff9a30234825e311c4c891))
* Add prompts support ([#72998](https://github.com/Arize-ai/arize/issues/72998)) ([1a3d5d3](https://github.com/Arize-ai/arize/commit/1a3d5d3b120b75037579c01ad71f7aaf80666dc0))


## [0.6.0](https://github.com/Arize-ai/arize/compare/arize-go-sdk-v2/v0.5.0...arize-go-sdk-v2/v0.6.0) (2026-05-27)


### 🎁 New Features

* add Roles client with List, Get, Create, Update, Delete and auto-generated Permissions namespace ([#71414](https://github.com/Arize-ai/arize/issues/71414)) ([4d33e77](https://github.com/Arize-ai/arize/commit/4d33e77cb7f94ab1c160e9dd617d4983b4df0dd3))
* add spaces subclient with full CRUD and membership operations ([#71416](https://github.com/Arize-ai/arize/issues/71416)) ([155eecb](https://github.com/Arize-ai/arize/commit/155eecb410d20c7960be9c5a8e6131d02a1a824f))


### 🐛 Bug Fixes

* raise AmbiguousNameError when multiple spaces share a name ([#72449](https://github.com/Arize-ai/arize/issues/72449)) ([6e71959](https://github.com/Arize-ai/arize/commit/6e71959db6b952aad77648532921355c028cfb98))

## [0.5.0](https://github.com/Arize-ai/arize/compare/arize-go-sdk-v2/v0.4.0...arize-go-sdk-v2/v0.5.0) (2026-05-20)


### 🎁 New Features

* accept resource names (not just IDs) across subclient request ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* **apikeys:** add CreateServiceKey method for creating service-scoped ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* **apikeys:** default List page size to 50 when limit is unspecified ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* **openapi:** extract inline enums for api_keys ([#71644](https://github.com/Arize-ai/arize/issues/71644)) ([4f91923](https://github.com/Arize-ai/arize/commit/4f9192351487624fb0eca3bbf90d4463bccb3e5b))
* **organizations:** add Delete, AddUser, and RemoveUser methods ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* **organizations:** return ErrNoUpdateFields when Update is called with ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* **spans:** add Annotate method for upserting human annotations on ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))


### 📚 Documentation

* document optional pointer fields and codify nil-behavior ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* add runnable examples programs for apikeys, organizations, ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* drop internal resolver references from public godocs ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))


### 💫 Code Refactoring

* collapse optional pointer fields to value types ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* **resourcerestrictions:** rename Create/Delete to ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))
* **sdk:** unify method signatures and request payload types across all subclients ([#72099](https://github.com/Arize-ai/arize/issues/72099)) ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))


### 🧪 Tests

* expand subclient coverage with default-limit, create-validation, ([52612cf](https://github.com/Arize-ai/arize/commit/52612cf9d495ee86ff2cd8b6f44e935d0c1ec27f))

## [0.4.0](https://github.com/Arize-ai/arize/compare/arize-go-sdk-v2/v0.3.0...arize-go-sdk-v2/v0.4.0) (2026-05-16)


### 🎁 New Features

* add organizations subclient with full CRUD support ([#71412](https://github.com/Arize-ai/arize/issues/71412)) ([22821e1](https://github.com/Arize-ai/arize/commit/22821e1aca1c137d51bc64d0e867cd2892b44666))
* add projects subclient with CRUD operations and tests ([#71413](https://github.com/Arize-ai/arize/issues/71413)) ([cd9b5f3](https://github.com/Arize-ai/arize/commit/cd9b5f3869225ba1be7db7e49795741a4a87e3cd))


### 💫 Code Refactoring

* remove sdkconfig dependency from apikeys, rolebindings, ([22821e1](https://github.com/Arize-ai/arize/commit/22821e1aca1c137d51bc64d0e867cd2892b44666))
* remove redundant package nouns from operation-scoped type names ([#71924](https://github.com/Arize-ai/arize/issues/71924)) ([fb02be3](https://github.com/Arize-ai/arize/commit/fb02be3217bc00462bdbdb4581617569111bc4f5))


### 🧪 Tests

* expand clients test coverage ([22821e1](https://github.com/Arize-ai/arize/commit/22821e1aca1c137d51bc64d0e867cd2892b44666))


### 🔀 Continuous Integration

* warm Go SDK pkg.go.dev proxy and canonicalize LICENSE ([#71950](https://github.com/Arize-ai/arize/issues/71950)) ([17798fe](https://github.com/Arize-ai/arize/commit/17798fe8fb906ceb4f8421db56851c2939cc532d))

## [0.3.0](https://github.com/Arize-ai/arize/compare/arize-go-sdk-v2/v0.2.0...arize-go-sdk-v2/v0.3.0) (2026-05-14)


### 🎁 New Features

* add spans subclient to Go SDK ([#71410](https://github.com/Arize-ai/arize/issues/71410)) ([c5dedb5](https://github.com/Arize-ai/arize/commit/c5dedb5c9aa3f6d05833585510ed0b51c3ee57ea))
* add API keys subclient ([#71411](https://github.com/Arize-ai/arize/issues/71411)) ([1d029f4](https://github.com/Arize-ai/arize/commit/1d029f48e3dd414f20e39ce0a0aa2e2c3da44d3f))
* create root package types ([#71411](https://github.com/Arize-ai/arize/issues/71411)) ([1d029f4](https://github.com/Arize-ai/arize/commit/1d029f48e3dd414f20e39ce0a0aa2e2c3da44d3f))


### 🐛 Bug Fixes

* add missing exported types in subclients ([#71411](https://github.com/Arize-ai/arize/issues/71411)) ([1d029f4](https://github.com/Arize-ai/arize/commit/1d029f48e3dd414f20e39ce0a0aa2e2c3da44d3f))

## [0.2.0](https://github.com/Arize-ai/arize/compare/arize-go-sdk-v2/v0.1.1...arize-go-sdk-v2/v0.2.0) (2026-05-14)


### 🎁 New Features

* add rolebindings subclient  ([#71409](https://github.com/Arize-ai/arize/issues/71409)) ([72e0dde](https://github.com/Arize-ai/arize/commit/72e0dde21a727ab3e32d8dc315674c7e7650c1b1))


### 🐛 Bug Fixes

* rename request types to remove `Body` prefix for naming consistency ([72e0dde](https://github.com/Arize-ai/arize/commit/72e0dde21a727ab3e32d8dc315674c7e7650c1b1))


### 📚 Documentation

* fix `InsecureSkipVerify` default in README ([f1e76e6](https://github.com/Arize-ai/arize/commit/f1e76e6a51274824bf409112a5eaf3acfd757c14))


### 💫 Code Refactoring

* convert tests to table-driven style ([#71774](https://github.com/Arize-ai/arize/issues/71774)) ([f1e76e6](https://github.com/Arize-ai/arize/commit/f1e76e6a51274824bf409112a5eaf3acfd757c14))

## [0.1.0](https://github.com/Arize-ai/arize/compare/client-go-v2/v0.1.0...client-go-v2/v0.0.0) (2026-05-12)


### 🎁 New Features

* add resource restrictions subclient ([#71408](https://github.com/Arize-ai/arize/issues/71408)) ([3c5ed49](https://github.com/Arize-ai/arize/commit/3c5ed499c0ba9322f5a9d50158cbd8d5739fab76))
* establish Go SDK v2 foundation and add core Arize Client ([#71407](https://github.com/Arize-ai/arize/issues/71407)) ([1999a82](https://github.com/Arize-ai/arize/commit/1999a82479d3f09b8bb472253aafa535cc394634))
* initialize foundational v2 Go SDK structure ([#71404](https://github.com/Arize-ai/arize/issues/71404)) ([c952ccc](https://github.com/Arize-ai/arize/commit/c952cccb62622373774f16bec4cf41e1ce018075))
