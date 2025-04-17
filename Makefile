ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

GIT_SHA := $(shell git rev-parse --short HEAD)
GIT_REPO := $(shell basename -s .git `git config --get remote.origin.url` | sed 's/.*://g')
SERVICES := user-api

.PHONY: dev emulator lint test cover build push deploy

dev:
	ifndef SERVICE
		@echo "Usage: make dev SERVICE=<service-name>"
		@echo "Example: make dev SERVICE=user-api"
		@exit 1
	endif
		@go mod tidy
		@go run cmd/$(SERVICE)/main.go


lint:
	@golangci-lint run --timeout 5m
	@hadolint Dockerfile

test:
	@go test -v ./...

build: lint
	@for service in $(SERVICES) ; do \
		@docker build --platform linux/amd64 --build-arg SRC_PATH=$(GIT_REPO) -t $(DOCKER_REPO)/$$service .
		@docker tag $(DOCKER_REPO)/$$service:latest $(DOCKER_REPO)/$$service:$(GIT_SHA)
	done

push:
	@docker push $(DOCKER_REPO)/$(SERVICE_NAME):latest
	@docker push $(DOCKER_REPO)/$(SERVICE_NAME):$(GIT_SHA)

deploy:
	@gcloud run deploy $(SERVICE_NAME) \
		--image $(DOCKER_REPO)/$(SERVICE_NAME):$(GIT_SHA) \
		--platform managed \
		--region $(GCP_REGION) \
		--allow-unauthenticated \
		--project $(GCP_PROJECT_ID) \
		--set-env-vars GIT_SHA=$(GIT_SHA) \
		--concurrency 20 \
		--cpu 1 \
		--memory 128Mi