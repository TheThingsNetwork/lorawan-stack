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
	stderrors "errors"
	"net/http"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"golang.org/x/oauth2"
)

var (
	errRefused       = errors.DefinePermissionDenied("refused", "refused by OAuth server", "reason")
	errNoStateParam  = errors.DefinePermissionDenied("no_state_param", "no state parameter present in request")
	errNoStateCookie = errors.DefinePermissionDenied("no_state_cookie", "no state cookie stored")
	errNoCodeParam   = errors.DefinePermissionDenied("no_code", "no code parameter present in request")
	errInvalidState  = errors.DefinePermissionDenied("invalid_state", "invalid state parameter")
	errExchange      = errors.DefinePermissionDenied("exchange", "token exchange refused")
)

type oauthAuthorizeResponse struct {
	Error            string `schema:"error"`
	ErrorDescription string `schema:"error_description"`
	State            string `schema:"state"`
	Code             string `schema:"code"`
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

var errParse = errors.DefineAborted("parse", "request body parsing")

// HandleCallback is a handler that takes the auth code and exchanges it for the
// access token.
func (oc *OAuthClient) HandleCallback(w http.ResponseWriter, r *http.Request) {
	var response oauthAuthorizeResponse
	if err := r.ParseForm(); err != nil {
		webhandlers.Error(w, r, errParse.WithCause(err))
		return
	}
	if err := oc.schemaDecoder.Decode(&response, r.Form); err != nil {
		webhandlers.Error(w, r, errParse.WithCause(err))
		return
	}
	if err := response.ValidateContext(r.Context()); err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	stateCookie, err := oc.getStateCookie(w, r)
	value, acErr := oc.getAuthCookie(w, r)
	if err != nil {
		// Running the callback without state cookie often occurs when re-running
		// the callback after successful token exchange (e.g. using the browser's
		// back button after logging in). If there is a valid auth cookie, we just
		// redirect back to the client mount instead of showing an error.
		if acErr != nil {
			webhandlers.Error(w, r, errNoStateCookie.WithCause(acErr))
			return
		}
		if value.AccessToken != "" {
			config := oc.configFromContext(r.Context())
			http.Redirect(w, r, config.RootURL, http.StatusFound)
			return
		}
		webhandlers.Error(w, r, err)
		return
	}
	if stateCookie.Secret != response.State {
		webhandlers.Error(w, r, errInvalidState.New())
		return
	}

	// Exchange token.
	ctx, err := oc.withHTTPClient(r.Context())
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	conf, err := oc.oauthConfig(r.Context())
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	token, err := conf.Exchange(ctx, response.Code)
	if err != nil {
		var retrieveError *oauth2.RetrieveError
		if stderrors.As(err, &retrieveError) {
			var ttnErr errors.Error
			if decErr := ttnErr.UnmarshalJSON(retrieveError.Body); decErr == nil {
				webhandlers.Error(w, r, errExchange.WithCause(&ttnErr))
				return
			}
		}
		webhandlers.Error(w, r, errExchange.WithCause(err))
		return
	}

	oc.removeStateCookie(w, r)

	err = oc.setAuthCookie(w, r, authCookie{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	if err := oc.callback(w, r, token, stateCookie.Next); err != nil {
		if _, ok := errors.From(err); !ok {
			err = errExchange.WithCause(err)
		}
		webhandlers.Error(w, r, err)
		return
	}
}

func (oc *OAuthClient) defaultCallback(w http.ResponseWriter, r *http.Request, _ *oauth2.Token, next string) error {
	config := oc.configFromContext(r.Context())
	http.Redirect(w, r, config.RootURL+next, http.StatusFound)
	return nil
}

func (oc *OAuthClient) defaultAuthCodeURLOptions(ctx context.Context) ([]oauth2.AuthCodeOption, error) {
	return nil, nil
}
