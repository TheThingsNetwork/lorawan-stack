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

package oauth

import (
	"net/http"
	"net/url"

	"github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	osin "github.com/openshift/osin"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
)

var (
	errInvalidLogoutRedirectURI = errors.DefineInvalidArgument(
		"invalid_logout_redirect_uri",
		"the redirect URI did not match the one(s) defined in the client",
	)
	errMissingAccessTokenIDParam = errors.DefinePermissionDenied(
		"missing_param_access_token_id",
		"access token ID was not provided",
	)
)

func (s *server) ClientLogout(c echo.Context) error {
	ctx := c.Request().Context()
	accessTokenID := c.QueryParam("access_token_id")
	redirectURI := s.config.UI.MountPath()
	if accessTokenID == "" {
		return errMissingAccessTokenIDParam
	}
	at, err := s.store.GetAccessToken(ctx, accessTokenID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if at != nil {
		client, err := s.store.GetClient(ctx, &at.ClientIds, &types.FieldMask{Paths: []string{"logout_redirect_uris"}})
		if err != nil {
			return err
		}
		if err = s.store.DeleteAccessToken(ctx, accessTokenID); err != nil {
			return err
		}
		events.Publish(evtAccessTokenDeleted.NewWithIdentifiersAndData(ctx, &at.UserIds, nil))
		err = s.store.DeleteSession(ctx, &at.UserIds, at.UserSessionID)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		events.Publish(EvtUserLogout.NewWithIdentifiersAndData(ctx, &at.UserIds, nil))
		redirectParam := c.QueryParam("post_logout_redirect_uri")
		if redirectParam == "" {
			if len(client.LogoutRedirectURIs) != 0 {
				redirectURI = client.LogoutRedirectURIs[0]
			}
		} else {
			for _, uri := range client.LogoutRedirectURIs {
				redirectURI, err = osin.ValidateUri(uri, redirectParam)
				if err == nil {
					break
				}
			}
			if err != nil {
				return errInvalidLogoutRedirectURI.WithCause(err)
			}
		}
	}
	session, err := s.session.Get(c)
	if err != nil && !errors.IsUnauthenticated(err) && !errors.IsNotFound(err) {
		return err
	}
	if session != nil {
		events.Publish(evtUserSessionTerminated.NewWithIdentifiersAndData(ctx, &session.UserIdentifiers, nil))
		if err = s.store.DeleteSession(ctx, &session.UserIdentifiers, session.SessionID); err != nil {
			return err
		}
	}
	s.session.RemoveAuthCookie(c)
	url, err := url.Parse(redirectURI)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, url.String())
}
