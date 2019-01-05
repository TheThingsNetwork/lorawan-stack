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

# Storybook executables
STORYBOOK_START ?= ./node_modules/.bin/start-storybook
STORYBOOK_BUILD ?= ./node_modules/.bin/build-storybook

STORYBOOK_CONFIG ?= config/storybook

# Storybook options
STORYBOOK_PORT ?= 9001
STORYBOOK_OUTPUT ?= stories

# Start the storybook and watch for changes
js.storybook:
	@$(log) watching stories...
	@STORYBOOK=1 $(STORYBOOK_START) --config-dir $(STORYBOOK_CONFIG) --port $(STORYBOOK_PORT) --static-dir $(PUBLIC_DIR)

# Build static storybook
js.storybook.build:
	@$(log) building stories
	@SOTRYBOOK=1 $(STORYBOOK_BUILD) --config-dir $(STORYBOOK_CONFIG) --output-dir $(STORYBOOK_OUTPUT)
