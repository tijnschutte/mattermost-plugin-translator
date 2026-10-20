# Plan: Mattermost translator plugin

## Goal

Each user picks a language in their Mattermost settings. Under every post in another
language, the plugin shows a translation into that language. The original post does
not change. The first users are one Portuguese colleague and the Dutch team.

## Design decisions

- **Translate on view, not on post.** The web app asks the server for a translation
  when it shows a post. The server never translates in a post hook. Only posts that a
  user with a language setting looks at leave the server. This keeps usage inside the
  Azure F0 quota and limits the text that goes to Azure.
- **One endpoint:** `GET /plugins/nl.tijnschutte.translator/api/v1/posts/{postId}/translation`.
  1. Read the user from the `Mattermost-User-Id` header.
  2. Load the post. Check `HasPermissionToChannel(userID, post.ChannelId, PermissionReadChannel)`.
     Without this check, any user can read any post through the endpoint.
  3. Read the target language from the user's preferences: category
     `pp_nl.tijnschutte.translator`, name `language`. No language means no translation.
  4. Look up the KV cache by `(postId, targetLang)`. On a miss, call the translator and
     store the result.
- **Cache.** Store the translated text and the detected source language. If the source
  language is the target language, store an empty result, so the next view makes no call.
  `MessageHasBeenUpdated` deletes the cache entries for the edited post.
- **Burst control.** One channel view mounts many post components at once. Use
  `golang.org/x/sync/singleflight` per cache key, and cap the parallel Azure calls.
- **Translator interface** in `server/translator`:
  `Translate(ctx, text, to string) (translated, detectedFrom string, err error)`.
  Azure AI Translator is the first implementation. DeepL or LibreTranslate can follow
  without changes elsewhere.
- **Quota.** Azure F0 answers with error `403001` when the month's quota is used. The
  plugin then shows nothing, logs one warning, and does not retry until the next hour.
- **Rejected: translations in `post.props`.** Every user sees post props, an update sends
  an edit event to every client, and one post needs one entry per language.

## Known gaps

- The mobile app does not run web app plugins. Mobile users see no translations. A
  `/translate` slash command can fill this gap later.
- Markdown, mentions, and code blocks can change in translation. Commit 4 sends code
  blocks through untranslated. Other cases wait until they cause a problem.

## Prerequisites (human steps)

- Create an Azure account, a Translator resource on the F0 tier, and copy its key and
  region. Azure asks for a credit card; an F0 resource does not charge it.
- Docker runs on the laptop (version 29.8.2). The integration tests and the local
  Mattermost need it.

## Commits

Each commit builds, and `SKIP_DOCKER_TESTS=1 go test ./server/...` passes. Each commit
waits for review and an explicit go.

1. **Agent setup.** Rewrite `AGENTS.md` for this plugin: purpose, design decisions, the
   endpoint's permission check, the quota rule, commands, and commit rules. Keep the
   template's test guidance. `CLAUDE.md` keeps `@AGENTS.md`. Add this plan. Add a
   `make quick` target in `build/custom.mk`: `go vet`, unit tests, and web app lint and
   types. Run `make install-go-tools` so `golangci-lint` works locally.
2. **Remove template code.** The hello endpoint and `TestHelloEndpoint`, the hello command
   and its mocks, `job.go` and the background job, `kvstore/startertemplate.go`,
   `public/hello.html`, and the template icon. Set `min_server_version` to `10.12.0`,
   the version that production runs.
3. **Local Mattermost.** Add a `docker-compose.yml` with `mattermost-team-edition:10.12.0`
   and Postgres, plus `make dev-up` and the `MM_SERVICESETTINGS_SITEURL` and admin token
   steps in `README.md`. `make deploy` and `make watch` point at this instance only.
4. **Translator package.** The interface, the Azure client, and unit tests against an
   `httptest` fake Azure: success, detected language, quota error `403001`, and code
   blocks left untranslated.
5. **Plugin settings.** Azure key (`secret: true`), region, and endpoint in
   `plugin.json`. Validate them in `OnConfigurationChange`.
6. **Language preference and cache.** Read the user's language, and the KV cache with the
   `MessageHasBeenUpdated` cleanup. Unit tests.
7. **Endpoint.** The route, the permission check, `singleflight`, and the call limit.
   Unit tests on `ServeHTTP` with a fake translator.
8. **Web app: language setting.** `registerUserSettings` with a language choice. The
   default is "Off".
9. **Web app: translation under the post.** `registerPostMessageAttachmentComponent`
   fetches and shows the translation, with a "Translated from Dutch" label and a
   "Hide" toggle. No request when the user has no language or wrote the post.
10. **Integration test.** With `MM_TEST_IMAGE=mattermost/mattermost-team-edition:10.12.0`:
    set a preference, create a post, call the endpoint, and check the result. Also check
    that a user outside the channel gets `403`.
11. **Release.** `README.md` with install and settings steps, then a `v0.1.0` tag. The
    `ci.yml` workflow builds the release bundle.

## Production install

Production is mattermost.hedgehogapplications.nl, Team Edition 10.12.0, with plugin
uploads turned on. The upload is outward-facing: it waits for its own explicit go, after
commit 11. Dev work never points at production.
