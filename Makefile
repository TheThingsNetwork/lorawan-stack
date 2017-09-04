# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

YEAR = 2017
HEADER = $(COMMENT) Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)
HEADER_EXTRA_FILES = Makefile

PRE_COMMIT = headers.check-staged
COMMIT_MSG = git.commit-msg-log git.commit-msg-length git.commit-msg-empty git.commit-msg-prefix git.commit-msg-phrase git.commit-msg-casing

include .make/log.make
include .make/general.make
include .make/git.make
include .make/headers.make
include .make/go/main.make

messages:
	@$(GO) run ./pkg/errors/generate_messages.go --filename config/messages.json

# Example builds
ttn-example: MAIN=./cmd/ttn-example/main.go
ttn-example: $(RELEASE_DIR)/ttn-example-$(GOOS)-$(GOARCH)
ttn-example: go.link

ttn-example-dev: MAIN=./cmd/ttn-example/main.go
ttn-example-dev: $(RELEASE_DIR)/ttn-example-$(GOOS)-$(GOARCH)-dev
ttn-example-dev: go.link

ttn-example-docker: MAIN=./cmd/ttn-example/main.go
ttn-example-docker: $(RELEASE_DIR)/ttn-example-linux-amd64
