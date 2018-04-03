# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

# go flags
GO_ENV = CGO_ENABLED=0
LD_FLAGS = -ldflags "-w $(GO_TAGS)"
GO_TAGS ?= -X github.com/TheThingsNetwork/ttn/pkg/version.GitCommit=$(GIT_COMMIT) -X github.com/TheThingsNetwork/ttn/pkg/version.BuildDate=$(BUILD_DATE) -X github.com/TheThingsNetwork/ttn/pkg/version.TTN=$(GIT_TAG) -X github.com/TheThingsNetwork/ttn/pkg/version.GitBranch=$(GIT_BRANCH)

# coverage
GO_COVER_FILE = coverage.out

# go test flags
GO_TEST_FLAGS ?= $(if $(CI),-cover -covermode=set -coverprofile=$(GO_COVER_FILE),-cover)

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

DEP_FLAGS ?= $(if $(CI),-vendor-only,)

# install dependencies
go.deps:
	@$(log) "Installing go dependencies"
	@dep ensure -v $(DEP_FLAGS)

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
go.init:
	@$(log) "Initializing go"
	@make go.dev-deps
	@make go.deps

# certificates
key.pem: dev-cert
cert.pem: dev-cert

INIT_RULES += go.init

include .make/go/build.make
include .make/go/test.make
include .make/go/quality.make

# vim: ft=make
