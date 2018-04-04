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

NODE = node
NPM = npm
YARN = yarn
ESLINT ?= ./node_modules/.bin/eslint
BABEL ?= ./node_modules/.bin/babel
JSON ?= ./node_modules/.bin/json
JEST ?= ./node_modules/.bin/jest
NSP ?= ./node_modules/.bin/nsp
TRANSLATIONS = .cache/make

NODE_ENV ?= production

ESLINT_CONFIG = config/eslintrc.yaml

YARN_FLAGS ?= --no-emoji --no-progress
ESLINT_FLAGS ?= --no-ignore --color --config $(ESLINT_CONFIG)
BABEL_FLAGS ?= -D --ignore '*.test.js'

CACHE_DIR ?= .cache
PUBLIC_DIR ?= public
CONFIG_DIR ?= config

SUPPORT_LOCALES ?= en

JS_ENV = \
	PUBLIC_DIR=$(PUBLIC_DIR) \
	CACHE_DIR=$(CACHE_DIR) \
	NODE_ENV=$(NODE_ENV) \
	VERSION=$(CURRENT_VERSION) \
	GIT_TAG=$(GIT_TAG) \
	SUPPORT_LOCALES=$(SUPPORT_LOCALES)

JS_SRC_DIR ?= pkg/webui
JS_FILES ?= $(ALL_FILES) | $(only_js)
JS_SRC_FILES ?= $(ALL_FILES) | $(only_js) | $(only_js_src)
JS_STAGED_FILES = $(STAGED_FILES) | $(only_js)
JS_TESTS ?= $(JS_FILES) | grep "\_test\.js$$"

# Filters

# select only js files
only_js = grep '\.js$$'

# select only source files
only_js_src = grep '^\./$(JS_SRC_DIR)'

# ignore proto files
no_pb = grep -v '_pb\.js$$'

# Rules

# install dev dependencies
js.dev-deps:
	@$(log) "fetching js tools"
	@command -v yarn > /dev/null || ($(log) Installing yarn && npm install -g yarn)

js_init_script = \
	var fs = require('fs'); \
	try { var pkg = require('./package.json') } catch(err) { pkg = {} }; \
	pkg.babel = pkg.babel || { presets: [ 'ttn' ] }; \
	pkg.eslintConfig = pkg.eslintConfig || { extends: 'ttn' }; \
	pkg.jest = pkg.jest || { preset: 'jest-preset-ttn' }; \
	fs.writeFileSync('package.json', JSON.stringify(pkg, null, 2) + '\n');

# initialize repository
js.init:
	@$(log) "initializing js"
	@echo "$(js_init_script)" | node

INIT_RULES += js.init

# install dependencies
js.deps:
	@$(log) "fetching js dependencies"
	@$(YARN) install $(YARN_FLAGS)

# clean build files
js.clean:
	@$(log) "cleaning js public dir" [rm -rf $(PUBLIC_DIR)]
	@rm -rf $(PUBLIC_DIR)

# list js files
js.list:
	@$(JS_FILES) | sort

js.list.src:
	@$(JS_SRC_FILES) | sort

js.list-staged:
	@$(JS_STAGED_FILES) | sort

include .make/js/build.make
include .make/js/quality.make

# vim: ft=make
