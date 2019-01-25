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

include .mage/mage.make

HEADER_EXTRA_FILES = Makefile

PRE_COMMIT = headers.check-staged js.lint-staged styl.lint-staged snap.lint-staged
COMMIT_MSG = git.commit-msg-log git.commit-msg-length git.commit-msg-empty git.commit-msg-prefix git.commit-msg-phrase git.commit-msg-casing git.commit-msg-imperative

SUPPORT_LOCALES = en

include .make/log.make
include .make/general.make
include .make/git.make
include .make/headers.make
include .make/go/main.make
include .make/protos/main.make
include .make/js/main.make
include .make/dev.make
include .make/ci.make
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
ttn-lw-stack: $(MAGE) js.build
	@GO_BINARIES=stack $(MAGE) go:build

# cli binary
ttn-lw-cli: $(MAGE)
	@GO_BINARIES=cli $(MAGE) go:build

# All binaries
build-all: $(MAGE) js.build
	@GO_BINARIES="" $(MAGE) go:build

# All supported platforms
build-all-platforms:
	@GO_BINARIES="" $(MAGE) go:buildCrossPlatform

DOCKER_IMAGE ?= thethingsnetwork/lorawan-stack

docker:
	GOOS=linux GOARCH=amd64 make build-all
	docker build -t $(DOCKER_IMAGE) .

translations: messages js.translations
