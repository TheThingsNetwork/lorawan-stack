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

NODE = node
NPM = npm

CACHE_DIR ?= .cache
PUBLIC_DIR ?= public
CONFIG_DIR ?= config
BINARIES_DIR ?= ./node_modules/.bin
DEV_SERVER_BUILD ?= false
YARN_VERSION ?= 1.12.3

YARN ?= $(BINARIES_DIR)/yarn
ESLINT ?= $(BINARIES_DIR)/eslint
BABEL ?= $(BINARIES_DIR)/babel
JSON ?= $(BINARIES_DIR)/json
JEST ?= $(BINARIES_DIR)/jest
NSP ?= $(BINARIES_DIR)/nsp
TRANSLATIONS = .cache/make

NODE_ENV ?= production

YARN_FLAGS ?= --no-emoji --no-progress
ESLINT_FLAGS ?= --no-ignore --color
BABEL_FLAGS ?= -D --ignore '*.test.js'

SUPPORT_LOCALES ?= en
DEFAULT_LOCALE ?= en

JS_ENV = \
	PUBLIC_DIR=$(PUBLIC_DIR) \
	CACHE_DIR=$(CACHE_DIR) \
	NODE_ENV=$(NODE_ENV) \
	VERSION=$(CURRENT_VERSION) \
	GIT_TAG=$(GIT_TAG) \
	SUPPORT_LOCALES=$(SUPPORT_LOCALES) \
	DEV_SERVER_BUILD=$(DEV_SERVER_BUILD) \
	DEFAULT_LOCALE=$(DEFAULT_LOCALE)

JS_SRC_DIR ?= pkg/webui
JS_FILES ?= $(ALL_FILES) | $(only_js)
JS_SRC_FILES ?= $(ALL_FILES) | $(only_js) | $(only_js_src)
JS_STAGED_FILES = $(STAGED_FILES) | $(only_js)
JS_TESTS ?= $(JS_FILES) | grep "\_test\.js$$"

LOCALES_DIR ?= $(JS_SRC_DIR)/locales
BACKEND_LOCALES_DIR ?= $(LOCALES_DIR)/.backend
DEFAULT_LOCALE_FILE ?= $(LOCALES_DIR)/$(DEFAULT_LOCALE).json
XX_LOCALE_FILE ?= $(LOCALES_DIR)/xx.json

# Filters

# select only js files
only_js = grep '\.js$$'

# select only source files
only_js_src = grep '^\./$(JS_SRC_DIR)'

# ignore proto files
no_pb = grep -v '_pb\.js$$'

# Rules

$(YARN): js.dev-deps

js.dev-deps: $(MAGE)
	@$(log) "Installing js dev dependencies"
	@$(MAGE) js:devDeps

# install dependencies
js.deps: $(MAGE)
	@$(log) "Installing js dependencies"
	@$(MAGE) js:deps

# init initializes js
js.init: $(MAGE)
	@$(log) "Initializing js"
	@make js.dev-deps
	@make sdk.js.deps
	@make sdk.js.build
	@make js.deps

INIT_RULES += js.init

# clean build files and cache
js.clean-public:
	@$(log) "cleaning js public dir" [rm -rf $(PUBLIC_DIR)]
	@rm -rf $(PUBLIC_DIR)

js.flush-cache:
	@$(log) "cleaning cache dir" [rm -rf $(CACHE_DIR)]
	@rm -rf $(CACHE_DIR)

js.clean-locale:
	@$(log) "cleaning locales dir" [rm -rf $(CACHE_DIR)]
	@rm -rf $(BACKEND_LOCALES_DIR)

js.clean: js.clean-public js.clean-locale js.flush-cache

# list js files
js.list:
	@$(JS_FILES) | sort

js.list.src:
	@$(JS_SRC_FILES) | sort

js.list-staged:
	@$(JS_STAGED_FILES) | sort

include .make/js/build.make
include .make/js/quality.make
include .make/js/storybook.make

# vim: ft=make
