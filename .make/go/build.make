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

# default main file
MAIN ?= ./main.go

# default vendor folder
VENDOR_DIR ?= $(PWD)/vendor
VENDOR_FILE ?= Gopkg.toml

LAZY_GOOS = `echo $@ | sed 's:$(RELEASE_DIR)/.*-\(.*\)-\(.*\):\1:'`
LAZY_GOARCH = `echo $@ | sed 's:$(RELEASE_DIR)/.*-\(.*\)-\(.*\):\2:'`
LAZY_GOEXE = $$(GOOS=$(LAZY_GOOS) go env GOEXE)

# Build the executable
$(RELEASE_DIR)/%: go.min-version $(shell $(GO_FILES)) $(GO_VENDOR_FILE)
	@$(log) "Building" [$(GO_ENV) GOOS=$(LAZY_GOOS) GOARCH=$(LAZY_GOARCH) $(GO) build $(GO_FLAGS) ...] to "$@$(LAZY_GOEXE)"
	@$(GO_ENV) GOOS=$(LAZY_GOOS) GOARCH=$(LAZY_GOARCH) $(GO) build -gcflags="all=-trimpath=$(GO_PATH)" -asmflags="all=-trimpath=$(GO_PATH)" -o "$@$(LAZY_GOEXE)" -v $(GO_FLAGS) $(LD_FLAGS) $(MAIN)

# link executables to a simplified name that is the same on all architectures.
go.link:
	@for i in $(wildcard $(RELEASE_DIR)/*-$(GOOS)-$(GOARCH)); do \
		ln -sfr $$i `echo $$i | sed 's:\(.*\)-.*-.*:\1:'`; \
	done

## initialize go dep
$(VENDOR_FILE):
	@$(log) "Initializing go deps"
	@mkdir -p $(VENDOR_DIR) && cd $(VENDOR_DIR)/.. && dep init
