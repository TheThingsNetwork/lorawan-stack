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
	"context"
	"strings"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

func TestValidatePasswordStrength(t *testing.T) {
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
		"āaa	AAA123 ": true,
		"āaaAAA123 ": true,
	} {
		t.Run(p, func(t *testing.T) {
			a := assertions.New(t)

			testWithIdentityServer(t, func(is *IdentityServer, _ *grpc.ClientConn) {
				conf := *is.config
				conf.UserRegistration.PasswordRequirements.MinLength = 8
				conf.UserRegistration.PasswordRequirements.MaxLength = 1000
				conf.UserRegistration.PasswordRequirements.MinUppercase = 1
				conf.UserRegistration.PasswordRequirements.MinDigits = 1
				conf.UserRegistration.PasswordRequirements.MinSpecial = 1
				conf.UserRegistration.PasswordRequirements.RejectUserID = true
				conf.UserRegistration.PasswordRequirements.RejectCommon = true
				ctx := context.WithValue(is.Context(), ctxKey, &conf)

				err := is.validatePasswordStrength(ctx, "username", p)
				if ok {
					a.So(err, should.BeNil)
				} else {
					a.So(err, should.NotBeNil)
				}
			})
		})
	}
}

func TestTemporaryValidPassword(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := population.Users[userTestUserIdx].GetIds()

		reg := ttnpb.NewUserRegistryClient(cc)

		_, err := reg.CreateTemporaryPassword(ctx, &ttnpb.CreateTemporaryPasswordRequest{
			UserIds: userID,
		})

		a.So(err, should.BeNil)

		_, err = reg.CreateTemporaryPassword(ctx, &ttnpb.CreateTemporaryPasswordRequest{
			UserIds: userID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}
	})
}

func TestUserCreate(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.UserRegistration.Enabled = false
		userID := &ttnpb.UserIdentifiers{UserId: "test-user-id"}
		reg := ttnpb.NewUserRegistryClient(cc)
		_, err := reg.Create(ctx, &ttnpb.CreateUserRequest{
			User: &ttnpb.User{
				Ids:                 userID,
				PrimaryEmailAddress: "test-user@example.com",
				Password:            "test password",
			},
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.Resemble(err, errUserRegistrationDisabled), should.BeTrue)
		}
	})

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.UserRegistration.Invitation.Required = true
		userID := &ttnpb.UserIdentifiers{UserId: "test-user-id"}
		reg := ttnpb.NewUserRegistryClient(cc)
		_, err := reg.Create(ctx, &ttnpb.CreateUserRequest{
			User: &ttnpb.User{
				Ids:                 userID,
				PrimaryEmailAddress: "test-user@example.com",
				Password:            "test password",
			},
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.Resemble(err, errInvitationTokenRequired), should.BeTrue)
		}
	})

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := &ttnpb.UserIdentifiers{UserId: "test-user-id"}
		reg := ttnpb.NewUserRegistryClient(cc)
		created, err := reg.Create(ctx, &ttnpb.CreateUserRequest{
			User: &ttnpb.User{
				Ids:                 userID,
				PrimaryEmailAddress: "test-user@example.com",
				Password:            "test password",
			},
		})
		a.So(err, should.BeNil)
		a.So(created, should.NotBeNil)
	})
}

func TestUserUpdateInvalidPassword(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := population.Users[userTestUserIdx].GetIds()

		reg := ttnpb.NewUserRegistryClient(cc)

		_, err := reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIds: userID,
			Old:     "wrong-user-password",
			New:     "new password",
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}
	})
}

// TODO: Add when 2FA is enabled (https://github.com/TheThingsNetwork/lorawan-stack/issues/2)
// func TestUsersPermissionDenied(t *testing.T) {
// 	a := assertions.New(t)
// 	ctx := test.Context()

// 	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
// 		user := population.Users[userTestUserIdx]
// 		userID := user.UserIdentifiers

// 		reg := ttnpb.NewUserRegistryClient(cc)

// 		_, err := reg.Get(ctx, &ttnpb.GetUserRequest{
// 			UserIdentifiers: userID,
// 			FieldMask:       &pbtypes.FieldMask{Paths: []string{"name"}},
// 		})

// 		if a.So(err, should.NotBeNil) {
// 			a.So(errors.IsUnauthenticated(err), should.BeTrue)
// 		}

// 		_, err = reg.Update(ctx, &ttnpb.UpdateUserRequest{
// 			User: ttnpb.User{
// 				UserIdentifiers: userID,
// 				Name:            "new name",
// 			},
// 			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
// 		})

// 		if a.So(err, should.NotBeNil) {
// 			a.So(errors.IsPermissionDenied(err), should.BeTrue)
// 		}

