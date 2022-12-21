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

	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestOAuthRegistry(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	usr := p.NewUser()
	usrKey, _ := p.NewAPIKey(usr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usrCreds := rpcCreds(usrKey)

	cli := p.NewClient(nil)
	cli.Rights = []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL}

	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		_, err := is.store.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
			UserIds:   usr.GetIds(),
			ClientIds: cli.GetIds(),
			Rights:    cli.GetRights(),
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		_, err = is.store.CreateAccessToken(ctx, &ttnpb.OAuthAccessToken{
			UserIds:       usr.GetIds(),
			ClientIds:     cli.GetIds(),
			UserSessionId: "12345678-1234-5678-1234-567812345678",
			Id:            "access_token_id",
			Rights:        cli.GetRights(),
			AccessToken:   "access_token",
			RefreshToken:  "refresh_token",
		}, "")
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		reg := ttnpb.NewOAuthAuthorizationRegistryClient(cc)

		authorizations, err := reg.List(ctx, &ttnpb.ListOAuthClientAuthorizationsRequest{
			UserIds: usr.GetIds(),
		}, usrCreds)
		if a.So(err, should.BeNil) && a.So(authorizations, should.NotBeNil) && a.So(authorizations.Authorizations, should.HaveLength, 1) {
			a.So(authorizations.Authorizations[0].GetClientIds().GetClientId(), should.Equal, cli.GetIds().GetClientId())
		}

		tokens, err := reg.ListTokens(ctx, &ttnpb.ListOAuthAccessTokensRequest{
			UserIds:   usr.GetIds(),
			ClientIds: cli.GetIds(),
		}, usrCreds)
		if a.So(err, should.BeNil) && a.So(tokens, should.NotBeNil) && a.So(tokens.Tokens, should.HaveLength, 1) {
			a.So(tokens.Tokens[0].Id, should.Equal, "access_token_id")
			a.So(tokens.Tokens[0].UserSessionId, should.Equal, "12345678-1234-5678-1234-567812345678")
		}

		_, err = reg.DeleteToken(ctx, &ttnpb.OAuthAccessTokenIdentifiers{
			UserIds:   usr.GetIds(),
			ClientIds: cli.GetIds(),
			Id:        "access_token_id",
		}, usrCreds)
		a.So(err, should.BeNil)

		tokens, err = reg.ListTokens(ctx, &ttnpb.ListOAuthAccessTokensRequest{
			UserIds:   usr.GetIds(),
			ClientIds: cli.GetIds(),
		}, usrCreds)
		if a.So(err, should.BeNil) && a.So(tokens, should.NotBeNil) {
			a.So(tokens.Tokens, should.BeEmpty)
		}

		_, err = reg.Delete(ctx, &ttnpb.OAuthClientAuthorizationIdentifiers{
			UserIds:   usr.GetIds(),
			ClientIds: cli.GetIds(),
		}, usrCreds)
		a.So(err, should.BeNil)

		authorizations, err = reg.List(ctx, &ttnpb.ListOAuthClientAuthorizationsRequest{
			UserIds: usr.GetIds(),
		}, usrCreds)
		if a.So(err, should.BeNil) && a.So(authorizations, should.NotBeNil) {
			a.So(authorizations.Authorizations, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}
