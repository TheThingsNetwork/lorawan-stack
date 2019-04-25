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
include .make/general.make
include .make/git.make
include .make/dev.make

docs:
	@rm -f doc/ttn-lw-{stack,cli}/*.{md,1,yaml}
	@$(GO) run ./cmd/ttn-lw-stack gen-man-pages --log.level=error -o doc/ttn-lw-stack
	@$(GO) run ./cmd/ttn-lw-stack gen-md-doc --log.level=error -o doc/ttn-lw-stack
	@$(GO) run ./cmd/ttn-lw-stack gen-yaml-doc --log.level=error -o doc/ttn-lw-stack
	@$(GO) run ./cmd/ttn-lw-cli gen-man-pages --log.level=error -o doc/ttn-lw-cli
	@$(GO) run ./cmd/ttn-lw-cli gen-md-doc --log.level=error -o doc/ttn-lw-cli
	@$(GO) run ./cmd/ttn-lw-cli gen-yaml-doc --log.level=error -o doc/ttn-lw-cli

dev-deps: go.deps

deps: go.deps

test: go.test

quality: go.quality

build-all: $(MAGE)
	@GO111MODULE=on $(GO) run github.com/goreleaser/goreleaser --snapshot --skip-publish

clean: js.clean
	rm -rf dist

translations: messages
