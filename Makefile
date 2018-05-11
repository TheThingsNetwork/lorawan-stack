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

PRE_COMMIT = headers.check-staged js.lint-staged
COMMIT_MSG = git.commit-msg-log git.commit-msg-length git.commit-msg-empty git.commit-msg-prefix git.commit-msg-phrase git.commit-msg-casing

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

ci.encrypt-variables:
	keybase encrypt -b -i ci/variables.yml -o ci/variables.yml.encrypted johanstokking htdvisser ericgo

ci.decrypt-variables:
	keybase decrypt -i ci/variables.yml.encrypted -o ci/variables.yml

messages:
	@$(GO) run ./pkg/errors/generate_messages.go --filename config/messages.json

dev-deps: go.dev-deps js.dev-deps

dev-cert:
	go run $$(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost

deps: go.deps js.deps

test: go.test js.test

quality: go.quality js.quality

# stack binary
ttn-lw-stack: MAIN=./cmd/ttn-lw-stack/main.go
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

# All binaries
build-all: ttn-lw-stack ttn-lw-identity-server ttn-lw-gateway-server ttn-lw-network-server ttn-lw-application-server ttn-lw-join-server

# All supported platforms
build-all-platforms:
	GOOS=linux GOARCH=amd64 make build-all
	GOOS=linux GOARCH=386 make build-all
	GOOS=linux GOARCH=arm make build-all
	GOOS=linux GOARCH=arm64 make build-all
	GOOS=darwin GOARCH=amd64 make build-all
	GOOS=windows GOARCH=amd64 make build-all
	GOOS=windows GOARCH=386 make build-all
