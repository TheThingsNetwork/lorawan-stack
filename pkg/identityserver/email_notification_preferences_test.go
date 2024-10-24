// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
	"os"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestEmailNotificationPreferences(t *testing.T) {
	p := &storetest.Population{}

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	admin.EmailNotificationPreferences = &ttnpb.EmailNotificationPreferences{
		Types: []ttnpb.NotificationType{
			ttnpb.NotificationType_API_KEY_CHANGED,
		},
	}
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	usr1.EmailNotificationPreferences = &ttnpb.EmailNotificationPreferences{
		Types: []ttnpb.NotificationType{
			ttnpb.NotificationType_API_KEY_CREATED,
			ttnpb.NotificationType_API_KEY_CHANGED,
		},
	}
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	limitedKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS,
	)
	limitedCreds := rpcCreds(limitedKey)

	appKey, _ := p.NewAPIKey(app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
	)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true
		is.config.Email.Provider = "dir"
		tempDir := t.TempDir()
		is.config.Email.Dir = tempDir

		reg := ttnpb.NewApplicationAccessClient(cc)

		// Test sending email to users that have API_KEY_CHANGED in their preferences.
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: app1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: appKey.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_APPLICATION_LINK,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Rights, should.Resemble, []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
				ttnpb.Right_RIGHT_APPLICATION_LINK,
			})
		}

		entries, err := os.ReadDir(tempDir)
		a.So(err, should.BeNil)
		a.So(entries, should.HaveLength, 1)

		for _, opts := range [][]grpc.CallOption{{adminCreds}, {usr1Creds}, {limitedCreds}} {
			created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
				ApplicationIds: app1.GetIds(),
				Name:           "api-key-name",
				Rights:         []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_INFO},
			}, opts...)
			if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
				a.So(created.Name, should.Equal, "api-key-name")
				a.So(created.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_INFO})
			}
		}
	}, withPrivateTestDatabase(p))
}
