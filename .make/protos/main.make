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

DOCKER ?= docker

API_PROTO_FILES = $(PWD)/api/'*.proto'

PROTOC_OUT ?= /out

PROTOC_DOCKER_IMAGE ?= thethingsindustries/protoc:3.1.3
PROTOC_DOCKER_ARGS = run --user `id -u` --rm \
                     --mount type=bind,src=$(PWD)/api,dst=$(PWD)/api \
                     --mount type=bind,src=$(PWD)/pkg/ttnpb,dst=$(PROTOC_OUT)/go.thethings.network/lorawan-stack/pkg/ttnpb \
                     --mount type=bind,src=$(PWD)/sdk/js,dst=$(PWD)/sdk/js \
                     -w $(PWD)
PROTOC ?= $(DOCKER) $(PROTOC_DOCKER_ARGS) $(PROTOC_DOCKER_IMAGE) -I$(shell dirname $(PWD))

protoc:
	$(DOCKER) pull $(PROTOC_DOCKER_IMAGE)

SWAGGER_PROTOC_FLAGS ?= --swagger_out=allow_merge,merge_file_name=api:$(PWD)/api

swagger.protos: $(wildcard api/*.proto)
	$(PROTOC) $(SWAGGER_PROTOC_FLAGS) $(API_PROTO_FILES)

swagger.protos.clean:
	rm -f $(PWD)/api/api.swagger.json

MARKDOWN_PROTOC_FLAGS ?= --doc_opt=$(PWD)/api/api.md.tmpl,api.md --doc_out=$(PWD)/api

markdown.protos: $(wildcard api/*.proto)
	$(PROTOC) $(MARKDOWN_PROTOC_FLAGS) $(API_PROTO_FILES)

markdown.protos.clean:
	rm -f $(PWD)/api/api.md

include .make/protos/go.make

protos: go.protos swagger.protos markdown.protos sdk.js.protos

protos.clean: go.protos.clean swagger.protos.clean markdown.protos.clean sdk.js.protos.clean

# vim: ft=make
