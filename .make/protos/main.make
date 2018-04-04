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

GOGO_REPO=github.com/gogo/protobuf
GRPC_GATEWAY_REPO=github.com/grpc-ecosystem/grpc-gateway

DOCKER ?= docker

PROTOC_VERSION ?= 3.0.5
PROTOC_DOCKER_IMAGE ?= thethingsindustries/protoc:$(PROTOC_VERSION)
PROTOC_DOCKER_ARGS = run --user `id -u` --rm -v$(GO_PATH):$(GO_PATH) -w`pwd`
PROTOC ?= $(DOCKER) $(PROTOC_DOCKER_ARGS) $(PROTOC_DOCKER_IMAGE) -I/usr/include
PROTOC += -I$(VENDOR_DIR) -I$(GOPATH)/src -I$(VENDOR_DIR)/$(GRPC_GATEWAY_REPO)/third_party/googleapis

protoc:
	$(DOCKER) pull $(PROTOC_DOCKER_IMAGE)

PROTO_DIR=$(PWD)/api

include .make/protos/go.make

# vim: ft=make
