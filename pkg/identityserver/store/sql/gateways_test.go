// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
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
		created, err := s.Gateways.Register(gtw)
		a.So(err, should.BeNil)
		a.So(created, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	// Attempt to recreate them should throw an error
	for _, gtw := range gateways {
		_, err := s.Gateways.Register(gtw)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 301)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestGatewayAttributes(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gtw := testGateways()["bob-gateway"]

	// Fetch attributes
	{
		attributes, err := s.Gateways.ListAttributes(gtw.ID)
		a.So(err, should.BeNil)
		if a.So(attributes, should.HaveLength, 2) {
			a.So(attributes, should.Resemble, gtw.Attributes)
		}
	}

	attributeKey := "Foo"
	attributeValue := "Bar"

	// Add attribute
	{
		err := s.Gateways.SetAttribute(gtw.ID, attributeKey, attributeValue)
		a.So(err, should.BeNil)

		attributes, err := s.Gateways.ListAttributes(gtw.ID)
		a.So(err, should.BeNil)
		if a.So(attributes, should.HaveLength, 3) && a.So(attributes, should.ContainKey, attributeKey) {
			a.So(attributes[attributeKey], should.Equal, attributeValue)
		}
	}

	// Overwrite attribute
	{
		attributeValue = "BarBar"

		err := s.Gateways.SetAttribute(gtw.ID, attributeKey, attributeValue)
		a.So(err, should.BeNil)

		attributes, err := s.Gateways.ListAttributes(gtw.ID)
		a.So(err, should.BeNil)
		if a.So(attributes, should.HaveLength, 3) && a.So(attributes, should.ContainKey, attributeKey) {
			a.So(attributes[attributeKey], should.Equal, attributeValue)
		}
	}

	// Delete attribute
	{
		err := s.Gateways.RemoveAttribute(gtw.ID, attributeKey)
		a.So(err, should.BeNil)

		attributes, err := s.Gateways.ListAttributes(gtw.ID)
		a.So(err, should.BeNil)
		if a.So(attributes, should.HaveLength, 2) {
			a.So(attributes, should.Resemble, gtw.Attributes)
		}
	}
}

func TestGatewayAntennas(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gtw := testGateways()["bob-gateway"]

	// Fetch antennas
	{
		antennas, err := s.Gateways.ListAntennas(gtw.ID)
		a.So(err, should.BeNil)
		if a.So(antennas, should.HaveLength, 1) {
			a.So(antennas[0], should.Resemble, gtw.Antennas[0])
		}
	}

	ant := gtw.Antennas[0]
	ant.Location = &ttnpb.Location{Longitude: 12.12, Latitude: 10.02, Altitude: 2}

	// Modify the antenna
	{
		err := s.Gateways.SetAntenna(gtw.ID, ant)
		a.So(err, should.BeNil)

		antennas, err := s.Gateways.ListAntennas(gtw.ID)
		a.So(err, should.BeNil)
		if a.So(antennas, should.HaveLength, 1) {
			a.So(antennas[0], should.Resemble, ant)
		}
	}

	// Delete the antenna
	{
		err := s.Gateways.RemoveAntenna(gtw.ID, ant.ID)
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

	// Add collaborator
	{
		err := s.Gateways.AddCollaborator(gtw.ID, collaborator)
		a.So(err, should.BeNil)
	}

	// Find collaborators
	{
		collaborators, err := s.Gateways.ListCollaborators(gtw.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		if len(collaborators) > 0 {
			a.So(collaborators[0], should.Resemble, collaborator)
		}
	}

	right := types.GatewaySettingsRight

	// Grant a right
	{
		err := s.Gateways.AddRight(gtw.ID, user.Username, right)
		a.So(err, should.BeNil)

		rights, err := s.Gateways.ListUserRights(gtw.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 2)
		if len(rights) > 0 {
			a.So(rights, should.Contain, right)
		}
	}

	// Revoke a right
	{
		err := s.Gateways.RemoveRight(gtw.ID, user.Username, right)
		a.So(err, should.BeNil)

		rights, err := s.Gateways.ListUserRights(gtw.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 1)
		if len(rights) > 0 {
			a.So(rights, should.NotContain, right)
		}
	}

	// Fetch user rights
	{
		rights, err := s.Gateways.ListUserRights(gtw.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(collaborator.Rights, should.Resemble, rights)
	}

	// Remove the collaborator
	{
		err := s.Gateways.RemoveCollaborator(gtw.ID, user.Username)
		a.So(err, should.BeNil)

		collaborators, err := s.Gateways.ListCollaborators(gtw.ID)
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

	// Add Alice as Gateway owner
	{
		collaborator := utils.Collaborator(alice.Username, rights)
		err := s.Gateways.AddCollaborator(gtw.ID, collaborator)
		a.So(err, should.BeNil)
	}

	// Add Bob as owner also
	{
		collaborator := utils.Collaborator(bob.Username, rights)
		err := s.Gateways.AddCollaborator(gtw.ID, collaborator)
		a.So(err, should.BeNil)
	}

	// Fetch owners
	{
		owners, err := s.Gateways.ListOwners(gtw.ID)
		a.So(err, should.BeNil)
		a.So(owners, should.HaveLength, 2)
		a.So(owners, should.Resemble, []string{"alice", "bob"})
	}
}

func TestGatewayEdit(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gtw := testGateways()["bob-gateway"]
	gtw.Description = "Fancy new description"

	updated, err := s.Gateways.Edit(gtw)
	a.So(err, should.BeNil)
	a.So(updated, test.ShouldBeGatewayIgnoringAutoFields, gtw)
}

func TestGatewayArchive(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	gtw := testGateways()["bob-gateway"]

	err := s.Gateways.Archive(gtw.ID)
	a.So(err, should.BeNil)

	found, err := s.Gateways.FindByID(gtw.ID)
	a.So(err, should.BeNil)
	a.So(found.GetGateway().Archived, should.NotBeNil)
}
