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

DOCKER ?= docker

PROTOC_OUT ?= /out

PROTOC_DOCKER_IMAGE ?= thethingsindustries/protoc:3.0.9
PROTOC_DOCKER_ARGS = run --user `id -u` --rm \
                     --mount type=bind,src=$(PWD)/api,dst=$(PWD)/api,ro=true \
                     --mount type=bind,src=$(PWD)/pkg/ttnpb,dst=$(PROTOC_OUT)/go.thethings.network/lorawan-stack/pkg/ttnpb \
                     -w $(PWD)
PROTOC ?= $(DOCKER) $(PROTOC_DOCKER_ARGS) $(PROTOC_DOCKER_IMAGE) -I$(shell dirname $(PWD))

protoc:
	$(DOCKER) pull $(PROTOC_DOCKER_IMAGE)

include .make/protos/go.make

# vim: ft=make
