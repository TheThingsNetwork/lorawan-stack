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

.PHONY: git.install-hooks
git.install-hooks: $(MAGE)
	@$(MAGE) git:installHooks

INIT_RULES += git.install-hooks

.PHONY: git.uninstall-hooks
git.uninstall-hooks: $(MAGE)
	@$(MAGE) git:uninstallHooks

.PHONY: git.pre-commit
git.pre-commit: $(MAGE) # NOTE: DO NOT CHANGE - will break previously installed git hooks.
	@HOOK=pre-commit $(MAGE) git:runHook

.PHONY: git.commit-msg
git.commit-msg: $(MAGE) # NOTE: DO NOT CHANGE - will break previously installed git hooks.
	@HOOK=commit-msg $(MAGE) git:runHook

.PHONY: git.pre-push
git.pre-push: $(MAGE) # NOTE: DO NOT CHANGE - will break previously installed git hooks.
	@HOOK=pre-push $(MAGE) git:runHook

.PHONY: go.check-version
go.check-version: $(MAGE)
	@$(MAGE) go:checkVersion

.PHONY: go.deps
go.deps:
	@GO111MODULE=on go mod vendor

DEPS_RULES += go.deps

coverage.out: go.cover

.PHONY: go.cover
go.cover: $(MAGE) dev.certs
	@$(MAGE) go:cover

.PHONY: go.coveralls
go.coveralls: $(MAGE) dev.certs
	@$(MAGE) go:coveralls

.PHONY: go.fmt
go.fmt: $(MAGE)
	@$(MAGE) go:fmt

.PHONY: go.lint
go.lint: $(MAGE)
	@$(MAGE) go:lint

.PHONY: go.misspell
go.misspell: $(MAGE)
	@$(MAGE) go:misspell

.PHONY: go.quality
go.quality: $(MAGE)
	@$(MAGE) go:quality

QUALITY_RULES += go.quality

.PHONY: go.test
go.test: $(MAGE) dev.certs
	@$(MAGE) go:test

TEST_RULES += go.test

.PHONY: go.unconvert
go.unconvert: $(MAGE)
	@$(MAGE) go:unconvert

.PHONY: go.messages
go.messages: $(MAGE)
	@$(MAGE) go:messages

.PHONY: headers.check
headers.check: $(MAGE)
	@$(MAGE) headers:check

.PHONY: js.backend-translations
js.backend-translations: $(MAGE)
	@$(MAGE) js:backendTranslations

.PHONY: js.build
js.build: $(MAGE)
	@$(MAGE) js:build

BUILD_RULES += js.build

.PHONY: js.build-dll
js.build-dll: $(MAGE)
	@$(MAGE) js:buildDll

.PHONY: js.build-main
js.build-main: $(MAGE)
	@$(MAGE) js:buildMain

.PHONY: js.clean
js.clean: $(MAGE)
	@$(MAGE) js:clean

CLEAN_RULES += js.clean

.PHONY: js.deps
js.deps: $(MAGE)
	@$(MAGE) js:deps

DEPS_RULES += js.deps

.PHONY: js.dev-deps
js.dev-deps: $(MAGE)
	@$(MAGE) js:devDeps

.PHONY: js.lint
js.lint: $(MAGE)
	@$(MAGE) js:lint

.PHONY: js.lint-all
js.lint-all: $(MAGE)
	@$(MAGE) js:lintAll

QUALITY_RULES += js.lint-all

.PHONY: js.lint-snap
js.lint-snap: $(MAGE)
	@$(MAGE) js:lintSnap

.PHONY: js.messages
js.messages: $(MAGE)
	@$(MAGE) js:messages

.PHONY: js.serve
js.serve: $(MAGE)
	@$(MAGE) js:serve

.PHONY: js.serve-main
js.serve-main: $(MAGE)
	@$(MAGE) js:serveMain

.PHONY: js.storybook
js.storybook: $(MAGE)
	@$(MAGE) js:storybook

.PHONY: js.test
js.test: $(MAGE)
	@$(MAGE) js:test

TEST_RULES += js.test

.PHONY: js.translations
js.translations: $(MAGE)
	@$(MAGE) js:translations

.PHONY: js.vulnerabilities
js.vulnerabilities: $(MAGE)
	@$(MAGE) js:vulnerabilities

.PHONY: proto.image
proto.image: $(MAGE)
	@$(MAGE) proto:image

.PHONY: proto.all
proto.all: $(MAGE)
	@$(MAGE) proto:all

.PHONY: proto.clean
proto.clean: $(MAGE)
	@$(MAGE) proto:clean

.PHONY: proto.go
proto.go: $(MAGE)
	@$(MAGE) proto:go

.PHONY: proto.go.clean
proto.go.clean: $(MAGE)
	@$(MAGE) proto:goClean

.PHONY: proto.markdown
proto.markdown: $(MAGE)
	@$(MAGE) proto:markdown

.PHONY: proto.markdown.clean
proto.markdown.clean: $(MAGE)
	@$(MAGE) proto:markdownClean

.PHONY: proto.sdk.js
proto.sdk.js: $(MAGE)
	@$(MAGE) proto:sdkJs

.PHONY: proto.js.sdk.clean
proto.js.sdk.clean: $(MAGE)
	@$(MAGE) proto:sdkJsClean

.PHONY: proto.swagger
proto.swagger: $(MAGE)
	@$(MAGE) proto:swagger

.PHONY: proto.swagger.clean
proto.swagger.clean: $(MAGE)
	@$(MAGE) proto:swaggerClean

.PHONY: js.sdk.build
js.sdk.build: $(MAGE)
	@$(MAGE) jsSDK:build

.PHONY: js.sdk.clean
js.sdk.clean: $(MAGE)
	@$(MAGE) jsSDK:clean

CLEAN_RULES += js.sdk.clean

.PHONY: js.sdk.deps
js.sdk.deps: $(MAGE)
	@$(MAGE) jsSDK:deps

DEPS_RULES += js.sdk.deps

.PHONY: js.sdk.dev-deps
js.sdk.dev-deps: $(MAGE)
	@$(MAGE) jsSDK:devDeps

.PHONY: js.sdk.test
js.sdk.test: $(MAGE)
	@$(MAGE) jsSDK:test

TEST_RULES += js.sdk.test

.PHONY: js.sdk.test-watch
js.sdk.test-watch: $(MAGE)
	@$(MAGE) jsSDK:testWatch

.PHONY: js.sdk.watch
js.sdk.watch: $(MAGE)
	@$(MAGE) jsSDK:watch

.PHONY: js.sdk.definitions
js.sdk.definitions: $(MAGE)
	@$(MAGE) jsSDK:definitions

.PHONY: styl.lint
styl.lint: $(MAGE)
	@$(MAGE) styl:lint

QUALITY_RULES += styl.lint
