// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/utils"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func testGateways() map[string]*types.DefaultGateway {
	return map[string]*types.DefaultGateway{
		"test-gateway": &types.DefaultGateway{
			ID:            "test-gateway",
			Description:   "My description",
			FrequencyPlan: "868_3",
			Key:           "1111",
			Brand:         utils.StringAddress("Kerklink"),
			Routers:       []string{"network.eu", "network.au"},
			Attributes: map[string]string{
				"foo": "bar",
			},
			Antennas: []types.GatewayAntenna{
				types.GatewayAntenna{
					ID: "test antenna",
					Location: &ttnpb.Location{
						Latitude:  11.11,
						Longitude: 22.22,
						Altitude:  10,
					},
				},
			},
		},
		"bob-gateway": &types.DefaultGateway{
			ID:            "bob-gateway",
			Description:   "My description",
			FrequencyPlan: "868_3",
			Key:           "1111",
			Routers:       []string{"network.eu", "network.au"},
			Attributes: map[string]string{
				"Modulation": "12345",
				"RFCH":       "111",
			},
			Antennas: []types.GatewayAntenna{
				types.GatewayAntenna{
					ID: "bobgw antenna",
				},
			},
		},
	}
}

func TestGatewayCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gateways := testGateways()

	for _, gtw := range gateways {
		created, err := s.Gateways.Create(gtw)
		a.So(err, should.BeNil)
		a.So(created, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	// recreate them should throw an error
	for _, gtw := range gateways {
		_, err := s.Gateways.Create(gtw)
		a.So(err, should.NotBeNil)
		a.So(err.Error(), should.Equal, ErrGatewayIDTaken.Error())
	}
}

func TestGatewayAntennas(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gtw := testGateways()["bob-gateway"]

	// fetch the gateway and check that the antenna has been registered with it
	{
		antennas, err := s.Gateways.Antennas(gtw.ID)
		a.So(err, should.BeNil)
		if a.So(antennas, should.HaveLength, 1) {
			a.So(antennas[0], should.Resemble, gtw.Antennas[0])
		}
	}

	ant := types.GatewayAntenna{
		ID:       "bobgw antenna",
		Location: &ttnpb.Location{Longitude: 12.12, Latitude: 10.02, Altitude: 2},
	}

	// modify antenna
	{
		err := s.Gateways.UpsertAntenna(gtw.ID, ant)
		a.So(err, should.BeNil)

		antennas, err := s.Gateways.Antennas(gtw.ID)
		a.So(err, should.BeNil)
		if a.So(antennas, should.HaveLength, 1) {
			a.So(antennas[0], should.Resemble, ant)
		}
	}

	// delete antenna
	{
		err := s.Gateways.DeleteAntenna(gtw.ID, ant.ID)
		a.So(err, should.BeNil)

		found, err := s.Gateways.FindByID(gtw.ID)
		a.So(err, should.BeNil)
		a.So(found.GetGateway().Antennas, should.HaveLength, 0)
	}
}

func TestGatewayCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gtw := testGateways()["bob-gateway"]
	user := testUsers()["alice"]
	rights := []types.Right{types.GatewayDeleteRight}

	collaborator := utils.Collaborator(user.Username, rights)

	// add collaborator
	{
		err := s.Gateways.AddCollaborator(gtw.ID, collaborator)
		a.So(err, should.BeNil)
	}

	// Find collaborators
	{
		collaborators, err := s.Gateways.Collaborators(gtw.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		if len(collaborators) > 0 {
			a.So(collaborators[0], should.Resemble, collaborator)
		}
	}

	right := types.GatewaySettingsRight

	// grant a right
	{
		err := s.Gateways.GrantRight(gtw.ID, user.Username, right)
		a.So(err, should.BeNil)

		rights, err := s.Gateways.UserRights(gtw.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 2)
		if len(rights) > 0 {
			a.So(rights, should.Contain, right)
		}
	}

	// revoke a right
	{
		err := s.Gateways.RevokeRight(gtw.ID, user.Username, right)
		a.So(err, should.BeNil)

		rights, err := s.Gateways.UserRights(gtw.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 1)
		if len(rights) > 0 {
			a.So(rights, should.NotContain, right)
		}
	}

	// fetch user rights
	{
		rights, err := s.Gateways.UserRights(gtw.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(collaborator.Rights, should.Resemble, rights)
	}

	// remove the collaborator
	{
		err := s.Gateways.RemoveCollaborator(gtw.ID, user.Username)
		a.So(err, should.BeNil)

		collaborators, err := s.Gateways.Collaborators(gtw.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}

func TestGatewayOwners(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gtw := testGateways()["test-gateway"]

	alice := testUsers()["alice"]
	bob := testUsers()["bob"]

	rights := []types.Right{types.GatewayOwnerRight}

	// add alice as owner
	{
		collaborator := utils.Collaborator(alice.Username, rights)
		err := s.Gateways.AddCollaborator(gtw.ID, collaborator)
		a.So(err, should.BeNil)
	}

	// add bob as owner
	{
		collaborator := utils.Collaborator(bob.Username, rights)
		err := s.Gateways.AddCollaborator(gtw.ID, collaborator)
		a.So(err, should.BeNil)
	}

	// fetch owners
	owners, err := s.Gateways.Owners(gtw.ID)
	a.So(err, should.BeNil)
	a.So(owners, should.HaveLength, 2)
	a.So(owners, should.Resemble, []string{"alice", "bob"})
}

func TestGatewayArchive(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gtw := testGateways()["bob-gateway"]

	err := s.Gateways.Archive(gtw.ID)
	a.So(err, should.BeNil)
}
