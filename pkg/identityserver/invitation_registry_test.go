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

func TestInvitationsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
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
	})
}

func TestInvitationsCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		creds := userCreds(adminUserIdx)

		reg := ttnpb.NewUserInvitationRegistryClient(cc)

		invit, err := reg.Send(ctx, &ttnpb.SendInvitationRequest{
			Email: "foobar@example.com",
		}, creds)

		a.So(err, should.BeNil)
		if a.So(invit, should.NotBeNil) {
			a.So(invit.Email, should.Equal, "foobar@example.com")
		}

		_, err = reg.Send(ctx, &ttnpb.SendInvitationRequest{
			Email: "foobar@example.com",
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}

		invits, err := reg.List(ctx, &ttnpb.ListInvitationsRequest{}, creds)

		a.So(err, should.BeNil)
		a.So(invits.Invitations[0].Email, should.Equal, "foobar@example.com")

		_, err = reg.Delete(ctx, &ttnpb.DeleteInvitationRequest{
			Email: "foobar@example.com",
		}, creds)

		a.So(err, should.BeNil)

		invits, err = reg.List(ctx, &ttnpb.ListInvitationsRequest{}, creds)

		a.So(err, should.BeNil)
		if a.So(invits, should.NotBeNil) {
			a.So(invits.Invitations, should.BeEmpty)
		}
	})
}
