# Copyright © 2017 The Things Network Foundation
# Use of this code is governed by the MIT license which can be found in the LICENSE file.


# run tests
go.test:
	@$(log) "testing `$(TEST_PACKAGES) | $(count)` go packages"
	@$(GO) test $(GO_TEST_FLAGS) `$(TEST_PACKAGES)`

# package coverage
$(GO_COVER_DIR)/%.out: GO_TEST_FLAGS=-cover -coverprofile="$(GO_COVER_FILE)"
$(GO_COVER_DIR)/%.out: %
	@$(log) "testing $<"
	@mkdir -p `dirname "$(GO_COVER_DIR)/$<"`
	@$(GO) test -cover -coverprofile="$@" "./$<"

## project coverage
$(GO_COVER_FILE): go.cover.clean $(patsubst ./%,./$(GO_COVER_DIR)/%.out,$(shell $(TEST_PACKAGES)))
	@echo "mode: set" > $(GO_COVER_FILE)
	@cat $(patsubst ./%,./$(GO_COVER_DIR)/%.out,$(shell $(TEST_PACKAGES))) | grep -vE "mode: set" | sort >> $(GO_COVER_FILE)

