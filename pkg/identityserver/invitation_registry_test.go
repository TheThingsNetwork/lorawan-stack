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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestInvitationsPermissionDenied(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserInvitationRegistryClient(cc)
		_, err := reg.Send(ctx, &ttnpb.SendInvitationRequest{
			Email: "foobar@example.com",
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		listInvit, err := reg.List(ctx, &ttnpb.ListInvitationsRequest{})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(listInvit, should.BeNil)

		_, err = reg.Delete(ctx, &ttnpb.DeleteInvitationRequest{
			Email: "foobar@example.com",
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	}, withPrivateTestDatabase(p))
}

func TestInvitationsCRUD(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminUsrKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminUsrCreds := rpcCreds(adminUsrKey)

	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserInvitationRegistryClient(cc)

		invitation, err := reg.Send(ctx, &ttnpb.SendInvitationRequest{
			Email: "foobar@example.com",
		}, adminUsrCreds)
		if a.So(err, should.BeNil) && a.So(invitation, should.NotBeNil) {
			a.So(invitation.Email, should.Equal, "foobar@example.com")
		}

		_, err = reg.Send(ctx, &ttnpb.SendInvitationRequest{
			Email: "foobar@example.com",
		}, adminUsrCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}

		invitations, err := reg.List(ctx, &ttnpb.ListInvitationsRequest{}, adminUsrCreds)
		if a.So(err, should.BeNil) && a.So(invitations, should.NotBeNil) && a.So(invitations.Invitations, should.HaveLength, 1) {
			a.So(invitations.Invitations[0].Email, should.Equal, "foobar@example.com")
		}

		_, err = reg.Delete(ctx, &ttnpb.DeleteInvitationRequest{
			Email: "foobar@example.com",
		}, adminUsrCreds)
		a.So(err, should.BeNil)

		invitations, err = reg.List(ctx, &ttnpb.ListInvitationsRequest{}, adminUsrCreds)
		if a.So(err, should.BeNil) && a.So(invitations, should.NotBeNil) {
			a.So(invitations.Invitations, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}
