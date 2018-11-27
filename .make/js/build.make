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
WEBPACK_SERVE ?= $(BINARIES_DIR)/webpack-dev-server

# The config file to use for client
WEBPACK_CONFIG ?= $(CONFIG_DIR)/webpack.config.js

# The config file for DLL builds
DLL_CONFIG ?= $(CONFIG_DIR)/webpack.dll.js

# Check changes in cache message directory
BABEL_EXTRACTED_MESSAGES ?= $(shell if [ -d "$(CACHE_DIR)/messages/" ]; then find $(CACHE_DIR)/messages/ -type f -name '*' && find $(CACHE_DIR)/messages/ -type d; fi)

# Pre-build config files for quicker builds
$(CACHE_DIR)/config/%.js: $(CONFIG_DIR)/%.js
	@$(log) pre-building config files [babel $<]
	@mkdir -p $(CACHE_DIR)/config
	@$(BABEL) $< >| $@

# The location of the cached config file
WEBPACK_CONFIG_BUILT = $(subst $(CONFIG_DIR)/,$(CACHE_DIR)/config/,$(WEBPACK_CONFIG))

## the location of the dll output
DLL_OUTPUT ?= $(PUBLIC_DIR)/libs.bundle.js
DLL_CONFIG_BUILT = $(subst $(CONFIG_DIR),$(CACHE_DIR)/config,$(DLL_CONFIG))

# Run webpack main bundle (webpack.config.js)
js.build-main: $(DLL_OUTPUT) $(shell $(JS_SRC_FILES)) yarn.lock
	$(MAKE) js.webpack-main

js.build-watch: NODE_ENV = development
js.build-watch: WEBPACK_FLAGS += -w
js.build-watch: js.webpack-main

js.build: js.build-dll js.build-main

js.watch: js.build-dll js.build-watch

js.serve: DEV_SERVER_BUILD = true
js.serve: $(WEBPACK_CONFIG_BUILT) $(DEFAULT_LOCALE_FILE) $(BACKEND_LOCALES_DIR) $(DLL_OUTPUT)
	@$(log) "Serving via webpack-dev-server, make sure stack is running for the api proxy to work"
	@$(JS_ENV) $(WEBPACK_SERVE) --config $(WEBPACK_CONFIG_BUILT)

js.webpack-main: $(WEBPACK_CONFIG_BUILT) $(DEFAULT_LOCALE_FILE) $(XX_LOCALE_FILE) $(BACKEND_LOCALES_DIR)
	@$(log) "Building client [webpack -c $(WEBPACK_CONFIG_BUILT) $(WEBPACK_FLAGS)]"
	@$(JS_ENV) $(WEBPACK) --config $(WEBPACK_CONFIG_BUILT) $(WEBPACK_FLAGS)

# build in dev mode
js.build-dev: NODE_ENV =
js.build-dev: js.build

# DLL for faster dev builds
$(DLL_OUTPUT): $(DLL_CONFIG_BUILT) yarn.lock
	$(MAKE) js.webpack-dll

js.webpack-dll:
	@$(log) "building dll file"
	@GIT_TAG=$(GIT_TAG) DLL_FILE=$(DLL_OUTPUT) NODE_ENV=$(NODE_ENV) CACHE_DIR=$(CACHE_DIR) $(WEBPACK) --config $(DLL_CONFIG_BUILT) $(WEBPACK_FLAGS)


# build dll for faster rebuilds
js.build-dll: $(DLL_CONFIG_BUILT) $(DLL_OUTPUT)

$(CACHE_DIR)/make/%.js: .make/js/%.js
	@$(log) "Pre-building translation scripts [babel $<]"
	@mkdir -p $(CACHE_DIR)/make
	@$(BABEL) $< >| $@

# Translations

$(CACHE_DIR)/messages: $(shell $(JS_SRC_FILES))
	@$(log) "Extracting frontend translation messages via babel"
	@rm -rf $(CACHE_DIR)/messages
	@mkdir -p $(LOCALES_DIR)
	@$(BABEL) -q $(JS_SRC_DIR) > /dev/null

$(DEFAULT_LOCALE_FILE): $(CACHE_DIR)/make/translations.js $(CACHE_DIR)/make/xx.js $(CACHE_DIR)/messages $(BABEL_EXTRACTED_MESSAGES)
	@$(log) "Gathering frontend translation messages"
	@$(NODE) $(CACHE_DIR)/make/translations.js --support $(SUPPORT_LOCALES)

$(XX_LOCALE_FILE): $(DEFAULT_LOCALE_FILE)

.PHONY: js.gather-locales
js.gather-locales: $(DEFAULT_LOCALE_FILE)
	@:

$(BACKEND_LOCALES_DIR): $(CACHE_DIR)/make/translations.js $(CACHE_DIR)/make/xx.js $(CACHE_DIR)/messages
	@$(log) "Gathering backend translation messages"
	@$(NODE) $(CACHE_DIR)/make/translations.js --support $(SUPPORT_LOCALES) --backend-messages $(CONFIG_DIR)/messages.json --locales $(BACKEND_LOCALES_DIR) --backend-only

js.translations: $(DEFAULT_LOCALE_FILE)

# vim: ft=make
