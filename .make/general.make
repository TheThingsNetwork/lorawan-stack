# Copyright © 2017 The Things Network Foundation
# Use of this code is governed by the MIT license which can be found in the LICENSE file.

# This makefile contains general variables that are used troughout the other makefiles

# set shell
SHELL = bash

# count the input
count = wc -w

GENERAL_FILE = $(shell echo $(MAKEFILE_LIST) | xargs -n 1 echo | grep 'general\.make')
MAKE_DIR = $(shell dirname $(GENERAL_FILE))

# init rules are the rules to invoke to initialize the repo
INIT_RULES ?= git.hooks

# init invokes the init rules
init:
	@make $(INIT_RULES)
