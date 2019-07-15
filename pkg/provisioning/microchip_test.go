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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/provisioning"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestMicrochip(t *testing.T) {
	a := assertions.New(t)

	provisioner := Get(Microchip)
	if !a.So(provisioner, should.NotBeNil) {
		t.FailNow()
	}

	entry := &pbtypes.Struct{
		Fields: map[string]*pbtypes.Value{
			"uniqueId": {
				Kind: &pbtypes.Value_StringValue{
					StringValue: "abcd",
				},
			},
		},
	}

	uniqueID, err := provisioner.UniqueID(entry)
	a.So(err, should.BeNil)
	a.So(uniqueID, should.Equal, "ABCD")
}
