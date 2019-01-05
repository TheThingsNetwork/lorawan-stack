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

STYLINT ?= ./node_modules/.bin/stylint

STYLINT_FLAGS ?= --config config/stylintrc.json

STYL_STAGED_FILES = $(STAGED_FILES) | $(only_styl)

# Filters

# select only styl files
only_styl = grep '\.styl$$'

styl.list:
	@$(ALL_FILES) | $(only_styl)

include .make/styl/quality.make
