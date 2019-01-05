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

package ttnpb

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestUserIdentifiersValidate(t *testing.T) {
	a := assertions.New(t)

	ids := UserIdentifiers{
		UserID: "foo",
		Email:  "foo@bar.com",
	}
	a.So(ids.Validate(), should.BeNil)

	ids = UserIdentifiers{
		UserID: "foo",
	}
	a.So(ids.Validate(), should.BeNil)

	ids = UserIdentifiers{
		Email: "foo@bar.com",
	}
	a.So(ids.Validate(), should.NotBeNil)

	ids = UserIdentifiers{}
	err := ids.Validate()
	a.So(err, should.NotBeNil)

	ids = UserIdentifiers{
		UserID: "foo",
		Email:  "foobar.com",
	}
	a.So(ids.Validate(), should.BeNil)

	ids = UserIdentifiers{
		UserID: "_foo",
		Email:  "foo@bar.com",
	}
	a.So(ids.Validate(), should.NotBeNil)
}

func TestGatewayIdentifiersValidate(t *testing.T) {
	a := assertions.New(t)

	ids := GatewayIdentifiers{
		GatewayID: "foo-gtw",
		EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.Validate(), should.BeNil)

	ids = GatewayIdentifiers{
		GatewayID: "foo-gtw",
	}
	a.So(ids.Validate(), should.BeNil)

	ids = GatewayIdentifiers{
		EUI: &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.Validate(), should.NotBeNil)

	ids = GatewayIdentifiers{}
	err := ids.Validate()
	a.So(err, should.NotBeNil)

	ids = GatewayIdentifiers{
		GatewayID: "_foo-gtw",
		EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.Validate(), should.NotBeNil)

	ids = GatewayIdentifiers{
		GatewayID: "foo-gtw",
		EUI:       &types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	a.So(ids.Validate(), should.BeNil)

	ids = GatewayIdentifiers{
		EUI: new(types.EUI64),
	}
	a.So(ids.Validate(), should.NotBeNil)
}
