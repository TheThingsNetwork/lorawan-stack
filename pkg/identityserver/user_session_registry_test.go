// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

func TestUserSessionsRegistry(t *testing.T) {
	a, ctx := test.New(t)

	randomUUID := uuid.NewV4().String()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		user, creds := population.Users[defaultUserIdx], userCreds(defaultUserIdx)
		credsWithoutRights := userCreds(defaultUserIdx, "key without rights")

		reg := ttnpb.NewUserSessionRegistryClient(cc)

		_, err := reg.List(ctx, &ttnpb.ListUserSessionsRequest{
			UserIdentifiers: user.UserIdentifiers,
		}, credsWithoutRights)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, &ttnpb.UserSessionIdentifiers{
			UserIdentifiers: user.UserIdentifiers,
			SessionID:       randomUUID,
		}, credsWithoutRights)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		sessions, err := reg.List(ctx, &ttnpb.ListUserSessionsRequest{
			UserIdentifiers: user.UserIdentifiers,
		}, creds)
		if a.So(err, should.BeNil) {
			a.So(sessions.Sessions, should.BeEmpty)
		}

		_, err = reg.Delete(ctx, &ttnpb.UserSessionIdentifiers{
			UserIdentifiers: user.UserIdentifiers,
			SessionID:       randomUUID,
		}, creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		var created *ttnpb.UserSession

		err = is.withDatabase(ctx, func(db *gorm.DB) error {
			created, err = store.GetUserSessionStore(db).CreateSession(ctx, &ttnpb.UserSession{
				UserIdentifiers: user.UserIdentifiers,
				SessionID:       randomUUID,
			})
			return err
		})
		if err != nil {
			t.Fatal(err)
		}

		sessions, err = reg.List(ctx, &ttnpb.ListUserSessionsRequest{
			UserIdentifiers: user.UserIdentifiers,
		}, creds)
		if a.So(err, should.BeNil) {
			a.So(sessions.Sessions, should.HaveLength, 1)
		}

		_, err = reg.Delete(ctx, &ttnpb.UserSessionIdentifiers{
			UserIdentifiers: user.UserIdentifiers,
			SessionID:       created.SessionID,
		}, creds)
		a.So(err, should.BeNil)

		sessions, err = reg.List(ctx, &ttnpb.ListUserSessionsRequest{
			UserIdentifiers: user.UserIdentifiers,
		}, creds)
		if a.So(err, should.BeNil) {
			a.So(sessions.Sessions, should.BeEmpty)
		}
	})
}
