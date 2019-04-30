# Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

# Databases

DEV_DATABASES ?= cockroach redis
DEV_DATABASE_NAME ?= ttn_lorawan_dev
DEV_DATA_DIR ?= .dev/data
DEV_DOCKER_COMPOSE := DEV_DATA_DIR=$(DEV_DATA_DIR) DEV_DATABASE_NAME=$(DEV_DATABASE_NAME) docker-compose -p lorawan-stack-dev

dev.databases.start:
	@$(DEV_DOCKER_COMPOSE) up -d $(DEV_DATABASES)
	@$(DEV_DOCKER_COMPOSE) ps

dev.databases.stop:
	@$(DEV_DOCKER_COMPOSE) stop $(DEV_DATABASES)

dev.databases.erase: dev.databases.stop
	rm -rf $(DEV_DATA_DIR)

dev.databases.sql: dev.databases.start
	@$(DEV_DOCKER_COMPOSE) exec cockroach ./cockroach sql --insecure -d $(DEV_DATABASE_NAME)

dev.databases.redis-cli: dev.databases.start
	@$(DEV_DOCKER_COMPOSE) exec redis redis-cli

dev.stack.init: dev.databases.start
	@$(DEV_DOCKER_COMPOSE) run --rm stack is-db init
	@$(DEV_DOCKER_COMPOSE) run --rm stack is-db create-admin-user --id admin --email admin@localhost --password admin
	@$(DEV_DOCKER_COMPOSE) run --rm stack is-db create-oauth-client --id cli --name "Command Line Interface" --owner admin --no-secret --redirect-uri 'local-callback' --redirect-uri 'code'
	@$(DEV_DOCKER_COMPOSE) run --rm stack is-db create-oauth-client --id console --name "Console" --owner admin --secret console --redirect-uri 'https://localhost:8885/console/oauth/callback' --redirect-uri 'http://localhost:1885/console/oauth/callback' --redirect-uri '/console/oauth/callback'

.PHONY: git.diff
git.diff:
	@if [[ ! -z "`git diff`" ]]; then \
		echo "Previous operations have created changes that were not recorded in the repository. Please make those changes on your local machine before pushing them to the repository:"; \
		git diff; \
		exit 1; \
	fi

# vim: ft=make
