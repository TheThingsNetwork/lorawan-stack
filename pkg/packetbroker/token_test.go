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

package packetbroker_test

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbroker"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"golang.org/x/oauth2"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func TestToken(t *testing.T) {
	for _, tc := range []struct {
		name string
		clientID,
		clientSecret string
		opts                       []packetbroker.TokenOption
		tokenRequestAssertion      func(a *assertions.Assertion, vars url.Values) bool
		tokenRequestErrorAssertion func(a *assertions.Assertion, err error) bool
		tokenClaims                func() packetbroker.IAMTokenClaims
		audience                   string
		tokenAssertion             func(a *assertions.Assertion, token string) bool
		tokenClaimsAssertion       func(a *assertions.Assertion, claims packetbroker.TokenClaims) bool
		tokenClaimsErrorAssertion  func(a *assertions.Assertion, err error) bool
	}{
		{
			name:         "Success",
			clientID:     "test",
			clientSecret: "secret",
			opts: []packetbroker.TokenOption{
				packetbroker.WithScope(packetbroker.ScopeNetworks),
				packetbroker.WithAudienceFromAddresses("iam.packetbroker.net:443"),
			},
			tokenRequestAssertion: func(a *assertions.Assertion, vars url.Values) bool {
				return a.So(vars["scope"], should.Resemble, []string{"networks"}) &&
					a.So(vars["audience"], should.Resemble, []string{"iam.packetbroker.net"})
			},
			tokenClaims: func() packetbroker.IAMTokenClaims {
				return packetbroker.IAMTokenClaims{
					Networks: []packetbroker.TokenNetworkClaim{
						{
							NetID:    0x000013,
							TenantID: "ttn",
						},
					},
				}
			},
			audience: "iam.packetbroker.net",
			tokenAssertion: func(a *assertions.Assertion, token string) bool {
				id, err := packetbroker.UnverifiedNetworkIdentifier(token)
				return a.So(err, should.BeNil) &&
					a.So(id.NetId, should.Equal, 0x000013) &&
					a.So(id.TenantId, should.Equal, "ttn")
			},
			tokenClaimsAssertion: func(a *assertions.Assertion, claims packetbroker.TokenClaims) bool {
				return a.So(claims.PacketBroker.Cluster, should.BeFalse)
			},
		},
		{
			name:         "SuccessWithDialerScheme",
			clientID:     "test",
			clientSecret: "secret",
			opts: []packetbroker.TokenOption{
				packetbroker.WithScope(packetbroker.ScopeNetworks),
				packetbroker.WithAudienceFromAddresses("passthrough:///iam.packetbroker.net:443"),
			},
			tokenRequestAssertion: func(a *assertions.Assertion, vars url.Values) bool {
				return a.So(vars["scope"], should.Resemble, []string{"networks"}) &&
					a.So(vars["audience"], should.Resemble, []string{"iam.packetbroker.net"})
			},
			tokenClaims: func() packetbroker.IAMTokenClaims {
				return packetbroker.IAMTokenClaims{
					Networks: []packetbroker.TokenNetworkClaim{
						{
							NetID:    0x000013,
							TenantID: "ttn",
						},
					},
				}
			},
			audience: "iam.packetbroker.net",
			tokenAssertion: func(a *assertions.Assertion, token string) bool {
				id, err := packetbroker.UnverifiedNetworkIdentifier(token)
				return a.So(err, should.BeNil) &&
					a.So(id.NetId, should.Equal, 0x000013) &&
					a.So(id.TenantId, should.Equal, "ttn")
			},
			tokenClaimsAssertion: func(a *assertions.Assertion, claims packetbroker.TokenClaims) bool {
				return a.So(claims.PacketBroker.Cluster, should.BeFalse)
			},
		},
		{
			name:         "BadRequest",
			clientID:     "test",
			clientSecret: "secret",
			opts: []packetbroker.TokenOption{
				packetbroker.WithScope(packetbroker.ScopeNetworks),
				packetbroker.WithAudienceFromAddresses("iam.packetbroker.net:443"),
			},
			tokenRequestAssertion: func(a *assertions.Assertion, vars url.Values) bool {
				if !a.So(vars["scope"], should.Resemble, []string{"networks"}) ||
					!a.So(vars["audience"], should.Resemble, []string{"iam.packetbroker.net"}) {
					return false
				}
				return false // The request is invalid
			},
			tokenRequestErrorAssertion: func(a *assertions.Assertion, err error) bool {
				var retrieveErr *oauth2.RetrieveError
				if !errors.As(err, &retrieveErr) {
					return false
				}
				return a.So(retrieveErr.Response.StatusCode, should.Equal, http.StatusBadRequest)
			},
		},
		{
			name:         "InvalidAudience",
			clientID:     "test",
			clientSecret: "secret",
			opts: []packetbroker.TokenOption{
				packetbroker.WithScope(packetbroker.ScopeNetworks),
				packetbroker.WithAudienceFromAddresses("iam.packetbroker.net:443"),
			},
			tokenRequestAssertion: func(a *assertions.Assertion, vars url.Values) bool {
				return a.So(vars["scope"], should.Resemble, []string{"networks"}) &&
					a.So(vars["audience"], should.Resemble, []string{"iam.packetbroker.net"})
			},
			tokenClaims: func() packetbroker.IAMTokenClaims {
				return packetbroker.IAMTokenClaims{
					Networks: []packetbroker.TokenNetworkClaim{
						{
							NetID:    0x000013,
							TenantID: "ttn",
						},
					},
				}
			},
			audience: "cp.packetbroker.net", // The audience is wrong, verification will fail
			tokenAssertion: func(a *assertions.Assertion, token string) bool {
				id, err := packetbroker.UnverifiedNetworkIdentifier(token)
				return a.So(err, should.BeNil) &&
					a.So(id.NetId, should.Equal, 0x000013) &&
					a.So(id.TenantId, should.Equal, "ttn")
			},
			tokenClaimsErrorAssertion: func(a *assertions.Assertion, err error) bool {
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				public, private, err := ed25519.GenerateKey(nil)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				const issuer = "https://thethings.example.com"
				var (
					publicJWK = jose.JSONWebKey{
						Key:       public,
						KeyID:     "test",
						Algorithm: "EdDSA",
					}
					privateJWK = jose.JSONWebKey{
						Key:       private,
						KeyID:     "test",
						Algorithm: "EdDSA",
					}
					tokenRequests,
					publicKeyRequests uint32
				)

				router := mux.NewRouter()
				router.Handle("/token", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					atomic.AddUint32(&tokenRequests, 1)
					if err := r.ParseForm(); err != nil {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					clientID, clientSecret, ok := r.BasicAuth()
					if !ok {
						clientID, clientSecret = r.PostFormValue("client_id"), r.PostFormValue("client_secret")
					}
					if clientID != tc.clientID || clientSecret != tc.clientSecret {
						t.Log("Wrong credentials")
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if !tc.tokenRequestAssertion(a, r.PostForm) {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					signer, err := jose.NewSigner(jose.SigningKey{
						Algorithm: jose.SignatureAlgorithm(privateJWK.Algorithm),
						Key:       privateJWK,
					}, new(jose.SignerOptions).WithType("JWT"))
					if err != nil {
						t.Errorf("Instantiate signer: %s", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					claims := packetbroker.TokenClaims{
						Claims: jwt.Claims{
							ID:       "test",
							Subject:  "test",
							IssuedAt: jwt.NewNumericDate(time.Now()),
							Expiry:   jwt.NewNumericDate(time.Now().Add(time.Hour)),
							Issuer:   issuer,
						},
						PacketBroker: tc.tokenClaims(),
					}
					if aud := r.PostFormValue("audience"); aud != "" {
						claims.Audience = jwt.Audience(strings.Split(aud, " "))
					}
					accessToken, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
					if err != nil {
						t.Errorf("Serialize token: %s", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					token := &oauth2.Token{
						AccessToken: accessToken,
						TokenType:   "bearer",
						Expiry:      claims.Expiry.Time(),
					}
					json.NewEncoder(w).Encode(token)
				})).Methods(http.MethodPost)
				router.Handle("/.well-known/jwks.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					atomic.AddUint32(&publicKeyRequests, 1)
					jwks := jose.JSONWebKeySet{
						Keys: []jose.JSONWebKey{publicJWK},
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(jwks)
				})).Methods(http.MethodGet)

				srv := httptest.NewServer(router)
				defer srv.Close()

				tokenSource := packetbroker.TokenSource(ctx, tc.clientID, tc.clientSecret,
					append(tc.opts, packetbroker.WithTokenURL(fmt.Sprintf("%s/token", srv.URL)))...,
				)

				keyProvider := packetbroker.CachePublicKey(
					packetbroker.PublicKeyFromURL(srv.Client(), fmt.Sprintf("%s/.well-known/jwks.json", srv.URL)),
					time.Hour,
				)

				// Repeat a couple of times to test token and public key cache.
				for i := 0; i < 10; i++ {
					token, err := tokenSource.Token()
					if err != nil {
						if tc.tokenRequestErrorAssertion == nil {
							t.Fatalf("Unexpected error: %s", err)
						}
						if !tc.tokenRequestErrorAssertion(a, err) {
							t.FailNow()
						}
						return
					} else if !a.So(err, should.BeNil) || tc.tokenRequestErrorAssertion != nil {
						t.FailNow()
					}

					if !tc.tokenAssertion(a, token.AccessToken) {
						t.FailNow()
					}

					claims, err := packetbroker.ParseAndVerify(ctx, token, keyProvider, issuer, tc.audience)
					if err != nil {
						if tc.tokenClaimsErrorAssertion == nil {
							t.Fatalf("Unexpected error: %s", err)
						}
						if !tc.tokenClaimsErrorAssertion(a, err) {
							t.FailNow()
						}
						return
					} else if !a.So(err, should.BeNil) || tc.tokenClaimsErrorAssertion != nil {
						t.FailNow()
					}
					if !tc.tokenClaimsAssertion(a, claims) {
						t.FailNow()
					}
				}

				// The token is cached by the reusable token source of the standard library.
				a.So(tokenRequests, should.Equal, 1)
				// The public key is cached by using CachePublicKey().
				a.So(publicKeyRequests, should.Equal, 1)
			},
		})
	}
}
