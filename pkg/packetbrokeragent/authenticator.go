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

package packetbrokeragent

import (
	"context"
	"crypto/tls"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/square/go-jose.v2/jwt"
)

type tlsConfigurator interface {
	GetTLSClientConfig(context.Context, ...component.TLSConfigOption) (*tls.Config, error)
}

type authenticator interface {
	AuthInfo(context.Context) (ttnpb.PacketBrokerNetworkIdentifier, error)
	DialOptions(context.Context) ([]grpc.DialOption, error)
}

type oauth2Authenticator struct {
	tokenSource oauth2.TokenSource
	tlsConfig   tlsConfigurator
}

func newOAuth2(ctx context.Context, oauth2Config OAuth2Config, tlsConfig tlsConfigurator) authenticator {
	config := clientcredentials.Config{
		ClientID:     oauth2Config.ClientID,
		ClientSecret: oauth2Config.ClientSecret,
		Scopes:       []string{"networks"},
		AuthStyle:    oauth2.AuthStyleInParams,
		TokenURL:     oauth2Config.TokenURL,
	}
	return &oauth2Authenticator{
		tokenSource: config.TokenSource(ctx),
		tlsConfig:   tlsConfig,
	}
}

var errOAuth2Token = errors.DefineUnauthenticated("oauth2_token", "invalid OAuth 2.0 token for network authentication")

func (a *oauth2Authenticator) AuthInfo(ctx context.Context) (ttnpb.PacketBrokerNetworkIdentifier, error) {
	token, err := a.tokenSource.Token()
	if err != nil {
		return ttnpb.PacketBrokerNetworkIdentifier{}, err
	}
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

func (a *oauth2Authenticator) DialOptions(ctx context.Context) (res []grpc.DialOption, err error) {
	var tlsConfig *tls.Config
	if a.tlsConfig != nil {
		tlsConfig, err = a.tlsConfig.GetTLSClientConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	res = make([]grpc.DialOption, 2)
	res[0] = grpc.WithPerRPCCredentials(rpcclient.OAuth2(a.tokenSource, tlsConfig == nil))
	if tlsConfig == nil {
		res[1] = grpc.WithInsecure()
	} else {
		res[1] = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}
	return
}
