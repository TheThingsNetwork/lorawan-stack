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

.PHONY: default
default: all

include .mage/mage.make
include .make/dev.make

.PHONY: all
all: init
	$(MAKE) $(BUILD_RULES)

.PHONY: init
init:
	$(MAKE) $(INIT_RULES)

.PHONY: deps
deps:
	$(MAKE) $(DEPS_RULES)

.PHONY: test
test:
	$(MAKE) $(TEST_RULES)

.PHONY: quality
quality:
	$(MAKE) $(QUALITY_RULES)

.PHONY: clean
clean:
	$(MAKE) $(CLEAN_RULES)

.PHONY: build
build:
	@echo "For development:"
	@echo "  run \"make $(BUILD_RULES)\""
	@echo "  and then \"go run ./cmd/ttn-lw-stack\" or \"go run ./cmd/ttn-lw-cli\"."
	@echo "For production:"
	@echo "  run \"make build-all\""

.PHONY: build-all
build-all:
	@GO111MODULE=on $(GO) run github.com/goreleaser/goreleaser --snapshot --skip-publish

.PHONY: docs
docs:
	@rm -f doc/ttn-lw-{stack,cli}/*.{md,1,yaml}
	@$(GO) run ./cmd/ttn-lw-stack gen-man-pages --log.level=error -o doc/ttn-lw-stack
	@$(GO) run ./cmd/ttn-lw-stack gen-md-doc --log.level=error -o doc/ttn-lw-stack
	@$(GO) run ./cmd/ttn-lw-stack gen-yaml-doc --log.level=error -o doc/ttn-lw-stack
	@$(GO) run ./cmd/ttn-lw-cli gen-man-pages --log.level=error -o doc/ttn-lw-cli
	@$(GO) run ./cmd/ttn-lw-cli gen-md-doc --log.level=error -o doc/ttn-lw-cli
	@$(GO) run ./cmd/ttn-lw-cli gen-yaml-doc --log.level=error -o doc/ttn-lw-cli

.PHONY: git.diff
git.diff:
	@if [[ ! -z "`git diff`" ]]; then \
		$(err) "Previous operations have created changes that were not recorded in the repository. Please make those changes on your local machine before pushing them to the repository:"; \
		git diff; \
		exit 1; \
	fi
