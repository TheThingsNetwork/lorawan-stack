// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestEmailValidationStore tests the email validation store operations.
func (st *StoreTest) TestEmailValidationStore(t *T) {
	usr1 := st.population.NewUser()
	usr1.PrimaryEmailAddress = "usr1@email.com"
	usr1.PrimaryEmailAddressValidatedAt = nil

	s, ok := st.PrepareDB(t).(interface {
		Store

		store.EmailValidationStore
		store.UserStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement ContactInfoStore")
	}
	defer s.Close()

	start := time.Now().Truncate(time.Second)
	validation := &ttnpb.EmailValidation{
		Id:        fmt.Sprintf("%s_%s_validation", usr1.EntityType(), usr1.IDString()),
		Token:     fmt.Sprintf("%s_%s_token", usr1.EntityType(), usr1.IDString()),
		Address:   usr1.PrimaryEmailAddress,
		ExpiresAt: timestamppb.New(start.Add(test.Delay << 10)),
	}

	t.Run("Create", func(t *T) {
		a, ctx := test.New(t)
		v, err := s.CreateEmailValidation(ctx, validation)
		a.So(err, should.BeNil)
		a.So(v.Id, should.Equal, validation.Id)
		a.So(v.Token, should.Equal, validation.Token)
		a.So(v.Address, should.Equal, validation.Address)

		_, err = s.CreateEmailValidation(ctx, validation)
		a.So(errors.IsAlreadyExists(err), should.BeTrue)
	})

	t.Run("Get", func(t *T) {
		a, ctx := test.New(t)
		v, err := s.GetEmailValidation(ctx, validation)
		a.So(err, should.BeNil)
		a.So(v.Id, should.Equal, validation.Id)
		a.So(v.Token, should.Equal, validation.Token)
		a.So(v.Address, should.Equal, validation.Address)

		// Taking the updateAt timestamp to compare it on the next test.
		validation.UpdatedAt = v.UpdatedAt
	})

	t.Run("Get Refreshable", func(t *T) {
		a, ctx := test.New(t)

		// Getting a email validation that was updated before (now() - test.Delay << 5).
		_, err := s.GetRefreshableEmailValidation(ctx, usr1.Ids, test.Delay<<5)
		a.So(errors.IsNotFound(err), should.BeTrue)

		time.Sleep(test.Delay)

		// Getting a email validation that was updated before (now() - test.Delay).
		v, err := s.GetRefreshableEmailValidation(ctx, usr1.Ids, test.Delay)
		a.So(err, should.BeNil)
		a.So(v.Id, should.Equal, validation.Id)
		a.So(v.Token, should.Equal, validation.Token)
		a.So(v.Address, should.Equal, validation.Address)
		a.So(v.UpdatedAt, should.Resemble, validation.UpdatedAt)
	})

	t.Run("Refresh", func(t *T) {
		a, ctx := test.New(t)

		err := s.RefreshEmailValidation(ctx, validation)
		a.So(err, should.BeNil)

		v, err := s.GetEmailValidation(ctx, validation)
		a.So(err, should.BeNil)
		a.So(v.Id, should.Equal, validation.Id)
		a.So(v.Token, should.Equal, validation.Token)
		a.So(v.Address, should.Equal, validation.Address)
		a.So(v.UpdatedAt.AsTime().After(validation.UpdatedAt.AsTime()), should.BeTrue)
	})

	t.Run("Expire", func(t *T) {
		a, ctx := test.New(t)

		err := s.ExpireEmailValidation(ctx, validation)
		a.So(err, should.BeNil)

		_, err = s.GetEmailValidation(ctx, validation)
		a.So(errors.IsFailedPrecondition(err), should.BeTrue)
	})
}
