// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package rpcclient

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc/credentials"
)

type clientCredentials struct {
	tokenSource oauth2.TokenSource
	accessToken string
	insecure    bool
}

var (
	errFetchOAuth2Token   = errors.DefineAborted("fetch_oauth2_token", "fetch OAuth 2.0 token")
	errInvalidOAuth2Token = errors.DefineAborted("invalid_oauth2_token", "invalid OAuth 2.0 token")
)

func (c *clientCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, errFetchOAuth2Token.WithCause(err)
	}
	if !token.Valid() {
		return nil, errInvalidOAuth2Token.WithCause(err)
	}
	return map[string]string{
		"authorization": fmt.Sprintf("%s %s", token.Type(), token.AccessToken),
	}, nil
}

func (c *clientCredentials) RequireTransportSecurity() bool {
	return !c.insecure
}

// OAuth2 returns per RPC client credentials using the OAuth Client Credentials flow.
// The token is being refreshed as-needed.
func OAuth2(ctx context.Context, tokenURL, clientID, clientSecret string, scopes []string, insecure bool) credentials.PerRPCCredentials {
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		AuthStyle:    oauth2.AuthStyleInParams,
		TokenURL:     tokenURL,
	}
	return &clientCredentials{
		tokenSource: config.TokenSource(ctx),
		insecure:    insecure,
	}
}
