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

# This makefile contains general variables that are used troughout the other makefiles

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
COMMA := ,

# set shell
SHELL = bash

# count the input
count = wc -w

GENERAL_FILE = $(shell echo $(MAKEFILE_LIST) | xargs -n 1 echo | grep 'general\.make')
MAKE_DIR = $(shell dirname $(GENERAL_FILE))

# init rules are the rules to invoke to initialize the repo
INIT_RULES ?= git.hooks

.PHONY: .FORCE
.FORCE:

# init invokes the init rules
init:
	@make $(INIT_RULES)

# vim: ft=make
