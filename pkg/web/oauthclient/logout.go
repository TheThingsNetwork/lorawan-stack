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

package oauthclient

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

func stripCommonRoot(targetURL string, rootURL string) string {
	trimmedURL := strings.TrimSuffix(rootURL, "/")
	if rootURL, err := url.Parse(rootURL); err == nil {
		rootURLSchemeHost := (&url.URL{Scheme: rootURL.Scheme, Host: rootURL.Host}).String()
		if strings.HasPrefix(targetURL, rootURLSchemeHost) {
			return strings.TrimPrefix(trimmedURL, rootURLSchemeHost)
		}
	}
	return trimmedURL
}

// Logout initiates the logout.
// Based on configuration, this method either returns a logout URL to redirect the user to complete the logout,
// or this method deletes the access token from the OAuth server.
func (oc *OAuthClient) Logout(ctx context.Context, token *oauth2.Token) (logoutURL string, err error) {
	config := oc.configFromContext(ctx)

	// If a logout URL is configured, return a decorated logout URI so the client
	// can decide to additionally logout of the OAuth server itself.
	if config.LogoutURL != "" {
		logoutURL, err := url.Parse(config.LogoutURL)
		if err != nil {
			return "", err
		}
		_, tokenID, _, err := auth.SplitToken(token.AccessToken)
		if err != nil {
			return "", err
		}
		redirectURL := stripCommonRoot(config.LogoutURL, config.RootURL)
		query := url.Values{
			"access_token_id":          []string{tokenID},
			"post_logout_redirect_uri": []string{redirectURL},
		}
		logoutURL.RawQuery = query.Encode()
		return logoutURL.String(), nil
	}

	// Otherwise, delete the access token in the OAuth server.
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     token.AccessToken,
		AllowInsecure: oc.component.AllowInsecureForCredentials(),
	})
	if cc, err := oc.component.GetPeerConn(ctx, ttnpb.ClusterRole_ACCESS, nil); err == nil {
		if res, err := ttnpb.NewEntityAccessClient(cc).AuthInfo(ctx, ttnpb.Empty, creds); err == nil {
			if tokenInfo := res.GetOauthAccessToken(); tokenInfo != nil {
				_, err := ttnpb.NewOAuthAuthorizationRegistryClient(cc).DeleteToken(ctx, &ttnpb.OAuthAccessTokenIdentifiers{
					UserIds:   tokenInfo.UserIds,
					ClientIds: tokenInfo.ClientIds,
					Id:        tokenInfo.Id,
				}, creds)
				if err != nil {
					log.FromContext(ctx).WithError(err).Error("Could not invalidate access token")
					return "", nil
				}
			}
		}
	}
	return "", nil
}

// HandleLogout invalidates the user's authorization, removes the auth
// cookie and may provide a URL to logout of the OAuth provider as well.
func (oc *OAuthClient) HandleLogout(w http.ResponseWriter, r *http.Request) {
	value, err := oc.getAuthCookie(w, r)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	oc.removeAuthCookie(w, r)

	token, err := oc.Token(r.Context(), &oauth2.Token{
		AccessToken:  value.AccessToken,
		RefreshToken: value.RefreshToken,
		Expiry:       time.Now(),
	})
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	logoutURL, err := oc.Logout(r.Context(), token)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	if logoutURL != "" {
		webhandlers.JSON(w, r, struct {
			OpLogoutURI string `json:"op_logout_uri"`
		}{
			OpLogoutURI: logoutURL,
		})
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
