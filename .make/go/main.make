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

.PHONY: go.min-version

# default release dir
RELEASE_DIR ?= release#

# default cache dir for intermediate files.
CACHE_DIR ?= .cache

# File where messages are written to.
GO_MESSAGES_FILE ?= $(RELEASE_DIR)/errors.json

# the first entry of the go path
GO_PATH ?= $(shell echo $(GOPATH) | awk -F':' '{ print $$1 }')

# the name of you go package (eg. github.com/foo/bar)
GO_PKG ?= $(shell echo $(PWD) | sed s:$(GO_PATH)/src/::)

# programs
GO = go
GO_FMT= gofmt
GO_MISSPELL= misspell
GO_UNCONVERT= unconvert
GO_METALINTER = gometalinter
GO_LINT_FILES = $(ALL_FILES) | $(only_go_lintable)
GO_MINIMUM_VERSION = 1.10
GO_VERSION := $(shell go version | cut -d' ' -f3 | cut -c 3-6)

# Get the minor Go version
get_go_major = cut -d'.' -f1
get_go_minor = cut -d'.' -f2

# go flags
GO_ENV = CGO_ENABLED=0
GO_STRINGVARS ?= -X go.thethings.network/lorawan-stack/pkg/version.GitCommit=$(GIT_COMMIT) -X go.thethings.network/lorawan-stack/pkg/version.BuildDate=$(BUILD_DATE) -X go.thethings.network/lorawan-stack/pkg/version.TTN=$(GIT_TAG) -X go.thethings.network/lorawan-stack/pkg/version.GitBranch=$(GIT_BRANCH)
LD_FLAGS = -ldflags "-w $(GO_STRINGVARS)"
ifdef ($(GO_TAGS))
	GO_FLAGS += "-tags $(GO_TAGS)"
endif

# coverage
GO_COVER_FILE = coverage.out
GO_FILTERED_COVER_FILE = coverage.filtered.out

# go test flags
GO_COVERALLS_FLAGS = -cover -covermode=atomic -coverprofile=$(GO_COVER_FILE)
GO_TEST_FLAGS ?= $(if $(CI),$(GO_COVERALLS_FLAGS),-cover)

# select only go files
only_go = grep '\.go$$'

# select/remove vendored files
no_vendor = grep -v 'vendor'
only_vendor = grep 'vendor'

# select/remove mock files
no_mock = grep -v '_mock\.go'
only_mock = grep '_mock\.go'

# select/remove protobuf generated files
no_pb = grep -Ev '\.pb\.go$$|\.pb\.gw\.go$$|pb_test.go$$'
only_pb = grep -E '\.pb\.go$$|\.pb\.gw\.go$$|pb_test.go$$'

# select/remove test files
no_test = grep -v '_test\.go$$'
only_test = grep '_test\.go$$'

# lintable files
only_go_lintable = $(only_go) | $(no_vendor) | $(no_mock) | $(no_pb)

# filter files to packages
to_packages = sed 's:/[^/]*$$::' | sort | uniq

# make packages local (prefix with ./)
to_local = sed 's:^:\./:' | sed 's:^\./.*\.go$$:.:'

# the govendor file
GO_VENDOR_FILE ?= Gopkg.toml

# all go files
GO_FILES = $(ALL_FILES) | $(only_go)

# local go packages
GO_PACKAGES = $(GO) list -v ./...

# local go packages as absolute paths
GO_PACKAGES_ABSOLUTE = $(GO) list -v -f '{{.Dir}}' ./...

# external go packages (in vendor)
EXTERNAL_PACKAGES = find ./vendor -name "*.go" | $(to_packages) | $(only_vendor)

# staged local packages
STAGED_PACKAGES = $(STAGED_FILES) | $(only_go) | $(no_vendor) | $(to_packages) | xargs $(GO) list -v 2>/dev/null

# staged local packages as absolute paths
STAGED_PACKAGES_ABSOLUTE = $(STAGED_FILES) | $(only_go) | $(no_vendor) | $(to_packages) | xargs $(GO) list -v -f '{{.Dir}}' 2>/dev/null

# packages for testing
TEST_PACKAGES = $(GO_FILES) | $(no_vendor) | $(only_test) | $(to_packages)

# get tools required for development
go.dev-deps:
	@$(log) "Installing go dev dependencies"
	@$(log) "Getting dep" && $(GO) get -u github.com/golang/dep/cmd/dep
	@if [[ ! -z "$(CI)" ]]; then $(log) "Getting goveralls" && $(GO) get -u github.com/mattn/goveralls; fi
	@$(log) "Getting gometalinter" && $(GO) get -u github.com/alecthomas/gometalinter
	@$(log) "Getting gometalinter linters" && $(GO_METALINTER) -i

# testing minimum version
go.min-version:
	@if [[ `echo $(GO_VERSION) | $(get_go_major)` -lt `echo $(GO_MINIMUM_VERSION) | $(get_go_major)` ]] || \
	 [[ `echo $(GO_VERSION) | $(get_go_major)` -eq `echo $(GO_MINIMUM_VERSION) | $(get_go_major)` && `echo $(GO_VERSION) | $(get_go_minor)` -lt `echo $(GO_MINIMUM_VERSION) | $(get_go_minor)` ]]; then \
		$(err) "Go is not up to date. Go $(GO_MINIMUM_VERSION) at least is required."; \
		exit 1; \
	fi

DEP_FLAGS ?= -v

# install dependencies
go.deps:
	@$(log) "Installing go dependencies"
	@dep ensure $(DEP_FLAGS)

# clean build files
go.clean:
	@$(log) "Cleaning release dir" [rm -rf $(RELEASE_DIR)]
	@rm -rf $(RELEASE_DIR)

# list all go files
go.list:
	@$(GO_FILES) | sort

# list all staged go files
go.list-staged: GO_FILES = $(STAGED_FILES) | $(only_go)
go.list-staged: go.list

# init initializes go
go.init: go.min-version
	@$(log) "Initializing go"
	@make go.dev-deps
	@make go.deps

INIT_RULES += go.init

include .make/go/build.make
include .make/go/test.make
include .make/go/quality.make

# vim: ft=make
