# The Things Stack v3

The Things Stack for LoRaWAN. Backend is Go; the web frontends (Console, Account App) are React. Development tooling is driven by [Mage](https://magefile.org/) via `tools/bin/mage`.

This file distills [CONTRIBUTING.md](CONTRIBUTING.md) and [DEVELOPMENT.md](DEVELOPMENT.md). Consult those for full details if in doubt. In case there is any mismatch with this file, the other files take precedence.

## Setup and common commands

```bash
make init                          # initialize tooling and dependencies (slow; run once, and after tooling changes)

# Run a development stack
tools/bin/mage js:build            # build frontend assets into public/ (slow, takes minutes)
tools/bin/mage dev:dbStart         # start databases in Docker (dev:dbStop also exists; dev:dbErase DESTROYS local data)
tools/bin/mage dev:initStack       # create DB, migrate, create admin/admin user
go run ./cmd/ttn-lw-stack -c ./config/stack/ttn-lw-stack.yml start   # Console at http://localhost:1885/

# Interactive frontend dev environment (stack + webpack-dev-server on :8080)
tools/bin/mage dev:serveDevWebui

# Tests
tools/bin/mage go:test             # full Go suite (slow — prefer targeted runs below)
tools/bin/mage js:test             # frontend unit tests (Jest over pkg/webui)
tools/bin/mage jsSDK:test          # JS SDK unit tests
go test ./pkg/<pkg>/...                # single package (preferred for targeted runs)
go test -run TestName ./pkg/<pkg>/...  # single Go test
node_modules/.bin/jest pkg/webui/<path-to-spec>  # single frontend test file
tools/bin/mage js:cypressHeadless  # end-to-end tests (see DEVELOPMENT.md for setup)
node_modules/.bin/cypress run --config-file config/cypress.config.js --spec <spec.js>  # single e2e spec

# Lint and format
golangci-lint run ./pkg/...        # Go lint (config in .golangci.yml)
golangci-lint run --new-from-rev=HEAD~1 ./<pkg_path>/... --fix # lint and fix only what has changed since the last commit
tools/bin/mage js:lint             # JS lint (eslint, config in config/eslintrc.yaml)
tools/bin/mage js:fmt              # JS format (prettier)

# Code generation — run after changing the corresponding sources, CI fails otherwise
tools/bin/mage proto:clean proto:all jsSDK:definitions   # after editing .proto files in api/ (needs Docker, slow)
tools/bin/mage go:messages         # after adding/changing errors, events or enums (updates config/messages.json)
tools/bin/mage js:translations     # after adding/changing frontend react-intl messages (updates pkg/webui/locales)
tools/bin/mage go:eventData        # after adding/changing events
```

Go tests that need Redis **skip themselves silently** unless `TEST_REDIS=1` is set and the databases are running (`dev:dbStart`) — passing output may mean skipped, not tested. `REDIS_ADDRESS`/`REDIS_DB` override the target instance; `TEST_SLOWDOWN` scales test timeouts on slow machines.

Add `-v` to mage for verbose output when a target fails silently. If a build is in a broken state, rebuild the frontend from scratch: `tools/bin/mage jsSDK:cleanDeps jsSDK:clean js:cleanDeps js:clean js:build`.

## Project structure

- `api/` — protocol buffer definitions (the source of truth for the API)
- `cmd/` — binaries: `ttn-lw-stack` and `ttn-lw-cli`
- `pkg/` — all Go libraries; one package per component (`networkserver`, `applicationserver`, `identityserver`, ...)
  - `pkg/ttnpb/` — generated code from protos (do not edit by hand)
  - `pkg/webui/` — frontend: `console/` and `account/` apps, shared `components/`, `containers/`, `lib/`, `locales/`
- `config/stack/` — development stack configuration files
- `cypress/` — frontend end-to-end tests
- `sdk/js/` — JavaScript SDK
- `data/` — data from external repositories (devices, frequency plans, webhook templates)
- `tools/` — Mage-based dev/test/build tooling
- `public/`, `release/` — build output, not committed

## Generated and vendored files — do not edit by hand

| Path                                        | Regenerate with                                          |
| ------------------------------------------- | -------------------------------------------------------- |
| `pkg/ttnpb/` and other proto-generated code | `tools/bin/mage proto:clean proto:all jsSDK:definitions` |
| `config/messages.json`                      | `tools/bin/mage go:messages`                             |
| `pkg/webui/locales/*.js`                    | `tools/bin/mage js:translations`                         |
| `data/`                                     | pulled from external repositories — do not edit          |

If CI complains about one of these files, fix the source (proto, error/event definition, message definition, `go.mod`) and regenerate — never patch the generated file directly.

## Go code style

- Formatting via `gofmt`/`goimports`; tabs for indentation (enforced by `.editorconfig`). Lint with `golangci-lint`.
- Line length: prefer a natural break past 80 columns, break lines past 120.
- Comments are full English sentences with a capital and a period. Every package and every top-level type, const, var and func gets a doc comment. TODO format: `// TODO: Description (https://github.com/TheThingsNetwork/lorawan-stack/issues/<number>).`
- Variable naming follows the standard library plus project conventions: `ctx`, `mu`, `conf`, `msg`, `srv`, `cnt`; entities `gtw`, `app`, `dev`, `usr`; component abbreviations `as`, `gs`, `is`, `js`, `ns`; well-known IDs `gtwID`, `appID`, `devEUI`, `joinEUI`, etc.
- API methods (protos): `VerbNoun` in upper camel case, imperative verb, CRUD order (`CreateType`, `GetType`, `ListTypes`, `UpdateType`, `DeleteType`). Short comments on every service, method, message and field in `.proto` files.
- Errors: define with `errors.Define<Type>(...)` close to the return statements that use them, preferably unexported. Names in snake case, short and unique within the package; no `failed_to_`/`_failed` affixes (prefer `missing_field` over `no_field`). Descriptions in lower case plain English, no trailing period. Error definitions are part of the API.
- Events: `events.Define("component.entity.action", "lowercase description")`, e.g. `ns.up.receive_duplicate`.
- Log field keys, event names, error names/attributes and task identifiers are snake case; LoRaWAN terms keep their separators (`DevAddr` → `dev_addr`, `AppSKey` → `app_s_key`).
- Errors, events and enum descriptions are user-visible and translated — after changing them run `tools/bin/mage go:messages`.
- Tests are mandatory: code without tests is incomplete. Write tests for all new code and update existing tests as necessary.
  - Keep tests simple, mutually exclusive and focused on the code they are testing.
  - Use `t.Context()`, not `context.Background()`.
  - Put tests in `_test` package unless instructed otherwise.
- Add logging selectively (warnings for recoverable errors, errors for unexpected failures). Log messages in imperative mood for actions starting, past tense with "Failed to" for errors.

## Frontend code style

- ES6+ transpiled with babel/webpack; prettier + eslint enforced in CI. Two-space indentation.
- Write new components as functional components with hooks. Do not introduce new HOCs or decorators.
- Component scopes: presentational (`components/`), container (`containers/`), view (`views/`), utility. Global components live in `pkg/webui/{components,containers,lib}`, app-specific ones under `pkg/webui/{console,account}/`. Presentational components get storybook stories.
- Import statement order is linter-enforced: builtins → external → internal (constants, api, components, utilities, store, assets) → parent → sibling → index, separated by blank lines.
- JSDoc comments on classes and functions; first line of a multi-line JSDoc block is empty; full sentences with periods; wrap code identifiers in backticks.
- All user-visible text uses `react-intl` messages (`intl.defineMessages` inline or in `pkg/webui/lib/shared-messages.js`), then run `tools/bin/mage js:translations`. Locale files are committed; discrepancies fail CI.

## Test style

- JS tests use `describe()`/`it()` (never bare `test()`), pattern "unit of work — expected behaviour when scenario": `describe('Login', ...)` capitalized, `it('succeeds when using correct credentials', ...)` lowercase, no `should`, no trailing period. React components are named `<MyComponent />` in describes.
- Cypress e2e specs live in `cypress/e2e/{account,console,smoke}` as `{context}.spec.js`, one entity/view per file. Select elements by label/role/text (Testing Library queries) and assert `be.visible`; use `data-test-id` only when realistic selection is not possible. Do test setup programmatically via custom commands, not by clicking through unrelated UI.

## Git conventions

- **Branch naming**: `fix/#-short-name`, `feature/#-short-name` or `issue/#-short-name` (issue number preferred). Keep branches small and scoped to a single task.
- **Commit messages**: `topic: Imperative subject starting with a capital`, e.g. `ns: Fix MIC check` or `ns,as,gs: Fix TLS check`. Accepted topics: `account`, `all`, `api`, `as`, `ci`, `cli`, `console`, `cs`, `data`, `dcs`, `dev`, `dr`, `dtc`, `es`, `gcs`, `gs`, `is`, `js`, `noc`, `ns`, `pba`, `qrg`, `tbs`, `util`. Body (optional) explains what and why, not how. Commits must be signed (`git commit -S`) and individually meaningful/cherry-pickable.
- **Changelog**: add user-facing changes to the `Unreleased` section of `CHANGELOG.md` (Keep a Changelog format: `Added`/`Changed`/`Deprecated`/`Removed`/`Fixed`/`Security`; bullets in imperative tense ending with a dot).

## Before you're done

Run through this checklist before considering a change complete:

1. Format and lint what you changed: `gofmt`/`goimports` and `golangci-lint run ./pkg/<changed>/...` for Go; `tools/bin/mage js:fmt js:lint` for frontend code.
2. Run the tests covering your change: `go test ./pkg/<changed>/...` (with databases running and `TEST_REDIS=1` if the package uses Redis) or the relevant Jest/Cypress specs. Add or update tests for the new behaviour.
3. Regenerate derived files if you touched their sources: protos → `proto:clean proto:all jsSDK:definitions`; errors/events/enums → `go:messages` (plus `go:eventData` for events); frontend `react-intl` messages → `js:translations`; `go.mod` → `go mod tidy`.
4. Add a `CHANGELOG.md` entry under `Unreleased` for user-facing changes.
5. Check that commit messages follow the `topic: Subject` format and are signed.
