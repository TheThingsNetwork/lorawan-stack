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
	"net/http"
	"net/url"
	"strings"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

func stripCommonRoot(URL string, rootURL string) string {
	trimmedURL := strings.TrimSuffix(rootURL, "/")
	if rootURL, err := url.Parse(rootURL); err == nil {
		rootURLSchemeHost := (&url.URL{Scheme: rootURL.Scheme, Host: rootURL.Host}).String()
		if strings.HasPrefix(URL, rootURLSchemeHost) {
			return strings.TrimPrefix(trimmedURL, rootURLSchemeHost)
		}
	}
	return trimmedURL
}

// HandleLogout invalidates the user's authorization, removes the auth
// cookie and provides a URL to logout of the OAuth provider as well.
func (oc *OAuthClient) HandleLogout(c echo.Context) error {
	token, err := oc.freshToken(c)
	if err != nil {
		return err
	}
	u, err := url.Parse(oc.config.LogoutURL)
	if err != nil {
		return err
	}
	logoutURL := oc.config.LogoutURL

	// If a logout URL is configured, return a decorated logout URI so the client
	// can decide to additionally logout of the OAuth server itself.
	if logoutURL != "" {
		_, tokenID, _, err := auth.SplitToken(token.AccessToken)
		if err != nil {
			return err
		}
		redirectURL := stripCommonRoot(logoutURL, oc.config.RootURL)
		query := url.Values{
			"access_token_id":          []string{tokenID},
			"post_logout_redirect_uri": []string{redirectURL},
		}
		u.RawQuery = query.Encode()
		oc.removeAuthCookie(c)
		return c.JSON(http.StatusOK, struct {
			OpLogoutURI string `json:"op_logout_uri"`
		}{
			OpLogoutURI: u.String(),
		})
	}

	// Otherwise, delete the access token in the OAuth server.
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     token.AccessToken,
		AllowInsecure: oc.component.AllowInsecureForCredentials(),
	})

	ctx := c.Request().Context()

	if cc, err := oc.component.GetPeerConn(ctx, ttnpb.ClusterRole_ACCESS, nil); err == nil {
		if res, err := ttnpb.NewEntityAccessClient(cc).AuthInfo(ctx, ttnpb.Empty, creds); err == nil {
			if tokenInfo := res.GetOAuthAccessToken(); tokenInfo != nil {
				_, err := ttnpb.NewOAuthAuthorizationRegistryClient(cc).DeleteToken(ctx, &ttnpb.OAuthAccessTokenIdentifiers{
					UserIDs:   tokenInfo.UserIDs,
					ClientIDs: tokenInfo.ClientIDs,
					ID:        tokenInfo.ID,
				}, creds)
				if err != nil {
					log.FromContext(ctx).WithError(err).Error("Could not invalidate access token")
				}
			}
		}
	}
	oc.removeAuthCookie(c)
	return c.NoContent(http.StatusNoContent)
}
