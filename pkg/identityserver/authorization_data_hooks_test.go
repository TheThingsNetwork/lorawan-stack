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
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/auth/oauth"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc/metadata"
)

func TestBuildauthorizationData(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	alice := testUsers()["alice"]
	client := testClient()

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
		CreatedAt:   time.Now().UTC().Truncate(time.Duration(time.Hour * 999)),
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
	for i, tc := range []struct {
		authType  string
		authValue string
		res       *authorizationData
		success   bool
	}{
		{
			// Returns empty authorization data as no authorization credentials were found.
			"",
			"",
			new(authorizationData),
			true,
		},
		{
			// Returns error as only valid AuthType is `Bearer`.
			"Key",
			"",
			nil,
			false,
		},
		{
			// Returns error as the token can not be decoded using `auth.DecodeTokenOrKey(string)`.
			"Bearer",
			"fake",
			nil,
			false,
		},
		{
			// Returns error because the token does not exist in the database.
			"Bearer",
			token2,
			nil,
			false,
		},
		{
			// Returns error because the token is expired.
			"Bearer",
			token3,
			nil,
			false,
		},
		{
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
			"Bearer",
			ukey2,
			nil,
			false,
		},
		{
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
			"Bearer",
			akey2,
			nil,
			false,
		},
		{
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
			"Bearer",
			gkey2,
			nil,
			false,
		},
		{
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
			"Bearer",
			okey2,
			nil,
			false,
		},
		{
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
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			ctx := metadata.NewIncomingContext(
				context.Background(),
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
