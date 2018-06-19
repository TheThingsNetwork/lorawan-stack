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

STYL_LINT_FILES = $(ALL_FILES) | $(only_styl)
STYL_LINT_STAGED_FILES = $(STYL_STAGED_FILES) | $(no_pb)

styl.lint:
	@$(log) "linting `$(STYL_LINT_FILES) | $(count)` styl files"
	@set -o pipefail;\
		files=`$(STYL_LINT_FILES)`;\
		[ -n "$${files}" ] && echo $${files} | tr " " "\n" | xargs -L 1 $(STYLINT) $(STYLINT_FLAGS)\
		|| exit 0

# lint staged styl files
styl.lint-staged: STYL_LINT_FILES = $(STYL_LINT_STAGED_FILES)
styl.lint-staged: styl.lint

# perform all styl quality checks
styl.quality: styl.lint

# perform styl quality checks on staged files
styl.quality-staged: styl.lint-staged
