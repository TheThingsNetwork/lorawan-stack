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

.PHONY: git.hooks

git.hooks: $(MAGE)
	@$(MAGE) git:installHooks

INIT_RULES += git.hooks

.PHONY: git.hooks.remove

git.hooks.remove: $(MAGE)
	@$(MAGE) git:uninstallHooks

.PHONY: git.pre-commit

git.pre-commit: $(MAGE)
	@HOOK=pre-commit $(MAGE) git:runHook

.PHONY: git.commit-msg

git.commit-msg: $(MAGE)
	@HOOK=commit-msg $(MAGE) git:runHook

.PHONY: git.pre-push

git.pre-push: $(MAGE)
	@HOOK=pre-push $(MAGE) git:runHook

.PHONY: go.min-version
go.min-version: $(MAGE)
	@$(MAGE) go:checkVersion

.PHONY: go.deps
go.deps:
	@GO111MODULE=on go mod vendor

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

.PHONY: go.test
go.test: $(MAGE) dev.certs
	@$(MAGE) go:test

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

.PHONY: js.build-dll
js.build-dll: $(MAGE)
	@$(MAGE) js:buildDll

.PHONY: js.build-main
js.build-main: $(MAGE)
	@$(MAGE) js:buildMain

.PHONY: js.clean
js.clean: $(MAGE)
	@$(MAGE) js:clean

.PHONY: js.deps
js.deps: $(MAGE)
	@$(MAGE) js:deps

.PHONY: js.dev-deps
js.dev-deps: $(MAGE)
	@$(MAGE) js:devDeps

.PHONY: js.lint
js.lint: $(MAGE)
	@$(MAGE) js:lint

.PHONY: js.lint-all
js.lint-all: $(MAGE)
	@$(MAGE) js:lintAll

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

.PHONY: proto.sdk.js.clean
proto.sdk.js.clean: $(MAGE)
	@$(MAGE) proto:sdkJsClean

.PHONY: proto.swagger
proto.swagger: $(MAGE)
	@$(MAGE) proto:swagger

.PHONY: proto.swagger.clean
proto.swagger.clean: $(MAGE)
	@$(MAGE) proto:swaggerClean

.PHONY: sdk.js.build
sdk.js.build: $(MAGE)
	@$(MAGE) sdkJs:build

.PHONY: sdk.js.clean
sdk.js.clean: $(MAGE)
	@$(MAGE) sdkJs:clean

.PHONY: sdk.js.deps
sdk.js.deps: $(MAGE)
	@$(MAGE) sdkJs:deps

.PHONY: sdk.js.dev-deps
sdk.js.dev-deps: $(MAGE)
	@$(MAGE) sdkJs:devDeps

.PHONY: sdk.js.test
sdk.js.test: $(MAGE)
	@$(MAGE) sdkJs:test

.PHONY: sdk.js.test-watch
sdk.js.test-watch: $(MAGE)
	@$(MAGE) sdkJs:testWatch

.PHONY: sdk.js.watch
sdk.js.watch: $(MAGE)
	@$(MAGE) sdkJs:watch

.PHONY: sdk.js.definitions
sdk.js.definitions: $(MAGE)
	@$(MAGE) sdkJs:definitions

.PHONY: styl.lint
styl.lint: $(MAGE)
	@$(MAGE) styl:lint
