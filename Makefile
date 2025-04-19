ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

GIT_SHA := $(shell git rev-parse --short HEAD)
GIT_REPO := $(shell git remote get-url origin 2>/dev/null | sed 's/.*[/:]//;s/\.git$$//' || echo "local")

.PHONY: dev emulator lint test build push deploy deploy-without-sidecar infrastructure-apply infrastructure-plan

dev:
	@go mod tidy
	@air

lint:
	@golangci-lint run --timeout 5m
	@hadolint Dockerfile

test:
	@go test -v ./...


build: lint
	@docker build --platform linux/amd64 --build-arg SRC_PATH=$(GIT_REPO) -t $(DOCKER_BASE_PATH)/apis/$(SERVICE_NAME) .
	@docker tag $(DOCKER_BASE_PATH)/apis/$(SERVICE_NAME):latest $(DOCKER_BASE_PATH)/apis/$(SERVICE_NAME):$(GIT_SHA)
	@docker build --platform linux/amd64 -t $(DOCKER_BASE_PATH)/utils/otel . -f metrics.dockerfile

push:
	@docker push $(DOCKER_BASE_PATH)/apis/$(SERVICE_NAME):latest
	@docker push $(DOCKER_BASE_PATH)/apis/$(SERVICE_NAME):$(GIT_SHA)
	@docker push $(DOCKER_BASE_PATH)/utils/otel:latest

	@DIGEST=$$(docker inspect --format='{{index .RepoDigests 0}}' $(DOCKER_BASE_PATH)/apis/$(SERVICE_NAME):latest | awk -F'@' '{print $$2}'); \
		sed -i '' "s|digest *= *\".*\"|digest = \"$$DIGEST\"|" .infrastructure/variables.auto.tfvars; \
		terraform fmt ./.infrastructure/variables.auto.tfvars

deploy:
	@envsubst < service.yaml.tmpl > service.yaml
	@gcloud run services replace service.yaml --project=$(GCP_PROJECT_ID) --region=$(GCP_REGION) --quiet
	@sed -i '' "s|digest *= *\".*\"|digest         = \"$DIGEST\"|" .infrastructure/variables.auto.tfvars

infrastructure-plan:
	@sed -i '' "s|git_sha *= *\".*\"|git_sha = \"$(GIT_SHA)\"|" .infrastructure/variables.auto.tfvars
	@terraform fmt ./.infrastructure/variables.auto.tfvars
	@cd .infrastructure && \
	terraform init && \
	terraform plan -out=plan.tfplan

infrastructure-apply:
	@cd .infrastructure && \
	terraform init && \
	terraform apply plan.tfplan
