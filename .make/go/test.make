# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# run tests
go.test:
	@$(log) "Testing `$(TEST_PACKAGES) | $(count)` go packages"
	@$(GO) test $(GO_TEST_FLAGS) `$(TEST_PACKAGES)`

# package coverage
$(GO_COVER_DIR)/%.out: GO_TEST_FLAGS=-cover -coverprofile="$(GO_COVER_FILE)"
$(GO_COVER_DIR)/%.out: %
	@$(log) "Testing $<"
	@mkdir -p `dirname "$(GO_COVER_DIR)/$<"`
	@$(GO) test -cover -coverprofile="$@" "./$<"

## project coverage
$(GO_COVER_FILE): go.cover.clean $(patsubst ./%,./$(GO_COVER_DIR)/%.out,$(shell $(TEST_PACKAGES)))
	@echo "mode: set" > $(GO_COVER_FILE)
	@cat $(patsubst ./%,./$(GO_COVER_DIR)/%.out,$(shell $(TEST_PACKAGES))) | grep -vE "mode: set" | sort >> $(GO_COVER_FILE)


# vim: ft=make
