// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package api_test

import (
	"context"
	"sort"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	. "github.com/TheThingsNetwork/ttn/pkg/identityserver/api"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/claims"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestGateway(t *testing.T) {
	a := assertions.New(t)
	g := getGRPC(t)

	user := testUsers()["bob"]

	gtw := ttnpb.Gateway{
		GatewayIdentifier: ttnpb.GatewayIdentifier{"foo-gtw"},
		ClusterAddress:    "localhost:1234",
		FrequencyPlanID:   "868.8",
		Attributes: map[string]string{
			"version": "1.2",
		},
		Antennas: []ttnpb.GatewayAntenna{
			ttnpb.GatewayAntenna{
				Gain: 1.1,
				Location: ttnpb.Location{
					Latitude:  1.1,
					Longitude: 1.1,
				},
			},
			ttnpb.GatewayAntenna{
				Gain: 2.2,
				Location: ttnpb.Location{
					Latitude:  2.2,
					Longitude: 2.2,
				},
			},
			ttnpb.GatewayAntenna{
				Gain: 3,
				Location: ttnpb.Location{
					Latitude:  3,
					Longitude: 3,
				},
			},
		},
	}

	ctx := claims.NewContext(context.Background(), &auth.Claims{
		EntityID:  user.UserID,
		EntityTyp: auth.EntityUser,
		Source:    auth.Token,
		Rights:    append(ttnpb.AllUserRights, ttnpb.AllGatewayRights...),
	})

	_, err := g.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
		Gateway: gtw,
	})
	a.So(err, should.BeNil)

	// check that a api key has been created
	keys, err := g.ListGatewayAPIKeys(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		k := keys.APIKeys[0]
		a.So(k.Name, should.Equal, APIKeyName)
		a.So(k.Key, should.NotBeEmpty)
		a.So(k.Rights, should.HaveLength, 1)
		a.So(k.Rights, should.Contain, ttnpb.RIGHT_GATEWAY_INFO)
	}

	// can't create gateways with blacklisted ids
	for _, id := range settings.BlacklistedIDs {
		_, err := g.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifier: ttnpb.GatewayIdentifier{id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(ErrBlacklistedID.Describes(err), should.BeTrue)
	}

	found, err := g.GetGateway(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)

	gtws, err := g.ListGateways(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(gtws.Gateways, should.HaveLength, 1) {
		a.So(gtws.Gateways[0], test.ShouldBeGatewayIgnoringAutoFields, gtw)
	}

	gtw.Description = "foo"
	_, err = g.UpdateGateway(ctx, &ttnpb.UpdateGatewayRequest{
		Gateway: gtw,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = g.GetGateway(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, gtw)

	// generate a new API key
	key, err := g.GenerateGatewayAPIKey(ctx, &ttnpb.GenerateGatewayAPIKeyRequest{
		GatewayIdentifier: gtw.GatewayIdentifier,
		Name:              "foo",
		Rights:            ttnpb.AllGatewayRights,
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllGatewayRights)

	// update api key
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = g.UpdateGatewayAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
		GatewayIdentifier: gtw.GatewayIdentifier,
		Name:              key.Name,
		Rights:            key.Rights,
	})
	a.So(err, should.BeNil)

	// can't generate another API Key with the same name
	_, err = g.GenerateGatewayAPIKey(ctx, &ttnpb.GenerateGatewayAPIKeyRequest{
		GatewayIdentifier: gtw.GatewayIdentifier,
		Name:              key.Name,
		Rights:            []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	keys, err = g.ListGatewayAPIKeys(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 2) {
		sort.Slice(keys.APIKeys[1].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[1], should.Resemble, key)
	}

	_, err = g.RemoveGatewayAPIKey(ctx, &ttnpb.RemoveGatewayAPIKeyRequest{
		GatewayIdentifier: gtw.GatewayIdentifier,
		Name:              key.Name,
	})

	keys, err = g.ListGatewayAPIKeys(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 1)

	// set new collaborator
	alice := testUsers()["alice"]
	collab := &ttnpb.GatewayCollaborator{
		UserIdentifier:    alice.UserIdentifier,
		GatewayIdentifier: gtw.GatewayIdentifier,
		Rights:            []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
	}

	_, err = g.SetGatewayCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	rights, err := g.ListGatewayRights(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(rights.Rights, should.Resemble, ttnpb.AllGatewayRights)

	collabs, err := g.ListGatewayCollaborators(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)
	a.So(collabs.Collaborators, should.Contain, collab)
	a.So(collabs.Collaborators, should.Contain, &ttnpb.GatewayCollaborator{
		UserIdentifier:    user.UserIdentifier,
		GatewayIdentifier: gtw.GatewayIdentifier,
		Rights:            ttnpb.AllGatewayRights,
	})

	// while there is two collaborators can't unset the only collab with COLLABORATORS right
	_, err = g.SetGatewayCollaborator(ctx, &ttnpb.GatewayCollaborator{
		GatewayIdentifier: gtw.GatewayIdentifier,
		UserIdentifier:    user.UserIdentifier,
	})
	a.So(err, should.NotBeNil)

	collabs, err = g.ListGatewayCollaborators(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)

	// unset the last added collaborator
	collab.Rights = []ttnpb.Right{}
	_, err = g.SetGatewayCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	collabs, err = g.ListGatewayCollaborators(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 1)

	_, err = g.DeleteGateway(ctx, &ttnpb.GatewayIdentifier{gtw.GatewayID})
	a.So(err, should.BeNil)
}
