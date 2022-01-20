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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	clientAccessUser.Admin = false
	clientAccessUser.State = ttnpb.State_STATE_APPROVED
	for _, apiKey := range userAPIKeys(clientAccessUser.GetIds()).ApiKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.Right_RIGHT_CLIENT_ALL,
		}
	}
}

func TestClientAccessRightsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := clientAccessUser.GetIds(), userCreds(clientAccessUserIdx)
		clientID := userClients(userID).Clients[0].GetIds()
		collaboratorID := collaboratorUser.GetIds().GetOrganizationOrUserIdentifiers()

		reg := ttnpb.NewClientAccessClient(cc)

		_, err := reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
			ClientIds: clientID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_ALL},
			},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestClientAccessPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.GetIds()
		collaboratorID := collaboratorUser.GetIds().GetOrganizationOrUserIdentifiers()
		clientID := userClients(userID).Clients[0].GetIds()

		reg := ttnpb.NewClientAccessClient(cc)

		rights, err := reg.ListRights(ctx, clientID)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
			ClientIds: clientID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}
		a.So(collaborators, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
			ClientIds: clientID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_ALL},
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestClientAccessClusterAuth(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.GetIds()
		clientID := userClients(userID).Clients[0].GetIds()

		reg := ttnpb.NewClientAccessClient(cc)

		rights, err := reg.ListRights(ctx, clientID, is.WithClusterAuth())

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllClientRights).Sub(rights).Rights, should.BeEmpty)
		}
	})
}

func TestClientAccessCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.GetIds(), userCreds(defaultUserIdx)
		collaboratorID := collaboratorUser.GetIds().OrganizationOrUserIdentifiers()
		clientID := userClients(userID).Clients[0].GetIds()

		reg := ttnpb.NewClientAccessClient(cc)

		rights, err := reg.ListRights(ctx, clientID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.Contain, ttnpb.Right_RIGHT_CLIENT_ALL)
		}

		modifiedClientID := &ttnpb.ClientIdentifiers{ClientId: reverse(clientID.GetClientId())}

		rights, err = reg.ListRights(ctx, modifiedClientID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
			ClientIds: clientID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(collaborators, should.NotBeNil) {
			a.So(collaborators.Collaborators, should.NotBeEmpty)
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
			ClientIds: clientID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_ALL},
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
			ClientIds:    clientID,
			Collaborator: collaboratorID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(res, should.NotBeNil) {
			a.So(res.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_ALL})
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
			ClientIds: clientID,
			Collaborator: &ttnpb.Collaborator{
				Ids: collaboratorID,
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err = reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
			ClientIds:    clientID,
			Collaborator: collaboratorID,
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}
