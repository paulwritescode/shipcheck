# ShipCheck workflow commands.
# Backend is Go; the SPA lives under web/ with its own npm scripts.

.PHONY: all fmt vet build test test-pbt lint tidy secret-scan hooks web-install web-test ci

# Go package set for the backend (excludes web/ which holds the SPA and its
# node_modules, some of which ship stray Go files).
GO_PKGS := ./cmd/... ./internal/...

all: fmt vet build test

fmt:
	gofmt -w cmd internal

vet:
	go vet $(GO_PKGS)

build:
	go build $(GO_PKGS)

# Run the full Go test suite (unit + property-based tests).
test:
	go test $(GO_PKGS)

# Run only the property-based tests (the readiness-engine PBT deliverable).
test-pbt:
	go test ./internal/domain/... -run 'Property' -v

# gofmt check used in CI (fails if any file needs formatting).
lint:
	@test -z "$$(gofmt -l cmd internal)" || (echo 'gofmt needed:'; gofmt -l cmd internal; exit 1)

tidy:
	go mod tidy

secret-scan:
	bash scripts/secret-scan.sh --all

hooks:
	bash scripts/install-hooks.sh

web-install:
	cd web && npm install

web-test:
	cd web && npm test

# Build the Lambda bootstrap binary for deployment.
lambda:
	bash scripts/build-lambda.sh

# Build and deploy the full stack to AWS via CDK (needs AWS creds + CDK CLI).
deploy:
	bash scripts/deploy.sh

# Local CI parity: everything that CI runs for the backend.
ci: lint vet build test secret-scan
