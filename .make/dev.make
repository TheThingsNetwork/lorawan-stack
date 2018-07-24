# Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This makefile contains utilities for development purposes.

DEV_DB_DATA_PATH ?= $(PWD)/.dev/databases
DEV_COCKROACH_DATA_PATH ?= $(DEV_DB_DATA_PATH)/cockroach-data
DEV_REDIS_DATA_PATH ?= $(DEV_DB_DATA_PATH)/redis

# Path to the Docker binary to be used
DOCKER_BINARY ?= docker

DEV_COCKROACH_IMAGE = cockroachdb/cockroach:v2.0.3
DEV_REDIS_IMAGE = redis:4.0-alpine

db_docker_prefix = ttn-lw-devdb
cockroach_docker_name = $(db_docker_prefix)-cockroach
redis_docker_name = $(db_docker_prefix)-redis

# Save data on disk by creating a volume if there's a DEV_*_DATA_PATH set
ifdef DEV_COCKROACH_DATA_PATH
	cockroach_docker_volumes = -v "$(DEV_COCKROACH_DATA_PATH):/cockroach/cockroach-data"
endif
ifdef DEV_REDIS_DATA_PATH
	redis_docker_volumes = -v "$(DEV_REDIS_DATA_PATH):/data"
	redis_command = redis-server --appendonly yes
endif

# Certificates

