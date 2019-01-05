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

YARN_SDK_FLAGS ?= --cwd=./sdk/js
YARN_SDK ?= $(YARN) $(YARN_SDK_FLAGS)

# install all sdk dependencies
sdk.deps: sdk.js.deps


# JS SDK make rules

sdk.js.deps:
	@$(log) "Fetching JS SDK dependencies"
	@$(YARN_SDK) install $(YARN_FLAGS) --production=false

include .make/sdk/build.make
include .make/sdk/quality.make
