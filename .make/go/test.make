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
