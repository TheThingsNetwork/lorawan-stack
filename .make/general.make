# This makefile contains general variables that are used troughout the other makefiles

# set shell
SHELL = bash

# count the input
count = wc -w

GENERAL_FILE = echo $(MAKEFILE_LIST) | xargs -n 1 echo | grep 'general\.make'
MAKE_DIR = $(GENERAL_FILE) | xargs dirname

# init rules are the rules to invoke to initialize the repo
INIT_RULES ?= git.hooks

# init invokes the init rules
init:
	@make $(INIT_RULES)
