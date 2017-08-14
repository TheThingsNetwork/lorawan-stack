# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# fmt all packages
go.fmt:
	@$(log) "Formatting `$(GO_PACKAGES) | $(count)` go packages"
	@[[ -z "`$(GO_PACKAGES) | xargs go fmt | tee -a /dev/stderr`" ]]

# fmt stages packages
go.fmt-staged: GO_PACKAGES = $(STAGED_PACKAGES)
go.fmt-staged: go.fmt

# vet all packages
go.vet:
	@$(log) "Vetting `$(GO_PACKAGES) | $(count)` go packages"
	@$(GO_PACKAGES) | xargs $(GO) vet

# vet staged packages
go.vet-staged: GO_PACKAGES = $(STAGED_PACKAGES)
go.vet-staged: go.vet

# lint all packages, exiting when errors occur
go.lint:
	@$(log) "Linting `$(GO_LINT_FILES) | $(count)` go files"
	@CODE=0; for pkg in `$(GO_LINT_FILES)`; do $(GOLINT) $(GOLINT_FLAGS) $$pkg 2>/dev/null || { CODE=1; }; done; exit $$CODE

# lint all packages, ignoring errors
go.lint-all: GOLINT_FLAGS =
go.lint-all: go.lint

# lint staged files
go.lint-staged: GO_LINT_FILES = $(GO_LINT_STAGED_FILES)
go.lint-staged: go.lint

# check if you have vendored packages in vendor
VENDOR_FILE = $(GO_VENDOR_FILE)
go.check-vendors: DOUBLY_VENDORED=$(shell cat $(VENDOR_FILE) | grep -n '^[\t ]*name = .*/vendor/' | awk '{ print $$1 $$3 }' | sed 's/["]//g')
go.check-vendors:
	@test $(VENDOR_FILE) != "/dev/null" && $(log) "Checking $(VENDOR_FILE) for bad packages" || true
	@if test $$(echo $(DOUBLY_VENDORED) | wc -w) -gt 0; then $(err) "Doubly vendored packages in $(VENDOR_FILE):" && echo $(DOUBLY_VENDORED) | xargs -n1 echo "       " | sed 's/:/  /' && exit 1; fi

# check if you have vendored packages in vendor (if it is staged)
go.check-vendors-staged: VENDOR_FILE=$(shell $(STAGED_FILES) | grep -q $(GO_VENDOR_FILE) || echo /dev/null)
go.check-vendors-staged: go.check-vendors

# run all quality on all files
go.quality: go.fmt go.vet go.lint go.check-vendors

# run all quality on staged files
go.quality-staged: go.fmt-staged go.vet-staged go.lint-staged go.check-vendors-staged

# vim: ft=make
