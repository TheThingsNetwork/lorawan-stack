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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestEntityAccess(t *testing.T) {
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.UserRegistration.ContactInfoValidation.Required = true

		cli := ttnpb.NewEntityAccessClient(cc)

		t.Run("New User", func(t *testing.T) {
			a := assertions.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, userCreds(newUserIdx), grpc.Header(&md))

			a.So(err, should.BeNil)
			a.So(md.Get("warning"), should.Contain, "Restricted rights until email address validated")
			a.So(md.Get("warning"), should.Contain, "Restricted rights while account pending")
			if a.So(authInfo.GetAPIKey(), should.NotBeNil) {
				rights := ttnpb.RightsFrom(authInfo.GetAPIKey().GetRights()...)
				a.So(rights.IncludesAll(ttnpb.RIGHT_USER_INFO, ttnpb.RIGHT_USER_SETTINGS_BASIC, ttnpb.RIGHT_USER_DELETE), should.BeTrue)
			}
		})

		t.Run("Rejected User", func(t *testing.T) {
			a := assertions.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, userCreds(rejectedUserIdx), grpc.Header(&md))

			a.So(err, should.BeNil)
			a.So(md.Get("warning"), should.Contain, "Restricted rights after account rejection")
			if a.So(authInfo.GetAPIKey(), should.NotBeNil) {
				rights := ttnpb.RightsFrom(authInfo.GetAPIKey().GetRights()...)
				a.So(rights.IncludesAll(ttnpb.RIGHT_USER_INFO, ttnpb.RIGHT_USER_DELETE), should.BeTrue)
			}
		})

		t.Run("Suspended User", func(t *testing.T) {
			a := assertions.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, userCreds(suspendedUserIdx), grpc.Header(&md))

			a.So(err, should.BeNil)
			a.So(md.Get("warning"), should.Contain, "Restricted rights after account suspension")
			if a.So(authInfo.GetAPIKey(), should.NotBeNil) {
				rights := ttnpb.RightsFrom(authInfo.GetAPIKey().GetRights()...)
				a.So(rights.IncludesAll(ttnpb.RIGHT_USER_INFO), should.BeTrue)
			}
		})

		t.Run("Admin User", func(t *testing.T) {
			a := assertions.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, userCreds(adminUserIdx), grpc.Header(&md))

			a.So(err, should.BeNil)
			a.So(authInfo.GetUniversalRights().GetRights(), should.NotBeEmpty)
		})

		t.Run("Cluster Peer", func(t *testing.T) {
			a := assertions.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, is.WithClusterAuth(), grpc.Header(&md))

			a.So(err, should.BeNil)
			a.So(authInfo.GetUniversalRights().GetRights(), should.NotBeEmpty)
		})
	})
}
