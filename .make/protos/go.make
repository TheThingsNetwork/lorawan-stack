# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# Go
GO_PROTO_TYPES = any duration empty struct timestamp field_mask
GO_PROTO_TYPE_CONVERSIONS = $(subst $(SPACE),$(COMMA),$(foreach type,$(GO_PROTO_TYPES),Mgoogle/protobuf/$(type).proto=$(GOGO_REPO)/types))
GO_PROTOC_FLAGS ?= \
	--gogottn_out=plugins=grpc,$(GO_PROTO_TYPE_CONVERSIONS):$(GO_PATH)/src \
	--grpc-gateway_out=:$(GO_PATH)/src

go.protos: $(wildcard $(PROTO_DIR)/*.proto)
	$(PROTOC) $(GO_PROTOC_FLAGS) $(PROTO_DIR)/*.proto
	$(MAKE_DIR)/protos/fix-grpc-gateway-names.sh $(PROTO_DIR)
	sed -i'' 's:golang.org/x/net/context:context:' `find ./pkg -name '*pb.go' | grep -v 'vendor'`

go.protos.clean:
	find ./pkg/ttnpb -name '*.pb.go' -delete -or -name '*.pb.gw.go' -delete -or -name '*pb_test.go' -delete

# vim: ft=make
