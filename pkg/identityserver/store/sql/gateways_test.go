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

package sql_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	. "github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var gatewaySpecializer = func(base ttnpb.Gateway) store.Gateway {
	return &base
}

func TestGateways(t *testing.T) {
	for _, tc := range []struct {
		tcname string
		ids    ttnpb.GatewayIdentifiers
	}{
		{
			"UsingGatewayID",
			ttnpb.GatewayIdentifiers{
				GatewayID: "test-gateway",
			},
		},
		{
			"UsingEUI",
			ttnpb.GatewayIdentifiers{
				EUI: &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
			},
		},
		{
			"UsingAllIdentifiers",
			ttnpb.GatewayIdentifiers{
				GatewayID: "test-gateway",
				EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
			},
		},
	} {
		t.Run(tc.tcname, func(t *testing.T) {
			testGateways(t, tc.ids)
		})
	}
}

func testGateways(t *testing.T, ids ttnpb.GatewayIdentifiers) {
	a := assertions.New(t)
	s := testStore(t, database)

	gateway := &ttnpb.Gateway{
		GatewayIdentifiers: ids,
		Description:        "My description",
		Platform:           "Kerklink",
		DisableTxDelay:     true,
		Attributes: map[string]string{
			"foo": "bar",
		},
		FrequencyPlanID: "868_3",
		Radios: []ttnpb.GatewayRadio{
			{
				Frequency: 10,
			},
		},
		ContactAccountIDs: &ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &alice.UserIdentifiers}},
		Antennas: []ttnpb.GatewayAntenna{
			{
				Location: ttnpb.Location{
					Latitude:  11.11,
					Longitude: 22.22,
					Altitude:  10,
				},
			},
		},
	}

	err := s.Gateways.Create(gateway)
	a.So(err, should.BeNil)

	err = s.Gateways.Create(gateway)
	a.So(err, should.NotBeNil)
	a.So(ErrGatewayIDTaken.Describes(err), should.BeTrue)

	found, err := s.Gateways.GetByID(gateway.GatewayIdentifiers, gatewaySpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gateway)

	gateway.Description = ""
	err = s.Gateways.Update(gateway)
	a.So(err, should.BeNil)

	found, err = s.Gateways.GetByID(gateway.GatewayIdentifiers, gatewaySpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gateway)

	collaborator := ttnpb.GatewayCollaborator{
		GatewayIdentifiers:            gateway.GatewayIdentifiers,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &alice.UserIdentifiers}},
		Rights: []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO},
	}

	collaborators, err := s.Gateways.ListCollaborators(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 0)

	err = s.Gateways.SetCollaborator(collaborator)
	a.So(err, should.BeNil)

	collaborators, err = s.Gateways.ListCollaborators(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 1)
	a.So(collaborators, should.Contain, collaborator)

	collaborators, err = s.Gateways.ListCollaborators(gateway.GatewayIdentifiers, ttnpb.Right(0))
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 0)

	collaborators, err = s.Gateways.ListCollaborators(gateway.GatewayIdentifiers, collaborator.Rights...)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 1)
	a.So(collaborators, should.Contain, collaborator)

	rights, err := s.Gateways.ListCollaboratorRights(gateway.GatewayIdentifiers, collaborator.OrganizationOrUserIdentifiers)
	a.So(err, should.BeNil)
	a.So(rights, should.Resemble, collaborator.Rights)

	has, err := s.Gateways.HasCollaboratorRights(gateway.GatewayIdentifiers, collaborator.OrganizationOrUserIdentifiers, ttnpb.Right(0))
	a.So(err, should.BeNil)
	a.So(has, should.BeFalse)

	has, err = s.Gateways.HasCollaboratorRights(gateway.GatewayIdentifiers, collaborator.OrganizationOrUserIdentifiers, collaborator.Rights...)
	a.So(err, should.BeNil)
	a.So(has, should.BeTrue)

	has, err = s.Gateways.HasCollaboratorRights(gateway.GatewayIdentifiers, collaborator.OrganizationOrUserIdentifiers, ttnpb.RIGHT_GATEWAY_INFO, ttnpb.Right(0))
	a.So(err, should.BeNil)
	a.So(has, should.BeFalse)

	collaborator.Rights = []ttnpb.Right{ttnpb.RIGHT_GATEWAY_LOCATION}

	err = s.Gateways.SetCollaborator(collaborator)
	a.So(err, should.BeNil)

	collaborators, err = s.Gateways.ListCollaborators(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 1)
	a.So(collaborators, should.Contain, collaborator)

	collaborator.Rights = []ttnpb.Right{}

	err = s.Gateways.SetCollaborator(collaborator)
	a.So(err, should.BeNil)

	collaborators, err = s.Gateways.ListCollaborators(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 0)

	key := ttnpb.APIKey{
		Name:   "foo",
		Key:    "bar",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	keys, err := s.Gateways.ListAPIKeys(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(keys, should.HaveLength, 0)

	err = s.Gateways.SaveAPIKey(gateway.GatewayIdentifiers, key)
	a.So(err, should.BeNil)

	keys, err = s.Gateways.ListAPIKeys(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)
	if a.So(keys, should.HaveLength, 1) {
		a.So(keys, should.Contain, key)
	}

	key.Rights = []ttnpb.Right{ttnpb.Right(1)}
	err = s.Gateways.UpdateAPIKeyRights(gateway.GatewayIdentifiers, key.Name, key.Rights)
	a.So(err, should.BeNil)

	ids, foundKey, err := s.Gateways.GetAPIKey(key.Key)
	a.So(err, should.BeNil)
	a.So(ids, should.Resemble, gateway.GatewayIdentifiers)
	a.So(foundKey, should.Resemble, key)

	foundKey, err = s.Gateways.GetAPIKeyByName(gateway.GatewayIdentifiers, key.Name)
	a.So(err, should.BeNil)
	a.So(foundKey, should.Resemble, key)

	err = s.Gateways.DeleteAPIKey(gateway.GatewayIdentifiers, key.Name)
	a.So(err, should.BeNil)

	keys, err = s.Gateways.ListAPIKeys(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(keys, should.HaveLength, 0)

	// Save the API Key again. Calling `Delete` afterwards will handle it.
	err = s.Gateways.SaveAPIKey(gateway.GatewayIdentifiers, key)
	a.So(err, should.BeNil)

	err = s.Gateways.Delete(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)

	_, err = s.Gateways.GetByID(gateway.GatewayIdentifiers, gatewaySpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrGatewayNotFound.Describes(err), should.BeTrue)
}
