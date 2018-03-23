// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb_test

import (
	"regexp"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var idRegexp = regexp.MustCompile("^[0-9a-z](?:[_-]?[0-9a-z]){1,35}$")

func TestNewPopulatedEndDeviceIdentifiers(t *testing.T) {
	id := NewPopulatedEndDeviceIdentifiers(test.Randy, false)
	assertions.New(t).So(id.DeviceID == "" || idRegexp.MatchString(id.DeviceID), should.BeTrue)
	assertions.New(t).So(id.ApplicationID == "" || idRegexp.MatchString(id.ApplicationID), should.BeTrue)
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
	a.So(ids.Validate(), should.BeNil)

	ids = GatewayIdentifiers{}
	a.So(ids.Validate(), should.NotBeNil)
	a.So(ErrEmptyIdentifiers.Describes(ids.Validate()), should.BeTrue)

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
	a.So(ids.Validate(), should.BeNil)
}

func TestGatewayIdentifiersIsZero(t *testing.T) {
	a := assertions.New(t)

	ids := GatewayIdentifiers{
		GatewayID: "foo-gtw",
		EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.IsZero(), should.BeFalse)

	ids = GatewayIdentifiers{
		GatewayID: "foo-gtw",
	}
	a.So(ids.IsZero(), should.BeFalse)

	ids = GatewayIdentifiers{
		EUI: &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.IsZero(), should.BeFalse)

	ids = GatewayIdentifiers{
		EUI: &types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	a.So(ids.IsZero(), should.BeFalse)

	ids = GatewayIdentifiers{}
	a.So(ids.IsZero(), should.BeTrue)
}

func TestGatewayIdentifiersContains(t *testing.T) {
	a := assertions.New(t)

	ids := GatewayIdentifiers{
		GatewayID: "foo-gtw",
		EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.Contains(GatewayIdentifiers{}), should.BeFalse)
	a.So(ids.Contains(GatewayIdentifiers{GatewayID: "foo-gtw"}), should.BeTrue)
	a.So(ids.Contains(GatewayIdentifiers{EUI: &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42}}), should.BeTrue)
	a.So(ids.Contains(ids), should.BeTrue)
	a.So(ids.Contains(GatewayIdentifiers{GatewayID: "bar"}), should.BeFalse)

	ids = GatewayIdentifiers{
		EUI: &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.Contains(GatewayIdentifiers{}), should.BeFalse)
	a.So(ids.Contains(GatewayIdentifiers{EUI: &types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}), should.BeFalse)
	a.So(ids.Contains(GatewayIdentifiers{GatewayID: "bar"}), should.BeFalse)
	a.So(ids.Contains(GatewayIdentifiers{EUI: &types.EUI64{0x99, 0x99, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42}}), should.BeFalse)
	a.So(ids.Contains(ids), should.BeTrue)
}
