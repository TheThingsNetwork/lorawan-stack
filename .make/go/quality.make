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

# fmt all packages
go.fmt:
	@$(log) "Formatting `$(GO_PACKAGES_ABSOLUTE) | $(count)` go packages"
	@[[ -z "`$(GO_PACKAGES_ABSOLUTE) | xargs $(GO_FMT) -w -s | tee -a /dev/stderr`" ]]

# lint all packages, exiting when errors occur
go.lint:
	@GO111MODULE=on $(GO) run github.com/mgechev/revive -config .revive.toml -formatter stylish ./pkg/... ./cmd/...

# fix misspellings in all packages
go.misspell:
	@$(log) "Fixing misspellings in `$(GO_PACKAGES) | $(count)` go packages"
	@[[ -z "`$(GO_PACKAGES_ABSOLUTE) | GO111MODULE=on xargs $(GO) run github.com/client9/misspell/cmd/misspell -w | tee -a /dev/stderr`" ]]

# unconvert all packages
go.unconvert:
	@$(log) "Unconverting `$(GO_PACKAGES) | $(count)` go packages"
	@[[ -z "`$(GO_PACKAGES) | GO111MODULE=on xargs $(GO) run github.com/mdempsky/unconvert -safe -apply | tee -a /dev/stderr`" ]]

# run all quality on all files
go.quality: go.fmt go.misspell go.unconvert go.lint

# vim: ft=make
