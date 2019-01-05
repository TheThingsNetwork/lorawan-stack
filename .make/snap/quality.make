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

SNAP_LINT_FILES = $(SNAP_FILES) | $(no_pb)
SNAP_LINT_STAGED_FILES = $(SNAP_STAGED_FILES) | $(no_pb)

# lint all snapshot files
snap.lint:
	@$(log) "linting `$(SNAP_LINT_FILES) | $(count)` snapshot files"
	@set -o pipefail;\
		files=`$(SNAP_LINT_FILES)`;\
		[ -n "$${files}" ] && echo $${files} | xargs $(ESLINT) $(ESLINT_FLAGS) | sed 's:$(PWD)/::'\
		|| exit 0

# lint staged snapshot files
snap.lint-staged: SNAP_LINT_FILES = $(SNAP_LINT_STAGED_FILES)
snap.lint-staged: snap.lint

# perform all snapshot quality checks
snap.quality: snap.lint

# perform snapshot quality checks on staged files
snap.quality-staged: snap.lint-staged
