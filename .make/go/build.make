# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# Infer GOOS and GOARCH
GOOS ?= `go env GOOS`
GOARCH ?= `go env GOARCH`

# default main file
MAIN ?= ./main.go

BUILD_TYPE = production

# default vendor folder
VENDOR_DIR ?= $(PWD)/vendor
VENDOR_FILE ?= Gopkg.toml

LAZY_GOOS=`echo $@ | sed 's:$(RELEASE_DIR)/.*-\(.*\)-\(.*\):\1:'`
LAZY_GOARCH=`echo $@ | sed 's:$(RELEASE_DIR)/.*-\(.*\)-\(.*\):\2:'`

# Build the executable
$(RELEASE_DIR)/%: $(shell $(GO_FILES)) $(GO_VENDOR_FILE)
	@$(log) "Building" [BUILD_TYPE=$(BUILD_TYPE) $(GO_ENV) GOOS=$(LAZY_GOOS) GOARCH=$(LAZY_GOARCH) $(GO) build $(GO_FLAGS) ...]
	@$(GO_ENV) GOOS=$(LAZY_GOOS) GOARCH=$(LAZY_GOARCH) $(GO) build -o "$@" -v $(GO_FLAGS) $(LD_FLAGS) $(MAIN)

# Enable development mode
.PHONY: go.dev
go.dev:
	$(eval GO_ENV := )
	$(eval GO_FLAGS := )
	$(eval BUILD_TYPE := development)

# link executables to a simplified name that is the same on all architectures.
go.link:
	@for i in $(wildcard $(RELEASE_DIR)/*-$(shell echo $(GOOS))-$(shell echo $(GOARCH))); do \
		ln -sfr $$i `echo $$i | sed 's:\(.*\)-.*-.*:\1:'`; \
	done

## initialize go dep
$(VENDOR_FILE):
	@$(log) "Initializing go deps"
	@mkdir -p $(VENDOR_DIR) && cd $(VENDOR_DIR)/.. && dep init

# vim: ft=make
