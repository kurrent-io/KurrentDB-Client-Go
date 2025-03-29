.DEFAULT_GOAL := help

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

OS := $(shell uname)

GENERATE_PROTOS_FLAG=
EVENTSTORE_DOCKER_REGISTRY ?= docker.eventstore.com/eventstore-ce
EVENTSTORE_DOCKER_IMAGE ?= eventstoredb-ce
EVENTSTORE_DOCKER_TAG ?= latest

.PHONY: build
build: ## Build based on the OS.
ifeq ($(OS),Linux)
	./build.sh $(GENERATE_PROTOS_FLAG)
else ifeq ($(OS),Darwin)
	./build.sh $(GENERATE_PROTOS_FLAG)
else
	pwsh.exe -File ".\build.ps1" $(GENERATE_PROTOS_FLAG)
endif

.PHONY: generate-protos-and-build
generate-protos-and-build: ## Generate protobuf and gRPC files while building.
ifeq ($(OS),Linux)
	$(MAKE) build GENERATE_PROTOS_FLAG=--generate-protos
else ifeq ($(OS),Darwin)
	$(MAKE) build GENERATE_PROTOS_FLAG=--generate-protos
else
	$(MAKE) build GENERATE_PROTOS_FLAG=-generateProtos
endif

.PHONY: start-kurrentdb
start-kurrentdb:
	@docker --version
	@docker compose up -d
	@echo "Waiting for containers to be healthy..."
	@until [ "$$(docker inspect -f '{{.State.Health.Status}}' $$(docker ps -q) | grep -c healthy)" -eq "$$(docker ps -q | wc -l)" ]; do \
		printf "."; \
		sleep 2; \
	done; \
	echo ""; \
	echo "✅ All containers are healthy!"
	@docker compose ps

.PHONY: stop-kurrentdb
stop-kurrentdb:
	@docker compose down -v --remove-orphans

.PHONY: test
test: ## Run tests
	go test --count=1 -v ./test
