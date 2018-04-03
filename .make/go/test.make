# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# run tests
go.test: key.pem cert.pem
	@$(log) "Testing `$(TEST_PACKAGES) | $(count)` go packages"
	@$(GO) test $(GO_TEST_FLAGS) `$(TEST_PACKAGES)`

# vim: ft=make
