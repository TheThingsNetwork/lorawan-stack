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

# This makefile contains rules to check and fix headers in files. This can be
# used to automatically check for copyright headers and fix them easily.
#
# You can override the following variables from this file:
# HEADER: The content of the header that every related file should have
# HEADER_FILES: The command to invoke to determine which file to check
# HEADER_EXTRA_FILES: Use this to add a hardcoded list of files

HEADER_FILE = ./.make/header.txt

empty = echo ""
no_blanks = sed '/^$$/d'

# fallbacks for
GO_LINT_FILES ?= $(empty)
JS_LINT_FILES ?= $(empty)

only_make = grep '\.make$$'
MAKE_LINT_FILES = $(ALL_FILES) | $(only_make)

only_proto = grep '\.proto$$'
PROTO_FILES = $(ALL_FILES) | $(only_proto)

ls:
	@$(MAKE_LINT_FILES)

HEADER_FILES ?= $(GO_LINT_FILES) && $(JS_LINT_FILES) && $(MAKE_LINT_FILES) && $(PROTO_FILES)
HEADER_EXTRA_FILES ?=

# the files to check for a header
_HEADER_FILES ?= { $(HEADER_FILES) && echo $(HEADER_EXTRA_FILES); } | $(no_blanks)
__HEADER_FILES = $(_HEADER_FILES)

# check files to see if they have the required header
headers.check:
	@$(log) "Checking headers in `echo $$($(__HEADER_FILES)) | $(count)` files"
	@FILES=`$(__HEADER_FILES)` $(GO) run $(MAKE_DIR)/headers.go check || { $(err) "Incorrect or missing header in $$file"; exit 1; }

# fix the headers in all the files
headers.fix:
	@$(log) "Fixing headers in `echo $$($(__HEADER_FILES)) | $(count)` files"
	@FILES=`$(__HEADER_FILES)` $(GO) run $(MAKE_DIR)/headers.go fix

# remove the headers in all the files
headers.remove:
	@$(log) "Removing headers in `echo $$($(__HEADER_FILES)) | $(count)` files"
	@FILES=`$(__HEADER_FILES)` $(GO) run $(MAKE_DIR)/headers.go remove

# check staged files
headers.check-staged: __HEADER_FILES = $(_HEADER_FILES) | $(only_staged)
headers.check-staged: headers.check

# check staged files
headers.fix-staged: __HEADER_FILES = $(_HEADER_FILES) | $(only_staged)
headers.fix-staged: headers.fix

# vim: ft=make
