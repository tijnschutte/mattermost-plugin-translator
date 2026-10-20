# Include custom targets and environment variables here

## Fast check while iterating: go vet, golangci-lint, server unit tests, webapp lint, types, and tests. No Docker.
.PHONY: quick
quick: apply webapp/node_modules
	$(GO) vet ./...
	$(GOBIN)/golangci-lint run ./...
	SKIP_DOCKER_TESTS=1 $(GO) test ./server/...
	cd webapp && npm run lint && npm run check-types && npx jest

# Integration tests run against the Mattermost version production runs.
export MM_TEST_IMAGE ?= mattermost/mattermost-team-edition:10.12.0

# make deploy and make watch target the local Mattermost from dev/ unless you override these.
export MM_SERVICESETTINGS_SITEURL ?= http://localhost:8065
export MM_ADMIN_USERNAME ?= admin
export MM_ADMIN_PASSWORD ?= Admin-dev-1!

## Starts the local Mattermost 10.12.0 on http://localhost:8065 and seeds a team and two users.
.PHONY: dev-up
dev-up:
	docker compose -f dev/docker-compose.yml up -d
	./dev/seed.sh

## Stops the local Mattermost. Add DEV_RESET=1 to delete its data too.
.PHONY: dev-down
dev-down:
	docker compose -f dev/docker-compose.yml down $(if $(DEV_RESET),--volumes)
