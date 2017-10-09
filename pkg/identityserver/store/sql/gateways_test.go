// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func testGateways() map[string]*ttnpb.Gateway {
	return map[string]*ttnpb.Gateway{
		"test-gateway": &ttnpb.Gateway{
			GatewayIdentifier: ttnpb.GatewayIdentifier{"test-gateway"},
			Description:       "My description",
			FrequencyPlanID:   "868_3",
			Token:             "1111",
			Platform:          "Kerklink",
			Attributes: map[string]string{
				"foo": "bar",
			},
			Antennas: []ttnpb.GatewayAntenna{
				ttnpb.GatewayAntenna{
					Location: ttnpb.Location{
						Latitude:  11.11,
						Longitude: 22.22,
						Altitude:  10,
					},
				},
			},
		},
		"bob-gateway": &ttnpb.Gateway{
			GatewayIdentifier: ttnpb.GatewayIdentifier{"bob-gateway"},
			Description:       "My description",
			FrequencyPlanID:   "868_3",
			Token:             "1111",
			ClusterAddress:    "network.eu",
			Attributes: map[string]string{
				"Modulation": "12345",
				"RFCH":       "111",
			},
			Antennas: []ttnpb.GatewayAntenna{
				ttnpb.GatewayAntenna{
					Gain: 12.22},
			},
		},
	}
}

func TestGatewayCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gateways := testGateways()

	for _, gtw := range gateways {
		err := s.Gateways.Create(gtw)
		a.So(err, should.BeNil)
	}

	// Attempt to recreate them should throw an error
	for _, gtw := range gateways {
		err := s.Gateways.Create(gtw)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 301)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestGatewayAttributes(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gtw := testGateways()["bob-gateway"]

	// fetch gateway and check that the attributes has been registered
	{
		found, err := s.Gateways.GetByID(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	attributeKey := "Foo"

	// add attribute
	{
		gtw.Attributes[attributeKey] = "bar"
		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)
		found, err := s.Gateways.GetByID(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	// delete attribute
	{
		delete(gtw.Attributes, attributeKey)
		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)
		found, err := s.Gateways.GetByID(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)

	}
}

func TestGatewayAntennas(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gtw := testGateways()["bob-gateway"]

	// check that all antennas were registered
	{
		found, err := s.Gateways.GetByID(gtw.GatewayID)
		a.So(err, should.BeNil)
		if a.So(found.GetGateway().Antennas, should.HaveLength, 1) {
			a.So(found.GetGateway().Antennas[0], should.Resemble, gtw.Antennas[0])
			gtw = found.GetGateway()
		}
	}

	// add a new antenna
	{
		gtw.Antennas = append(gtw.Antennas, ttnpb.GatewayAntenna{Gain: 12.12})
		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)

		found, err := s.Gateways.GetByID(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	// delete previously added antenna
	{
		gtw.Antennas = gtw.Antennas[:1]

		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)

		found, err := s.Gateways.GetByID(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}
}

func TestGatewayCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	user := testUsers()["bob"]
	gtw := testGateways()["bob-gateway"]

	// check indeed that the gateway has no collaborator
	{
		collaborators, err := s.Gateways.ListCollaborators(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}

	collaborator := ttnpb.Collaborator{
		UserIdentifier: ttnpb.UserIdentifier{user.UserID},
		Rights: []ttnpb.Right{
			ttnpb.Right(1),
			ttnpb.Right(2),
		},
	}

	// set the collaborator
	{
		err := s.Gateways.SetCollaborator(gtw.GatewayID, collaborator)
		a.So(err, should.BeNil)

		collaborators, err := s.Gateways.ListCollaborators(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)

	}

	// fetch gateways where Bob is collaborator
	{
		gtws, err := s.Gateways.ListByUser(user.UserID)
		a.So(err, should.BeNil)
		if a.So(gtws, should.HaveLength, 1) {
			a.So(gtws[0], test.ShouldBeGatewayIgnoringAutoFields, gtw)
		}
	}

	// modify rights
	{
		collaborator.Rights = append(collaborator.Rights, ttnpb.Right(3))
		err := s.Gateways.SetCollaborator(gtw.GatewayID, collaborator)
		a.So(err, should.BeNil)

		collaborators, err := s.Gateways.ListCollaborators(gtw.GatewayID)
		a.So(err, should.BeNil)
		if a.So(collaborators, should.HaveLength, 1) {
			a.So(collaborators[0].Rights, should.Resemble, collaborator.Rights)
		}
	}

	// fetch user rights
	{
		rights, err := s.Gateways.ListUserRights(gtw.GatewayID, user.UserID)
		a.So(err, should.BeNil)
		if a.So(rights, should.HaveLength, 3) {
			a.So(rights, should.Resemble, collaborator.Rights)
		}
	}

	// remove collaborator
	{
		collaborator.Rights = []ttnpb.Right{}
		err := s.Gateways.SetCollaborator(gtw.GatewayID, collaborator)
		a.So(err, should.BeNil)

		collaborators, err := s.Gateways.ListCollaborators(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}

func TestGatewayUpdate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gtw := testGateways()["bob-gateway"]

	gtw.Description = "Fancy new description"
	err := s.Gateways.Update(gtw)
	a.So(err, should.BeNil)

	found, err := s.Gateways.GetByID(gtw.GatewayID)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
}

func TestGatewayArchive(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gtw := testGateways()["bob-gateway"]

	err := s.Gateways.Archive(gtw.GatewayID)
	a.So(err, should.BeNil)

	found, err := s.Gateways.GetByID(gtw.GatewayID)
	a.So(err, should.BeNil)

	a.So(found.GetGateway().ArchivedAt.IsZero(), should.BeFalse)
}
