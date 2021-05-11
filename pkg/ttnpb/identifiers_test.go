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

package ttnpb_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestIdentifiersIsZero(t *testing.T) {
	a := assertions.New(t)

	for _, ids := range []interface{ IsZero() bool }{
		ApplicationIdentifiers{},
		&ApplicationIdentifiers{},
		ClientIdentifiers{},
		&ClientIdentifiers{},
		EndDeviceIdentifiers{},
		&EndDeviceIdentifiers{},
		GatewayIdentifiers{},
		&GatewayIdentifiers{},
		OrganizationIdentifiers{},
		&OrganizationIdentifiers{},
		UserIdentifiers{},
		&UserIdentifiers{},
	} {
		a.So(ids.IsZero(), should.BeTrue)
	}

	eui := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	devAddr := types.DevAddr{1, 2, 3, 4}

	for _, ids := range []interface{ IsZero() bool }{
		ApplicationIdentifiers{ApplicationId: "foo"},
		ClientIdentifiers{ClientId: "foo"},
		EndDeviceIdentifiers{ApplicationIdentifiers: ApplicationIdentifiers{ApplicationId: "foo"}, DeviceId: "foo"},
		EndDeviceIdentifiers{JoinEui: &eui, DevEui: &eui},
		EndDeviceIdentifiers{DevAddr: &devAddr},
		GatewayIdentifiers{GatewayId: "foo"},
		GatewayIdentifiers{Eui: &eui},
		OrganizationIdentifiers{OrganizationId: "foo"},
		UserIdentifiers{UserId: "foo"},
		UserIdentifiers{Email: "foo@example.com"},
	} {
		a.So(ids.IsZero(), should.BeFalse)
	}
}

func TestOrganizationOrUserIdentifiers(t *testing.T) {
	a := assertions.New(t)

	usrID := NewPopulatedUserIdentifiers(test.Randy, true)
	ouID := usrID.OrganizationOrUserIdentifiers()
	a.So(ouID, should.NotBeNil)
	a.So(ouID.GetUserIds(), should.Resemble, usrID)

	orgID := NewPopulatedOrganizationIdentifiers(test.Randy, true)
	ouID = orgID.OrganizationOrUserIdentifiers()
	a.So(ouID, should.NotBeNil)
	a.So(ouID.GetOrganizationIds(), should.Resemble, orgID)
}

func TestEntityIdentifiers(t *testing.T) {
	a := assertions.New(t)

	appID := NewPopulatedApplicationIdentifiers(test.Randy, true)
	eID := appID.GetEntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.GetApplicationIds(), should.Resemble, appID)

	cliID := NewPopulatedClientIdentifiers(test.Randy, true)
	eID = cliID.GetEntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.GetClientIds(), should.Resemble, cliID)

	devID := NewPopulatedEndDeviceIdentifiers(test.Randy, true)
	eID = devID.GetEntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.GetDeviceIds(), should.Resemble, devID)

	gtwID := NewPopulatedGatewayIdentifiers(test.Randy, true)
	eID = gtwID.GetEntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.GetGatewayIds(), should.Resemble, gtwID)

	orgID := NewPopulatedOrganizationIdentifiers(test.Randy, true)
	eID = orgID.GetEntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.GetOrganizationIds(), should.Resemble, orgID)

	ouID := orgID.OrganizationOrUserIdentifiers()
	eID = ouID.GetEntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.GetOrganizationIds(), should.Resemble, orgID)

	usrID := NewPopulatedUserIdentifiers(test.Randy, true)
	eID = usrID.GetEntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.GetUserIds(), should.Resemble, usrID)

	ouID = usrID.OrganizationOrUserIdentifiers()
	eID = ouID.GetEntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.GetUserIds(), should.Resemble, usrID)
}

func TestUserIdentifiersValidate(t *testing.T) {
	a := assertions.New(t)

	ids := UserIdentifiers{
		UserId: "foo",
		Email:  "foo@bar.com",
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = UserIdentifiers{
		UserId: "foo",
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = UserIdentifiers{
		Email: "foo@bar.com",
	}
	a.So(ids.ValidateFields(), should.NotBeNil)

	ids = UserIdentifiers{}
	err := ids.ValidateFields()
	a.So(err, should.NotBeNil)

	ids = UserIdentifiers{
		UserId: "foo",
		Email:  "foobar.com",
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = UserIdentifiers{
		UserId: "_foo",
		Email:  "foo@bar.com",
	}
	a.So(ids.ValidateFields(), should.NotBeNil)
}

func TestGatewayIdentifiersValidate(t *testing.T) {
	a := assertions.New(t)

	ids := GatewayIdentifiers{
		GatewayId: "foo-gtw",
		Eui:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = GatewayIdentifiers{
		GatewayId: "foo-gtw",
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = GatewayIdentifiers{
		Eui: &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.ValidateFields(), should.NotBeNil)

	ids = GatewayIdentifiers{}
	err := ids.ValidateFields()
	a.So(err, should.NotBeNil)

	ids = GatewayIdentifiers{
		GatewayId: "_foo-gtw",
		Eui:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.ValidateFields(), should.NotBeNil)

	ids = GatewayIdentifiers{
		GatewayId: "foo-gtw",
		Eui:       &types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = GatewayIdentifiers{
		Eui: new(types.EUI64),
	}
	a.So(ids.ValidateFields(), should.NotBeNil)
}
