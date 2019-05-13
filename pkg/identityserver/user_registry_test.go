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
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func TestValidatePasswordStrength(t *testing.T) {
	a := assertions.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, _ *grpc.ClientConn) {
		conf := *is.config
		conf.UserRegistration.PasswordRequirements.MinLength = 8
		conf.UserRegistration.PasswordRequirements.MinUppercase = 1
		conf.UserRegistration.PasswordRequirements.MinDigits = 1
		conf.UserRegistration.PasswordRequirements.MinSpecial = 1
		ctx := context.WithValue(is.Context(), ctxKey, &conf)

		for p, ok := range map[string]bool{
			"aA0$":         false, // Fails length check.
			"aaa123!@#":    false, // Fails uppercase check.
			"aaaAAA!@#":    false, // Fails digits check.
			"aaaAAA123":    false, // Fails special check.
			"aaaAAA123!@#": true,
		} {
			err := is.validatePasswordStrength(ctx, p)
			if ok {
				a.So(err, should.BeNil)
			} else {
				a.So(err, should.NotBeNil)
			}
		}
	})
}

func TestTemporaryValidPassword(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := population.Users[defaultUserIdx].UserIdentifiers

		reg := ttnpb.NewUserRegistryClient(cc)

		_, err := reg.CreateTemporaryPassword(ctx, &ttnpb.CreateTemporaryPasswordRequest{
			UserIdentifiers: userID,
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
	})
}

func TestUsersPermissionNotRequired(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := ttnpb.UserIdentifiers{UserID: "test-user-id"}

		reg := ttnpb.NewUserRegistryClient(cc)

		created, err := reg.Create(ctx, &ttnpb.CreateUserRequest{
			User: ttnpb.User{
				UserIdentifiers:     userID,
				PrimaryEmailAddress: "test-user@example.com",
				Password:            "test password",
			},
		})

		a.So(created, should.NotBeNil)
		a.So(err, should.BeNil)
	})
}

func TestUserUpdateInvalidPassword(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := population.Users[defaultUserIdx].UserIdentifiers

		reg := ttnpb.NewUserRegistryClient(cc)

		_, err := reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIdentifiers: userID,
			Old:             "wrong-user-password",
			New:             "new password",
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsInternal(err), should.BeTrue)
	})
}

func TestUsersPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		user := population.Users[defaultUserIdx]
		userID := user.UserIdentifiers

		reg := ttnpb.NewUserRegistryClient(cc)

		_, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: userID,
			FieldMask:       types.FieldMask{Paths: []string{"name"}},
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsUnauthenticated(err), should.BeTrue)

		_, err = reg.Update(ctx, &ttnpb.UpdateUserRequest{
			User: ttnpb.User{
				UserIdentifiers: userID,
				Name:            "new name",
			},
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIdentifiers: userID,
			Old:             user.Password,
			New:             "new password",
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = reg.Delete(ctx, &user.UserIdentifiers)

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestUsersWeakPassword(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserRegistryClient(cc)

		weakPassword := "weak" // Does not meet minimum length requirement of 10 characters.

		_, err := reg.Create(ctx, &ttnpb.CreateUserRequest{
			User: ttnpb.User{
				UserIdentifiers: ttnpb.UserIdentifiers{UserID: "test-user-id"},
				Password:        weakPassword,
			},
		})
		a.So(err, should.NotBeNil)
		a.So(errors.IsInvalidArgument(err), should.BeTrue)

		user, creds := population.Users[defaultUserIdx], userCreds(defaultUserIdx)

		oldPassword := user.Password
		newPassword := weakPassword

		_, err = reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIdentifiers: user.UserIdentifiers,
			Old:             oldPassword,
			New:             newPassword,
		}, creds)

		a.So(err, should.NotBeNil)
		a.So(errors.IsInvalidArgument(err), should.BeTrue)

		afterUpdate, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"password_updated_at"}},
		}, creds)

		a.So(err, should.BeNil)
		a.So(afterUpdate.PasswordUpdatedAt, should.Equal, user.PasswordUpdatedAt)
	})
}

func TestUsersCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserRegistryClient(cc)

		user, creds := population.Users[defaultUserIdx], userCreds(defaultUserIdx)
		credsWithoutRights := userCreds(defaultUserIdx, "key without rights")

		got, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"name", "admin", "created_at", "updated_at"}},
		}, creds)

		a.So(got, should.NotBeNil)
		a.So(got.Name, should.Equal, user.Name)
		a.So(got.Admin, should.Equal, user.Admin)
		a.So(got.CreatedAt, should.Equal, user.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, user.UpdatedAt)
		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"ids"}},
		}, credsWithoutRights)

		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"attributes"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updateTime := time.Now()
		updatedName := "updated user name"

		updated, err := reg.Update(ctx, &ttnpb.UpdateUserRequest{
			User: ttnpb.User{
				UserIdentifiers: user.UserIdentifiers,
				Name:            updatedName,
			},
			FieldMask: types.FieldMask{Paths: []string{"name", "created_at", "updated_at"}},
		}, creds)

		a.So(updated, should.NotBeNil)
		a.So(updated.Name, should.Equal, updatedName)
		a.So(updated.CreatedAt, should.Equal, user.CreatedAt)
		a.So(updated.UpdatedAt, should.HappenAfter, updateTime)
		a.So(err, should.BeNil)

		passwordUpdateTime := time.Now()
		oldPassword := user.Password
		newPassword := "updated user password" // Meets minimum length requirement of 10 characters.

		_, err = reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIdentifiers: user.UserIdentifiers,
			Old:             oldPassword,
			New:             newPassword,
		}, creds)

		a.So(err, should.BeNil)

		afterUpdate, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"password_updated_at"}},
		}, creds)

		a.So(err, should.BeNil)
		a.So(afterUpdate.PasswordUpdatedAt, should.HappenAfter, passwordUpdateTime)

		_, err = reg.Delete(ctx, &user.UserIdentifiers, creds)
		a.So(err, should.BeNil)

		empty, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(empty, should.BeNil)
		a.So(err, should.NotBeNil)

		// NOTE: For other entities, this would be a NotFound, but in this case
		// the user's credentials become invalid when the user is deleted.
		a.So(errors.IsUnauthenticated(err), should.BeTrue)

		empty, err = reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"name"}},
		}, userCreds(adminUserIdx))

		a.So(empty, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
}
