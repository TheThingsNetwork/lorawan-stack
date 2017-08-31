# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# Infer GOOS and GOARCH
GOOS   ?= $(or $(word 1,$(subst -, ,${TARGET_PLATFORM})), $(shell echo "`go env GOOS`"))
GOARCH ?= $(or $(word 2,$(subst -, ,${TARGET_PLATFORM})), $(shell echo "`go env GOARCH`"))

# default main file
MAIN ?= ./main.go

# default vendor folder
VENDOR_DIR ?= $(PWD)/vendor
VENDOR_FILE ?= Gopkg.toml

LAZY_GOOS=`echo $@ | sed 's:-dev::' | sed 's:$(RELEASE_DIR)/.*-\(.*\)-\(.*\):\1:'`
LAZY_GOARCH=`echo $@ | sed 's:-dev::' | sed 's:$(RELEASE_DIR)/.*-\(.*\)-\(.*\):\2:'`
OUTPUT=`echo $@ | sed 's:-dev::'`

# Build the executable
$(RELEASE_DIR)/%: $(shell $(GO_FILES)) $(GO_VENDOR_FILE)
	@$(log) "Building" [$(GO_ENV) GOOS=$(LAZY_GOOS) GOARCH=$(LAZY_GOARCH) $(GO) build $(GO_FLAGS) ...]
	@$(GO_ENV) GOOS=$(LAZY_GOOS) GOARCH=$(LAZY_GOARCH) $(GO) build -o "$(OUTPUT)" -v $(GO_FLAGS) $(LD_FLAGS) $(MAIN)

# Build the executable in dev mode (much faster)
$(RELEASE_DIR)/%-dev: GO_ENV =
$(RELEASE_DIR)/%-dev: GO_FLAGS =
$(RELEASE_DIR)/%-dev: BUILD_TYPE = dev
$(RELEASE_DIR)/%-dev: $(RELEASE_DIR)/%

# link executables to a simplified name that is the same on all architectures.
go.link:
		@for i in $(wildcard $(RELEASE_DIR)/*-$(GOOS)-$(GOARCH)); do \
			ln -sfr $$i `echo $$i | sed 's:\(.*\)-.*-.*:\1:'` \
		; done

## initialize go dep
$(VENDOR_FILE):
	@$(log) "Initializing go deps"
	@mkdir -p $(VENDOR_DIR) && cd $(VENDOR_DIR)/.. && dep init

# vim: ft=make
