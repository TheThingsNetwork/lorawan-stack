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
	"fmt"
	"sort"
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestContactInfoStoreCRUD(t *T) {
	app1 := st.population.NewApplication(nil)
	cli1 := st.population.NewClient(nil)
	gtw1 := st.population.NewGateway(nil)
	org1 := st.population.NewOrganization(nil)
	usr1 := st.population.NewUser()
	usr1.PrimaryEmailAddressValidatedAt = nil

	s, ok := st.PrepareDB(t).(interface {
		Store

		is.ContactInfoStore
	})
	defer st.DestroyDB(t, false)
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement ContactInfoStore")
	}

	for _, ids := range []*ttnpb.EntityIdentifiers{
		app1.GetEntityIdentifiers(),
		cli1.GetEntityIdentifiers(),
		gtw1.GetEntityIdentifiers(),
		org1.GetEntityIdentifiers(),
		usr1.GetEntityIdentifiers(),
	} {
		t.Run(ids.EntityType(), func(t *T) {
			start := time.Now().Truncate(time.Second)
			info := &ttnpb.ContactInfo{
				ContactType:   ttnpb.ContactType_CONTACT_TYPE_TECHNICAL,
				ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
				Value:         "foo@example.com",
				Public:        true,
			}
			if ids.EntityType() == "user" && ids.IDString() == usr1.IDString() {
				info.Value = usr1.PrimaryEmailAddress
			}

			t.Run("SetContactInfo", func(t *T) {
				a, ctx := test.New(t)
				created, err := s.SetContactInfo(ctx, ids, []*ttnpb.ContactInfo{
					info,
				})
				if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) && a.So(created, should.HaveLength, 1) {
					a.So(created[0], should.Resemble, info)
				}
			})

			t.Run("GetContactInfo", func(t *T) {
				a, ctx := test.New(t)
				got, err := s.GetContactInfo(ctx, ids)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
					a.So(got[0], should.Resemble, info)
				}
			})

			validationID := fmt.Sprintf("%s_%s_validation", ids.EntityType(), ids.IDString())

			t.Run("CreateValidation_Expired", func(t *T) {
				a, ctx := test.New(t)
				_, err := s.CreateValidation(ctx, &ttnpb.ContactInfoValidation{
					Id:          validationID,
					Token:       "EXPIRED_TOKEN",
					Entity:      ids,
					ContactInfo: []*ttnpb.ContactInfo{info},
					ExpiresAt:   ttnpb.ProtoTimePtr(start.Add(-1 * time.Minute)),
				})
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
			})

			t.Run("CreateValidation", func(t *T) {
				a, ctx := test.New(t)
				created, err := s.CreateValidation(ctx, &ttnpb.ContactInfoValidation{
					Id:          validationID,
					Token:       "TOKEN",
					Entity:      ids,
					ContactInfo: []*ttnpb.ContactInfo{info},
					ExpiresAt:   ttnpb.ProtoTimePtr(start.Add(5 * time.Minute)),
				})
				if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
					a.So(created.Id, should.Equal, validationID)
					a.So(created.Token, should.Equal, "TOKEN")
					a.So(created.Entity, should.Resemble, ids)
					a.So(created.ContactInfo, should.Resemble, []*ttnpb.ContactInfo{info})
					a.So(*ttnpb.StdTime(created.ExpiresAt), should.Equal, start.Add(5*time.Minute))
					a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
				}
			})

			t.Run("CreateValidation_AfterCreate", func(t *T) {
				a, ctx := test.New(t)
				_, err := s.CreateValidation(ctx, &ttnpb.ContactInfoValidation{
					Id:          validationID,
					Token:       "OTHER_TOKEN",
					Entity:      ids,
					ContactInfo: []*ttnpb.ContactInfo{info},
					ExpiresAt:   ttnpb.ProtoTimePtr(start.Add(5 * time.Minute)),
				})
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsAlreadyExists(err), should.BeTrue)
				}
			})

			t.Run("Validate_Expired", func(t *T) {
				a, ctx := test.New(t)
				err := s.Validate(ctx, &ttnpb.ContactInfoValidation{
					Id:    validationID,
					Token: "EXPIRED_TOKEN",
				})
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})

			t.Run("Validate", func(t *T) {
				a, ctx := test.New(t)
				err := s.Validate(ctx, &ttnpb.ContactInfoValidation{
					Id:    validationID,
					Token: "TOKEN",
				})
				a.So(err, should.BeNil)
			})

			t.Run("Validate_AfterValidate", func(t *T) {
				a, ctx := test.New(t)
				err := s.Validate(ctx, &ttnpb.ContactInfoValidation{
					Id:    validationID,
					Token: "TOKEN",
				})
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsAlreadyExists(err), should.BeTrue)
				}
			})

			t.Run("Validate_Other", func(t *T) {
				a, ctx := test.New(t)
				err := s.Validate(ctx, &ttnpb.ContactInfoValidation{
					Id:    validationID,
					Token: "OTHER_TOKEN",
				})
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})

			t.Run("GetContactInfo_AfterValidate", func(t *T) {
				a, ctx := test.New(t)
				got, err := s.GetContactInfo(ctx, ids)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
					a.So(got[0].ValidatedAt, should.NotBeNil)
				}
			})

			if ids.EntityType() == "user" && ids.IDString() == usr1.IDString() {
				t.Run("GetUser_AfterValidate", func(t *T) {
					a, ctx := test.New(t)
					got, err := s.(store.UserStore).GetUser(ctx, usr1.GetIds(), fieldMask("primary_email_address_validated_at"))
					if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
						a.So(got.PrimaryEmailAddressValidatedAt, should.NotBeNil)
					}
				})
			}

			updatedInfo := &ttnpb.ContactInfo{
				ContactType:   ttnpb.ContactType_CONTACT_TYPE_TECHNICAL,
				ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
				Value:         "foo@example.com",
				Public:        false,
				ValidatedAt:   ttnpb.ProtoTimePtr(start.Add(time.Minute)),
			}
			extraInfo := &ttnpb.ContactInfo{
				ContactType:   ttnpb.ContactType_CONTACT_TYPE_TECHNICAL,
				ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
				Value:         "extra@example.com",
				Public:        true,
				ValidatedAt:   ttnpb.ProtoTimePtr(start.Add(time.Minute)),
			}

			t.Run("UpdateContactInfo", func(t *T) {
				a, ctx := test.New(t)
				updated, err := s.SetContactInfo(ctx, ids, []*ttnpb.ContactInfo{
					updatedInfo,
					extraInfo,
				})
				if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) && a.So(updated, should.HaveLength, 2) {
					sort.Slice(updated, func(i, j int) bool { return updated[i].Value < updated[j].Value })
					a.So(updated[0], should.Resemble, extraInfo)
					a.So(updated[1], should.Resemble, updatedInfo)
				}
			})

			t.Run("GetContactInfo_AfterUpdate", func(t *T) {
				a, ctx := test.New(t)
				got, err := s.GetContactInfo(ctx, ids)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 2) {
					sort.Slice(got, func(i, j int) bool { return got[i].Value < got[j].Value })
					a.So(got[0], should.Resemble, extraInfo)
					a.So(got[1], should.Resemble, updatedInfo)
				}
			})

			t.Run("DeleteEntityContactInfo", func(t *T) {
				a, ctx := test.New(t)
				err := s.DeleteEntityContactInfo(ctx, ids)
				a.So(err, should.BeNil)
			})
		})
	}
}

// TODO: Test Pagination (https://github.com/TheThingsNetwork/lorawan-stack/issues/5047).
