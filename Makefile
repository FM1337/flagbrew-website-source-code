.DEFAULT_GOAL := build

BINARY=flagbrew
DIRS=bin

VERSION=$(shell git describe --tags --always --abbrev=0 --match=v* 2> /dev/null | sed -r "s:^v::g" || echo 0)
VERSION_FULL=$(shell git describe --tags --always --dirty --match=v* 2> /dev/null | sed -r "s:^v::g" || echo 0)

$(info $(shell mkdir -p $(DIRS)))
BIN=$(CURDIR)/bin
export GOBIN=$(CURDIR)/bin

debug-go: fetch-go generate-go
	go run *.go --http "0.0.0.0:8081" --debug

debug-frontend: fetch-node ## Generate public html/css/js when files change (faster, but larger files). Also spins up go server.
	/bin/rm -rfv "public/dist"
	cd public && echo BUILD_DATE=$$(date) > .env && npm run watch

debug-cypress: ## Launches cypress
	cd public && npm run cy:open

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

fetch-go: ## Fetches the necessary Go dependencies to build.
	go mod download
	go mod tidy
	go mod vendor

fetch-node: ## Fetches the necessary NodeJS dependencies to build.
	test -d public/node_modules || (cd public && npm install)

upgrade-deps: ## Upgrade all dependencies to the latest version.
	go get -u ./...

upgrade-deps-patch: ## Upgrade all dependencies to the latest patch release.
	go get -u=patch ./...

clean: ## Cleans up generated files/folders from the build.
	/bin/rm -rfv "public/dist" "${BINARY}"

clean-cache: ## Cleans up generated cache (speeds up during dev time).
	/bin/rm -rfv "public/.cache"

generate-watch: ## Generate public html/css/js when files change (faster, but larger files.)
	cd public && npm run watch

generate-node: ## Generate public html/css/js files for use in production (slower, smaller/minified files.)
	cd public && echo BUILD_DATE=$$(date) > .env && npm run build

generate-go: ## Generate go bundled files from frontend
	go generate -x ./...

compile:
	go build -ldflags '-s -w' -tags netgo -installsuffix netgo -v -o "${BINARY}"

build: fetch-go fetch-node clean clean-cache generate-node generate-go compile ## Builds the application (with generate.)
	echo