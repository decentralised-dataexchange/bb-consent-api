PROJECT := igrant
APP     := api
NAME    = $(PROJECT)-$(APP)

PROJECT_PACKAGE         := github.com/igrant/$(APP)
PKG_LIST_CMD = go list ./... | grep -v '/vendor/\|/mocks/\|/tmp/'
SOURCE_FILES = $(shell /usr/bin/find . -type f -name '*.go' -not \( -path './vendor/*' -or -path './mocks/*' -or -path './tmp/*' -or -path './resources/*' \))

TERM_FLAGS ?= -ti

EXTRA_RUN_ARGS ?=

PKGS = $(shell $(PKG_LIST_CMD))

VERSION   ?= $(shell git describe --tags --abbrev=0)
CANDIDATE ?= "dev"
CONTAINER_API ?= "igrant_api_dev"
DB_CONTAINER_NAME = "mongo"
KAFKA_BROKER_CONTAINER_NAME = "broker"

CONTAINER_DEFAULT_RUN_FLAGS := \
	--rm $(TERM_FLAGS) \
	$(EXTRA_RUN_ARGS) \
	--env GOOGLE_APPLICATION_CREDENTIALS=/opt/igrant/api/kubernetes-config/keyfile.json \
	-v "$(CURDIR)":/go/src/$(PROJECT_PACKAGE) \
	-v $(CURDIR)/resources/config/:/opt/igrant/api/config/:ro \
	-v $(CURDIR)/resources/kubernetes-config/:/opt/igrant/api/kubernetes-config/:ro \
	-w /go/src/$(PROJECT_PACKAGE)

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD | sed -E 's/[^a-zA-Z0-9]+/-/g')
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# jenkins specific
ifdef BRANCH_NAME
    GIT_BRANCH = $(shell echo $(BRANCH_NAME) | tr '[:upper:]' '[:lower:]' | tr -cd '[[:alnum:]]_-')
endif

DEPLOY_VERSION_FILE = ./deploy_version
DEPLOY_VERSION = $(shell test -f $(DEPLOY_VERSION_FILE) && cat $(DEPLOY_VERSION_FILE))

GCLOUD_HOSTNAME = eu.gcr.io
GCLOUD_PROJECTID = jenkins-189019
DOCKER_IMAGE := ${GCLOUD_HOSTNAME}/${GCLOUD_PROJECTID}/$(NAME)

# tag based on git branch, date and commit
DOCKER_TAG := $(GIT_BRANCH)-$(shell date +%Y%m%d%H%M%S)-$(GIT_COMMIT)

DIST_FILE := dist/linux_amd64/$(NAME)

UNAME := $(shell uname -m)

.DEFAULT_GOAL := help
.PHONY: help
help:
	@echo "------------------------------------------------------------------------"
	@echo "iGrant API"
	@echo "------------------------------------------------------------------------"
	@grep -E '^[0-9a-zA-Z_/%\-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.bootstrap:
	git clone git@github.com:L3-iGrant/bootstrap.git "$(CURDIR)/.bootstrap"
	git --git-dir=$(CURDIR)/.bootstrap/.git --work-tree=$(CURDIR)/.bootstrap checkout cors

.PHONY: bootstrap
bootstrap: .bootstrap ## Boostraps development environment
	git -C $(CURDIR)/.bootstrap fetch --all --prune
	@if [ -d $(CURDIR)/.bootstrap/scripts/docker-proxy/vhost.d ] ; then \
		sudo rm -rf $(CURDIR)/.bootstrap/scripts/docker-proxy/vhost.d; \
	fi
	git -C $(CURDIR)/.bootstrap reset --hard origin/cors
	make -C .bootstrap bootstrap

setup: bootstrap build/docker/builder ## Sets up development environment
	@$(CURDIR)/resources/scripts/setup-development-environment.sh

console: ## Runs bash shell, i.e. builder/console
	@bash

builder/%:: DOCKER_BIN=$(shell command -v docker 2>/dev/null || echo 'docker')
builder/%:: ## Runs make target in builder container (example: make builder/console)
	@if [ -x $(DOCKER_BIN) ] && $(DOCKER_BIN) info >/dev/null 2>&1 ; then \
		docker run \
			$(shell env | grep ^TRAVIS | cut -d= -f1  | awk '{print "-e", $$1}') \
			$(CONTAINER_DEFAULT_RUN_FLAGS) \
			$(DOCKER_IMAGE):builder \
			make $*; \
	else \
		make $*; \
	fi

api/build_debug: builder/_api/build_debug ## Builds API without optimizations
_api/build_debug: GOFLAGS := -gcflags "-N -l"
_api/build_debug: _api/build

api/build: builder/_api/build ## Builds API
_api/build:
	go build \
		-ldflags " \
			-X $(PROJECT_PACKAGE)/src/version.version=$(VERSION) \
			-X $(PROJECT_PACKAGE)/src/version.candidate=$(CANDIDATE) \
			-X $(PROJECT_PACKAGE)/src/version.gitCommit=$(GIT_COMMIT)" \
		$(GOFLAGS) \
		-o $(CURDIR)/bin/$(NAME) $(PROJECT_PACKAGE)/src/main

