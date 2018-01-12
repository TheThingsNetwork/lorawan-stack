# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# This makefile contains rules to check and fix headers in files. This can be
# used to automatically check for copyright headers and fix them easily.
#
# You can override the following variables from this file:
# HEADER: The content of the header that every related file should have
# HEADER_FILES: The command to invoke to determine which file to check
# HEADER_EXTRA_FILES: Use this to add a hardcoded list of files

YEAR ?= $(shell date +%Y)
COMMENT ?= \\(\#\\|//\\)
HEADER_PREFIX = $(COMMENT) Copyright © $(YEAR)
HEADER = $(HEADER_PREFIX) The Things Network Foundation, distributed under the MIT license (see LICENSE file)

empty = echo ""
no_blanks = sed '/^$$/d'

# fallbacks for
GO_LINT_FILES ?= $(empty)
JS_LINT_FILES ?= $(empty)

only_make = grep '\.make$$'
MAKE_LINT_FILES = $(ALL_FILES) | $(only_make)

ls:
	@$(MAKE_LINT_FILES)

HEADER_FILES ?= $(GO_LINT_FILES) && $(JS_LINT_FILES) && $(MAKE_LINT_FILES)
HEADER_EXTRA_FILES ?=

# the files to check for a header
_HEADER_FILES ?= { $(HEADER_FILES) && echo $(HEADER_EXTRA_FILES); } | $(no_blanks)
__HEADER_FILES = $(_HEADER_FILES)

# check files to see if they have the required header
headers.check:
	@$(log) "Checking headers in `echo $$($(__HEADER_FILES)) | $(count)` files"
	@CODE=0; \
	for file in `$(__HEADER_FILES)`; do \
		"$(MAKE_DIR)/headers.sh" check "$(HEADER)" "$$file" || { $(err) "Incorrect or missing header in $$file"; CODE=1; }; \
	done; \
	exit $$CODE

# fix the headers in all the files
headers.fix:
	@$(log) "Fixing headers in `echo $$($(__HEADER_FILES)) | $(count)` files"
	@for file in `$(__HEADER_FILES)`; do \
		"$(MAKE_DIR)/headers.sh" fix "$(HEADER)" "$$file" "$(COMMENT)"; \
		code=$$?; \
		if [[ $$code -eq 2 ]]; then \
			$(log) "Fixed header in \`$$file\`"; \
		elif [[ $$code -ne 0 ]]; then \
			$(err) "Could not fix header in \`$$file\`"; exit 1; \
		fi; \
	done

# remove the headers in all the files
headers.remove:
	@$(log) "Removing headers in `echo $$($(__HEADER_FILES)) | $(count)` files"
	@for file in `$(__HEADER_FILES)`; do \
		"$(MAKE_DIR)/headers.sh" remove "$(HEADER)" "$$file" "$(COMMENT)"; \
		code=$$?; \
		if [[ $$code -eq 2 ]]; then \
			$(log) "Removed header in \`$$file\`"; \
		elif [[ $$code -ne 0 ]]; then \
			$(err) "Could not remove header in \`$$file\`"; exit 1; \
		fi; \
	done

# check staged files
headers.check-staged: __HEADER_FILES = $(_HEADER_FILES) | $(only_staged)
headers.check-staged: headers.check

# check staged files
headers.fix-staged: __HEADER_FILES = $(_HEADER_FILES) | $(only_staged)
headers.fix-staged: headers.fix

# vim: ft=make
