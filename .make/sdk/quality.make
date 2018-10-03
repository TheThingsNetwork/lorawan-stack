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

# Run all SDK tests
sdk.test: sdk.js.test


# JS SDK tests

sdk.js.test:
	@$(log) "Running JS SDK tests"
	@$(YARN_SDK) run test

sdk.js.test-watch:
	@$(log) "Watching JS SDK tests"
	@$(YARN_SDK) run test:watch


