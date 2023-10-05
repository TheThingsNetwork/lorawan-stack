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
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

// TODO: This set of tests have to be updated after (https://github.com/TheThingsNetwork/lorawan-stack/issues/6567) is
// resolved.
func TestContactInfoValidation(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}
	usr1 := p.NewUser()
	usr1.ContactInfo = []*ttnpb.ContactInfo{ // nolint:staticcheck
		{
			ContactType:   ttnpb.ContactType_CONTACT_TYPE_OTHER,
			ContactMethod: ttnpb.ContactMethod_CONTACT_METHOD_EMAIL,
			Value:         "usr@test.com",
			Public:        false,
			ValidatedAt:   nil,
		},
	}
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewContactInfoRegistryClient(cc)

		retryInterval := test.Delay << 5
		is.config.UserRegistration.ContactInfoValidation.Required = true
		is.config.UserRegistration.ContactInfoValidation.TokenTTL = retryInterval * 2
		is.config.UserRegistration.ContactInfoValidation.RetryInterval = retryInterval

		t.Run("Request Validation", func(t *testing.T) { // nolint:paralleltest
			t.Run("Insufficient Rights", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				validation, err := reg.RequestValidation(ctx, usr1.GetEntityIdentifiers())
				a.So(validation, should.BeNil)
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			})
			t.Run("Valid Request for Validation", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				validation, err := reg.RequestValidation(ctx, usr1.GetEntityIdentifiers(), usr1Creds)
				a.So(err, should.BeNil)
				a.So(validation, should.NotBeNil)
			})
			t.Run("Request before email interval", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				validation, err := reg.RequestValidation(ctx, usr1.GetEntityIdentifiers(), usr1Creds)
				a.So(errors.IsAlreadyExists(err), should.BeTrue)
				a.So(validation, should.BeNil)
			})

			// Sleep enough time for the email interval to pass but not the expire time.
			time.Sleep(retryInterval)

			t.Run("Request after email interval", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				validation, err := reg.RequestValidation(ctx, usr1.GetEntityIdentifiers(), usr1Creds)
				a.So(err, should.BeNil)
				a.So(validation, should.NotBeNil)
			})
		})
	}, withPrivateTestDatabase(p))
}
