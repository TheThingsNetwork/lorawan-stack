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
	"strings"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestValidatePasswordStrength(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	testWithIdentityServer(t, func(is *IdentityServer, _ *grpc.ClientConn) {
		is.config.UserRegistration.PasswordRequirements.MinLength = 8
		is.config.UserRegistration.PasswordRequirements.MaxLength = 1000
		is.config.UserRegistration.PasswordRequirements.MinUppercase = 1
		is.config.UserRegistration.PasswordRequirements.MinDigits = 1
		is.config.UserRegistration.PasswordRequirements.MinSpecial = 1
		is.config.UserRegistration.PasswordRequirements.RejectUserID = true
		is.config.UserRegistration.PasswordRequirements.RejectCommon = true

		for p, ok := range map[string]bool{
			"āA0$": false, // Too short
			strings.Repeat("āaaAAA➉23!@#aaaAAA12aaaAAA123!@#aaaAAA12aaaAAA123!@#aaaAAA12aaaAAA123!@#aaaAAA12aaaAAA123!@#aaaAAA12aaa", 10): false, // Too long.
			"āaabbb➉23":    false, // No uppercase and special characters.
			"āaabbb➉AA":    false, // No digits and special characters.
			"āaabbb➉@#":    false, // No digits and uppercase characters.
			"āaa123➉@#":    false, // No uppercase characters.
			"āaaAAA➉@#":    false, // No digits.
			"āaaAAA➉23":    false, // No special characters.
			"myusername":   false, // Contains username.
			"password1":    false, // Too common.
			"āaaAAA123!@#": true,
			"       1A":    true,
			"āaa	AAA123 ":  true,
			"āaaAAA123 ":   true,
		} {
			t.Run(p, func(t *testing.T) {
				a, ctx := test.New(t)

				err := is.validatePasswordStrength(ctx, "username", p)
				if ok {
					a.So(err, should.BeNil)
				} else {
					a.So(err, should.NotBeNil)
				}
			})
		}
	}, withPrivateTestDatabase(p))
}

func TestTemporaryValidPassword(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	usr1 := p.NewUser()

	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserRegistryClient(cc)

		_, err := reg.CreateTemporaryPassword(ctx, &ttnpb.CreateTemporaryPasswordRequest{
			UserIds: usr1.GetIds(),
		})
		a.So(err, should.BeNil)

		_, err = reg.CreateTemporaryPassword(ctx, &ttnpb.CreateTemporaryPasswordRequest{
			UserIds: usr1.GetIds(),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}
	}, withPrivateTestDatabase(p))
}

func TestUserCreate(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	a, ctx := test.New(t)

	req := &ttnpb.CreateUserRequest{
		User: &ttnpb.User{
			Ids:                 &ttnpb.UserIdentifiers{UserId: "test-user-id"},
			PrimaryEmailAddress: "test-user@example.com",
			Password:            "test password",
		},
	}

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserRegistryClient(cc)

		is.config.UserRegistration.Enabled = false

		_, err := reg.Create(ctx, req)
		if a.So(err, should.NotBeNil) {
			a.So(errors.Resemble(err, errUserRegistrationDisabled), should.BeTrue)
		}

		is.config.UserRegistration.Enabled = true
		is.config.UserRegistration.Invitation.Required = true

		_, err = reg.Create(ctx, req)
		if a.So(err, should.NotBeNil) {
			a.So(errors.Resemble(err, errInvitationTokenRequired), should.BeTrue)
		}

		is.config.UserRegistration.Invitation.Required = false

		created, err := reg.Create(ctx, req)
		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.GetIds(), should.Resemble, req.GetUser().GetIds())
		}
	}, withPrivateTestDatabase(p))
}

