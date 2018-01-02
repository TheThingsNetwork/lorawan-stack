# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# This makefile has tools for logging

LOG_NAME ?= $(shell basename $$(pwd))
LOG_LEN = $(shell echo $(LOG_NAME) | wc -c)

# log colors
log_black = \033[30m
log_red = \033[31m
log_green = \033[32m
log_yellow = \033[33m
log_blue = \033[34m
log_magenta = \033[35m
log_cyan = \033[36m
log_bold = \033[1m

log_clear = \033[0m

# default log colors
log_color ?= $(log_bold)$(log_blue)
log_error ?= $(log_bold)$(log_red)
log_warn ?= $(log_bold)$(log_yellow)
log_meta ?= $(log_black)

ECHO = echo -e

log = $(ECHO) "$(log_color)$(LOG_NAME)$(log_clear) "
err = $(ECHO) "$(log_error)$(LOG_NAME)$(log_clear) "
warn = $(ECHO) "$(log_warn)$(LOG_NAME)$(log_clear) "

meta = printf "$(log_meta)[ %s ]$(log_clear)\n"

# vim: ft=make
