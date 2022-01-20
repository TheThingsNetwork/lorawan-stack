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
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestOAuthStore(t *T) {
	usr1 := st.population.NewUser()
	ses1 := st.population.NewUserSession(usr1.GetIds())
	cli1 := st.population.NewClient(nil)

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.OAuthStore
	})
	defer st.DestroyDB(t, true, "users", "accounts", "user_sessions", "clients")
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement OAuthStore")
	}

	var createdAuthorization *ttnpb.OAuthClientAuthorization

	t.Run("Authorize", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		createdAuthorization, err = s.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
			UserIds:   usr1.GetIds(),
			ClientIds: cli1.GetIds(),
			Rights:    []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
		})
		if a.So(err, should.BeNil) && a.So(createdAuthorization, should.NotBeNil) {
			a.So(createdAuthorization.UserIds, should.Resemble, usr1.GetIds())
			a.So(createdAuthorization.ClientIds, should.Resemble, cli1.GetIds())
			a.So(createdAuthorization.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL})
			a.So(*ttnpb.StdTime(createdAuthorization.CreatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(*ttnpb.StdTime(createdAuthorization.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("GetAuthorization", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetAuthorization(ctx, usr1.GetIds(), cli1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, createdAuthorization)
		}
	})

	t.Run("ListAuthorizations", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListAuthorizations(ctx, usr1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, createdAuthorization)
		}
	})

	var updatedAuthorization *ttnpb.OAuthClientAuthorization

	t.Run("Authorize_again", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		updatedAuthorization, err = s.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
			UserIds:   usr1.GetIds(),
			ClientIds: cli1.GetIds(),
			Rights:    []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
		})
		if a.So(err, should.BeNil) && a.So(updatedAuthorization, should.NotBeNil) {
			a.So(updatedAuthorization.UserIds, should.Resemble, usr1.GetIds())
			a.So(updatedAuthorization.ClientIds, should.Resemble, cli1.GetIds())
			a.So(updatedAuthorization.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL})
			a.So(*ttnpb.StdTime(updatedAuthorization.CreatedAt), should.Equal, *ttnpb.StdTime(createdAuthorization.CreatedAt))
			a.So(*ttnpb.StdTime(updatedAuthorization.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("GetAuthorization_AfterUpdate", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetAuthorization(ctx, usr1.GetIds(), cli1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, updatedAuthorization)
		}
	})

	t.Run("ListAuthorizations_AfterUpdate", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListAuthorizations(ctx, usr1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, updatedAuthorization)
		}
	})

	t.Run("DeleteAuthorization", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteAuthorization(ctx, usr1.GetIds(), cli1.GetIds())
		a.So(err, should.BeNil)
	})

	t.Run("GetAuthorization_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetAuthorization(ctx, usr1.GetIds(), cli1.GetIds())
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("ListAuthorizations_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListAuthorizations(ctx, usr1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("DeleteUserAuthorizations", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
			UserIds:   usr1.GetIds(),
			ClientIds: cli1.GetIds(),
			Rights:    []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		err = s.DeleteUserAuthorizations(ctx, usr1.GetIds())
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		_, err = s.GetAuthorization(ctx, usr1.GetIds(), cli1.GetIds())
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("DeleteClientAuthorizations", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
			UserIds:   usr1.GetIds(),
			ClientIds: cli1.GetIds(),
			Rights:    []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		err = s.DeleteClientAuthorizations(ctx, cli1.GetIds())
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		_, err = s.GetAuthorization(ctx, usr1.GetIds(), cli1.GetIds())
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	var createdAuthorizationCode *ttnpb.OAuthAuthorizationCode

	t.Run("CreateAuthorizationCode", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		createdAuthorizationCode, err = s.CreateAuthorizationCode(ctx, &ttnpb.OAuthAuthorizationCode{
			UserIds:       usr1.GetIds(),
			UserSessionId: ses1.GetSessionId(),
			ClientIds:     cli1.GetIds(),
			Rights:        []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
			Code:          "CODE",
			RedirectUri:   "https://example.com",
			State:         "state",
			ExpiresAt:     ttnpb.ProtoTimePtr(start.Add(5 * time.Minute)),
		})
		if a.So(err, should.BeNil) && a.So(createdAuthorizationCode, should.NotBeNil) {
			a.So(createdAuthorizationCode.UserIds, should.Resemble, usr1.GetIds())
			a.So(createdAuthorizationCode.UserSessionId, should.Equal, ses1.GetSessionId())
			a.So(createdAuthorizationCode.ClientIds, should.Resemble, cli1.GetIds())
			a.So(createdAuthorizationCode.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL})
			a.So(createdAuthorizationCode.Code, should.Equal, "CODE")
			a.So(createdAuthorizationCode.RedirectUri, should.Equal, "https://example.com")
			a.So(createdAuthorizationCode.State, should.Equal, "state")
			a.So(*ttnpb.StdTime(createdAuthorizationCode.ExpiresAt), should.Equal, start.Add(5*time.Minute))
			a.So(*ttnpb.StdTime(createdAuthorizationCode.CreatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("GetAuthorizationCode", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetAuthorizationCode(ctx, createdAuthorizationCode.Code)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, createdAuthorizationCode)
		}
	})

	t.Run("GetAuthorizationCode_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetAuthorizationCode(ctx, "OTHER_CODE")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.GetAuthorizationCode(ctx, "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("DeleteAuthorizationCode", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteAuthorizationCode(ctx, createdAuthorizationCode.Code)
		a.So(err, should.BeNil)
	})

	t.Run("DeleteAuthorizationCode_Other", func(t *T) {
		// FIXME: DeleteAuthorizationCode does not return NotFound error when code not found (https://github.com/TheThingsNetwork/lorawan-stack/issues/5046).
		t.Skip("DeleteAuthorizationCode does not return NotFound error when code not found")
		// a, ctx := test.New(t)
		// err := s.DeleteAuthorizationCode(ctx, "OTHER_CODE")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.DeleteAuthorizationCode(ctx, "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetAuthorizationCode_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetAuthorizationCode(ctx, createdAuthorizationCode.Code)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	var createdAccessToken *ttnpb.OAuthAccessToken

	t.Run("CreateAccessToken", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		createdAccessToken, err = s.CreateAccessToken(ctx, &ttnpb.OAuthAccessToken{
			UserIds:       usr1.GetIds(),
			UserSessionId: ses1.GetSessionId(),
			ClientIds:     cli1.GetIds(),
			Id:            "token_id",
			AccessToken:   "access_token",
			RefreshToken:  "refresh_token",
			Rights:        []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
			ExpiresAt:     ttnpb.ProtoTimePtr(start.Add(5 * time.Minute)),
		}, "")
		if a.So(err, should.BeNil) && a.So(createdAccessToken, should.NotBeNil) {
			a.So(createdAccessToken.UserIds, should.Resemble, usr1.GetIds())
			a.So(createdAccessToken.UserSessionId, should.Equal, ses1.GetSessionId())
			a.So(createdAccessToken.ClientIds, should.Resemble, cli1.GetIds())
			a.So(createdAccessToken.Id, should.Equal, "token_id")
			a.So(createdAccessToken.AccessToken, should.Equal, "access_token")
			a.So(createdAccessToken.RefreshToken, should.Equal, "refresh_token")
			a.So(createdAccessToken.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL})
			a.So(*ttnpb.StdTime(createdAccessToken.ExpiresAt), should.Equal, start.Add(5*time.Minute))
			a.So(*ttnpb.StdTime(createdAccessToken.CreatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("GetAccessToken", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetAccessToken(ctx, "token_id")
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, createdAccessToken)
		}
	})

	t.Run("GetAccessToken_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetAccessToken(ctx, "other_id")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.GetAccessToken(ctx, "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("ListAccessTokens", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListAccessTokens(ctx, usr1.GetIds(), cli1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, createdAccessToken)
		}
	})

	t.Run("DeleteAccessToken", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteAccessToken(ctx, "token_id")
		a.So(err, should.BeNil)
	})

	t.Run("DeleteAccessToken_Other", func(t *T) {
		// FIXME: DeleteAccessToken does not return NotFound error when token not found (https://github.com/TheThingsNetwork/lorawan-stack/issues/5046).
		t.Skip("DeleteAccessToken does not return NotFound error when token not found")
		// a, ctx := test.New(t)
		// err := s.DeleteAccessToken(ctx, "other_id")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.DeleteAccessToken(ctx, "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetAccessToken_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetAccessToken(ctx, "token_id")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("ListAccessTokens_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListAccessTokens(ctx, usr1.GetIds(), cli1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})
}

// TODO: Test Pagination (https://github.com/TheThingsNetwork/lorawan-stack/issues/5047).
