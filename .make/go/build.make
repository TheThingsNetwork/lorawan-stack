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

# Infer GOOS and GOARCH
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Default vendor folder
VENDOR_DIR ?= vendor

# Default main file
MAIN ?= ./main.go

# Build the executable
$(RELEASE_DIR)/%: go.min-version $(shell $(GO_FILES))
	@$(log) "Building" [$(GO_ENV) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GO_FLAGS) ...] to "$@$(LAZY_GOEXE)"
	@$(GO_ENV) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -gcflags="all=-trimpath=$(GO_PATH)" -asmflags="all=-trimpath=$(GO_PATH)" -o "$@$(LAZY_GOEXE)" -v $(GO_FLAGS) $(LD_FLAGS) $(MAIN)
