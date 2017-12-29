// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var gatewayFactory = func() types.Gateway {
	return &ttnpb.Gateway{}
}

func testGateways() map[string]*ttnpb.Gateway {
	return map[string]*ttnpb.Gateway{
		"test-gateway": {
			GatewayIdentifier: ttnpb.GatewayIdentifier{"test-gateway"},
			Description:       "My description",
			Platform:          "Kerklink",
			Attributes: map[string]string{
				"foo": "bar",
			},
			FrequencyPlanID: "868_3",
			Antennas: []ttnpb.GatewayAntenna{
				{
					Location: ttnpb.Location{
						Latitude:  11.11,
						Longitude: 22.22,
						Altitude:  10,
					},
				},
			},
		},
		"bob-gateway": {
			GatewayIdentifier: ttnpb.GatewayIdentifier{"bob-gateway"},
			Description:       "My description",
			Attributes: map[string]string{
				"Modulation": "12345",
				"RFCH":       "111",
			},
			FrequencyPlanID: "868_3",
			ClusterAddress:  "network.eu",
			Antennas: []ttnpb.GatewayAntenna{
				{
					Gain: 12.12,
				},
			},
			Radios: []ttnpb.GatewayRadio{
				{
					Frequency: 10,
				},
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
		a.So(ErrGatewayIDTaken.Describes(err), should.BeTrue)
	}
}

func TestGatewayAPIKeys(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gtwID := testGateways()["test-gateway"].GatewayID
	key := &ttnpb.APIKey{
		Key:    "abcabcabc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	list, err := s.Gateways.ListAPIKeys(gtwID)
	a.So(err, should.BeNil)
	a.So(list, should.HaveLength, 0)

	err = s.Gateways.SaveAPIKey(gtwID, key)
	a.So(err, should.BeNil)

	key2 := &ttnpb.APIKey{
		Key:    "123abc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Gateways.SaveAPIKey(gtwID, key2)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	found, err := s.Gateways.GetAPIKey(gtwID, key.Name)
	a.So(err, should.BeNil)
	a.So(found, should.Resemble, key)

	key.Rights = append(key.Rights, ttnpb.Right(5))
	err = s.Gateways.UpdateAPIKeyRights(gtwID, key.Name, key.Rights)
	a.So(err, should.BeNil)

	list, err = s.Gateways.ListAPIKeys(gtwID)
	a.So(err, should.BeNil)
	if a.So(list, should.HaveLength, 1) {
		a.So(list[0], should.Resemble, key)
	}

	err = s.Gateways.DeleteAPIKey(gtwID, key.Name)
	a.So(err, should.BeNil)

	found, err = s.Gateways.GetAPIKey(gtwID, key.Name)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNotFound.Describes(err), should.BeTrue)
}

func TestGatewayAttributes(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gtw := testGateways()["bob-gateway"]

	// fetch gateway and check that the attributes has been registered
	{
		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	attributeKey := "Foo"

	// add attribute
	{
		gtw.Attributes[attributeKey] = "bar"
		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)
		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	// delete attribute
	{
		delete(gtw.Attributes, attributeKey)
		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)
		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
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
		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
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

		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	// delete previously added antenna
	{
		gtw.Antennas = gtw.Antennas[:1]

		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)

		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}
}

func TestGatewayRadios(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	gtw := testGateways()["bob-gateway"]

	// check that all radios were registered
	{
		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
		a.So(err, should.BeNil)
		if a.So(found.GetGateway().Radios, should.HaveLength, 1) {
			a.So(found.GetGateway().Radios[0], should.Resemble, gtw.Radios[0])
			gtw = found.GetGateway()
		}
	}

	// add a new radio
	{
		gtw.Radios = append(gtw.Radios, ttnpb.GatewayRadio{Frequency: 10})
		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)

		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	// delete previously added radio
	{
		gtw.Radios = gtw.Radios[:1]

		err := s.Gateways.Update(gtw)
		a.So(err, should.BeNil)

		found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
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

	collaborator := &ttnpb.GatewayCollaborator{
		GatewayIdentifier: ttnpb.GatewayIdentifier{gtw.GatewayID},
		UserIdentifier:    ttnpb.UserIdentifier{user.UserID},
		Rights: []ttnpb.Right{
			ttnpb.Right(1),
			ttnpb.Right(2),
		},
	}

	// set the collaborator
	{
		err := s.Gateways.SetCollaborator(collaborator)
		a.So(err, should.BeNil)

		collaborators, err := s.Gateways.ListCollaborators(gtw.GatewayID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)
	}

	// test ListCollaborators filter
	{
		collaborators, err := s.Gateways.ListCollaborators(gtw.GatewayID, ttnpb.Right(999))
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)

		collaborators, err = s.Gateways.ListCollaborators(gtw.GatewayID, ttnpb.Right(1))
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)

		collaborators, err = s.Gateways.ListCollaborators(gtw.GatewayID, ttnpb.Right(1), ttnpb.Right(3))
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)

	}

	// test HasUserRights method
	{
		yes, err := s.Gateways.HasUserRights(gtw.GatewayID, user.UserID, ttnpb.Right(0))
		a.So(yes, should.BeFalse)
		a.So(err, should.BeNil)

		yes, err = s.Gateways.HasUserRights(gtw.GatewayID, user.UserID, collaborator.Rights...)
		a.So(yes, should.BeTrue)
		a.So(err, should.BeNil)
	}

	// fetch gateways where Bob is collaborator
	{
		gtws, err := s.Gateways.ListByUser(user.UserID, gatewayFactory)
		a.So(err, should.BeNil)
		if a.So(gtws, should.HaveLength, 1) {
			a.So(gtws[0], test.ShouldBeGatewayIgnoringAutoFields, gtw)
		}
	}

	// modify rights
	{
		collaborator.Rights = append(collaborator.Rights, ttnpb.Right(3))
		err := s.Gateways.SetCollaborator(collaborator)
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
		err := s.Gateways.SetCollaborator(collaborator)
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

	found, err := s.Gateways.GetByID(gtw.GatewayID, gatewayFactory)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)
}

func TestGatewayDelete(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	userID := testUsers()["bob"].UserID
	gtwID := "delete-test"

	testGatewayDeleteFeedDatabase(t, userID, gtwID)

	err := s.Gateways.Delete(gtwID)
	a.So(err, should.BeNil)

	found, err := s.Gateways.GetByID(gtwID, gatewayFactory)
	a.So(err, should.NotBeNil)
	a.So(ErrGatewayNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func testGatewayDeleteFeedDatabase(t *testing.T, userID, gtwID string) {
	a := assertions.New(t)
	s := testStore(t)

	gtw := testGateways()["test-gateway"]
	gtw.GatewayID = gtwID

	err := s.Gateways.Create(gtw)
	a.So(err, should.BeNil)

	collaborator := &ttnpb.GatewayCollaborator{
		GatewayIdentifier: gtw.GatewayIdentifier,
		UserIdentifier:    ttnpb.UserIdentifier{userID},
		Rights:            []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Gateways.SetCollaborator(collaborator)
	a.So(err, should.BeNil)

	key := &ttnpb.APIKey{
		Name:   "foo",
		Key:    "123",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Gateways.SaveAPIKey(gtwID, key)
	a.So(err, should.BeNil)
}
