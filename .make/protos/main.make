# Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

GOGO_REPO=github.com/gogo/protobuf
GRPC_GATEWAY_REPO=github.com/grpc-ecosystem/grpc-gateway

DOCKER ?= docker

PROTOC_VERSION ?= 3.0.4
PROTOC_DOCKER_IMAGE ?= thethingsindustries/protoc:$(PROTOC_VERSION)
PROTOC_DOCKER_ARGS = run --user `id -u` --rm -v$(GO_PATH):$(GO_PATH) -w`pwd`
PROTOC ?= $(DOCKER) $(PROTOC_DOCKER_ARGS) $(PROTOC_DOCKER_IMAGE) -I/usr/include
PROTOC += -I$(VENDOR_DIR) -I$(GOPATH)/src -I$(VENDOR_DIR)/$(GRPC_GATEWAY_REPO)/third_party/googleapis

protoc:
	$(DOCKER) pull $(DOCKER_IMAGE)

PROTO_DIR=$(PWD)/api

include .make/protos/go.make

# vim: ft=make
