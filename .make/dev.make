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

# Certificates

dev.certs:
	@$(log) "Generating certificates"
	@go run $(shell go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost

INIT_RULES += dev.certs

# Databases

DEV_DATABASES ?= cockroach redis
DEV_DATABASE_NAME ?= ttn_lorawan_dev
DEV_DATA_DIR ?= .dev/data
DEV_DOCKER_COMPOSE := DEV_DATA_DIR=$(DEV_DATA_DIR) DEV_DATABASE_NAME=$(DEV_DATABASE_NAME) docker-compose -p lorawan-stack-dev

dev.check-docker-compose:
	@command -v docker > /dev/null || ($(err) It looks like Docker is not installed)
	@command -v docker-compose > /dev/null || ($(err) It looks like Docker Compose is not installed)

dev.databases.start: dev.check-docker-compose
	@$(DEV_DOCKER_COMPOSE) up -d $(DEV_DATABASES)
	@$(DEV_DOCKER_COMPOSE) ps

dev.databases.stop: dev.check-docker-compose
	@$(DEV_DOCKER_COMPOSE) stop $(DEV_DATABASES)

dev.databases.erase: dev.databases.stop
	rm -rf $(DEV_DATA_DIR)

dev.databases.sql: dev.databases.start
	@$(DEV_DOCKER_COMPOSE) exec cockroach ./cockroach sql --insecure -d $(DEV_DATABASE_NAME)

dev.databases.redis-cli: dev.databases.start
	@$(DEV_DOCKER_COMPOSE) exec redis redis-cli

# Binaries

dev.stack.init: dev.databases.start
	go run ./cmd/ttn-lw-identity-server db init
	go run ./cmd/ttn-lw-identity-server create-admin-user --id admin --email admin@localhost --password admin
	go run ./cmd/ttn-lw-identity-server create-oauth-client --id cli --name "Command Line Interface" --owner admin --no-secret --redirect-uri 'http://localhost:11885/oauth/callback'
	go run ./cmd/ttn-lw-identity-server create-oauth-client --id console --name "Console" --owner admin --secret console --redirect-uri 'http://localhost:1885/console/oauth/callback'

# vim: ft=make
