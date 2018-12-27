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

HEADER_EXTRA_FILES = Makefile

PRE_COMMIT = headers.check-staged js.lint-staged styl.lint-staged snap.lint-staged
COMMIT_MSG = git.commit-msg-log git.commit-msg-length git.commit-msg-empty git.commit-msg-prefix git.commit-msg-phrase git.commit-msg-casing git.commit-msg-imperative

SUPPORT_LOCALES = en

include .make/log.make
include .make/general.make
include .make/git.make
include .make/versions.make
include .make/headers.make
include .make/go/main.make
include .make/protos/main.make
include .make/js/main.make
include .make/dev.make
include .make/styl/main.make
include .make/snap/main.make
include .make/sdk/main.make

messages:
	@$(GO) run ./cmd/internal/generate_i18n.go

dev-deps: go.dev-deps js.dev-deps

deps: go.deps js.deps sdk.deps

test: go.test js.test sdk.test

quality: go.quality js.quality styl.quality snap.quality

clean: go.clean js.clean

# stack binary
ttn-lw-stack: MAIN=./cmd/ttn-lw-stack/main.go
ttn-lw-stack: js.build
ttn-lw-stack: $(RELEASE_DIR)/ttn-lw-stack-$(GOOS)-$(GOARCH)

# identity-server binary
ttn-lw-identity-server: MAIN=./cmd/ttn-lw-identity-server/main.go
ttn-lw-identity-server: $(RELEASE_DIR)/ttn-lw-identity-server-$(GOOS)-$(GOARCH)

# gateway-server binary
ttn-lw-gateway-server: MAIN=./cmd/ttn-lw-gateway-server/main.go
ttn-lw-gateway-server: $(RELEASE_DIR)/ttn-lw-gateway-server-$(GOOS)-$(GOARCH)

# network-server binary
ttn-lw-network-server: MAIN=./cmd/ttn-lw-network-server/main.go
ttn-lw-network-server: $(RELEASE_DIR)/ttn-lw-network-server-$(GOOS)-$(GOARCH)

# application-server binary
ttn-lw-application-server: MAIN=./cmd/ttn-lw-application-server/main.go
ttn-lw-application-server: $(RELEASE_DIR)/ttn-lw-application-server-$(GOOS)-$(GOARCH)

# join-server binary
ttn-lw-join-server: MAIN=./cmd/ttn-lw-join-server/main.go
ttn-lw-join-server: $(RELEASE_DIR)/ttn-lw-join-server-$(GOOS)-$(GOARCH)

# console binary
ttn-lw-console: MAIN=./cmd/ttn-lw-console/main.go
ttn-lw-console: js.build
ttn-lw-console: $(RELEASE_DIR)/ttn-lw-console-$(GOOS)-$(GOARCH)

# cli binary
ttn-lw-cli: MAIN=./cmd/ttn-lw-cli/main.go
ttn-lw-cli: $(RELEASE_DIR)/ttn-lw-cli-$(GOOS)-$(GOARCH)

# All binaries
build-all: ttn-lw-stack ttn-lw-identity-server ttn-lw-gateway-server ttn-lw-network-server ttn-lw-application-server ttn-lw-join-server ttn-lw-console ttn-lw-cli

# All supported platforms
build-all-platforms:
	GOOS=linux GOARCH=amd64 make build-all
	GOOS=linux GOARCH=386 make build-all
	GOOS=linux GOARCH=arm make build-all
	GOOS=linux GOARCH=arm64 make build-all
	GOOS=darwin GOARCH=amd64 make build-all
	GOOS=windows GOARCH=amd64 make build-all
	GOOS=windows GOARCH=386 make build-all

DOCKER_IMAGE ?= thethingsnetwork/lorawan-stack

docker:
	GOOS=linux GOARCH=amd64 make build-all
	docker build -t $(DOCKER_IMAGE) .

translations: messages js.translations
