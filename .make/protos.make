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

ALL_FILES ?= (git ls-files . && git ls-files . --exclude-standard --others) | grep -v node_modules | sed 's:^:./:'
PROTO_FILES ?= $(ALL_FILES) | grep "\.proto$$"

# Go
GO_PROTO_TARGETS ?= $(patsubst %.proto,%.pb.go,$(shell $(PROTO_FILES)))
GO_PROTO_TYPES = any duration empty struct timestamp
GO_PROTO_TYPE_CONVERSIONS = $(subst $(SPACE),$(COMMA),$(foreach type,$(GO_PROTO_TYPES),Mgoogle/protobuf/$(type).proto=$(GOGO_REPO)/types))
GO_PROTOC_FLAGS ?= \
	--gogottn_out=plugins=grpc,$(GO_PROTO_TYPE_CONVERSIONS):$(GO_PATH)/src \
	--grpc-gateway_out=:$(GO_PATH)/src

%.pb.go: %.proto protoc
	$(PROTOC) $(GO_PROTOC_FLAGS) $(PWD)/$<

go.protos: $(GO_PROTO_TARGETS)

# vim: ft=make
