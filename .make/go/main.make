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

.PHONY: go.min-version

# default release dir
RELEASE_DIR ?= release#

# the first entry of the go path
GO_PATH ?= $(shell echo $(GOPATH) | awk -F':' '{ print $$1 }')

# programs
GO = go
GO_FMT= gofmt
GO_MISSPELL= misspell
GO_UNCONVERT= unconvert
GO_METALINTER = gometalinter
GO_LINT_FILES = $(ALL_FILES) | $(only_go_lintable)

# go flags
ifdef ($(GO_TAGS))
	GO_FLAGS += "-tags $(GO_TAGS)"
endif

# coverage
GO_COVER_FILE = coverage.out
GO_FILTERED_COVER_FILE = coverage.filtered.out

# go test flags
GO_COVERALLS_FLAGS = -cover -covermode=atomic -coverprofile=$(GO_COVER_FILE)
GO_TEST_FLAGS ?= $(if $(CI),$(GO_COVERALLS_FLAGS),-cover) -timeout 5m

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

# all go files
GO_FILES = $(ALL_FILES) | $(only_go)

# local go packages
GO_PACKAGES = $(GO) list -v ./...

# local go packages as absolute paths
GO_PACKAGES_ABSOLUTE = $(GO) list -v -f '{{.Dir}}' ./...

# packages for testing
TEST_PACKAGES = $(GO_FILES) | $(no_vendor) | $(only_test) | $(to_packages)

# get tools required for development
go.dev-deps:
	@$(log) "Installing go dev dependencies"
	@if [[ ! -z "$(CI)" ]]; then $(log) "Getting goveralls" && GO111MODULE=off go get -u github.com/mattn/goveralls; fi
	@$(log) "Getting gometalinter" && GO111MODULE=off go get -u github.com/alecthomas/gometalinter
	@$(log) "Getting gometalinter linters" && GO111MODULE=off $(GO_METALINTER) -i

go.min-version: $(MAGE)
	@$(MAGE) go:checkVersion

DEP_FLAGS ?= -v

# install dependencies
go.deps:
	@$(log) "Installing go dependencies"
	@GO111MODULE=on go mod vendor

# clean build files
go.clean:
	@$(log) "Cleaning release dir" [rm -rf $(RELEASE_DIR)]
	@rm -rf $(RELEASE_DIR)

# init initializes go
go.init: go.min-version
	@$(log) "Initializing go"
	@make go.dev-deps
	@make go.deps

INIT_RULES += go.init

include .make/go/test.make
include .make/go/quality.make

# vim: ft=make
