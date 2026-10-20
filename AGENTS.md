# AGENTS.md

Mattermost plugin that shows each post translated into the language each user picks.
The original post does not change. The plugin has a Go server part (`server/`) and a
React/TypeScript web app part (`webapp/`). It targets Mattermost Team Edition 10.12.0.

For the design and the commit sequence, see [`docs/PLAN.md`](docs/PLAN.md). Work through
the plan one commit at a time.

## Rules

- **Translate on view, never on post.** The web app requests a translation when it shows
  a post. Do not call the translator from `MessageWillBePosted`, `MessageHasBeenPosted`,
  or any other post hook. This keeps usage inside the Azure F0 quota and limits the text
  that leaves the server.
- **Check channel access first.** The translation endpoint reads `Mattermost-User-Id`,
  loads the post, and calls `HasPermissionToChannel(userID, post.ChannelId,
  PermissionReadChannel)` before anything else. Every new endpoint that takes a post or
  channel ID does the same.
- **Never change the original post.** Translations live in the plugin's KV store, not in
  `post.Message` or `post.Props`.
- **The translator is an interface.** Code outside `server/translator` does not know which
  service it talks to.
- **Never point a dev loop at production.** `make deploy` and `make watch` target the
  local Mattermost only. Production is mattermost.hedgehogapplications.nl; uploading
  there needs an explicit go from the user.
- **Small commits.** One step of the plan per commit. `make quick` passes before each
  commit. Wait for the user's review and an explicit go before every commit and push.

## Commands

Run `make quick` before you ask for review. It runs `go vet`, `golangci-lint`, the server
unit tests, and the web app lint, type check, and tests. It needs no Docker.

```bash
make quick                                     # fast check before review
SKIP_DOCKER_TESTS=1 go test ./server/...       # server unit tests only
go test ./server/translator -run TestAzure     # single test
cd webapp && npx jest src/manifest.test.tsx    # single webapp test
make test                                      # full check, with Docker, before a release
```

## Local Mattermost

`dev/docker-compose.yml` runs Mattermost Team Edition 10.12.0 with Postgres on
http://localhost:8065. `build/custom.mk` points `make deploy` and `make watch` at it by
default.

```bash
make dev-up            # start it, and seed team `dev` with users `admin` and `colleague`
make deploy            # build the plugin and install it on the local Mattermost
make dev-down          # stop it; DEV_RESET=1 also deletes its data
```

The logins are `admin` / `Admin-dev-1!` (system admin) and `colleague` /
`Colleague-dev-1!`. To check a change as a user, log in as `colleague` in a private
browser window. To run admin commands, use
`docker compose -f dev/docker-compose.yml exec mattermost mmctl --local <command>`.

`make help` lists the other targets (deploy, watch, release bumps, debugger attach).
Project-specific targets go in `build/custom.mk`.

## Architecture

- `plugin.json` is the source of truth for ID, version, and settings. `make apply`
  generates `server/manifest.go` and `webapp/src/manifest.ts` from it. Both are
  gitignored, so never edit them.
- Package layout: stay in package `main` unless there is a reason not to. The translator
  client is in its own package, `server/translator`.
- The user's target language is a Mattermost preference: category
  `pp_nl.tijnschutte.translator`, name `language`. The web app sets it through
  `registerUserSettings`.

## Backend tests

- **Integration tests** (`server/integration_test.go`) are the default for anything that
  goes through the Mattermost server: plugin API calls, hooks, KV persistence,
  permissions, HTTP routing, and activation. `testhelper.Setup(t)` runs Postgres and
  Mattermost through testcontainers and deploys the bundle from `dist/`. For fixtures and
  environment variables, see
  `go doc github.com/mattermost/mattermost/server/public/pluginapi/testhelper`. Each
  `Setup` resets the server (about 10 seconds), so group related assertions into one test.
- **Plain unit tests** (`server/plugin_test.go`) for logic that does not touch
  `p.API` or `p.client`. For example, drive `ServeHTTP` with `httptest` after
  `initRouter()`, with a fake translator.
- **Translator tests** run the real client against an `httptest` server that acts as
  Azure. They never call the real Azure API.
- **Mocked-API unit tests** only for error paths a real server cannot easily produce: wrap
  a `plugintest.API` in `pluginapi.NewClient` and set `api.On(...)` expectations. Avoid
  them otherwise; they encode assumptions about the server instead of testing it.

Integration test caveats:

- There is no build tag, so `go test ./...` includes them and fails without Docker.
  `SKIP_DOCKER_TESTS=1` skips them.
- `build/custom.mk` sets `MM_TEST_IMAGE=mattermost/mattermost-team-edition:10.12.0`, the
  version production runs. When you run `go test` by hand, set it yourself; the
  testhelper default is `mattermost-enterprise-edition:latest`.
- They test the bundle in `dist/`, which must be linux/amd64. After a server change,
  rebuild it before you run one by hand:
  ```bash
  make apply server bundle MM_SERVICESETTINGS_ENABLEDEVELOPER=true DEFAULT_GOOS=linux DEFAULT_GOARCH=amd64
  go test ./server -run TestTranslationEndpoint
  ```
  This skips the web app build but needs an existing `webapp/dist`. The first time, use
  `make dist` with the same variables.
