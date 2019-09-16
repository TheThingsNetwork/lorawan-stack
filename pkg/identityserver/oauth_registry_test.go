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

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func TestOAuthRegistry(t *testing.T) {
	ctx := test.Context()
	a := assertions.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		oauthStore := store.GetOAuthStore(is.db)
		user, creds := defaultUser, userCreds(defaultUserIdx)
		client := population.Clients[0]

		_, err := oauthStore.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
			UserIDs:   user.UserIdentifiers,
			ClientIDs: client.ClientIdentifiers,
			Rights:    client.Rights,
		})
		if err != nil {
			panic(err)
		}

		err = oauthStore.CreateAccessToken(ctx, &ttnpb.OAuthAccessToken{
			UserIDs:      user.UserIdentifiers,
			ClientIDs:    client.ClientIdentifiers,
			ID:           "access_token_id",
			Rights:       client.Rights,
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}, "")
		if err != nil {
			panic(err)
		}

		reg := ttnpb.NewOAuthAuthorizationRegistryClient(cc)

		authorizations, err := reg.List(ctx, &ttnpb.ListOAuthClientAuthorizationsRequest{
			UserIdentifiers: user.UserIdentifiers,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(authorizations, should.NotBeNil) && a.So(authorizations.Authorizations, should.HaveLength, 1) {
			a.So(authorizations.Authorizations[0].ClientIDs.ClientID, should.Equal, client.ClientID)
		}

		tokens, err := reg.ListTokens(ctx, &ttnpb.ListOAuthAccessTokensRequest{
			UserIDs:   user.UserIdentifiers,
			ClientIDs: client.ClientIdentifiers,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(tokens, should.NotBeNil) && a.So(tokens.Tokens, should.HaveLength, 1) {
			a.So(tokens.Tokens[0].ID, should.Equal, "access_token_id")
		}

		_, err = reg.DeleteToken(ctx, &ttnpb.OAuthAccessTokenIdentifiers{
			UserIDs:   user.UserIdentifiers,
			ClientIDs: client.ClientIdentifiers,
			ID:        "access_token_id",
		}, creds)

		a.So(err, should.BeNil)

		tokens, err = reg.ListTokens(ctx, &ttnpb.ListOAuthAccessTokensRequest{
			UserIDs:   user.UserIdentifiers,
			ClientIDs: client.ClientIdentifiers,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(tokens, should.NotBeNil) {
			a.So(tokens.Tokens, should.BeEmpty)
		}

		_, err = reg.Delete(ctx, &ttnpb.OAuthClientAuthorizationIdentifiers{
			UserIDs:   user.UserIdentifiers,
			ClientIDs: client.ClientIdentifiers,
		}, creds)

		a.So(err, should.BeNil)

		authorizations, err = reg.List(ctx, &ttnpb.ListOAuthClientAuthorizationsRequest{
			UserIdentifiers: user.UserIdentifiers,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(authorizations, should.NotBeNil) {
			a.So(authorizations.Authorizations, should.BeEmpty)
		}
	})
}
