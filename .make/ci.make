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

ci.js.lint.simple:
	@$(MAKE) js.dev-deps js.deps sdk.deps
	@$(MAKE) js.lint
	@$(MAKE) git.diff
	
ci.js.lint.full: ci.js.lint.simple
	@$(MAKE) go.dev-deps go.deps
	@$(MAKE) translations
	@$(MAKE) git.diff
	
ci.js.test.simple:
	@$(MAKE) js.dev-deps js.deps sdk.deps
	@$(MAKE) js.test sdk.test
	
ci.js.test.full: ci.js.test.simple
	
ci.go.lint.simple:
	@$(MAKE) go.dev-deps go.deps
	@$(MAKE) go.misspell
	@$(MAKE) go.fmt
	@$(MAKE) go.depfmt
	@$(MAKE) headers.check
	@$(MAKE) git.diff
	
ci.go.lint.full: ci.go.lint.simple
	@$(MAKE) protos.clean protos
	@$(MAKE) go.unconvert
	@$(MAKE) go.lint-travis || true
	@$(MAKE) git.diff
  
ci.js.simple: ci.js.lint.simple ci.js.test.simple

ci.js.full: ci.js.lint.full ci.js.test.full
  
ci.go.test.pre:
	@$(MAKE) go.dev-deps go.deps
	@$(MAKE) dev.certs
	@$(MAKE) dev.databases.start
  
ci.go.test.amd64.simple: ci.go.test.pre
	@GOARCH=amd64 TEST_SLOWDOWN=8 TEST_REDIS=1 $(MAKE) go.test
	
ci.go.test.amd64.full: ci.go.test.amd64.simple

ci.go.test.386.simple:

ci.go.test.386.full: ci.go.test.pre
	@GOARCH=386 TEST_SLOWDOWN=8 TEST_REDIS=1 $(MAKE) go.test
   
ci.build-all.simple:

ci.build-all.full:
	@$(MAKE) dev-deps deps
	@$(MAKE) build-all
	