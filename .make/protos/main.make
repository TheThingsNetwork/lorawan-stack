# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

GOGO_REPO=github.com/gogo/protobuf
GRPC_GATEWAY_REPO=github.com/grpc-ecosystem/grpc-gateway

DOCKER ?= docker
DOCKER_ARGS = run --user `id -u` --rm -v$(GO_PATH):$(GO_PATH) -w`pwd`
DOCKER_IMAGE ?= thethingsindustries/protoc
PROTOC ?= $(DOCKER) $(DOCKER_ARGS) $(DOCKER_IMAGE) -I/usr/include
PROTOC += -I$(VENDOR_DIR) -I$(GOPATH)/src -I$(VENDOR_DIR)/$(GRPC_GATEWAY_REPO)/third_party/googleapis

protoc:
	$(DOCKER) pull $(DOCKER_IMAGE)

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
COMMA := ,
SED = $(shell command -v gsed || command -v sed)

PROTO_DIR=$(PWD)/api
PROTO_OUT=$(PWD)/pkg/ttnpb

include .make/protos/go.make

# vim: ft=make
