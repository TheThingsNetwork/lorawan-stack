# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# run tests
go.test: key.pem cert.pem
	@$(log) "Testing `$(TEST_PACKAGES) | $(count)` go packages"
	@$(GO) test $(GO_TEST_FLAGS) `$(TEST_PACKAGES)`

$(GO_COVER_FILE): GO_TEST_FLAGS = $(GO_COVERALLS_FLAGS)
$(GO_COVER_FILE):
	$(MAKE) go.test

$(GO_FILTERED_COVER_FILE): $(GO_COVER_FILE)
	@cat $(GO_COVER_FILE) | grep -vE '.pb(.gw)?.go' > $(GO_FILTERED_COVER_FILE)

go.coveralls: $(GO_FILTERED_COVER_FILE)
	@goveralls -coverprofile=$(GO_FILTERED_COVER_FILE) -service=$${COVERALLS_SERVICE:-travis-ci} -repotoken $${COVERALLS_TOKEN:-""}

# vim: ft=make