func TestUserUpdateInvalidPassword(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	usr := p.NewUser()
	usr.Password = "SuperSecretPassword"

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserRegistryClient(cc)

		t.Run("Incorrect", func(t *testing.T) {
			a, ctx := test.New(t)

			_, err := reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
				UserIds: usr.GetIds(),
				Old:     "WrongPassword",
				New:     "NewPassword",
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsUnauthenticated(err), should.BeTrue)
			}
		})

		t.Run("Weak", func(t *testing.T) {
			a, ctx := test.New(t)

			_, err := reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
				UserIds: usr.GetIds(),
				Old:     "SuperSecretPassword",
				New:     "Weak",
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsInvalidArgument(err), should.BeTrue)
			}
		})
	}, withPrivateTestDatabase(p))
}

func TestUsersCRUD(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminUsrKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminUsrCreds := rpcCreds(adminUsrKey)

	usr1 := p.NewUser()
	usr1.Password = "OldPassword"

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	keyWithoutRights, _ := p.NewAPIKey(usr1.GetEntityIdentifiers())
	credsWithoutRights := rpcCreds(keyWithoutRights)

	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserRegistryClient(cc)

		got, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   usr1.GetIds(),
			FieldMask: ttnpb.FieldMask("name", "admin", "created_at", "updated_at"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, usr1.Name)
			a.So(got.Admin, should.Equal, usr1.Admin)
			a.So(got.CreatedAt, should.Resemble, usr1.CreatedAt)
		}

		got, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   usr1.GetIds(),
			FieldMask: ttnpb.FieldMask("ids"),
		}, credsWithoutRights)
		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   usr1.GetIds(),
			FieldMask: ttnpb.FieldMask("attributes"),
		}, credsWithoutRights)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateUserRequest{
			User: &ttnpb.User{
				Ids:  usr1.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		updated, err = reg.Update(ctx, &ttnpb.UpdateUserRequest{
			User: &ttnpb.User{
				Ids:              usr1.GetIds(),
				State:            ttnpb.State_STATE_FLAGGED,
				StateDescription: "something is wrong",
			},
			FieldMask: ttnpb.FieldMask("state", "state_description"),
		}, adminUsrCreds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.State, should.Equal, ttnpb.State_STATE_FLAGGED)
			a.So(updated.StateDescription, should.Equal, "something is wrong")
		}

		updated, err = reg.Update(ctx, &ttnpb.UpdateUserRequest{
			User: &ttnpb.User{
				Ids:   usr1.GetIds(),
				State: ttnpb.State_STATE_APPROVED,
			},
			FieldMask: ttnpb.FieldMask("state"),
		}, adminUsrCreds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.State, should.Equal, ttnpb.State_STATE_APPROVED)
		}

		got, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   usr1.GetIds(),
			FieldMask: ttnpb.FieldMask("state", "state_description"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got.State, should.Equal, ttnpb.State_STATE_APPROVED)
			a.So(got.StateDescription, should.Equal, "")
		}

		passwordUpdateTime := time.Now().Truncate(time.Millisecond)

		_, err = reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIds: usr1.GetIds(),
			Old:     "OldPassword",
			New:     "NewPassword", // Meets minimum length requirement of 10 characters.
		}, creds)
		a.So(err, should.BeNil)

		afterUpdate, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   usr1.GetIds(),
			FieldMask: ttnpb.FieldMask("password_updated_at"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(afterUpdate, should.NotBeNil) {
			a.So(afterUpdate.PasswordUpdatedAt, should.NotBeNil)
			a.So(*ttnpb.StdTime(afterUpdate.PasswordUpdatedAt), should.HappenAfter, passwordUpdateTime)
		}

		_, err = reg.Delete(ctx, usr1.GetIds(), creds)
		a.So(err, should.BeNil)

		_, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   usr1.GetIds(),
			FieldMask: ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.NotBeNil) {
			// NOTE: For other entities, this would be a NotFound, but in this case
			// the user's credentials become invalid when the user is deleted.
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   usr1.GetIds(),
			FieldMask: ttnpb.FieldMask("name"),
		}, adminUsrCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		_, err = reg.Purge(ctx, usr1.GetIds(), creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Purge(ctx, usr1.GetIds(), adminUsrCreds)
		a.So(err, should.BeNil)
	}, withPrivateTestDatabase(p))
}
