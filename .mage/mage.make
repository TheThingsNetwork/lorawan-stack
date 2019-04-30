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

MAGE ?= ./mage

$(MAGE): magefile.go $(wildcard .mage/*.go)
	GO111MODULE=on go install github.com/magefile/mage
	GO111MODULE=on go run github.com/magefile/mage -compile $(MAGE)

.PHONY: init
init: $(MAGE)
	@$(MAGE) init
	@echo "Run \"./mage -l\" for a list of build targets"

.PHONY: git.pre-commit
git.pre-commit: $(MAGE) # NOTE: DO NOT CHANGE - will break previously installed git hooks.
	@HOOK=pre-commit $(MAGE) git:runHook

.PHONY: git.commit-msg
git.commit-msg: $(MAGE) # NOTE: DO NOT CHANGE - will break previously installed git hooks.
	@HOOK=commit-msg $(MAGE) git:runHook

.PHONY: git.pre-push
git.pre-push: $(MAGE) # NOTE: DO NOT CHANGE - will break previously installed git hooks.
	@HOOK=pre-push $(MAGE) git:runHook

# vim: ft=make
