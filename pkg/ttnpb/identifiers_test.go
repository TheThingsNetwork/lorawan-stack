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
	. "go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
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
		ApplicationIdentifiers{ApplicationID: "foo"},
		ClientIdentifiers{ClientID: "foo"},
		EndDeviceIdentifiers{ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo"}, DeviceID: "foo"},
		EndDeviceIdentifiers{JoinEUI: &eui, DevEUI: &eui},
		EndDeviceIdentifiers{DevAddr: &devAddr},
		GatewayIdentifiers{GatewayID: "foo"},
		GatewayIdentifiers{EUI: &eui},
		OrganizationIdentifiers{OrganizationID: "foo"},
		UserIdentifiers{UserID: "foo"},
		UserIdentifiers{Email: "foo@example.com"},
	} {
		a.So(ids.IsZero(), should.BeFalse)
	}
}
func TestCombinedIdentifiers(t *testing.T) {
	a := assertions.New(t)

	for _, msg := range []interface{ CombinedIdentifiers() *CombinedIdentifiers }{
		NewPopulatedApplicationIdentifiers(test.Randy, true),
		NewPopulatedClientIdentifiers(test.Randy, true),
		NewPopulatedEndDeviceIdentifiers(test.Randy, true),
		NewPopulatedGatewayIdentifiers(test.Randy, true),
		NewPopulatedOrganizationIdentifiers(test.Randy, true),
		NewPopulatedUserIdentifiers(test.Randy, true),
		NewPopulatedUserSessionIdentifiers(test.Randy, true),
		NewPopulatedEntityIdentifiers(test.Randy, true),
		NewPopulatedCombinedIdentifiers(test.Randy, true),

		NewPopulatedCreateApplicationRequest(test.Randy, true),
		NewPopulatedCreateClientRequest(test.Randy, true),
		NewPopulatedCreateEndDeviceRequest(test.Randy, true),
		NewPopulatedCreateGatewayRequest(test.Randy, true),
		NewPopulatedCreateOrganizationRequest(test.Randy, true),
		NewPopulatedCreateUserRequest(test.Randy, true),

		NewPopulatedGetApplicationRequest(test.Randy, true),
		NewPopulatedGetClientRequest(test.Randy, true),
		NewPopulatedGetEndDeviceRequest(test.Randy, true),
		NewPopulatedGetGatewayRequest(test.Randy, true),
		NewPopulatedGetOrganizationRequest(test.Randy, true),
		NewPopulatedGetUserRequest(test.Randy, true),

		NewPopulatedListApplicationsRequest(test.Randy, true),
		NewPopulatedListClientsRequest(test.Randy, true),
		NewPopulatedListEndDevicesRequest(test.Randy, true),
		NewPopulatedListGatewaysRequest(test.Randy, true),
		NewPopulatedListOrganizationsRequest(test.Randy, true),
		NewPopulatedListUserSessionsRequest(test.Randy, true),

		NewPopulatedUpdateApplicationRequest(test.Randy, true),
		NewPopulatedUpdateClientRequest(test.Randy, true),
		NewPopulatedUpdateEndDeviceRequest(test.Randy, true),
		NewPopulatedUpdateGatewayRequest(test.Randy, true),
		NewPopulatedUpdateOrganizationRequest(test.Randy, true),
		NewPopulatedUpdateUserRequest(test.Randy, true),

		NewPopulatedSetEndDeviceRequest(test.Randy, true),

		NewPopulatedCreateApplicationAPIKeyRequest(test.Randy, true),
		NewPopulatedCreateGatewayAPIKeyRequest(test.Randy, true),
		NewPopulatedCreateOrganizationAPIKeyRequest(test.Randy, true),
		NewPopulatedCreateUserAPIKeyRequest(test.Randy, true),

		NewPopulatedUpdateApplicationAPIKeyRequest(test.Randy, true),
		NewPopulatedUpdateGatewayAPIKeyRequest(test.Randy, true),
		NewPopulatedUpdateOrganizationAPIKeyRequest(test.Randy, true),
		NewPopulatedUpdateUserAPIKeyRequest(test.Randy, true),

		NewPopulatedSetApplicationCollaboratorRequest(test.Randy, true),
		NewPopulatedSetClientCollaboratorRequest(test.Randy, true),
		NewPopulatedSetGatewayCollaboratorRequest(test.Randy, true),
		NewPopulatedSetOrganizationCollaboratorRequest(test.Randy, true),

		NewPopulatedCreateTemporaryPasswordRequest(test.Randy, true),
		NewPopulatedUpdateUserPasswordRequest(test.Randy, true),

		NewPopulatedPullGatewayConfigurationRequest(test.Randy, true),

		NewPopulatedDownlinkQueueRequest(test.Randy, true),

		NewPopulatedGetApplicationLinkRequest(test.Randy, true),
		NewPopulatedSetApplicationLinkRequest(test.Randy, true),

		NewPopulatedProcessDownlinkMessageRequest(test.Randy, true),
		NewPopulatedProcessUplinkMessageRequest(test.Randy, true),

		NewPopulatedListOAuthAccessTokensRequest(test.Randy, true),
		NewPopulatedListOAuthClientAuthorizationsRequest(test.Randy, true),
		NewPopulatedOAuthAccessTokenIdentifiers(test.Randy, true),
		NewPopulatedOAuthClientAuthorizationIdentifiers(test.Randy, true),

		NewPopulatedStreamEventsRequest(test.Randy, true),
	} {
		combined := msg.CombinedIdentifiers()
		a.So(combined, should.NotBeNil)
	}
}

