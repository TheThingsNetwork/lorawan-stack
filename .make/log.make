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
