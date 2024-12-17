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

package identityserver

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestEmailValidation(t *testing.T) {
	p := &storetest.Population{}

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	usr1 := p.NewUser()
	usr1.PrimaryEmailAddress = "usr1@email.com"
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	usr2 := p.NewUser()
	usr2.PrimaryEmailAddress = "usr2@email.com"
	usr2Key, _ := p.NewAPIKey(usr2.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr2Creds := rpcCreds(usr2Key)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		// Configuration necessary for testing refresh of validation email.
		retryInterval := test.Delay << 6
		tokenTTL := test.Delay << 8
		is.config.UserRegistration.ContactInfoValidation.Required = true
		is.config.UserRegistration.ContactInfoValidation.TokenTTL = tokenTTL
		is.config.UserRegistration.ContactInfoValidation.RetryInterval = retryInterval

		reg := ttnpb.NewEmailValidationRegistryClient(cc)
		usrReg := ttnpb.NewUserRegistryClient(cc)

		t.Run("Request validation", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)

			// No rights.
			_, err := reg.RequestValidation(ctx, usr1.Ids)
			a.So(err, should.NotBeNil)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)

			// Only reader rights.
			_, err = reg.RequestValidation(ctx, usr1.Ids, readOnlyAdminKeyCreds)
			a.So(err, should.NotBeNil)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)

			// Proper Request.
			validation, err := reg.RequestValidation(ctx, usr1.Ids, usr1Creds)
			a.So(err, should.BeNil)
			a.So(validation.Address, should.Equal, usr1.PrimaryEmailAddress)
			a.So(len(validation.Id), should.BeGreaterThan, 0)
			a.So(len(validation.Id), should.BeLessThanOrEqualTo, 64)
			a.So(len(validation.Token), should.Equal, 0) // Token is emptied before sending to the response.

			// Request again before retry interval is passed, expect already exists error for this user's email address.
			_, err = reg.RequestValidation(ctx, usr1.Ids, usr1Creds)
			a.So(err, should.NotBeNil)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)

			time.Sleep(retryInterval)

			// This should trigger the refresh of the validation email.
			newValidation, err := reg.RequestValidation(ctx, usr1.Ids, usr1Creds)
			a.So(err, should.BeNil)
			a.So(newValidation.Address, should.Equal, validation.Address)
			a.So(newValidation.Id, should.Equal, validation.Id)
		})

		t.Run("Validate", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)

			// Token's value is a secret which is only known to the user and the DB. For testing purposes, we fetch the
			// validation information based on the email address and then validate the token.
			val, err := is.store.GetRefreshableEmailValidation(ctx, usr1.Ids, 0)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			_, err = reg.Validate(ctx, &ttnpb.ValidateEmailRequest{Id: val.Id, Token: val.Token})
			a.So(err, should.BeNil)

			usr, err := usrReg.Get(
				ctx,
				&ttnpb.GetUserRequest{
					UserIds:   usr1.Ids,
					FieldMask: ttnpb.FieldMask("primary_email_address_validated_at"),
				},
				usr1Creds,
			)
			a.So(err, should.BeNil)
			a.So(usr.PrimaryEmailAddressValidatedAt, should.NotBeNil)
		})

		t.Run("Use expired token", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			validation, err := reg.RequestValidation(ctx, usr2.Ids, usr2Creds)
			a.So(err, should.BeNil)
			a.So(validation.Address, should.Equal, usr2.PrimaryEmailAddress)

			// Token's value is a secret which is only known to the user and the DB. For testing purposes, we fetch the
			// validation information based on the email address and then validate the token.
			val, err := is.store.GetRefreshableEmailValidation(ctx, usr2.Ids, 0)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			time.Sleep(tokenTTL)

			_, err = reg.Validate(ctx, &ttnpb.ValidateEmailRequest{Id: val.Id, Token: val.Token})
			a.So(errors.IsFailedPrecondition(err), should.BeTrue)
		})
	}, withPrivateTestDatabase(p))
}
