// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	. "github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var gatewaySpecializer = func(base ttnpb.Gateway) store.Gateway {
	return &base
}

func TestGateways(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gateway := &ttnpb.Gateway{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
		Description:        "My description",
		Platform:           "Kerklink",
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

	// save it again. Call to `Delete` will handle it
	err = s.Gateways.SaveAPIKey(gateway.GatewayIdentifiers, key)
	a.So(err, should.BeNil)

	err = s.Gateways.Delete(gateway.GatewayIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Gateways.GetByID(gateway.GatewayIdentifiers, gatewaySpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrGatewayNotFound.Describes(err), should.BeTrue)
}
