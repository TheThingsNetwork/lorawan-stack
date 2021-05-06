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

func TestOAuthStore(t *testing.T) {
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db,
			&ClientAuthorization{},
			&AuthorizationCode{},
			&AccessToken{},
			&User{},
			&Client{},
			&Account{},
		)

		s := newStore(db)
		store := GetOAuthStore(db)

		s.createEntity(ctx, &User{Account: Account{UID: "test-user"}})
		userIDs := &ttnpb.UserIdentifiers{UserId: "test-user"}

		s.createEntity(ctx, &Client{ClientID: "test-client"})
		clientIDs := &ttnpb.ClientIdentifiers{ClientId: "test-client"}

		rights := []ttnpb.Right{ttnpb.RIGHT_ALL}

		t.Run("Authorize", func(t *testing.T) {
			a := assertions.New(t)

			empty, err := store.GetAuthorization(ctx, userIDs, clientIDs)

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}
			a.So(empty, should.BeNil)

			start := cleanTime(time.Now())

			created, err := store.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
				ClientIds: *clientIDs,
				UserIds:   *userIDs,
				Rights:    rights,
			})

			a.So(err, should.BeNil)
			a.So(created, should.NotBeNil)
			if a.So(created, should.NotBeNil) {
				a.So(created.UserIds.UserId, should.Equal, "test-user")
				a.So(created.ClientIds.ClientId, should.Equal, "test-client")
				a.So(created.CreatedAt, should.HappenOnOrAfter, start)
				a.So(created.UpdatedAt, should.HappenOnOrAfter, start)
			}

			got, err := store.GetAuthorization(ctx, userIDs, clientIDs)

			a.So(err, should.BeNil)
			a.So(got, should.NotBeNil)
			if a.So(got, should.NotBeNil) {
				a.So(created.UserIds.UserId, should.Equal, "test-user")
				a.So(created.ClientIds.ClientId, should.Equal, "test-client")
				a.So(got.CreatedAt, should.HappenOnOrAfter, start)
				a.So(got.UpdatedAt, should.HappenOnOrAfter, start)
			}

			list, err := store.ListAuthorizations(ctx, userIDs)

			a.So(err, should.BeNil)
			if a.So(list, should.NotBeNil) && a.So(list, should.HaveLength, 1) {
				a.So(list[0], should.Resemble, got)
			}

			err = store.DeleteAuthorization(ctx, userIDs, clientIDs)

			a.So(err, should.BeNil)

			deleted, err := store.GetAuthorization(ctx, userIDs, clientIDs)

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}
			a.So(deleted, should.BeNil)
		})

		t.Run("Authorization Code", func(t *testing.T) {
			a := assertions.New(t)

			code := "test-authorization-code"
			redirectURI := "http://test-redirect-url:8080/callback"
			state := "test-state"

			authCode, err := store.GetAuthorizationCode(ctx, "")

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}
			a.So(authCode, should.BeNil)

			empty, err := store.GetAuthorizationCode(ctx, code)

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}
			a.So(empty, should.BeNil)

			err = store.DeleteAuthorizationCode(ctx, "")

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}

			start := cleanTime(time.Now())

			err = store.CreateAuthorizationCode(ctx, &ttnpb.OAuthAuthorizationCode{
				ClientIds:   *clientIDs,
				UserIds:     *userIDs,
				Rights:      rights,
				Code:        code,
				RedirectURI: redirectURI,
				State:       state,
			})

			a.So(err, should.BeNil)

			got, err := store.GetAuthorizationCode(ctx, code)

			a.So(err, should.BeNil)
			if a.So(got, should.NotBeNil) {
				a.So(got.UserIds.UserId, should.Equal, userIDs.UserId)
				a.So(got.ClientIds.ClientId, should.Equal, clientIDs.ClientId)
				a.So(got.Code, should.Equal, code)
				a.So(got.RedirectURI, should.Equal, redirectURI)
				a.So(got.State, should.Equal, state)
				a.So(got.CreatedAt, should.HappenOnOrAfter, start)
				if a.So(got.Rights, should.HaveLength, len(rights)) {
					for _, right := range rights {
						a.So(got.Rights, should.Contain, right)
					}
				}
			}

			err = store.DeleteAuthorizationCode(ctx, code)

			a.So(err, should.BeNil)

			deleted, err := store.GetAuthorizationCode(ctx, code)

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}
			a.So(deleted, should.BeNil)
		})

		t.Run("Access Token", func(t *testing.T) {
			a := assertions.New(t)

			tokenID := "test-token-id"
			access := "test-access-token"
			refresh := "test-refresh-token"
			prevID := ""

			accessToken, err := store.GetAccessToken(ctx, "")

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}
			a.So(accessToken, should.BeNil)

			err = store.DeleteAccessToken(ctx, "")

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}

			empty, err := store.GetAccessToken(ctx, tokenID)

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}
			a.So(empty, should.BeNil)

			start := cleanTime(time.Now())

			err = store.CreateAccessToken(ctx, &ttnpb.OAuthAccessToken{
				UserIds:      *userIDs,
				ClientIds:    *clientIDs,
				ID:           tokenID,
				AccessToken:  access,
				RefreshToken: refresh,
				Rights:       rights,
			}, prevID)

			a.So(err, should.BeNil)

			got, err := store.GetAccessToken(ctx, tokenID)

			a.So(err, should.BeNil)
			if a.So(got, should.NotBeNil) {
				a.So(got.UserIds.UserId, should.Equal, userIDs.UserId)
				a.So(got.ClientIds.ClientId, should.Equal, clientIDs.ClientId)
				a.So(got.ID, should.Equal, tokenID)
				a.So(got.AccessToken, should.Equal, access)
				a.So(got.RefreshToken, should.Equal, refresh)
				a.So(got.CreatedAt, should.HappenOnOrAfter, start)
				if a.So(got.Rights, should.HaveLength, len(rights)) {
					for _, right := range rights {
						a.So(got.Rights, should.Contain, right)
					}
				}
			}

			list, err := store.ListAccessTokens(ctx, userIDs, clientIDs)

			a.So(err, should.BeNil)
			if a.So(list, should.NotBeNil) && a.So(list, should.HaveLength, 1) {
				a.So(list[0], should.Resemble, got)
			}

			err = store.DeleteAccessToken(ctx, tokenID)

			a.So(err, should.BeNil)

			deleted, err := store.GetAccessToken(ctx, tokenID)

			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}
			a.So(deleted, should.BeNil)
		})

		t.Run("Delete user authorizations", func(t *testing.T) {
			a := assertions.New(t)

			code := "test-authorization-code"
			redirectURI := "http://test-redirect-url:8080/callback"
			state := "test-state"

			tokenID := "test-token-id"
			access := "test-access-token"
			refresh := "test-refresh-token"
			prevID := ""

			store.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
				ClientIds: *clientIDs,
				UserIds:   *userIDs,
				Rights:    rights,
			})
			authorizationList, _ := store.ListAuthorizations(ctx, userIDs)

			a.So(authorizationList, should.HaveLength, 1)

			store.CreateAuthorizationCode(ctx, &ttnpb.OAuthAuthorizationCode{
				ClientIds:   *clientIDs,
				UserIds:     *userIDs,
				Rights:      rights,
				Code:        code,
				RedirectURI: redirectURI,
				State:       state,
			})

			_, err := store.GetAuthorizationCode(ctx, code)

			store.CreateAccessToken(ctx, &ttnpb.OAuthAccessToken{
				UserIds:      *userIDs,
				ClientIds:    *clientIDs,
				ID:           tokenID,
				AccessToken:  access,
				RefreshToken: refresh,
				Rights:       rights,
			}, prevID)

			tokenList, _ := store.ListAccessTokens(ctx, userIDs, clientIDs)
			a.So(tokenList, should.HaveLength, 1)

			err = store.DeleteUserAuthorizations(ctx, userIDs)

			a.So(err, should.BeNil)

			authorizationList, _ = store.ListAuthorizations(ctx, userIDs)

			a.So(authorizationList, should.HaveLength, 0)

			_, err = store.GetAuthorizationCode(ctx, code)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsNotFound(err), should.BeTrue)
			}

			tokenList, _ = store.ListAccessTokens(ctx, userIDs, clientIDs)
			a.So(tokenList, should.HaveLength, 0)
		})
	})
}
