// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"context"
	"testing"

	"github.com/google/uuid"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestUserSessionsRegistry(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	usr1 := p.NewUser()

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	keyWithoutRights, _ := p.NewAPIKey(usr1.GetEntityIdentifiers())
	credsWithoutRights := rpcCreds(keyWithoutRights)

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	a, ctx := test.New(t)

	randomUUID := uuid.NewString()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserSessionRegistryClient(cc)

		// Invalid credentials
		for _, opts := range [][]grpc.CallOption{nil, {credsWithoutRights}, {readOnlyAdminKeyCreds}} {
			_, err := reg.List(ctx, &ttnpb.ListUserSessionsRequest{
				UserIds: usr1.GetIds(),
			}, opts...)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Delete(ctx, &ttnpb.UserSessionIdentifiers{
				UserIds:   usr1.GetIds(),
				SessionId: randomUUID,
			}, opts...)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}
		}

		sessions, err := reg.List(ctx, &ttnpb.ListUserSessionsRequest{
			UserIds: usr1.GetIds(),
		}, creds)
		if a.So(err, should.BeNil) {
			a.So(sessions.Sessions, should.BeEmpty)
		}

		_, err = reg.Delete(ctx, &ttnpb.UserSessionIdentifiers{
			UserIds:   usr1.GetIds(),
			SessionId: randomUUID,
		}, creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		var created *ttnpb.UserSession

		err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
			created, err = st.CreateSession(ctx, &ttnpb.UserSession{
				UserIds:   usr1.GetIds(),
				SessionId: randomUUID,
			})
			return err
		})
		if err != nil {
			t.Fatal(err)
		}

		sessions, err = reg.List(ctx, &ttnpb.ListUserSessionsRequest{
			UserIds: usr1.GetIds(),
		}, creds)
		if a.So(err, should.BeNil) && a.So(sessions, should.NotBeNil) {
			a.So(sessions.Sessions, should.HaveLength, 1)
		}

		_, err = reg.Delete(ctx, &ttnpb.UserSessionIdentifiers{
			UserIds:   usr1.GetIds(),
			SessionId: created.SessionId,
		}, creds)
		a.So(err, should.BeNil)

		sessions, err = reg.List(ctx, &ttnpb.ListUserSessionsRequest{
			UserIds: usr1.GetIds(),
		}, creds)
		if a.So(err, should.BeNil) && a.So(sessions, should.NotBeNil) {
			a.So(sessions.Sessions, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}
