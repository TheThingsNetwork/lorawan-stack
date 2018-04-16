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

	"github.com/RangelReale/osin"
	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// authorize is an echo handler for the authorize endpoint.
func (s *Server) authorize(authorizePage echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()

		// Only POST and GET are supported.
		if req.Method != http.MethodGet && req.Method != http.MethodPost {
			return c.NoContent(http.StatusMethodNotAllowed)
		}

		resp := s.oauth.NewResponse()
		defer resp.Close()

		ar := s.oauth.HandleAuthorizeRequest(resp, req)
		if ar == nil {
			return s.output(c, resp)
		}
		client := ar.Client.(store.Client).GetClient()

		// Make sure client supports authorization code.
		if !client.HasGrant(ttnpb.GRANT_AUTHORIZATION_CODE) {
			resp.SetError(osin.E_INVALID_CLIENT, "")
			s.oauth.FinishAuthorizeRequest(resp, req, ar)
			return s.output(c, resp)
		}

		user, err := s.getUser(c)
		if err != nil {
			return err
		}
		uids := user.GetUser().UserIdentifiers

		ar.Authorized, err = s.config.Store.OAuth.IsClientAuthorized(uids, client.ClientIdentifiers)
		if err != nil {
			return err
		}

		if !ar.Authorized {
			switch c.Request().Method {
			case http.MethodPost:
				a := req.Form.Get("authorize")
				ar.Authorized = a == "1" || a == "true"
			case http.MethodGet:
				return authorizePage(c)
			}
		}

		ar.UserData = &UserData{
			UserID: uids.UserID,
		}

		// The requested scope is always fixed to the scope of the third-party client.
		ar.Scope = Scope(client.Rights)

		s.oauth.FinishAuthorizeRequest(resp, req, ar)
		return s.output(c, resp)
	}
}
