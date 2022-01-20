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

package storetest

import (
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestUserStoreCRUD(t *T) {
	s, ok := st.PrepareDB(t).(interface {
		Store
		is.UserStore
	})
	defer st.DestroyDB(t, true, "pictures")
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement UserStore")
	}

	mask := fieldMask(
		"name",
		"description",
		"attributes",
		"primary_email_address",
		"primary_email_address_validated_at",
		"password",
		"password_updated_at",
		"require_password_update",
		"state",
		"state_description",
		"admin",
		"temporary_password",
		"temporary_password_created_at",
		"temporary_password_expires_at",
		"profile_picture",
	)

	picture := &ttnpb.Picture{
		Embedded: &ttnpb.Picture_Embedded{
			MimeType: "image/png",
			Data:     []byte("foobarbaz"),
		},
	}
	var created *ttnpb.User

	t.Run("CreateUser", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)
		stamp := start.Add(-1 * time.Minute)

		created, err = s.CreateUser(ctx, &ttnpb.User{
			Ids:                            &ttnpb.UserIdentifiers{UserId: "foo"},
			Name:                           "Foo Name",
			Description:                    "Foo Description",
			Attributes:                     attributes,
			PrimaryEmailAddress:            "foo@example.com",
			PrimaryEmailAddressValidatedAt: ttnpb.ProtoTimePtr(stamp),
			Password:                       "password_hash",
			PasswordUpdatedAt:              ttnpb.ProtoTimePtr(stamp),
			RequirePasswordUpdate:          true,
			State:                          ttnpb.State_STATE_APPROVED,
			StateDescription:               "welcome!",
			Admin:                          true,
			TemporaryPassword:              "temporary_hash",
			TemporaryPasswordCreatedAt:     ttnpb.ProtoTimePtr(stamp),
			TemporaryPasswordExpiresAt:     ttnpb.ProtoTimePtr(stamp.Add(time.Hour)),
			ProfilePicture:                 picture,
		})

		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.GetIds().GetUserId(), should.Equal, "foo")
			a.So(created.Name, should.Equal, "Foo Name")
			a.So(created.Description, should.Equal, "Foo Description")
			a.So(created.Attributes, should.Resemble, attributes)
			a.So(created.PrimaryEmailAddress, should.Equal, "foo@example.com")
			a.So(*ttnpb.StdTime(created.PrimaryEmailAddressValidatedAt), should.Equal, stamp)
			a.So(created.Password, should.Equal, "password_hash")
			a.So(*ttnpb.StdTime(created.PasswordUpdatedAt), should.Equal, stamp)
			a.So(created.RequirePasswordUpdate, should.BeTrue)
			a.So(created.State, should.Equal, ttnpb.State_STATE_APPROVED)
			a.So(created.StateDescription, should.Equal, "welcome!")
			a.So(created.Admin, should.BeTrue)
			a.So(created.TemporaryPassword, should.Equal, "temporary_hash")
			a.So(*ttnpb.StdTime(created.TemporaryPasswordCreatedAt), should.Equal, stamp)
			a.So(*ttnpb.StdTime(created.TemporaryPasswordExpiresAt), should.Equal, stamp.Add(time.Hour))
			a.So(created.ProfilePicture, should.Resemble, picture)
			a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("CreateUser_AfterCreate", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.CreateUser(ctx, &ttnpb.User{
			Ids: &ttnpb.UserIdentifiers{UserId: "foo"},
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}

		_, err = s.CreateUser(ctx, &ttnpb.User{
			Ids:                 &ttnpb.UserIdentifiers{UserId: "other"},
			PrimaryEmailAddress: "foo@example.com",
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}
	})

	t.Run("GetUser", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetUser(ctx, &ttnpb.UserIdentifiers{UserId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetUser_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetUser(ctx, &ttnpb.UserIdentifiers{UserId: "other"}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.GetUser(ctx, &ttnpb.UserIdentifiers{UserId: ""}, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetUserByPrimaryEmailAddress", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetUserByPrimaryEmailAddress(ctx, "foo@example.com", mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("FindUsers", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindUsers(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	t.Run("ListAdmins", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListAdmins(ctx, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	updatedPicture := &ttnpb.Picture{
		Sizes: map[uint32]string{0: "https://example.com/profile_picture.jpg"},
	}
	var updated *ttnpb.User

	t.Run("UpdateUser", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)
		stamp := start.Add(time.Minute)

		updated, err = s.UpdateUser(ctx, &ttnpb.User{
			Ids:                            &ttnpb.UserIdentifiers{UserId: "foo"},
			Name:                           "New Foo Name",
			Description:                    "New Foo Description",
			Attributes:                     updatedAttributes,
			PrimaryEmailAddress:            "updated@example.com",
			PrimaryEmailAddressValidatedAt: ttnpb.ProtoTimePtr(stamp),
			Password:                       "updated_password_hash",
			PasswordUpdatedAt:              ttnpb.ProtoTimePtr(stamp),
			RequirePasswordUpdate:          false,
			State:                          ttnpb.State_STATE_FLAGGED,
			StateDescription:               "flagged",
			Admin:                          false,
			TemporaryPassword:              "updated_temporary_hash",
			TemporaryPasswordCreatedAt:     ttnpb.ProtoTimePtr(stamp),
			TemporaryPasswordExpiresAt:     ttnpb.ProtoTimePtr(stamp.Add(time.Hour)),
			ProfilePicture:                 updatedPicture,
		}, mask)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.GetIds().GetUserId(), should.Equal, "foo")
			a.So(updated.Name, should.Equal, "New Foo Name")
			a.So(updated.Description, should.Equal, "New Foo Description")
			a.So(updated.Attributes, should.Resemble, updatedAttributes)
			a.So(updated.PrimaryEmailAddress, should.Equal, "updated@example.com")
			a.So(*ttnpb.StdTime(updated.PrimaryEmailAddressValidatedAt), should.Equal, stamp)
			a.So(updated.Password, should.Equal, "updated_password_hash")
			a.So(*ttnpb.StdTime(updated.PasswordUpdatedAt), should.Equal, stamp)
			a.So(updated.RequirePasswordUpdate, should.BeFalse)
			a.So(updated.State, should.Equal, ttnpb.State_STATE_FLAGGED)
			a.So(updated.StateDescription, should.Equal, "flagged")
			a.So(updated.Admin, should.BeFalse)
			a.So(updated.TemporaryPassword, should.Equal, "updated_temporary_hash")
			a.So(*ttnpb.StdTime(updated.TemporaryPasswordCreatedAt), should.Equal, stamp)
			a.So(*ttnpb.StdTime(updated.TemporaryPasswordExpiresAt), should.Equal, stamp.Add(time.Hour))
			a.So(updated.ProfilePicture, should.Resemble, updatedPicture)
			a.So(*ttnpb.StdTime(updated.CreatedAt), should.Equal, *ttnpb.StdTime(created.CreatedAt))
			a.So(*ttnpb.StdTime(updated.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("UpdateUser_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.UpdateUser(ctx, &ttnpb.User{
			Ids: &ttnpb.UserIdentifiers{UserId: "other"},
		}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.UpdateUser(ctx, &ttnpb.User{
		// 	Ids: &ttnpb.UserIdentifiers{UserId: ""},
		// }, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetUser_AfterUpdate", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetUser(ctx, &ttnpb.UserIdentifiers{UserId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("DeleteUser", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteUser(ctx, &ttnpb.UserIdentifiers{UserId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("DeleteUser_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteUser(ctx, &ttnpb.UserIdentifiers{UserId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.DeleteUser(ctx, &ttnpb.UserIdentifiers{UserId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetUser_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetUser(ctx, &ttnpb.UserIdentifiers{UserId: "foo"}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("FindUsers_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindUsers(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("GetDeletedUser", func(t *T) {
		a, ctx := test.New(t)
		ctx = store.WithSoftDeleted(ctx, true)
		got, err := s.GetUser(ctx, &ttnpb.UserIdentifiers{UserId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			if a.So(got.DeletedAt, should.NotBeNil) {
				got.DeletedAt = nil // Unset DeletedAt for the should.Resemble below.
			}
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("FindDeletedUsers", func(t *T) {
		a, ctx := test.New(t)
		ctx = store.WithSoftDeleted(ctx, true)
		got, err := s.FindUsers(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			if a.So(got[0].DeletedAt, should.NotBeNil) {
				got[0].DeletedAt = nil // Unset DeletedAt for the should.Resemble below.
			}
			a.So(got[0], should.Resemble, updated)
		}
	})

	t.Run("RestoreUser", func(t *T) {
		a, ctx := test.New(t)
		err := s.RestoreUser(ctx, &ttnpb.UserIdentifiers{UserId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("RestoreUser_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.RestoreUser(ctx, &ttnpb.UserIdentifiers{UserId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.RestoreUser(ctx, &ttnpb.UserIdentifiers{UserId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetUser_AfterRestore", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetUser(ctx, &ttnpb.UserIdentifiers{UserId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("PurgeUser", func(t *T) {
		a, ctx := test.New(t)
		err := s.PurgeUser(ctx, &ttnpb.UserIdentifiers{UserId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("PurgeUser_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.PurgeUser(ctx, &ttnpb.UserIdentifiers{UserId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.PurgeUser(ctx, &ttnpb.UserIdentifiers{UserId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})
}

// TODO: Test Pagination (https://github.com/TheThingsNetwork/lorawan-stack/issues/5047).
