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

GO = go
GOBIN = $(PWD)/.bin
export GOBIN

.PHONY: deps
deps:
	$(GO) install github.com/client9/misspell/cmd/misspell
	$(GO) install --tags extended github.com/gohugoio/hugo
	$(GO) install github.com/goreleaser/goreleaser
	$(GO) install github.com/magefile/mage
	$(GO) install github.com/mattn/goveralls
	$(GO) install github.com/mdempsky/unconvert
	$(GO) install github.com/mgechev/revive
	$(GO) install golang.org/x/tools/cmd/goimports
