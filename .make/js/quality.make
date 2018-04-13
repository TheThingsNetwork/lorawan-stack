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

.PHONY: js.quality js.quality-staged js.lint js.lint-staged

JS_LINT_FILES = $(JS_FILES) | $(no_pb)
JS_LINT_STAGED_FILES = $(JS_STAGED_FILES) | $(no_pb)

# Lint

# lint all js files
js.lint:
	@$(log) "linting `$(JS_LINT_FILES) | $(count)` js files"
	@set -o pipefail;\
		files=`$(JS_LINT_FILES)`;\
		[ -n "$${files}" ] && echo $${files} | xargs $(ESLINT) $(ESLINT_FLAGS) | sed 's:$(PWD)/::'\
		|| exit 0

js.lintfix: ESLINT_FLAGS += --fix
js.lintfix: js.lint


# lint staged js files
js.lint-staged: JS_LINT_FILES = $(JS_LINT_STAGED_FILES)
js.lint-staged: js.lint

# perform all js quality checks
js.quality: js.lint

# perform js quality checks on staged files
js.quality-staged: js.lint-staged

# test all js files
js.test:
	@$(log) "testing `$(JS_TESTS) | $(count)` js files"
	@$(JEST) `$(JS_TESTS)`

js.vulnerabilities:
	@$(log) "checking js dependencies for vulnerabilities"
	@$(NSP) check

# vim: ft=make
