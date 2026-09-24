ifeq ($(OS),Windows_NT)
  BASH_EXE := $(shell where bash.exe 2>NUL)
  ifeq ($(strip $(BASH_EXE)),)
    $(warning bash.exe not found in PATH.)
    $(warning Install Git for Windows and ensure bash is in PATH, or use WSL.)
    $(error bash is required to run this Makefile on Windows)
  endif
  SHELL := $(firstword $(BASH_EXE))
else
  SHELL := /bin/bash
endif

.DEFAULT_GOAL := help

ENV ?= .env

COMPOSE = docker compose --env-file $(ENV) -f docker-compose.yml

SERVICE ?=
LOG_TAIL ?= 120

.PHONY: help init \
    build up down destroy restart ps logs health \
    test test-gateway test-core check check-compose proto service \
    backup verify-backup verify-backup-full restore-db-all \
    bundle restore-bundle \
    k8s-cluster k8s-deploy k8s-status k8s-undeploy \
    _check-env

BACKEND_SERVICES = gateway core-service

help: ## Show commands
	@printf '\nAvailable commands\n\n'
	@awk 'BEGIN {FS = ":.*## "}; /^[A-Za-z0-9_. -]+:.*## / {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf '\n'

_check-env:
	@if [ ! -f $(ENV) ]; then printf "\033[31mError: $(ENV) not found. Run 'make init'.\033[0m\n"; exit 1; fi
	@bash ./scripts/env-check.sh $(ENV)

init: ## Create root local environment file
	@if [ ! -f $(ENV) ]; then cp .env.example $(ENV); echo "Created $(ENV)"; else echo "$(ENV) already exists"; fi

build: _check-env ## Build all service images
	$(COMPOSE) build $(BACKEND_SERVICES)

up: _check-env ## Start stack or SERVICE=name (waits until healthy)
	$(COMPOSE) up -d --remove-orphans --wait --wait-timeout 180 $(SERVICE)

down: _check-env ## Stop stack
	$(COMPOSE) down

destroy: _check-env ## Stop stack and remove volumes
	$(COMPOSE) down -v

restart: _check-env ## Restart stack or SERVICE=name
	$(COMPOSE) restart $(SERVICE)

ps: _check-env ## Show stack status
	$(COMPOSE) ps

logs: _check-env ## Tail logs, optionally SERVICE=name
	$(COMPOSE) logs -f --tail=$(LOG_TAIL) $(SERVICE)

health: _check-env ## Check containers and public endpoints
	@bash ./scripts/health-check.sh "$(ENV)"

test: test-gateway test-core ## Run all service tests

test-gateway: ## Run gateway tests
	cd backend/gateway && uv run --frozen pytest

test-core: ## Run core-service tests
	cd backend/core-service && uv run --frozen pytest

check: check-compose test ## Run project checks

check-compose: _check-env ## Validate compose config
	$(COMPOSE) config --quiet

proto: ## Regenerate gRPC stubs from backend/proto
	@bash ./scripts/gen-proto.sh

service: ## Stamp a new service from templates, usage: make service NAME=orders
	@if [ -z "$(NAME)" ]; then echo "Error: set NAME=<service-name>"; exit 1; fi
	@bash ./scripts/new-service.sh $(NAME)

backup: ## Backup keycloak database
	@echo "=== Database Backup ==="
	@ENV_FILE="$(ENV)" bash ./scripts/db-backup.sh

verify-backup: ## Verify latest backup checksums and archive structure
	@echo "=== Backup Verification ==="
	@ENV_FILE="$(ENV)" bash ./scripts/backup-verify.sh

verify-backup-full: ## Restore latest backup into an isolated temporary database
	@echo "=== Full Isolated Restore Verification ==="
	@ENV_FILE="$(ENV)" bash ./scripts/backup-verify.sh --full

restore-db-all: ## Restore database by selecting backup date from list
	@ENV_FILE="$(ENV)" bash ./scripts/db-restore-all.sh

bundle: ## Pack a backup set into one portable file, usage: make bundle [RUN_ID=2026-09-24_12-00-00]
	@echo "=== Bundling backup set ==="
	@bash ./scripts/db-bundle.sh $(RUN_ID)

restore-bundle: ## Restore from a single bundle file, usage: make restore-bundle BUNDLE=backups/template-backup-<id>.tar.gz CONFIRM_RESTORE=yes
	@echo "=== Restore From Bundle ==="
	@if [ "$(CONFIRM_RESTORE)" != "yes" ]; then echo "Error: set CONFIRM_RESTORE=yes"; exit 1; fi
	@if [ -z "$(BUNDLE)" ]; then echo "Error: set BUNDLE=path/to/template-backup-<id>.tar.gz"; exit 1; fi
	@ENV_FILE="$(ENV)" CONFIRM_RESTORE=yes bash ./scripts/db-restore-bundle.sh "$(BUNDLE)"

# ==================== Kubernetes ====================

k8s-cluster: ## Create kind cluster from kind-config.yaml
	@kind create cluster --config kind-config.yaml

k8s-deploy: ## Deploy stack to Kubernetes (kind)
	@bash ./scripts/deploy-k8s.sh

k8s-status: ## Show Kubernetes resources status
	@bash ./scripts/status-k8s.sh

k8s-undeploy: ## Remove stack from Kubernetes
	@bash ./scripts/undeploy-k8s.sh

%:
	@printf "\e[31mCommand not found\e[0m"
	@$(MAKE) -s help