dev.certs:
	@$(log) "Generating certificates"
	@go run $(shell go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost

INIT_RULES += dev.certs

# Docker

dev.docker.installed:
	@if [[ -z "$$(which $(DOCKER_BINARY))" ]]; then \
		$(err) "Could not find Docker binary. Please install Docker or specify the location of the Docker binary with DOCKER_BINARY."; \
		exit 1; \
	fi

# All databases

dev.databases.start: dev.cockroach.start dev.redis.start

dev.databases.stop: dev.cockroach.stop dev.redis.stop

dev.databases.erase: dev.cockroach.erase dev.redis.erase

# Cockroach

dev.cockroach.start: dev.docker.installed dev.cockroach.remove-container
	@$(log) "Start Cockroach container as $(cockroach_docker_name)"
	@$(log) "WebUI exposed on http://localhost:26256"
	@if [[ ! -z "$(DEV_COCKROACH_DATA_PATH)" ]]; then mkdir -p $(DEV_COCKROACH_DATA_PATH); fi
	@$(DOCKER_BINARY) run -d -p 127.0.0.1:26257:26257 -p 127.0.0.1:26256:26256 --name $(cockroach_docker_name) $(cockroach_docker_volumes) $(DEV_COCKROACH_IMAGE) start --http-port 26256 --insecure > /dev/null

dev.cockroach.stop: dev.docker.installed
	@$(DOCKER_BINARY) kill $(cockroach_docker_name) > /dev/null 2> /dev/null && $(log) "Cockroach container killed" || $(warn) "Cockroach container was not killed"
	@$(DOCKER_BINARY) rm $(cockroach_docker_name) > /dev/null && $(log) "Cockroach container removed" || $(warn) "Cockroach container was not removed"

dev.cockroach.erase: dev.docker.installed
	@if [[ "$$($(DOCKER_BINARY) ps)" =~ "$(cockroach_docker_name)" ]]; then \
		$(err) "Cockroach container is still running, aborted erase operation."; \
	else \
		rm -rf "$(DEV_COCKROACH_DATA_PATH)"; \
	fi

dev.cockroach.sql: dev.docker.installed
	@$(log) "Opening a Cockroach shell"
	@$(DOCKER_BINARY) exec -it $(cockroach_docker_name) ./cockroach sql --insecure

dev.cockroach.drop: dev.docker.installed
	@if [[ -z "$(NAME)" ]]; then \
		$(err) "No NAME specified for the database to drop."; \
		exit 1; \
	fi
	@$(log) "Dropping $(NAME) Cockroach database"
	@$(DOCKER_BINARY) exec $(cockroach_docker_name) ./cockroach sql --insecure --execute="DROP DATABASE $(NAME) CASCADE;"

dev.cockroach.remove-container:
	@if [[ `$(DOCKER_BINARY) container ls -a` =~ $(cockroach_docker_name) ]]; then \
		$(log) "Removing old Cockroach container $(cockroach_docker_name)"; \
		$(DOCKER_BINARY) rm $(cockroach_docker_name) > /dev/null; \
	fi

# Redis

dev.redis.start: dev.docker.installed dev.redis.remove-container
	@$(log) "Start Redis container as $(redis_docker_name)"
	@if [[ ! -z "$(DEV_REDIS_DATA_PATH)" ]]; then mkdir -p $(DEV_REDIS_DATA_PATH); fi
	@$(DOCKER_BINARY) run -d -p 127.0.0.1:6379:6379 --name $(redis_docker_name) $(redis_docker_volumes) $(DEV_REDIS_IMAGE) $(redis_command) > /dev/null

dev.redis.stop: dev.docker.installed
	@$(DOCKER_BINARY) kill $(redis_docker_name) > /dev/null 2> /dev/null && $(log) "Redis container killed" || $(warn) "Redis container was not killed"
	@$(DOCKER_BINARY) rm $(redis_docker_name) > /dev/null && $(log) "Redis container removed" || $(warn) "Redis container was not removed"

dev.redis.erase: dev.docker.installed
	@if [[ "$$($(DOCKER_BINARY) ps)" =~ "$(redis_docker_name)" ]]; then \
		$(err) "Redis container is still running, aborted erase operation."; \
	else \
		rm -rf "$(DEV_REDIS_DATA_PATH)"; \
	fi

dev.redis.remove-container:
	@if [[ `$(DOCKER_BINARY) container ls -a` =~ $(redis_docker_name) ]]; then \
		$(log) "Removing old Redis container $(redis_docker_name)"; \
		$(DOCKER_BINARY) rm $(redis_docker_name); \
	fi

dev.git-diff:
	@if [[ ! -z "`git diff`" ]]; then \
		$(err) "Previous operations have created changes that were not recorded in the repository. Please make those changes on your local machine before pushing them to the repository:"; \
		git diff; \
		exit 1; \
	fi

# Binaries

dev.ttn-lw-stack.start: ttn-lw-stack
	./release/ttn-lw-stack-$(GOOS)-$(GOARCH) start --log.level=debug

dev.ttn-lw-stack.start-dev: ttn-lw-stack
	./release/ttn-lw-stack-$(GOOS)-$(GOARCH) start --log.level=debug \
	--console.public-url=http://localhost:8080/console \
	--is.oauth.public-url=http://localhost:8080/oauth \
	--console.oauth-url=http://localhost:8080/oauth

dev.ttn-lw-identity-server.start: ttn-lw-identity-server
	./release/ttn-lw-identity-server-$(GOOS)-$(GOARCH) start --log.level=debug

dev.ttn-lw-identity-server.init: ttn-lw-identity-server
	./release/ttn-lw-identity-server-$(GOOS)-$(GOARCH) init

dev.ttn-lw-identity-server.init-dev: ttn-lw-identity-server
	./release/ttn-lw-identity-server-$(GOOS)-$(GOARCH) init \
	--initial-data.console.skip-authorization=false \
	--initial-data.console.redirect-uri=http://localhost:8080/console/oauth/callback

dev.ttn-lw-gateway-server.start: ttn-lw-gateway-server
	./release/ttn-lw-gateway-server-$(GOOS)-$(GOARCH) start --log.level=debug

dev.ttn-lw-network-server.start: ttn-lw-network-server
	./release/ttn-lw-network-server-$(GOOS)-$(GOARCH) start --log.level=debug

dev.ttn-lw-application-server.start: ttn-lw-application-server
	./release/ttn-lw-application-server-$(GOOS)-$(GOARCH) start --log.level=debug

dev.ttn-lw-join-server.start: ttn-lw-join-server
	./release/ttn-lw-join-server-$(GOOS)-$(GOARCH) start --log.level=debug

dev.ttn-lw-console.start: ttn-lw-console
	./release/ttn-lw-console-$(GOOS)-$(GOARCH) start --log.level=debug

# vim: ft=make
