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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"golang.org/x/oauth2"
)

var errRefresh = errors.DefinePermissionDenied("refresh", "token refresh refused")

// Token returns the OAuth 2.0 token.
// If the given token is about to expire, this method refreshes the token and returns the new token.
func (oc *OAuthClient) Token(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	conf, err := oc.oauthConfig(ctx)
	if err != nil {
		return nil, err
	}
	ctx, err = oc.withHTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	freshToken, err := conf.TokenSource(ctx, token).Token()
	if err != nil {
		var retrieveError *oauth2.RetrieveError
		if stderrors.As(err, &retrieveError) {
			var ttnErr errors.Error
			if decErr := ttnErr.UnmarshalJSON(retrieveError.Body); decErr == nil {
				return nil, errRefresh.WithCause(&ttnErr)
			}
		}
		return nil, errRefresh.WithCause(err)
	}
	return freshToken, nil
}

// HandleToken is a handler that returns a valid OAuth token.
// It reads the token from the authorization cookie and refreshes it if needed.
// If the authorization cookie is not there, it returns a 401 Unauthorized error.
func (oc *OAuthClient) HandleToken(w http.ResponseWriter, r *http.Request) {
	value, err := oc.getAuthCookie(w, r)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	currentToken := &oauth2.Token{
		AccessToken:  value.AccessToken,
		RefreshToken: value.RefreshToken,
		Expiry:       time.Now(),
	}

	freshToken, err := oc.Token(r.Context(), currentToken)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	if freshToken != currentToken {
		err = oc.setAuthCookie(w, r, authCookie{
			AccessToken:  freshToken.AccessToken,
			RefreshToken: freshToken.RefreshToken,
			Expiry:       freshToken.Expiry,
		})
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
	}

	webhandlers.JSON(w, r, struct {
		AccessToken string    `json:"access_token"`
		Expiry      time.Time `json:"expiry"`
	}{
		AccessToken: freshToken.AccessToken,
		Expiry:      freshToken.Expiry,
	})
}