$(DIST_FILE): builder/_$(DIST_FILE) ## Builds deployable API executable (statically linked and optimized)
_$(DIST_FILE):
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
	go build \
		-ldflags " \
			-X $(PROJECT_PACKAGE)/src/version.version=$(VERSION) \
			-X $(PROJECT_PACKAGE)/src/version.candidate="release" \
			-X $(PROJECT_PACKAGE)/src/version.gitCommit=$(GIT_COMMIT)" \
		-o $(CURDIR)/$(DIST_FILE) $(PROJECT_PACKAGE)/src/main

api/run: ## Run API locally for development purposes
	docker run \
		$(CONTAINER_DEFAULT_RUN_FLAGS) \
		--expose 80 \
		-p 8080:80 \
		--link=${DB_CONTAINER_NAME} \
		-e VIRTUAL_HOST=$(APP).$(PROJECT).dev \
		--name "${CONTAINER_API}" \
		$(DOCKER_IMAGE):builder \
		$(BEFORE_ARGS) ./bin/$(NAME) -config config-development.json

api/run_with_kafka: ## Run API locally for development purposes connected to a kafka broker container
	docker run \
		$(CONTAINER_DEFAULT_RUN_FLAGS) \
		--expose 80 \
		-p 8080:80 \
		--link=${DB_CONTAINER_NAME} \
		--link=${KAFKA_BROKER_CONTAINER_NAME} \
		-e VIRTUAL_HOST=$(APP).$(PROJECT).dev \
		--name "${CONTAINER_API}" \
		$(DOCKER_IMAGE):builder \
		$(BEFORE_ARGS) ./bin/$(NAME) -config config-development.json

# go-stack causes SIGSEGV intentionally, so let's hide it, see
# https://github.com/go-stack/stack/issues/4
api/gdb: BEFORE_ARGS := gdb --quiet -ex "handle SIGSEGV nostop noprint" --args ## Debug API with GDB
api/gdb: api/run

clean: mock/clean
	rm -rf ./bin

# QA
qa: builder/test/static builder/test/unit ## Performs all QA checks

mock/clean: ## Remove mocks
	@/bin/bash -c 'find src -name "mock_*.go" -delete -o -name "mock.goconvey" -delete'
	@rm -rf ./mocks

mock/build: ## Build mocks
	#mkdir -p ./mocks/net/http
	#mockgen -destination=./mocks/net/http/http.go net/http RoundTripper,ResponseWriter
	#mkdir -p ./mocks/io
	#mockgen -destination=./mocks/io/io.go io ReadCloser
	resources/scripts/generate_mocks.sh "$(CURDIR)" "src" > /dev/null

test/unit: mock/build ## Run unit tests
	@go test -tags=mock -cover -race -v $(shell $(PKG_LIST_CMD))

test/unit/%: mock/build ## Run unit test for a specific dir (e.g. `lib/container`)
	@go test -tags=mock -cover -race -v "$(PROJECT_PACKAGE)/$*"

test/static: lint ## Run static analysis

lint: mock/build # Run static analysis on API code
	@go test -tags=linter resources/scripts/lint_test.go --args $(shell $(PKG_LIST_CMD))

format:
	@go fmt $(shell $(PKG_LIST_CMD))

build/docker/builder: ## Builds docker image containing dependency for building the project
	docker build --platform=linux/amd64 -t $(DOCKER_IMAGE):builder -f resources/docker/builder/Dockerfile .

.PHONY: build/docker/deployable
build/docker/deployable: $(DIST_FILE) ## Builds deployable docker image for preview, staging and production
	docker build --platform=linux/amd64 -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f resources/docker/production/Dockerfile .
	echo "$(DOCKER_IMAGE):$(DOCKER_TAG)" > $(DEPLOY_VERSION_FILE)

.PHONY: publish
publish: $(DEPLOY_VERSION_FILE) ## Publish latest production Docker image to docker hub
	gcloud docker -- push $(DEPLOY_VERSION)

deploy/production: $(DEPLOY_VERSION_FILE) ## Deploy to K8s cluster (e.g. make deploy/{preview,staging,production})
	kubectl set image deployment/igrant-api-demo igrant-api-demo=$(DEPLOY_VERSION) -n demo

deploy/staging: $(DEPLOY_VERSION_FILE) ## Deploy to K8s cluster (e.g. make deploy/{preview,staging,production})
	kubectl set image deployment/igrant-api-staging igrant-api-staging=$(DEPLOY_VERSION) -n staging

.PHONY: release
release:  ## Produces binaries needed for a release
	GOOS=linux GOARCH=amd64 CANDIDATE="" make builder/clean _api/build
	@mkdir -p dist
	mv bin/$(NAME) dist/$(NAME)-production
	GOOS=linux GOARCH=amd64 make builder/clean _api/build
	mv bin/$(NAME) dist/$(NAME)-staging

$(DEPLOY_VERSION_FILE):
	@echo "Missing '$(DEPLOY_VERSION_FILE)' file. Run 'make build/docker/deployable'" >&2
	exit 1

