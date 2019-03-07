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

GIT_TAG ?= `git describe --abbrev=0 --tags 2>/dev/null`

GIT_RELATIVE_DIR=git rev-parse --show-prefix

only_existing = (xargs ls -d 2>/dev/null || true)
dot_prefixed = sed 's:^:./:'

# All files that are not ignored by git
ALL_FILES ?= git ls-files --cached --modified --others --killed --exclude-standard | $(only_existing) | $(dot_prefixed)

# Get all files that are currently staged, except for deleted files
STAGED_FILES = git diff --staged --name-only --diff-filter=d --relative=$$($(GIT_RELATIVE_DIR)) | $(only_existing) | $(dot_prefixed)

.PHONY: git.hooks

git.hooks: $(MAGE)
	@$(MAGE) git:installHooks

INIT_RULES += git.hooks

.PHONY: git.hooks.remove

git.hooks.remove: $(MAGE)
	@$(MAGE) git:uninstallHooks

.PHONY: git.pre-commit

git.pre-commit: $(MAGE)
	@HOOK=pre-commit $(MAGE) git:runHook

.PHONY: git.commit-msg

git.commit-msg: $(MAGE)
	@HOOK=commit-msg $(MAGE) git:runHook

.PHONY: git.pre-push

git.pre-push: $(MAGE)
	@HOOK=pre-push $(MAGE) git:runHook

.PHONY: git.diff

git.diff:
	@if [[ ! -z "`git diff`" ]]; then \
		$(err) "Previous operations have created changes that were not recorded in the repository. Please make those changes on your local machine before pushing them to the repository:"; \
		git diff; \
		exit 1; \
	fi

# vim: ft=make