// 		_, err = reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
// 			UserIdentifiers: userID,
// 			Old:             user.Password,
// 			New:             "new password",
// 		})

// 		if a.So(err, should.NotBeNil) {
// 			a.So(errors.IsPermissionDenied(err), should.BeTrue)
// 		}

// 		_, err = reg.Delete(ctx, &user.UserIdentifiers)

// 		if a.So(err, should.NotBeNil) {
// 			a.So(errors.IsPermissionDenied(err), should.BeTrue)
// 		}
// 	})
// }

func TestUsersWeakPassword(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserRegistryClient(cc)

		weakPassword := "weak" // Does not meet minimum length requirement of 10 characters.

		_, err := reg.Create(ctx, &ttnpb.CreateUserRequest{
			User: &ttnpb.User{
				Ids:      &ttnpb.UserIdentifiers{UserId: "test-user-id"},
				Password: weakPassword,
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}

		user, creds := population.Users[userTestUserIdx], userCreds(userTestUserIdx)

		oldPassword := user.Password
		newPassword := weakPassword

		_, err = reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIds: user.GetIds(),
			Old:     oldPassword,
			New:     newPassword,
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}

		afterUpdate, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   user.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"password_updated_at"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(afterUpdate, should.NotBeNil) && a.So(afterUpdate.PasswordUpdatedAt, should.NotBeNil) {
			a.So(*ttnpb.StdTime(afterUpdate.PasswordUpdatedAt), should.Equal, *ttnpb.StdTime(user.PasswordUpdatedAt))
		}
	})
}

func TestUsersCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserRegistryClient(cc)

		user, creds := population.Users[userTestUserIdx], userCreds(userTestUserIdx)
		credsWithoutRights := userCreds(userTestUserIdx, "key without rights")

		got, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   user.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name", "admin", "created_at", "updated_at"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, user.Name)
			a.So(got.Admin, should.Equal, user.Admin)
			a.So(got.CreatedAt, should.Resemble, user.CreatedAt)
		}

		got, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   user.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"ids"}},
		}, credsWithoutRights)

		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   user.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"attributes"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateUserRequest{
			User: &ttnpb.User{
				Ids:  user.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		updated, err = reg.Update(ctx, &ttnpb.UpdateUserRequest{
			User: &ttnpb.User{
				Ids:              user.GetIds(),
				State:            ttnpb.State_STATE_FLAGGED,
				StateDescription: "something is wrong",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"state", "state_description"}},
		}, userCreds(adminUserIdx))

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.State, should.Equal, ttnpb.State_STATE_FLAGGED)
			a.So(updated.StateDescription, should.Equal, "something is wrong")
		}

		updated, err = reg.Update(ctx, &ttnpb.UpdateUserRequest{
			User: &ttnpb.User{
				Ids:   user.GetIds(),
				State: ttnpb.State_STATE_APPROVED,
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"state"}},
		}, userCreds(adminUserIdx))

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.State, should.Equal, ttnpb.State_STATE_APPROVED)
		}

		got, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   user.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"state", "state_description"}},
		}, creds)

		if a.So(err, should.BeNil) {
			a.So(got.State, should.Equal, ttnpb.State_STATE_APPROVED)
			a.So(got.StateDescription, should.Equal, "")
		}

		passwordUpdateTime := time.Now()
		oldPassword := user.Password
		newPassword := "updated user password" // Meets minimum length requirement of 10 characters.

		_, err = reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIds: user.GetIds(),
			Old:     oldPassword,
			New:     newPassword,
		}, creds)

		a.So(err, should.BeNil)

		afterUpdate, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   user.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"password_updated_at"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(afterUpdate, should.NotBeNil) && a.So(afterUpdate.PasswordUpdatedAt, should.NotBeNil) {
			a.So(*ttnpb.StdTime(afterUpdate.PasswordUpdatedAt), should.HappenAfter, passwordUpdateTime)
		}

		_, err = reg.Delete(ctx, user.GetIds(), creds)

		a.So(err, should.BeNil)

		empty, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   user.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.NotBeNil)
		a.So(empty, should.BeNil)

		// NOTE: For other entities, this would be a NotFound, but in this case
		// the user's credentials become invalid when the user is deleted.
		a.So(errors.IsUnauthenticated(err), should.BeTrue)

		empty, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIds:   user.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, userCreds(adminUserIdx))

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(empty, should.BeNil)

		_, err = reg.Purge(ctx, user.GetIds(), creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Purge(ctx, user.GetIds(), userCreds(adminUserIdx))
		a.So(err, should.BeNil)
	})
}
