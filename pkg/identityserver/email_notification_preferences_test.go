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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEmailNotificationPreferences(t *testing.T) {
	p := &storetest.Population{}

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)
	admin.EmailNotificationPreferences = &ttnpb.EmailNotificationPreferences{
		Types: []ttnpb.NotificationType{
			ttnpb.NotificationType_API_KEY_CREATED,
		},
	}

	usr1 := p.NewUser()
	usr1.EmailNotificationPreferences = &ttnpb.EmailNotificationPreferences{
		Types: []ttnpb.NotificationType{
			ttnpb.NotificationType_API_KEY_CREATED,
		},
	}
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS,
	)
	usr1Creds := rpcCreds(usr1Key)

	app1 := p.NewApplication(admin.GetOrganizationOrUserIdentifiers())
	p.NewMembership(
		usr1.GetOrganizationOrUserIdentifiers(),
		app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
	)
	appKey, _ := p.NewAPIKey(app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS,
	)

	now := timestamppb.Now()
	usrIDs := &ttnpb.UserIdentifiers{
		UserId: "foo-usr",
	}

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true
		is.config.Email.Provider = "dir"
		tempDir := t.TempDir()
		is.config.Email.Dir = tempDir

		reg := ttnpb.NewApplicationAccessClient(cc)
		userReg := ttnpb.NewUserRegistryClient(cc)

		// Test user not receiving email notification because this
		// notification type is not in the list of preferences.
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: app1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: appKey.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS,
					ttnpb.Right_RIGHT_APPLICATION_LINK,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, adminCreds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Rights, should.Resemble, []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
				ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS,
				ttnpb.Right_RIGHT_APPLICATION_LINK,
			})
		}

		entries, err := os.ReadDir(tempDir)
		a.So(err, should.BeNil)
		a.So(entries, should.HaveLength, 0)

		// Test admin receiving email notification in spite of the list of preferences.
		updatedUser, err := userReg.Create(ctx, &ttnpb.CreateUserRequest{
			User: &ttnpb.User{
				Ids:                 usrIDs,
				Password:            "test password",
				CreatedAt:           now,
				UpdatedAt:           now,
				Name:                "Foo User",
				Description:         "Foo User Description",
				PrimaryEmailAddress: "foo@example.com",
				State:               ttnpb.State_STATE_REQUESTED,
			},
		}, adminCreds)
		if a.So(err, should.BeNil) && a.So(updatedUser, should.NotBeNil) {
			a.So(updatedUser.State, should.Equal, ttnpb.State_STATE_REQUESTED)
		}

		entries, err = os.ReadDir(tempDir)
		a.So(err, should.BeNil)
		a.So(entries, should.HaveLength, 1)

		time.Sleep(test.Delay)

		// Test users receiving email notification because this notification type is in the list of preferences.
		for _, opts := range [][]grpc.CallOption{{adminCreds}, {usr1Creds}} {
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

		entries, err = os.ReadDir(tempDir)
		a.So(err, should.BeNil)
		a.So(entries, should.HaveLength, 3)
	}, withPrivateTestDatabase(p))
}
