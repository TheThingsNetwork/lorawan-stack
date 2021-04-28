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
	"encoding/json"
	stderrors "errors"
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"golang.org/x/oauth2"
)

var (
	errRefused      = errors.DefinePermissionDenied("refused", "refused by OAuth server", "reason")
	errNoStateParam = errors.DefinePermissionDenied("no_state", "no state parameter present in request")
	errNoCodeParam  = errors.DefinePermissionDenied("no_code", "no code parameter present in request")
	errInvalidState = errors.DefinePermissionDenied("invalid_state", "invalid state parameter")
	errExchange     = errors.DefinePermissionDenied("exchange", "token exchange refused")
)

type oauthAuthorizeResponse struct {
	Error            string `form:"error" query:"error"`
	ErrorDescription string `form:"error_description" query:"error_description"`
	State            string `form:"state" query:"state"`
	Code             string `form:"code" query:"code"`
}

func (res *oauthAuthorizeResponse) ValidateContext(c context.Context) error {
	if res.Error != "" {
		return errRefused.WithAttributes("reason", res.ErrorDescription)
	}
	if res.State == "" {
		return errNoStateParam.New()
	}
	if res.Code == "" {
		return errNoCodeParam.New()
	}
	return nil
}

// HandleCallback is a handler that takes the auth code and exchanges it for the
// access token.
func (oc *OAuthClient) HandleCallback(c echo.Context) error {
	var response oauthAuthorizeResponse
	if err := c.Bind(&response); err != nil {
		return err
	}
	if err := response.ValidateContext(c.Request().Context()); err != nil {
		return err
	}

	stateCookie, err := oc.getStateCookie(c)
	if err != nil {
		// Running the callback without state cookie often occurs when re-running
		// the callback after successful token exchange (e.g. using the browser's
		// back button after logging in). If there is a valid auth cookie, we just
		// redirect back to the client mount instead of showing an error.
		value, err := oc.getAuthCookie(c)
		if err == nil && value.AccessToken != "" {
			config := oc.configFromContext(c.Request().Context())
			return c.Redirect(http.StatusFound, config.RootURL)
		}
		return err
	}
	if stateCookie.Secret != response.State {
		return errInvalidState.New()
	}

	// Exchange token.
	ctx, err := oc.withHTTPClient(c.Request().Context())
	if err != nil {
		return err
	}
	conf, err := oc.oauth(c)
	if err != nil {
		return err
	}
	token, err := conf.Exchange(ctx, response.Code)
	if err != nil {
		var retrieveError *oauth2.RetrieveError
		if stderrors.As(err, &retrieveError) {
			var ttnErr errors.Error
			if decErr := json.Unmarshal(retrieveError.Body, &ttnErr); decErr == nil {
				return errExchange.WithCause(ttnErr)
			}
		}
		return errExchange.WithCause(err)
	}

	oc.removeStateCookie(c)

	err = oc.setAuthCookie(c, authCookie{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})
	if err != nil {
		return err
	}

	return oc.callback(c, token, stateCookie.Next)
}

func (oc *OAuthClient) defaultCallback(c echo.Context, _ *oauth2.Token, next string) error {
	config := oc.configFromContext(c.Request().Context())
	return c.Redirect(http.StatusFound, config.RootURL+next)
}

func (oc *OAuthClient) defaultAuthCodeURLOptions(echo.Context) ([]oauth2.AuthCodeOption, error) {
	return nil, nil
}
