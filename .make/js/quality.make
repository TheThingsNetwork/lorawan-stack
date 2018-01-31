# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

.PHONY: js.quality js.quality-staged js.lint js.lint-staged

JS_LINT_FILES = $(JS_FILES) | $(no_pb)
JS_LINT_STAGED_FILES = $(JS_STAGED_FILES) | $(no_pb)

# Lint

# lint all js files
js.lint:
	@$(log) "linting `$(JS_LINT_FILES) | $(count)` js files"
	@set -o pipefail; ($(JS_LINT_FILES) || exit 0) | xargs $(ESLINT) $(ESLINT_FLAGS) | sed 's:$(PWD)/::'

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
