ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
# Compose file lives under deploy/; project-directory keeps volume paths correct
DOCKER_COMPOSE = cd "$(ROOT)" && docker compose -f deploy/docker-compose.yml --project-directory deploy

.PHONY: preview-isr release \
	docker-restart-all docker-restart \
	docker-restart-nginx docker-restart-server docker-restart-renderer \
	docker-restart-postgres docker-restart-redis \
	docker-rebuild-all docker-rebuild \
	docker-rebuild-server docker-rebuild-renderer

preview-isr:
	@bash ./scripts/preview-isr.sh

release:
ifndef VERSION
	$(error VERSION is required, e.g. make release VERSION=v1.2.3 [PUSH=1])
endif
ifeq ($(PUSH),1)
	@bash ./scripts/release.sh $(VERSION) --push
else
	@bash ./scripts/release.sh $(VERSION)
endif

# Restart every service (server, renderer, nginx, postgres, redis)
docker-restart-all:
	$(DOCKER_COMPOSE) restart

# Restart one service: make docker-restart SERVICE=nginx
# Services: server | renderer | nginx | postgres | redis
docker-restart:
ifndef SERVICE
	$(error SERVICE is required, e.g. make docker-restart SERVICE=nginx. Or: make docker-restart-all)
endif
	$(DOCKER_COMPOSE) restart $(SERVICE)

docker-restart-nginx:
	$(DOCKER_COMPOSE) restart nginx

docker-restart-server:
	$(DOCKER_COMPOSE) restart server

docker-restart-renderer:
	$(DOCKER_COMPOSE) restart renderer

docker-restart-postgres:
	$(DOCKER_COMPOSE) restart postgres

docker-restart-redis:
	$(DOCKER_COMPOSE) restart redis

# Full image rebuild (no cache) + recreate containers. Rebuilds all services that define `build:` (server, renderer).
docker-rebuild-all:
	$(DOCKER_COMPOSE) build --no-cache
	$(DOCKER_COMPOSE) up -d --force-recreate

# Rebuild one service: make docker-rebuild SERVICE=server
docker-rebuild:
ifndef SERVICE
	$(error SERVICE is required, e.g. make docker-rebuild SERVICE=server. Or: make docker-rebuild-all)
endif
	$(DOCKER_COMPOSE) build --no-cache $(SERVICE)
	$(DOCKER_COMPOSE) up -d --force-recreate $(SERVICE)

docker-rebuild-server:
	$(DOCKER_COMPOSE) build --no-cache server
	$(DOCKER_COMPOSE) up -d --force-recreate server

docker-rebuild-renderer:
	$(DOCKER_COMPOSE) build --no-cache renderer
	$(DOCKER_COMPOSE) up -d --force-recreate renderer
