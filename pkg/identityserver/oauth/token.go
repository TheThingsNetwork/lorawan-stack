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
	"net/url"

	"github.com/RangelReale/osin"
	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// tokenRequest is a request for the OAuth token endpoint.
type tokenRequest struct {
	GrantType   string `json:"grant_type" form:"grant_type"`
	Code        string `json:"code" form:"code"`
	RedirectURI string `json:"redirect_uri" form:"redirect_uri"`
}

// token is the echo.Handler for getting an OAuth token.
func (s *Server) token(c echo.Context) error {
	req := c.Request()
	resp := s.oauth.NewResponse()
	defer resp.Close()

	tr := &tokenRequest{}
	err := c.Bind(tr)
	if err != nil {
		return err
	}

	if req.Form == nil {
		req.Form = make(url.Values)
	}

	req.Form.Set("grant_type", tr.GrantType)
	req.Form.Set("code", tr.Code)
	req.Form.Set("redirect_uri", tr.RedirectURI)

	ar := s.oauth.HandleAccessRequest(resp, req)
	if ar == nil {
		return s.output(c, resp)
	}

	client := ar.Client.(store.Client).GetClient()

	switch ar.Type {
	case osin.AUTHORIZATION_CODE:
		ar.Authorized = client != nil && client.HasGrant(ttnpb.GRANT_AUTHORIZATION_CODE)
	case osin.REFRESH_TOKEN:
		ar.Authorized = client != nil && client.HasGrant(ttnpb.GRANT_REFRESH_TOKEN)
	case osin.PASSWORD:
		ar.Authorized = client != nil && client.HasGrant(ttnpb.GRANT_PASSWORD)
	case osin.CLIENT_CREDENTIALS, osin.ASSERTION, osin.IMPLICIT:
		// Not supported.
		ar.Authorized = false
	default:
		ar.Authorized = false
	}

	s.oauth.FinishAccessRequest(resp, req, ar)

	return s.output(c, resp)
}
