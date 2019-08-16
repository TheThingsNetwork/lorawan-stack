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

package identityserver

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	clientAccessUser.Admin = false
	clientAccessUser.State = ttnpb.STATE_APPROVED
	for _, apiKey := range userAPIKeys(&clientAccessUser.UserIdentifiers).APIKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.RIGHT_CLIENT_ALL,
		}
	}
}

func TestClientAccessRightsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := clientAccessUser.UserIdentifiers, userCreds(clientAccessUserIdx)
		clientID := userClients(&userID).Clients[0].ClientIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewClientAccessClient(cc)

		_, err := reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
			ClientIdentifiers: clientID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_ALL},
			},
		}, creds)

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestClientAccessPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.UserIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()
		clientID := userClients(&userID).Clients[0].ClientIdentifiers

		reg := ttnpb.NewClientAccessClient(cc)

		rights, err := reg.ListRights(ctx, &clientID)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
			ClientIdentifiers: clientID,
		})

		a.So(collaborators, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsUnauthenticated(err), should.BeTrue)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
			ClientIdentifiers: clientID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
			},
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestClientAccessClusterAuth(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.UserIdentifiers
		clientID := userClients(&userID).Clients[0].ClientIdentifiers

		reg := ttnpb.NewClientAccessClient(cc)

		rights, err := reg.ListRights(ctx, &clientID, is.WithClusterAuth())

		a.So(rights, should.NotBeNil)
		a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllClientRights).Sub(rights).Rights, should.BeEmpty)
		a.So(err, should.BeNil)
	})
}

func TestClientAccessCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()
		clientID := userClients(&userID).Clients[0].ClientIdentifiers

		reg := ttnpb.NewClientAccessClient(cc)

		rights, err := reg.ListRights(ctx, &clientID, creds)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.Contain, ttnpb.RIGHT_CLIENT_ALL)
		a.So(err, should.BeNil)

		modifiedClientID := clientID
		modifiedClientID.ClientID = reverse(modifiedClientID.ClientID)

		rights, err = reg.ListRights(ctx, &modifiedClientID, creds)
		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
			ClientIdentifiers: clientID,
		}, creds)

		a.So(collaborators, should.NotBeNil)
		a.So(collaborators.Collaborators, should.NotBeEmpty)
		a.So(err, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
			ClientIdentifiers: clientID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
			ClientIdentifiers:             clientID,
			OrganizationOrUserIdentifiers: *collaboratorID,
		}, creds)

		a.So(err, should.BeNil)
		a.So(res.Rights, should.Resemble, []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL})
	})
}
