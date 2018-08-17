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
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/identityserver/oauth"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc/metadata"
)

func TestBuildauthorizationData(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	alice := newTestUsers()["alice"]
	client := newTestClient()

	// Generate access token and save it in the store.
	token, err := auth.GenerateAccessToken("issuer")
	a.So(err, should.BeNil)

	tdata := store.AccessData{
		AccessToken: token,
		UserID:      alice.UserID,
		ClientID:    client.ClientID,
		CreatedAt:   time.Now().UTC(),
		ExpiresIn:   time.Duration(time.Hour),
		Scope:       oauth.Scope(ttnpb.AllUserRights()),
		RedirectURI: "http://localhost/auth/callback",
	}
	err = is.store.OAuth.SaveAccessToken(tdata)
	a.So(err, should.BeNil)
	defer func() {
		is.store.OAuth.DeleteAccessToken(token)
	}()

	// Generate a different access token but do not save it.
	token2, err := auth.GenerateAccessToken("issuer")
	a.So(err, should.BeNil)

	// Generate another access token but save it as expired token.
	token3, err := auth.GenerateAccessToken("issuer")
	a.So(err, should.BeNil)

	tdata = store.AccessData{
		AccessToken: token3,
		UserID:      alice.UserID,
		ClientID:    client.ClientID,
		CreatedAt:   time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		ExpiresIn:   time.Duration(time.Minute * 10),
		Scope:       oauth.Scope([]ttnpb.Right{ttnpb.RIGHT_USER_AUTHORIZED_CLIENTS}),
		RedirectURI: "http://localhost/auth/callback",
	}
	err = is.store.OAuth.SaveAccessToken(tdata)
	a.So(err, should.BeNil)
	defer func() {
		is.store.OAuth.DeleteAccessToken(token3)
	}()

	// Prepare user API keys.
	ukey, err := auth.GenerateUserAPIKey("issuer")
	a.So(err, should.BeNil)

	err = is.store.Users.SaveAPIKey(alice.UserIdentifiers, ttnpb.APIKey{
		Name:   "key",
		Key:    ukey,
		Rights: []ttnpb.Right{ttnpb.Right(0)},
	})
	a.So(err, should.BeNil)
	defer func() {
		is.store.Users.DeleteAPIKey(alice.UserIdentifiers, "key")
	}()

	ukey2, err := auth.GenerateUserAPIKey("issuer")
	a.So(err, should.BeNil)

	// Prepare application API keys.
	app := &ttnpb.Application{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "foo-app",
		},
	}
	err = is.store.Applications.Create(app)
	a.So(err, should.BeNil)
	defer func() {
		is.store.Applications.Delete(app.ApplicationIdentifiers)
	}()

	akey, err := auth.GenerateApplicationAPIKey("issuer")
	a.So(err, should.BeNil)

	err = is.store.Applications.SaveAPIKey(app.ApplicationIdentifiers, ttnpb.APIKey{
		Name:   "key",
		Key:    akey,
		Rights: []ttnpb.Right{ttnpb.Right(0)},
	})
	a.So(err, should.BeNil)
	defer func() {
		is.store.Applications.DeleteAPIKey(app.ApplicationIdentifiers, "key")
	}()

	akey2, err := auth.GenerateApplicationAPIKey("issuer")
	a.So(err, should.BeNil)

	// Prepare gateway API keys.
	gtw := &ttnpb.Gateway{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{
			GatewayID: "foo-gtw",
		},
	}
	err = is.store.Gateways.Create(gtw)
	a.So(err, should.BeNil)
	defer func() {
		is.store.Gateways.Delete(gtw.GatewayIdentifiers)
	}()

	gkey, err := auth.GenerateGatewayAPIKey("issuer")
	a.So(err, should.BeNil)

	err = is.store.Gateways.SaveAPIKey(gtw.GatewayIdentifiers, ttnpb.APIKey{
		Name:   "key",
		Key:    gkey,
		Rights: []ttnpb.Right{ttnpb.Right(0)},
	})
	a.So(err, should.BeNil)
	defer func() {
		is.store.Gateways.DeleteAPIKey(gtw.GatewayIdentifiers, "key")
	}()

	gkey2, err := auth.GenerateGatewayAPIKey("issuer")
	a.So(err, should.BeNil)

	// Prepare organization API keys.
	org := &ttnpb.Organization{
		OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{
			OrganizationID: "foo-org",
		},
	}
	err = is.store.Organizations.Create(org)
	a.So(err, should.BeNil)
	defer func() {
		is.store.Organizations.Delete(org.OrganizationIdentifiers)
	}()

	okey, err := auth.GenerateOrganizationAPIKey("issuer")
	a.So(err, should.BeNil)

	err = is.store.Organizations.SaveAPIKey(org.OrganizationIdentifiers, ttnpb.APIKey{
		Name:   "key",
		Key:    okey,
		Rights: []ttnpb.Right{ttnpb.Right(0)},
	})
	a.So(err, should.BeNil)
	defer func() {
		is.store.Organizations.DeleteAPIKey(org.OrganizationIdentifiers, "key")
	}()

	okey2, err := auth.GenerateOrganizationAPIKey("issuer")
	a.So(err, should.BeNil)

	// Actual test cases section.
	for _, tc := range []struct {
		name      string
		authType  string
		authValue string
		res       *authorizationData
		success   bool
	}{
		{
			// Returns empty authorization data as no authorization credentials were found.
			"Empty",
			"",
			"",
			new(authorizationData),
			true,
		},
		{
			// Returns error as only valid AuthType is `Bearer`.
			"InvalidAuthType",
			"Key",
			"",
			nil,
			false,
		},
		{
			// Returns error as the token can not be decoded using `auth.DecodeTokenOrKey(string)`.
			"InvalidAuthValue",
			"Bearer",
			"fake",
			nil,
			false,
		},
		{
			// Returns error because the token does not exist in the database.
			"NotExistingToken",
			"Bearer",
			token2,
			nil,
			false,
		},
		{
			// Returns error because the token is expired.
			"ExpiredToken",
			"Bearer",
			token3,
			nil,
			false,
		},
		{
			"NormalWithUserToken",
			"Bearer",
			token,
			&authorizationData{
				EntityIdentifiers: ttnpb.UserIdentifiers{
					UserID: alice.UserID,
				},
				Source: auth.Token,
				Rights: ttnpb.AllUserRights(),
			},
			true,
		},
		{
			// Returns error because the user API does not exist.
			"NotExistingApplicationAPIKey",
			"Bearer",
			ukey2,
			nil,
			false,
		},
		{
			"NormalWithUserAPIKey",
			"Bearer",
			ukey,
			&authorizationData{
				EntityIdentifiers: alice.UserIdentifiers,
				Source:            auth.Key,
				Rights:            []ttnpb.Right{ttnpb.Right(0)},
			},
			true,
		},
		{
			// Returns error because the application API does not exist.
			"NotExistingUserAPIKey",
			"Bearer",
			akey2,
			nil,
			false,
		},
		{
			"NormalWithApplicationAPIKey",
			"Bearer",
			akey,
			&authorizationData{
				EntityIdentifiers: app.ApplicationIdentifiers,
				Source:            auth.Key,
				Rights:            []ttnpb.Right{ttnpb.Right(0)},
			},
			true,
		},
		{
			// Returns error because the gateway API does not exist.
			"NotExistingGatewayAPIKey",
			"Bearer",
			gkey2,
			nil,
			false,
		},
		{
			"NormalWithGatewayAPIKey",
			"Bearer",
			gkey,
			&authorizationData{
				EntityIdentifiers: gtw.GatewayIdentifiers,
				Source:            auth.Key,
				Rights:            []ttnpb.Right{ttnpb.Right(0)},
			},
			true,
		},
		{
			// Returns error because the organization API does not exist.
			"NotExistingOrganizationAPIKey",
			"Bearer",
			okey2,
			nil,
			false,
		},
		{
			"NormalWithOrganizationAPIKey",
			"Bearer",
			okey,
			&authorizationData{
				EntityIdentifiers: org.OrganizationIdentifiers,
				Source:            auth.Key,
				Rights:            []ttnpb.Right{ttnpb.Right(0)},
			},
			true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := metadata.NewIncomingContext(
				test.Context(),
				metadata.Pairs("authorization", fmt.Sprintf("%s %s", tc.authType, tc.authValue)),
			)

			authorizationData, err := is.buildAuthorizationData(ctx)
			if tc.success {
				a.So(err, should.BeNil)
				a.So(authorizationData, should.Resemble, tc.res)
			} else {
				a.So(err, should.NotBeNil)
				a.So(authorizationData, should.BeNil)
			}
		})
	}
}
