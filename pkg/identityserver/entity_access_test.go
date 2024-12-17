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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEntityAccess(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	newUsr := p.NewUser()
	newUsr.State = ttnpb.State_STATE_REQUESTED
	newUsr.PrimaryEmailAddressValidatedAt = nil
	newUsrKey, _ := p.NewAPIKey(newUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	newUsrCreds := rpcCreds(newUsrKey)

	rejectedUsr := p.NewUser()
	rejectedUsr.State = ttnpb.State_STATE_REJECTED
	rejectedUsrKey, _ := p.NewAPIKey(rejectedUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	rejectedUsrCreds := rpcCreds(rejectedUsrKey)

	suspendedUsr := p.NewUser()
	suspendedUsr.State = ttnpb.State_STATE_SUSPENDED
	suspendedUsrKey, _ := p.NewAPIKey(suspendedUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	suspendedUsrCreds := rpcCreds(suspendedUsrKey)

	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminUsrKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminUsrCreds := rpcCreds(adminUsrKey)

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdmin.UniversalRights = ttnpb.AllReadAdminRights.GetRights()
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	expiredKey, storedKey := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	storedKey.ExpiresAt = timestamppb.New(time.Now().Add(-10 * time.Minute))
	expiredCreds := rpcCreds(expiredKey)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.UserRegistration.ContactInfoValidation.Required = true

		cli := ttnpb.NewEntityAccessClient(cc)

		t.Run("New User", func(t *testing.T) {
			a, ctx := test.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, newUsrCreds, grpc.Header(&md))
			if a.So(err, should.BeNil) {
				a.So(md.Get("warning"), should.Contain, "Restricted rights until email address validated")
				a.So(md.Get("warning"), should.Contain, "Restricted rights while account pending")
				if a.So(authInfo, should.NotBeNil) && a.So(authInfo.GetApiKey(), should.NotBeNil) {
					rights := ttnpb.RightsFrom(authInfo.GetApiKey().GetApiKey().GetRights()...)
					a.So(
						rights.IncludesAll(
							ttnpb.Right_RIGHT_USER_INFO,
							ttnpb.Right_RIGHT_USER_SETTINGS_BASIC,
							ttnpb.Right_RIGHT_USER_DELETE,
						),
						should.BeTrue,
					)
				}
			}
		})

		t.Run("Rejected User", func(t *testing.T) {
			a, ctx := test.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, rejectedUsrCreds, grpc.Header(&md))
			if a.So(err, should.BeNil) {
				if warnings := md.Get("warning"); a.So(warnings, should.HaveLength, 1) {
					a.So(warnings[0], should.ContainSubstring, "Restricted rights after account rejection")
				}
				if a.So(authInfo, should.NotBeNil) && a.So(authInfo.GetApiKey(), should.NotBeNil) {
					rights := ttnpb.RightsFrom(authInfo.GetApiKey().GetApiKey().GetRights()...)
					a.So(rights.IncludesAll(ttnpb.Right_RIGHT_USER_INFO, ttnpb.Right_RIGHT_USER_DELETE), should.BeTrue)
				}
			}
		})

		t.Run("Suspended User", func(t *testing.T) {
			a, ctx := test.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, suspendedUsrCreds, grpc.Header(&md))
			if a.So(err, should.BeNil) {
				if warnings := md.Get("warning"); a.So(warnings, should.HaveLength, 1) {
					a.So(warnings[0], should.ContainSubstring, "Restricted rights after account suspension")
				}
				if a.So(authInfo, should.NotBeNil) && a.So(authInfo.GetApiKey(), should.NotBeNil) {
					rights := ttnpb.RightsFrom(authInfo.GetApiKey().GetApiKey().GetRights()...)
					a.So(rights.IncludesAll(ttnpb.Right_RIGHT_USER_INFO), should.BeTrue)
				}
			}
		})

		t.Run("Admin User", func(t *testing.T) {
			a, ctx := test.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, adminUsrCreds, grpc.Header(&md))
			if a.So(err, should.BeNil) && a.So(authInfo, should.NotBeNil) {
				a.So(authInfo.GetIsAdmin(), should.BeTrue)
				a.So(authInfo.GetUniversalRights().GetRights(), should.NotBeEmpty)
			}
		})

		t.Run("Read-only Admin User", func(t *testing.T) {
			a, ctx := test.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, readOnlyAdminKeyCreds, grpc.Header(&md))
			if a.So(err, should.BeNil) && a.So(authInfo, should.NotBeNil) {
				a.So(authInfo.GetIsAdmin(), should.BeTrue)
				a.So(authInfo.GetUniversalRights().GetRights(), should.NotBeEmpty)
				a.So(
					authInfo.GetUniversalRights().IncludesAll(ttnpb.AllReadAdminRights.GetRights()...),
					should.BeTrue,
				)
				a.So(
					authInfo.GetUniversalRights().IncludesAll(ttnpb.Right_RIGHT_USER_APPLICATIONS_CREATE),
					should.BeFalse,
				)

			}
		})

		t.Run("Cluster Peer", func(t *testing.T) {
			a, ctx := test.New(t)
			var md metadata.MD
			authInfo, err := cli.AuthInfo(ctx, ttnpb.Empty, is.WithClusterAuth(), grpc.Header(&md))
			if a.So(err, should.BeNil) && a.So(authInfo, should.NotBeNil) {
				a.So(authInfo.GetUniversalRights().GetRights(), should.NotBeEmpty)
			}
		})

		t.Run("Expired API Key User", func(t *testing.T) {
			a, ctx := test.New(t)
			var md metadata.MD
			_, err := cli.AuthInfo(ctx, ttnpb.Empty, expiredCreds, grpc.Header(&md))
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsUnauthenticated(err), should.BeTrue)
			}
		})
	}, withPrivateTestDatabase(p))
}
