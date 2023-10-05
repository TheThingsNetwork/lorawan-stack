// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package apitest contains common API definition test utilities.
package apitest

import (
	"fmt"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// RunTestFieldMaskSpecified runs a test that checks whether all RPCs have their allowed field mask paths set.
func RunTestFieldMaskSpecified(t *testing.T, pkg protoreflect.FullName, paths map[string]ttnpb.RPCFieldMaskPathValue) {
	t.Helper()
	a := assertions.New(t)
	protoregistry.GlobalFiles.RangeFilesByPackage(pkg, func(fd protoreflect.FileDescriptor) bool {
		t.Helper()
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			methods := services.Get(i).Methods()
			for j := 0; j < methods.Len(); j++ {
				method := methods.Get(j)
				fields := method.Input().Fields()
				for k := 0; k < fields.Len(); k++ {
					field := fields.Get(k)
					if field.Name() != "field_mask" {
						continue
					}
					message := field.Message()
					if message == nil || message.FullName() != "google.protobuf.FieldMask" {
						continue
					}
					a.So(
						paths,
						should.ContainKey,
						fmt.Sprintf("/%s/%s", method.FullName().Parent(), method.Name()),
					)
				}
			}
		}
		return true
	})
}
