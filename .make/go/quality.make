# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# fmt all packages
go.fmt:
	@$(log) "Formatting `$(GO_PACKAGES) | $(count)` go packages"
	@[[ -z "`$(GO_PACKAGES) | xargs go fmt | tee -a /dev/stderr`" ]]

# fmt stages packages
go.fmt-staged: GO_PACKAGES = $(STAGED_PACKAGES)
go.fmt-staged: go.fmt

# lint all packages, exiting when errors occur
go.lint:
	@$(log) "Linting `$(GO_PACKAGES) | $(count)` go packages"
	@CODE=0; for pkg in `$(GO_PACKAGES)`; do $(GO_METALINTER) $(GO_METALINTER_FLAGS) $(GOPATH)/src/$$pkg 2> /dev/null || { CODE=1; }; done; exit $$CODE

go.lint-full: GO_METALINTER_FLAGS=$(GO_METALINTER_FLAGS_FULL)
go.lint-full: go.lint

# lint staged packages
go.lint-staged: GO_PACKAGES = $(STAGED_PACKAGES)
go.lint-staged: go.lint

# lint staged packages with all linters
go.lint-staged-full: GOMETALINTER_FLAGS=$(GO_METALINTER_FLAGS_FULL)
go.lint-staged-full: go.lint-staged

go.lint-travis: GO_PACKAGES = git diff --name-only HEAD $(TRAVIS_BRANCH) |  $(to_packages)
go.lint-travis: log = true
go.lint-travis: go.lint-full

go.lint-travis-comment:
	@if [ "$$TRAVIS_PULL_REQUEST" != "false" ]; then \
	 	REMARKS=`make go.lint-travis 2>/dev/null` || go run .make/comment.go '`gometalinter` has some remarks:' '```' "$$REMARKS" '```'; \
	else \
		make go.lint || true; \
	fi

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
go.quality: go.fmt go.lint go.check-vendors

# run all quality on staged files
go.quality-staged: go.fmt-staged go.lint-staged go.check-vendors-staged

# vim: ft=make
