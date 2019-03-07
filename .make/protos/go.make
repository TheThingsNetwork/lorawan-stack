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

# Go
GO_PROTO_TYPES := any duration empty field_mask struct timestamp wrappers
GO_PROTO_TYPE_CONVERSIONS = $(subst $(SPACE),$(COMMA),$(foreach type,$(GO_PROTO_TYPES),Mgoogle/protobuf/$(type).proto=github.com/gogo/protobuf/types))
GO_PROTOC_FLAGS ?= \
	--fieldmask_out=lang=gogo,$(GO_PROTO_TYPE_CONVERSIONS):$(PROTOC_OUT) \
	--gogottn_out=plugins=grpc,$(GO_PROTO_TYPE_CONVERSIONS):$(PROTOC_OUT) \
	--grpc-gateway_out=$(GO_PROTO_TYPE_CONVERSIONS):$(PROTOC_OUT)

go.protos: $(wildcard api/*.proto)
	$(PROTOC) $(GO_PROTOC_FLAGS) $(API_PROTO_FILES) 2>&1 | grep -vE ' protoc-gen-gogo: WARNING: failed finding publicly imported dependency for \.ttn\.lorawan\.v3\..* used in' || true
	$(MAKE_DIR)/protos/fix-grpc-gateway-names.sh api
	perl -i -pe 's:golang.org/x/net/context:context:' `find ./pkg -name '*pb.go' -or -name '*pb.gw.go' | grep -v 'vendor'`
	GO111MODULE=on $(GO) run golang.org/x/tools/cmd/goimports -w $(PWD)/pkg/ttnpb
	GO111MODULE=on $(GO) run github.com/mdempsky/unconvert -apply ./pkg/ttnpb
	gofmt -w -s $(PWD)/pkg/ttnpb

go.protos.clean:
	find ./pkg/ttnpb -name '*.pb.go' -delete -or -name '*.pb.gw.go' -delete -or -name '*.pb.fm.go' -delete -or -name '*.pb.util.fm.go' -delete

# vim: ft=make
