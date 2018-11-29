// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"strconv"
	"strings"

	"github.com/RangelReale/osin"
	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func RightsToScope(rights ...ttnpb.Right) string {
	rights = ttnpb.RightsFrom(rights...).Sorted().GetRights()
	rightStrings := make([]string, len(rights))
	for i, right := range rights {
		rightStrings[i] = right.String()
	}
	return strings.Join(rightStrings, " ")
}

func RightsFromScope(scope string) []ttnpb.Right {
	scopes := strings.Split(scope, " ")
	rights := make([]ttnpb.Right, 0, len(scopes))
	for _, scope := range scopes {
		if right, ok := ttnpb.Right_value[scope]; ok {
			rights = append(rights, ttnpb.Right(right))
		}
	}
	return ttnpb.RightsFrom(rights...).Sorted().GetRights()
}

func (s *server) Authorize(authorizePage echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		if req.Method != http.MethodGet && req.Method != http.MethodPost {
			return c.NoContent(http.StatusMethodNotAllowed)
		}
		session, err := s.getSession(c)
		if err != nil {
			return err
		}
		oauth2 := s.oauth2(req.Context())
		resp := oauth2.NewResponse()
		defer resp.Close()
		ar := oauth2.HandleAuthorizeRequest(resp, req)
		if ar == nil {
			return s.output(c, resp)
		}
		ar.UserData = userData{UserIdentifiers: session.UserIdentifiers}
		client := ttnpb.Client(ar.Client.(osinClient))
		if !clientHasGrant(&client, ttnpb.GRANT_AUTHORIZATION_CODE) {
			resp.SetError(osin.E_INVALID_CLIENT, "OAuth client does not have authorization code grant")
			oauth2.FinishAuthorizeRequest(resp, req, ar)
			return s.output(c, resp)
		}
		switch client.State {
		case ttnpb.STATE_REJECTED:
			resp.SetError(osin.E_INVALID_CLIENT, "OAuth client was rejected")
			oauth2.FinishAuthorizeRequest(resp, req, ar)
			return s.output(c, resp)
		case ttnpb.STATE_SUSPENDED:
			resp.SetError(osin.E_INVALID_CLIENT, "OAuth client was suspended")
			oauth2.FinishAuthorizeRequest(resp, req, ar)
			return s.output(c, resp)
		case ttnpb.STATE_REQUESTED:
			// TODO: Allow if user is collaborator.
			resp.SetError(osin.E_INVALID_CLIENT, "OAuth client is not yet approved")
			oauth2.FinishAuthorizeRequest(resp, req, ar)
			return s.output(c, resp)
		}
		ar.Authorized = client.SkipAuthorization
		ar.Scope = RightsToScope(client.Rights...)
		if !ar.Authorized {
			authorization, err := s.store.GetAuthorization(
				req.Context(),
				&session.UserIdentifiers,
				&client.ClientIdentifiers,
			)
			if err != nil && !errors.IsNotFound(err) {
				return err
			}
			if ttnpb.RightsFrom(authorization.GetRights()...).IncludesAll(client.Rights...) {
				ar.Authorized = true
			}
		}
		if !ar.Authorized {
			switch c.Request().Method {
			case http.MethodPost:
				ar.Authorized, _ = strconv.ParseBool(req.Form.Get("authorize")) // TODO: Replace with PostForm
			case http.MethodGet:
				return authorizePage(c)
			}
		}
		if ar.Authorized {
			events.Publish(evtAuthorize(req.Context(), ttnpb.CombineIdentifiers(session.UserIdentifiers, client.ClientIdentifiers), nil))
		}
		oauth2.FinishAuthorizeRequest(resp, req, ar)
		return s.output(c, resp)
	}
}

func (s *server) Token(c echo.Context) error {
	req := c.Request()
	oauth2 := s.oauth2(req.Context())
	resp := oauth2.NewResponse()
	defer resp.Close()
	ar := oauth2.HandleAccessRequest(resp, req)
	if ar == nil {
		return s.output(c, resp)
	}
	client := ttnpb.Client(ar.Client.(osinClient))
	userIDs := ar.UserData.(userData).UserIdentifiers
	ar.GenerateRefresh = clientHasGrant(&client, ttnpb.GRANT_REFRESH_TOKEN)
	switch ar.Type {
	case osin.AUTHORIZATION_CODE:
		ar.Authorized = clientHasGrant(&client, ttnpb.GRANT_AUTHORIZATION_CODE)
	case osin.REFRESH_TOKEN:
		ar.Authorized = clientHasGrant(&client, ttnpb.GRANT_REFRESH_TOKEN)
	case osin.PASSWORD:
		if clientHasGrant(&client, ttnpb.GRANT_PASSWORD) {
			err := s.doLogin(req.Context(), ar.Username, ar.Password)
			if err != nil {
				return err
			}
			ar.Authorized = true
		}
	}
	if ar.Authorized {
		events.Publish(evtTokenExchange(req.Context(), ttnpb.CombineIdentifiers(userIDs, client.ClientIdentifiers), nil))
	}
	oauth2.FinishAccessRequest(resp, req, ar)
	delete(resp.Output, "scope")
	return s.output(c, resp)
}

func clientHasGrant(cli *ttnpb.Client, wanted ttnpb.GrantType) bool {
	for _, grant := range cli.Grants {
		if grant == wanted {
			return true
		}
	}
	return false
}
