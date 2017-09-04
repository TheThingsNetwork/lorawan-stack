# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# default release dir
RELEASE_DIR ?= release#

# the first entry of the go path
GO_PATH ?= $(shell echo $(GOPATH) | awk -F':' '{ print $$1 }')

# the name of you go package (eg. github.com/foo/bar)
GO_PKG ?= $(shell echo $(PWD) | sed s:$(GO_PATH)/src/::)

# programs
GO = go
GOLINT = golint

# go flags
GO_FLAGS = -a
GO_ENV = CGO_ENABLED=0
LD_FLAGS = -ldflags "-w $(GO_TAGS)"
GO_TAGS ?= -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE) -X main.tag=$(GIT_TAG) -X main.branch=$(GIT_BRANCH)

# golint flags
GOLINT_FLAGS = -set_exit_status

# go test flags
GO_TEST_FLAGS = -cover

# coverage
GO_COVER_FILE = coverage.out
GO_COVER_DIR  = .coverage

# select only go files
only_go = grep '\.go$$'

# select/remove vendored files
no_vendor = grep -v 'vendor'
only_vendor = grep 'vendor'

# select/remove mock files
no_mock = grep -v '_mock\.go'
only_mock = grep '_mock\.go'

# select/remove protobuf generated files
no_pb = grep -Ev '\.pb\.go$$|\.pb\.gw\.go$$'
only_pb = grep -E '\.pb\.go$$|\.pb\.gw\.go$$'

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
GO_PACKAGES = go list -v ./...

# lintable go files
GO_LINT_FILES = $(ALL_FILES) | $(only_go_lintable)

# staged lintable go files
GO_LINT_STAGED_FILES = $(ALL_FILES) | $(only_go_lintable) | $(only_staged)

# external go packages (in vendor)
EXTERNAL_PACKAGES = find ./vendor -name "*.go" | $(to_packages) | $(only_vendor)

# staged local packages
STAGED_PACKAGES = $(STAGED_FILES) | $(only_go) | $(no_vendor) | $(to_packages) | xargs $(GO) list -v 2>/dev/null

# packages for testing
TEST_PACKAGES = $(GO_FILES) | $(no_vendor) | $(only_test) | $(to_packages)

# get tools required for development
go.dev-deps:
	@$(log) "Installing go dev dependencies"
	@command -v dep  >/dev/null || { $(log) "Installing dep" && $(GO) get -u github.com/golang/dep/cmd/dep; }
	@command -v golint >/dev/null || { $(log) "Installing golint" && $(GO) get -u github.com/golang/lint/golint; }

# install dependencies
go.deps:
	@$(log) "Installing go dependencies"
	@dep ensure -v

# install packages for faster rebuilds
go.install:
	@$(log) "Installing `$(EXTERNAL_PACKAGES) | $(count)` go packages"
	@$(GO) install -v ./...

# pre-build local files, ignoring failures (from unused packages or files for example)
# use this to improve build speed
go.pre:
	@$(log) "Installing go packages"
	@$(GO_FILES) | $(to_packages) | xargs $(GO) install -v || true

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

INIT_RULES += go.init

include .make/go/build.make
include .make/go/test.make
include .make/go/quality.make

# vim: ft=make