func TestOrganizationOrUserIdentifiers(t *testing.T) {
	a := assertions.New(t)

	usrID := NewPopulatedUserIdentifiers(test.Randy, true)
	ouID := usrID.OrganizationOrUserIdentifiers()
	a.So(ouID, should.NotBeNil)
	a.So(ouID.Identifiers(), should.Resemble, usrID)

	orgID := NewPopulatedOrganizationIdentifiers(test.Randy, true)
	ouID = orgID.OrganizationOrUserIdentifiers()
	a.So(ouID, should.NotBeNil)
	a.So(ouID.Identifiers(), should.Resemble, orgID)
}

func TestEntityIdentifiers(t *testing.T) {
	a := assertions.New(t)

	appID := NewPopulatedApplicationIdentifiers(test.Randy, true)
	eID := appID.EntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.Identifiers(), should.Resemble, appID)

	cliID := NewPopulatedClientIdentifiers(test.Randy, true)
	eID = cliID.EntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.Identifiers(), should.Resemble, cliID)

	devID := NewPopulatedEndDeviceIdentifiers(test.Randy, true)
	eID = devID.EntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.Identifiers(), should.Resemble, devID)

	gtwID := NewPopulatedGatewayIdentifiers(test.Randy, true)
	eID = gtwID.EntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.Identifiers(), should.Resemble, gtwID)

	orgID := NewPopulatedOrganizationIdentifiers(test.Randy, true)
	eID = orgID.EntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.Identifiers(), should.Resemble, orgID)

	ouID := orgID.OrganizationOrUserIdentifiers()
	eID = ouID.EntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.Identifiers(), should.Resemble, orgID)

	usrID := NewPopulatedUserIdentifiers(test.Randy, true)
	eID = usrID.EntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.Identifiers(), should.Resemble, usrID)

	ouID = usrID.OrganizationOrUserIdentifiers()
	eID = ouID.EntityIdentifiers()
	a.So(eID, should.NotBeNil)
	a.So(eID.Identifiers(), should.Resemble, usrID)
}

func TestUserIdentifiersValidate(t *testing.T) {
	a := assertions.New(t)

	ids := UserIdentifiers{
		UserID: "foo",
		Email:  "foo@bar.com",
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = UserIdentifiers{
		UserID: "foo",
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
		UserID: "foo",
		Email:  "foobar.com",
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = UserIdentifiers{
		UserID: "_foo",
		Email:  "foo@bar.com",
	}
	a.So(ids.ValidateFields(), should.NotBeNil)
}

func TestGatewayIdentifiersValidate(t *testing.T) {
	a := assertions.New(t)

	ids := GatewayIdentifiers{
		GatewayID: "foo-gtw",
		EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = GatewayIdentifiers{
		GatewayID: "foo-gtw",
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = GatewayIdentifiers{
		EUI: &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.ValidateFields(), should.NotBeNil)

	ids = GatewayIdentifiers{}
	err := ids.ValidateFields()
	a.So(err, should.NotBeNil)

	ids = GatewayIdentifiers{
		GatewayID: "_foo-gtw",
		EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
	}
	a.So(ids.ValidateFields(), should.NotBeNil)

	ids = GatewayIdentifiers{
		GatewayID: "foo-gtw",
		EUI:       &types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	a.So(ids.ValidateFields(), should.BeNil)

	ids = GatewayIdentifiers{
		EUI: new(types.EUI64),
	}
	a.So(ids.ValidateFields(), should.NotBeNil)
}
