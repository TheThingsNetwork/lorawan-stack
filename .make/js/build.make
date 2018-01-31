# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# Include this makefile to enable webpack-related rules

# The location of the config files
CONFIG_DIR ?= config

# The place where we keep intermediate build files
CACHE_DIR ?= .cache

# Webpack
WEBPACK ?= ./node_modules/.bin/webpack
WEBPACK_FLAGS ?= --colors $(if $(CI),,--progress)

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

js.build: $(PUBLIC_DIR)/console.html

$(PUBLIC_DIR)/console.html: $(WEBPACK_CONFIG_BUILT) $(shell $(JS_SRC_FILES)) $(JS_SRC_DIR)/index.html package.json yarn.lock
	@$(log) "building client [webpack -c $(WEBPACK_CONFIG_BUILT) $(WEBPACK_FLAGS)]"
	@$(JS_ENV) $(WEBPACK) --config $(WEBPACK_CONFIG_BUILT) $(WEBPACK_FLAGS)

# build in dev mode
js.build-dev: NODE_ENV =
js.build-dev: js.dll js.build

# watch files
.PHONY: js.watch
js.watch: NODE_ENV = development
js.watch: js.dll js.watch_

js.watch_: WEBPACK_FLAGS += -w
js.watch_: js.build

## the location of the dll output
DLL_OUTPUT ?= $(PUBLIC_DIR)/libs.bundle.js

DLL_CONFIG_BUILT = $(subst $(CONFIG_DIR),$(CACHE_DIR)/config,$(DLL_CONFIG))

# DLL for faster dev builds
$(DLL_OUTPUT): $(DLL_CONFIG_BUILT) package.json yarn.lock
	@$(log) "building dll file"
	@GIT_TAG=$(GIT_TAG) DLL_FILE=$(DLL_OUTPUT) NODE_ENV=$(NODE_ENV) CACHE_DIR=$(CACHE_DIR) $(WEBPACK) --config $(DLL_CONFIG_BUILT) $(WEBPACK_FLAGS)

# build dll for faster rebuilds
js.dll: $(DLL_OUTPUT)

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
