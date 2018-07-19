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

# Include this makefile to enable webpack-related rules

# The location of the config files
CONFIG_DIR ?= config

# The place where we keep intermediate build files
CACHE_DIR ?= .cache

# Webpack
WEBPACK ?= $(BINARIES_DIR)/webpack
WEBPACK_FLAGS ?= --colors $(if $(CI),,--progress)
WEBPACK_SERVE ?= $(BINARIES_DIR)/webpack-serve

# The config file to use for client
WEBPACK_CONFIG ?= $(CONFIG_DIR)/webpack.config.js

# The config file for DLL builds
DLL_CONFIG ?= $(CONFIG_DIR)/webpack.dll.js

# Pre-build config files for quicker builds
$(CACHE_DIR)/config/%.js: $(CONFIG_DIR)/%.js
	@$(log) pre-building config files [babel $<]
	@mkdir -p $(CACHE_DIR)/config
	@$(BABEL) $< >| $@

# The location of the cached config file
WEBPACK_CONFIG_BUILT = $(subst $(CONFIG_DIR)/,$(CACHE_DIR)/config/,$(WEBPACK_CONFIG))

# Run webpack main bundle (webpack.config.js)
js.build-main: $(PUBLIC_DIR)/console.html

js.build-watch: NODE_ENV = development
js.build-watch: WEBPACK_FLAGS += -w
js.build-watch: js.webpack-main

js.build: js.build-dll js.build-main

js.watch: js.build-dll js.build-watch

js.serve: DEV_SERVER_BUILD = true
js.serve: $(WEBPACK_CONFIG_BUILT)
	@$(log) "Serving via webpack-serve, make sure stack is running for the api proxy to work"
	@$(JS_ENV) $(WEBPACK_SERVE) $(WEBPACK_CONFIG_BUILT)

js.webpack-main:
	@$(log) "building client [webpack -c $(WEBPACK_CONFIG_BUILT) $(WEBPACK_FLAGS)]"
	@$(JS_ENV) $(WEBPACK) --config $(WEBPACK_CONFIG_BUILT) $(WEBPACK_FLAGS)

$(PUBLIC_DIR)/console.html: $(WEBPACK_CONFIG_BUILT) $(shell $(JS_SRC_FILES)) $(JS_SRC_DIR)/index.html yarn.lock
	$(MAKE) js.webpack-main

# build in dev mode
js.build-dev: NODE_ENV =
js.build-dev: js.build

## the location of the dll output
DLL_OUTPUT ?= $(PUBLIC_DIR)/libs.bundle.js

DLL_CONFIG_BUILT = $(subst $(CONFIG_DIR),$(CACHE_DIR)/config,$(DLL_CONFIG))

# DLL for faster dev builds
$(DLL_OUTPUT): $(DLL_CONFIG_BUILT) yarn.lock
	$(MAKE) js.webpack-dll

js.webpack-dll:
	@$(log) "building dll file"
	@GIT_TAG=$(GIT_TAG) DLL_FILE=$(DLL_OUTPUT) NODE_ENV=$(NODE_ENV) CACHE_DIR=$(CACHE_DIR) $(WEBPACK) --config $(DLL_CONFIG_BUILT) $(WEBPACK_FLAGS)


# build dll for faster rebuilds
js.build-dll: $(DLL_OUTPUT)

$(CACHE_DIR)/make/%.js: .make/js/%.js
	@$(log) "pre-building translation scrips [babel $<]"
	@mkdir -p $(CACHE_DIR)/make
	@$(BABEL) $< >| $@

SUPPORT_LOCALES ?= en,ja
DEFAULT_LOCALE ?= en
OUTPUT_MESSAGES ?= messages.yml,messages.xlsx
UPDATES_FILES ?= messages.xlsx

# update translations
js.translations: $(CACHE_DIR)/make/translations.js $(CACHE_DIR)/make/xls.js $(CACHE_DIR)/make/xx.js
	@$(log) "gathering translations [translations --output messages.{yml,xlsx]"
	@$(NODE) $(CACHE_DIR)/make/translations.js --output $(OUTPUT_MESSAGES) --support $(SUPPORT_LOCALES) --updates $(UPDATES_FILES)

# vim: ft=make
