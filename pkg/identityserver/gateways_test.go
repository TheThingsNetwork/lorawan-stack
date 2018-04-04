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

package identityserver

import (
	"sort"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ ttnpb.IsGatewayServer = new(gatewayService)

func TestGateway(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	user := testUsers()["bob"]

	gtw := ttnpb.Gateway{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gtw"},
		ClusterAddress:     "localhost:1234",
		FrequencyPlanID:    "868.8",
		Attributes: map[string]string{
			"version": "1.2",
		},
		Antennas: []ttnpb.GatewayAntenna{
			{
				Gain: 1.1,
				Location: ttnpb.Location{
					Latitude:  1.1,
					Longitude: 1.1,
				},
			},
			{
				Gain: 2.2,
				Location: ttnpb.Location{
					Latitude:  2.2,
					Longitude: 2.2,
				},
			},
			{
				Gain: 3,
				Location: ttnpb.Location{
					Latitude:  3,
					Longitude: 3,
				},
			},
		},
		Radios: []ttnpb.GatewayRadio{},
	}

	ctx := testCtx(user.UserID)

	_, err := is.gatewayService.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
		Gateway: gtw,
	})
	a.So(err, should.BeNil)

	// can't create gateways with blacklisted ids
	for _, id := range testSettings().BlacklistedIDs {
		_, err = is.gatewayService.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(ErrBlacklistedID.Describes(err), should.BeTrue)
	}

	found, err := is.gatewayService.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)

	gtws, err := is.gatewayService.ListGateways(ctx, &ttnpb.ListGatewaysRequest{})
	a.So(err, should.BeNil)
	if a.So(gtws.Gateways, should.HaveLength, 1) {
		a.So(gtws.Gateways[0], test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	gtw.Description = "foo"
	_, err = is.gatewayService.UpdateGateway(ctx, &ttnpb.UpdateGatewayRequest{
		Gateway: gtw,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.gatewayService.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)

	// generate a new API key
	key, err := is.gatewayService.GenerateGatewayAPIKey(ctx, &ttnpb.GenerateGatewayAPIKeyRequest{
		GatewayIdentifiers: gtw.GatewayIdentifiers,
		Name:               "foo",
		Rights:             ttnpb.AllGatewayRights(),
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllGatewayRights())

	// update api key
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = is.gatewayService.UpdateGatewayAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
		GatewayIdentifiers: gtw.GatewayIdentifiers,
		Name:               key.Name,
		Rights:             key.Rights,
	})
	a.So(err, should.BeNil)

	// can't generate another API Key with the same name
	_, err = is.gatewayService.GenerateGatewayAPIKey(ctx, &ttnpb.GenerateGatewayAPIKeyRequest{
		GatewayIdentifiers: gtw.GatewayIdentifiers,
		Name:               key.Name,
		Rights:             []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	keys, err := is.gatewayService.ListGatewayAPIKeys(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = is.gatewayService.RemoveGatewayAPIKey(ctx, &ttnpb.RemoveGatewayAPIKeyRequest{
		GatewayIdentifiers: gtw.GatewayIdentifiers,
		Name:               key.Name,
	})
	a.So(err, should.BeNil)

	keys, err = is.gatewayService.ListGatewayAPIKeys(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// set a new collaborator with SETTINGS_COLLABORATOR and INFO rights
	alice := testUsers()["alice"]
	collab := &ttnpb.GatewayCollaborator{
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &alice.UserIdentifiers}},
		GatewayIdentifiers:            gtw.GatewayIdentifiers,
		Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS},
	}

	_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	rights, err := is.gatewayService.ListGatewayRights(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(rights.Rights, should.Resemble, ttnpb.AllGatewayRights())

	collabs, err := is.gatewayService.ListGatewayCollaborators(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)
	a.So(collabs.Collaborators, should.Contain, collab)
	a.So(collabs.Collaborators, should.Contain, &ttnpb.GatewayCollaborator{
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
		GatewayIdentifiers:            gtw.GatewayIdentifiers,
		Rights:                        ttnpb.AllGatewayRights(),
	})

	// the new collaborator can't grant himself more rights
	{
		collab.Rights = append(collab.Rights, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)

		ctx := testCtx(alice.UserID)

		_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
		a.So(err, should.BeNil)

		rights, err := is.gatewayService.ListGatewayRights(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 2)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)

		// but it can't revoke itself the INFO right
		collab.Rights = []ttnpb.Right{ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS}
		_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
		a.So(err, should.BeNil)

		rights, err = is.gatewayService.ListGatewayRights(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 1)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_GATEWAY_INFO)

		collab.Rights = []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS}
		_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
		a.So(err, should.BeNil)
	}

	// try to unset the main collaborator will result in error as the gateway
	// will become unmanageable
	_, err = is.gatewayService.SetGatewayCollaborator(ctx, &ttnpb.GatewayCollaborator{
		GatewayIdentifiers:            gtw.GatewayIdentifiers,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
	})
	a.So(err, should.NotBeNil)
	a.So(ErrSetGatewayCollaboratorFailed.Describes(err), should.BeTrue)

	// but we can revoke a shared right between the two collaborators
	_, err = is.gatewayService.SetGatewayCollaborator(ctx, &ttnpb.GatewayCollaborator{
		GatewayIdentifiers:            gtw.GatewayIdentifiers,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
		Rights: ttnpb.DifferenceRights(ttnpb.AllGatewayRights(), []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO}),
	})
	a.So(err, should.NotBeNil)

	collabs, err = is.gatewayService.ListGatewayCollaborators(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)

	// unset the last added collaborator
	collab.Rights = []ttnpb.Right{}
	_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	collabs, err = is.gatewayService.ListGatewayCollaborators(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 1)

	_, err = is.gatewayService.DeleteGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: gtw.GatewayID})
	a.So(err, should.BeNil)
}
