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
	"net"
	"net/url"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Scope defines a scope of claims to request in the token.
type Scope string

const (
	ScopeNetworks Scope = "networks"
)

const (
	defaultTokenURL        = "https://iam.packetbroker.net/token"
	defaultTokenPublicKeys = "https://iam.packetbroker.net/.well-known/jwks.json"
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
		tokenURL: defaultTokenURL,
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

var errOAuth2Token = errors.DefineUnauthenticated("oauth2_token", "invalid OAuth 2.0 token for network authentication")

// UnverifiedNetworkIdentifier returns the Packet Broker network identifier from the given token.
// This function does not verify the token.
func UnverifiedNetworkIdentifier(token *oauth2.Token) (ttnpb.PacketBrokerNetworkIdentifier, error) {
	parsed, err := jwt.ParseSigned(token.AccessToken)
	if err != nil {
		return ttnpb.PacketBrokerNetworkIdentifier{}, errOAuth2Token.WithCause(err)
	}
	var claims struct {
		PacketBroker struct {
			Networks []struct {
				NetID    uint32 `json:"nid"`
				TenantID string `json:"tid"`
			} `json:"ns"`
		} `json:"https://iam.packetbroker.net/claims"`
	}
	if err := parsed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return ttnpb.PacketBrokerNetworkIdentifier{}, errOAuth2Token.WithCause(err)
	}
	if len(claims.PacketBroker.Networks) == 0 {
		return ttnpb.PacketBrokerNetworkIdentifier{}, errOAuth2Token.New()
	}
	return ttnpb.PacketBrokerNetworkIdentifier{
		NetId:    claims.PacketBroker.Networks[0].NetID,
		TenantId: claims.PacketBroker.Networks[0].TenantID,
	}, nil
}
