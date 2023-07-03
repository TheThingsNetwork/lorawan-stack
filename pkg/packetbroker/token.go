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

package packetbroker

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Scope defines a scope of claims to request in the token.
type Scope string

const (
	ScopeNetworks Scope = "networks"
)

type tokenOptions struct {
	tokenURL string
	scopes   []Scope
	audience []string
}

// TokenOption customizes fetching a Packet Broker token.
type TokenOption func(o *tokenOptions)

// WithTokenURL customizes the token URL.
func WithTokenURL(tokenURL string) TokenOption {
	return func(o *tokenOptions) {
		o.tokenURL = tokenURL
	}
}

// WithScope customizes the scope.
func WithScope(scopes ...Scope) TokenOption {
	return func(o *tokenOptions) {
		o.scopes = append(o.scopes, scopes...)
	}
}

// WithAudienceFromAddresses provides the service addresses for which the token will be valid.
// The host parts of the addresses are used as the token audience.
func WithAudienceFromAddresses(addresses ...string) TokenOption {
	return func(o *tokenOptions) {
		hosts := make(map[string]bool, len(addresses))
		for _, addr := range addresses {
			if addr == "" {
				continue
			}
			// If the address is a URL with a scheme, check if it's a gRPC dialer scheme and remove it.
			// gRPC dialer schemes in the target address look like passthrough:///host:port.
			if u, err := url.Parse(addr); err == nil && u.Scheme != "" && strings.HasPrefix(addr, u.Scheme+":///") {
				addr = addr[len(u.Scheme)+4:]
			}
			if h, _, err := net.SplitHostPort(addr); err == nil {
				addr = h
			}
			hosts[addr] = true
		}
		for h := range hosts {
			o.audience = append(o.audience, h)
		}
	}
}

// TokenSource returns a new OAuth 2.0 token source using Packet Broker credentials.
func TokenSource(ctx context.Context, clientID, clientSecret string, opts ...TokenOption) oauth2.TokenSource {
	o := tokenOptions{
		tokenURL: DefaultTokenURL,
	}
	for _, opt := range opts {
		opt(&o)
	}
	scopes := make([]string, len(o.scopes))
	for i, s := range o.scopes {
		scopes[i] = string(s)
	}
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		AuthStyle:    oauth2.AuthStyleInParams,
		TokenURL:     o.tokenURL,
	}
	if len(o.audience) > 0 {
		config.EndpointParams = url.Values{
			"audience": []string{strings.Join(o.audience, " ")},
		}
	}
	return config.TokenSource(ctx)
}

// TokenNetworkClaim defines a Packet Broker network identifier.
type TokenNetworkClaim struct {
	NetID    uint32 `json:"nid"`
	TenantID string `json:"tid"`
}

// IAMTokenClaims defines the claims from Packet Broker IAM.
type IAMTokenClaims struct {
	Cluster  bool                `json:"c,omitempty"`
	Networks []TokenNetworkClaim `json:"ns,omitempty"`
	Rights   []int32             `json:"rights,omitempty"`
}

// TokenClaims defines the Packet Broker JSON Web Token (JWT) claims.
type TokenClaims struct {
	jwt.Claims
	PacketBroker IAMTokenClaims `json:"https://iam.packetbroker.net/claims,omitempty"`
}

var (
	errOAuth2Token   = errors.DefineUnauthenticated("oauth2_token", "invalid OAuth 2.0 token")
	errNotAuthorized = errors.DefinePermissionDenied("not_authorized", "not authorized")
)

