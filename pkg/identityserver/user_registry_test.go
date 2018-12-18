// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

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
				UserIdentifiers: userID,
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

func TestUsersCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		user, creds := population.Users[defaultUserIdx], userCreds(defaultUserIdx)

		reg := ttnpb.NewUserRegistryClient(cc)

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
		newPassword := "updated user password"

		_, err = reg.UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
			UserIdentifiers: user.UserIdentifiers,
			Old:             oldPassword,
			New:             newPassword,
		}, creds)

		a.So(err, should.BeNil)

		pwUpdated, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"password", "password_updated_at"}},
		}, creds)

		oldPasswordMatch, err := auth.Password(pwUpdated.Password).Validate(oldPassword)
		if err != nil {
			panic(err)
		}
		newPasswordMatch, _ := auth.Password(pwUpdated.Password).Validate(newPassword)
		if err != nil {
			panic(err)
		}

		a.So(pwUpdated, should.NotBeNil)
		a.So(oldPasswordMatch, should.BeFalse)
		a.So(newPasswordMatch, should.BeTrue)
		a.So(pwUpdated.PasswordUpdatedAt, should.HappenAfter, passwordUpdateTime)
		a.So(err, should.BeNil)

		_, err = reg.Delete(ctx, &user.UserIdentifiers, creds)
		a.So(err, should.BeNil)

		empty, err := reg.Get(ctx, &ttnpb.GetUserRequest{
			UserIdentifiers: user.UserIdentifiers,
			FieldMask:       types.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(empty, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
}
