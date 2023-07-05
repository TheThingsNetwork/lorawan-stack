// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package provisioning_test

import (
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/provisioning"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMicrochip(t *testing.T) {
	a := assertions.New(t)

	provisioner := Get(Microchip)
	if !a.So(provisioner, should.NotBeNil) {
		t.FailNow()
	}

	entry := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"uniqueId": {
				Kind: &structpb.Value_StringValue{
					StringValue: "abcd",
				},
			},
		},
	}

	uniqueID, err := provisioner.UniqueID(entry)
	a.So(err, should.BeNil)
	a.So(uniqueID, should.Equal, "ABCD")
}
