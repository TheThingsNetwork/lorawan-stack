// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"
	"sort"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/claims"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestApplication(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	user := testUsers()["bob"]

	app := ttnpb.Application{
		ApplicationIdentifier: ttnpb.ApplicationIdentifier{"foo-app"},
	}

	ctx := claims.NewContext(context.Background(), &auth.Claims{
		EntityID:   user.UserID,
		EntityType: auth.EntityUser,
		Source:     auth.Token,
		Rights:     append(ttnpb.AllUserRights, ttnpb.AllApplicationRights...),
	})

	_, err := is.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
		Application: app,
	})
	a.So(err, should.BeNil)

	// can't create applications with blacklisted ids
	for _, id := range testSettings().BlacklistedIDs {
		_, err := is.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
			Application: ttnpb.Application{
				ApplicationIdentifier: ttnpb.ApplicationIdentifier{id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(ErrBlacklistedID.Describes(err), should.BeTrue)
	}

	found, err := is.GetApplication(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, app)

	apps, err := is.ListApplications(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(apps.Applications, should.HaveLength, 1) {
		a.So(apps.Applications[0], test.ShouldBeApplicationIgnoringAutoFields, app)
	}

	app.Description = "foo"
	_, err = is.UpdateApplication(ctx, &ttnpb.UpdateApplicationRequest{
		Application: app,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.GetApplication(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, app)

	// generate a new API key
	key, err := is.GenerateApplicationAPIKey(ctx, &ttnpb.GenerateApplicationAPIKeyRequest{
		ApplicationIdentifier: app.ApplicationIdentifier,
		Name:   "foo",
		Rights: ttnpb.AllApplicationRights,
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllApplicationRights)

	// update api key
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = is.UpdateApplicationAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
		ApplicationIdentifier: app.ApplicationIdentifier,
		Name:   key.Name,
		Rights: key.Rights,
	})
	a.So(err, should.BeNil)

	// can't generate another API Key with the same name
	_, err = is.GenerateApplicationAPIKey(ctx, &ttnpb.GenerateApplicationAPIKeyRequest{
		ApplicationIdentifier: app.ApplicationIdentifier,
		Name:   key.Name,
		Rights: []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	keys, err := is.ListApplicationAPIKeys(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = is.RemoveApplicationAPIKey(ctx, &ttnpb.RemoveApplicationAPIKeyRequest{
		ApplicationIdentifier: app.ApplicationIdentifier,
		Name: key.Name,
	})

	keys, err = is.ListApplicationAPIKeys(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// set new collaborator
	alice := testUsers()["alice"]
	collab := &ttnpb.ApplicationCollaborator{
		UserIdentifier:        alice.UserIdentifier,
		ApplicationIdentifier: app.ApplicationIdentifier,
		Rights:                []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
	}

	_, err = is.SetApplicationCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	rights, err := is.ListApplicationRights(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(rights.Rights, should.Resemble, ttnpb.AllApplicationRights)

	collabs, err := is.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)
	a.So(collabs.Collaborators, should.Contain, collab)
	a.So(collabs.Collaborators, should.Contain, &ttnpb.ApplicationCollaborator{
		UserIdentifier:        user.UserIdentifier,
		ApplicationIdentifier: app.ApplicationIdentifier,
		Rights:                ttnpb.AllApplicationRights,
	})

	// while there is two collaborators can't unset the only collab with COLLABORATORS right
	_, err = is.SetApplicationCollaborator(ctx, &ttnpb.ApplicationCollaborator{
		ApplicationIdentifier: app.ApplicationIdentifier,
		UserIdentifier:        user.UserIdentifier,
	})
	a.So(err, should.NotBeNil)

	collabs, err = is.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)

	// unset the last added collaborator
	collab.Rights = []ttnpb.Right{}
	_, err = is.SetApplicationCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	collabs, err = is.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 1)

	_, err = is.DeleteApplication(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
}
