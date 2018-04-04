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

# fmt all packages
go.fmt:
	@$(log) "Formatting `$(GO_PACKAGES_ABSOLUTE) | $(count)` go packages"
	@[[ -z "`$(GO_PACKAGES_ABSOLUTE) | xargs $(GO_FMT) -w -s | tee -a /dev/stderr`" ]]

# lint all packages, exiting when errors occur
go.lint:
	@$(log) "Linting `$(GO_PACKAGES) | $(count)` go packages"
	@CODE=0; $(GO_METALINTER) `$(GO_PACKAGES_ABSOLUTE)` 2> /dev/null || { CODE=1; }; exit $$CODE

# fix misspellings in all packages
go.misspell:
	@$(log) "Fixing misspellings in `$(GO_PACKAGES) | $(count)` go packages"
	@[[ -z "`$(GO_PACKAGES_ABSOLUTE) | xargs $(GO_MISSPELL) -w | tee -a /dev/stderr`" ]]

# unconvert all packages
go.unconvert:
	@$(log) "Unconverting `$(GO_PACKAGES) | $(count)` go packages"
	@[[ -z "`$(GO_PACKAGES) | xargs $(GO_UNCONVERT) -safe -apply | tee -a /dev/stderr`" ]]

# fmt staged packages
go.fmt-staged: GO_PACKAGES_ABSOLUTE = $(STAGED_PACKAGES_ABSOLUTE)
go.fmt-staged: go.fmt

# lint staged packages
go.lint-staged: GO_PACKAGES = $(STAGED_PACKAGES)
go.lint-staged: go.lint

# fix misspellings in all staged packages
go.misspell-staged: GO_PACKAGES_ABSOLUTE = $(STAGED_PACKAGES_ABSOLUTE)
go.misspell-staged: go.misspell

# unconvert all staged packages
go.unconvert-staged: GO_PACKAGES = $(STAGED_PACKAGES)
go.unconvert-staged: go.unconvert

# lint changed packages in travis
go.lint-travis: GO_PACKAGES = git diff --name-only HEAD $(TRAVIS_BRANCH) | $(to_packages)
go.lint-travis: go.lint

# check if you have vendored packages in vendor
VENDOR_FILE = $(GO_VENDOR_FILE)
go.check-vendors: DOUBLY_VENDORED=$(shell cat $(VENDOR_FILE) | grep -n '^[\t ]*name = .*/vendor/' | awk '{ print $$1 $$3 }' | sed 's/["]//g')
go.check-vendors:
	@test $(VENDOR_FILE) != "/dev/null" && $(log) "Checking $(VENDOR_FILE) for bad packages" || true
	@if test $$(echo $(DOUBLY_VENDORED) | wc -w) -gt 0; then $(err) "Doubly vendored packages in $(VENDOR_FILE):" && echo $(DOUBLY_VENDORED) | xargs -n1 echo "       " | sed 's/:/  /' && exit 1; fi

# check if you have vendored packages in vendor (if it is staged)
go.check-vendors-staged: VENDOR_FILE=$(shell $(STAGED_FILES) | grep -q $(GO_VENDOR_FILE) || echo /dev/null)
go.check-vendors-staged: go.check-vendors

go.depfmt:
	@go run $(MAKE_DIR)/go/depfmt.go

# run all quality on all files
go.quality: go.fmt go.misspell go.unconvert go.lint go.check-vendors go.depfmt

# run all quality on staged files
go.quality-staged: go.fmt-staged go.misspell-staged go.unconvert-staged go.lint-staged go.check-vendors-staged go.depfmt

# vim: ft=make