// UnverifiedNetworkIdentifier returns the Packet Broker network identifier from the given token.
// This function does not verify the token.
func UnverifiedNetworkIdentifier(token string) (*ttnpb.PacketBrokerNetworkIdentifier, error) {
	parsed, err := jwt.ParseSigned(token)
	if err != nil {
		return nil, errOAuth2Token.WithCause(err)
	}
	var claims TokenClaims
	if err := parsed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, errOAuth2Token.WithCause(err)
	}
	if len(claims.PacketBroker.Networks) == 0 {
		return nil, errOAuth2Token.New()
	}
	return &ttnpb.PacketBrokerNetworkIdentifier{
		NetId:    claims.PacketBroker.Networks[0].NetID,
		TenantId: claims.PacketBroker.Networks[0].TenantID,
	}, nil
}

// PublicKeyProvider provides a set of public keys.
type PublicKeyProvider interface {
	PublicKeys(context.Context) (*jose.JSONWebKeySet, error)
}

// PublicKeyProviderFunc is a function that implements PublicKeyProvider.
type PublicKeyProviderFunc func(context.Context) (*jose.JSONWebKeySet, error)

// PublicKeys implements PublicKeyProvider.
func (f PublicKeyProviderFunc) PublicKeys(ctx context.Context) (*jose.JSONWebKeySet, error) {
	return f(ctx)
}

// ParseAndVerify parses and verifies the token and returns the claims.
// See Verify for the verification process.
func ParseAndVerify(ctx context.Context, token *oauth2.Token, keyProvider PublicKeyProvider, issuer, audience string) (TokenClaims, error) {
	t, err := jwt.ParseSigned(token.AccessToken)
	if err != nil {
		return TokenClaims{}, errOAuth2Token.WithCause(err)
	}
	return Verify(ctx, t, keyProvider, issuer, audience)
}

// Verify verifies the token and returns the claims.
// If issuer is non-empty, the token's issuer must match the issuer.
// If audience is non-empty, one of the token's audiences must match the audience.
// The current system timestamp is used as reference to verify not before, issued at and expiry.
func Verify(ctx context.Context, token *jwt.JSONWebToken, keyProvider PublicKeyProvider, issuer, audience string) (TokenClaims, error) {
	keys, err := keyProvider.PublicKeys(ctx)
	if err != nil {
		return TokenClaims{}, err
	}
	var claims TokenClaims
	if err := token.Claims(keys, &claims); err != nil {
		return TokenClaims{}, errOAuth2Token.WithCause(err)
	}

	exp := jwt.Expected{
		Issuer: issuer,
		Time:   time.Now(),
	}
	if audience != "" {
		exp.Audience = jwt.Audience{audience}
	}
	if err := claims.Validate(exp); err != nil {
		return TokenClaims{}, errNotAuthorized.WithCause(err)
	}
	return claims, nil
}

// CachePublicKey caches the result from the given PublicKeyProvider with the TTL.
func CachePublicKey(provider PublicKeyProvider, ttl time.Duration) PublicKeyProvider {
	var (
		key *jose.JSONWebKeySet
		err error
		t   time.Time
		mu  sync.Mutex
	)
	return PublicKeyProviderFunc(func(ctx context.Context) (*jose.JSONWebKeySet, error) {
		mu.Lock()
		defer mu.Unlock()
		if key == nil && err == nil || time.Since(t) > ttl {
			key, err = provider.PublicKeys(ctx)
			t = time.Now()
		}
		return key, err
	})
}

var errFetchToken = errors.DefineAborted("fetch_token", "fetch token")

// PublicKeyFromURL loads the public keys from the given URL.
func PublicKeyFromURL(client *http.Client, url string) PublicKeyProvider {
	return PublicKeyProviderFunc(func(ctx context.Context) (*jose.JSONWebKeySet, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, errFetchToken.WithCause(err)
		}
		res, err := client.Do(req)
		if err != nil {
			return nil, errFetchToken.WithCause(err)
		}
		defer res.Body.Close()
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errFetchToken.WithCause(err)
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return nil, errFetchToken.New()
		}
		key := new(jose.JSONWebKeySet)
		if err := json.Unmarshal(buf, key); err != nil {
			return nil, errFetchToken.WithCause(err)
		}
		return key, nil
	})
}
