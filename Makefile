# Copyright © 2017 The Things Network Foundation
# Use of this code is governed by the MIT license which can be found in the LICENSE file.

YEAR = 2017
HEADER = $(COMMENT) Copyright © $(YEAR) The Things Network Foundation\n$(COMMENT) Use of this code is governed by the MIT license which can be found in the LICENSE file.
HEADER_EXTRA_FILES = Makefile

PRE_COMMIT = headers.check-staged
COMMIT_MSG = git.commit-msg-length git.commit-msg-empty git.commit-msg-prefix

include .make/log.make
include .make/general.make
include .make/git.make
include .make/headers.make

