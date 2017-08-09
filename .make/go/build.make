# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# Infer GOOS and GOARCH
GOOS   ?= $(or $(word 1,$(subst -, ,${TARGET_PLATFORM})), $(shell echo "`go env GOOS`"))
GOARCH ?= $(or $(word 2,$(subst -, ,${TARGET_PLATFORM})), $(shell echo "`go env GOARCH`"))

# build
go.build: $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)

# default main file
MAIN ?= ./main.go

# default vendor folder
VENDOR_DIR ?= ./vendor

LAZY_GOOS=`echo $@ | sed 's:$(RELEASE_DIR)/$(NAME)-::' | sed 's:-.*::'`
LAZY_GOARCH=`echo $@ | sed 's:$(RELEASE_DIR)/$(NAME)-::' | sed 's:.*-::'`

# Build the executable
$(RELEASE_DIR)/$(NAME)-%: $(shell $(GO_FILES)) $(VENDOR_DIR)/vendor.json
	@$(log) "building" [$(GO_ENV) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GO_FLAGS) ...]
	$(GO_ENV) GOOS=$(LAZY_GOOS) GOARCH=$(LAZY_GOARCH) $(GO) build -o "$(RELEASE_DIR)/$(NAME)-$(LAZY_GOOS)-$(LAZY_GOARCH)" -v $(GO_FLAGS) $(LD_FLAGS) $(MAIN)

# Build the executable in dev mode (much faster)
go.dev: GO_FLAGS =
go.dev: GO_ENV =
go.dev: BUILD_TYPE = dev
go.dev: $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)

## link the executable to a simple name
$(RELEASE_DIR)/$(NAME): $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)
	@$(log) "linking binary" [ln -sf $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH) $(RELEASE_DIR)/$(NAME)]
	@ln -sf $(NAME)-$(GOOS)-$(GOARCH) $(RELEASE_DIR)/$(NAME)

go.link: $(RELEASE_DIR)/$(NAME)

go.link-dev: GO_FLAGS =
go.link-dev: GO_ENV =
go.link-dev: BUILD_TYPE = dev
go.link-dev: go.link

## initialize govendor
$(VENDOR_DIR)/vendor.json:
	@$(log) initializing govendor
	@mkdir -p $(VENDOR_DIR) && cd $(VENDOR_DIR)/.. && govendor init

# vim: ft=make
