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
)

// HandleLogout invalidates the user's authorization, removes the auth
// cookie and provides a URL to logout of the OAuth provider as well.
func (oc *OAuthClient) HandleLogout(c echo.Context) error {
	token, err := oc.freshToken(c)
	if err != nil {
		return err
	}
	oc.removeAuthCookie(c)
	u, err := url.Parse(oc.config.LogoutURL)
	if err != nil {
		return err
	}
	logoutURL := oc.config.LogoutURL
	redirectURL := strings.TrimSuffix(oc.config.RootURL, "/")
	if oauthRootURL, err := url.Parse(oc.config.RootURL); err == nil {
		rootURL := (&url.URL{Scheme: oauthRootURL.Scheme, Host: oauthRootURL.Host}).String()
		if strings.HasPrefix(logoutURL, rootURL) {
			redirectURL = strings.TrimPrefix(redirectURL, rootURL)
		}
	}
	query := url.Values{
		"access_token":             []string{token.AccessToken},
		"post_logout_redirect_uri": []string{redirectURL},
	}
	u.RawQuery = query.Encode()
	return c.JSON(http.StatusOK, struct {
		OpLogoutURI string `json:"op_logout_uri"`
	}{
		OpLogoutURI: u.String(),
	})
}
