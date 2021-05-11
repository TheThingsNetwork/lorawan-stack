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

package store

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestContactInfoStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	now := cleanTime(time.Now())

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &ContactInfo{}, &Application{}, &ContactInfoValidation{})

		appStore := GetApplicationStore(db)

		app, err := appStore.CreateApplication(ctx, &ttnpb.Application{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "foo"},
		})

		a.So(err, should.BeNil)

		s := GetContactInfoStore(db)

		contactInfo, err := s.GetContactInfo(ctx, app.ApplicationIdentifiers)

		a.So(err, should.BeNil)
		a.So(contactInfo, should.BeEmpty)

		_, err = s.SetContactInfo(ctx, app.ApplicationIdentifiers, []*ttnpb.ContactInfo{
			{ContactType: ttnpb.CONTACT_TYPE_TECHNICAL, ContactMethod: ttnpb.CONTACT_METHOD_EMAIL, Value: "foo@example.com", ValidatedAt: &now},
			{ContactType: ttnpb.CONTACT_TYPE_BILLING, ContactMethod: ttnpb.CONTACT_METHOD_EMAIL, Value: "admin@example.com"},
		})

		a.So(err, should.BeNil)

		contactInfo, err = s.GetContactInfo(ctx, app.ApplicationIdentifiers)

		a.So(err, should.BeNil)
		a.So(contactInfo, should.HaveLength, 2)

		_, err = s.SetContactInfo(ctx, app.ApplicationIdentifiers, []*ttnpb.ContactInfo{
			{ContactType: ttnpb.CONTACT_TYPE_TECHNICAL, ContactMethod: ttnpb.CONTACT_METHOD_EMAIL, Value: "bar@example.com"},
			{ContactType: ttnpb.CONTACT_TYPE_TECHNICAL, ContactMethod: ttnpb.CONTACT_METHOD_EMAIL, Value: "foo@example.com"},
			{ContactType: ttnpb.CONTACT_TYPE_ABUSE, ContactMethod: ttnpb.CONTACT_METHOD_EMAIL, Value: "foo@example.com"},
			{ContactType: ttnpb.CONTACT_TYPE_BILLING, ContactMethod: ttnpb.CONTACT_METHOD_EMAIL, Value: "admin@example.com"},
			{ContactType: ttnpb.CONTACT_TYPE_BILLING, ContactMethod: ttnpb.CONTACT_METHOD_EMAIL, Value: "other_admin@example.com"},
		})

		a.So(err, should.BeNil)

		contactInfo, err = s.GetContactInfo(ctx, app.ApplicationIdentifiers)

		a.So(err, should.BeNil)
		if a.So(contactInfo, should.HaveLength, 5) {
			for _, contactInfo := range contactInfo {
				if contactInfo.ContactType == ttnpb.CONTACT_TYPE_TECHNICAL && contactInfo.Value == "foo@example.com" {
					a.So(contactInfo.ValidatedAt, should.NotBeNil)
				}
			}
		}

		err = s.DeleteEntityContactInfo(ctx, app.ApplicationIdentifiers)

		a.So(err, should.BeNil)

		contactInfo, err = s.GetContactInfo(ctx, app.ApplicationIdentifiers)

		a.So(contactInfo, should.HaveLength, 0)
	})
}

func TestContactInfoValidation(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &ContactInfo{}, &Account{}, &User{}, &ContactInfoValidation{})

		usrStore := GetUserStore(db)

		usr, err := usrStore.CreateUser(ctx, &ttnpb.User{
			UserIdentifiers: ttnpb.UserIdentifiers{UserId: "foo"},
		})

		a.So(err, should.BeNil)

		s := GetContactInfoStore(db)

		info, err := s.SetContactInfo(ctx, usr.UserIdentifiers, []*ttnpb.ContactInfo{
			{ContactMethod: ttnpb.CONTACT_METHOD_EMAIL, Value: "foo@example.com"},
		})

		a.So(err, should.BeNil)

		expiresAt := time.Now().Add(time.Hour)

		_, err = s.CreateValidation(ctx, &ttnpb.ContactInfoValidation{
			ID:          "validation-id",
			Token:       "validation-token",
			Entity:      usr.GetEntityIdentifiers(),
			ContactInfo: info,
			ExpiresAt:   &expiresAt,
		})

		a.So(err, should.BeNil)

		_, err = s.CreateValidation(ctx, &ttnpb.ContactInfoValidation{
			ID:          "validation-id",
			Token:       "other-token",
			Entity:      usr.GetEntityIdentifiers(),
			ContactInfo: info,
			ExpiresAt:   &expiresAt,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}

		err = s.Validate(ctx, &ttnpb.ContactInfoValidation{
			ID:    "validation-id",
			Token: "validation-token",
		})

		a.So(err, should.BeNil)

		info, err = s.GetContactInfo(ctx, usr)

		a.So(err, should.BeNil)
		if a.So(info, should.HaveLength, 1) {
			a.So(info[0].ValidatedAt, should.NotBeNil)
		}

		err = s.Validate(ctx, &ttnpb.ContactInfoValidation{
			ID:    "validation-id",
			Token: "validation-token",
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}
	})
}
