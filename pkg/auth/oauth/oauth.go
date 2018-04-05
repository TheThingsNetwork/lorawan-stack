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
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/RangelReale/osin"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/web"
	"github.com/labstack/echo"
)

const (
	// AuthorizationExpiration is the authorization code expiration in seconds (default 5 minutes).
	AuthorizationExpiration = 300

	// AccessExpiration is the access token expiration in seconds (default 1 hour).
	AccessExpiration = 3600

	// TokenType is the access token type to return.
	TokenType = "bearer"
)

// Server represents an OAuth 2.0 Server.
type Server struct {
	logger     log.Interface
	iss        string
	oauth      *osin.Server
	authorizer Authorizer
}

// New returns a new *Server that is ready to use.
func New(logger log.Interface, iss string, store *sql.Store, authorizer Authorizer) *Server {
	config := &osin.ServerConfig{
		AuthorizationExpiration:     AuthorizationExpiration,
		AccessExpiration:            AccessExpiration,
		ErrorStatusCode:             http.StatusUnauthorized,
		RequirePKCEForPublicClients: false,
		RedirectUriSeparator:        "",
		RetainTokenAfterRefresh:     false,
		AllowClientSecretInParams:   false,
		TokenType:                   "bearer",
		AllowedAuthorizeTypes: osin.AllowedAuthorizeType{
			osin.CODE,
		},
		AllowedAccessTypes: osin.AllowedAccessType{
			osin.AUTHORIZATION_CODE,
			osin.REFRESH_TOKEN,
			osin.PASSWORD,
		},
	}

	storage := &storage{
		store: store,
	}

	s := &Server{
		logger:     logger,
		iss:        iss,
		oauth:      osin.NewServer(config, storage),
		authorizer: authorizer,
	}

	s.oauth.AuthorizeTokenGen = s
	s.oauth.AccessTokenGen = s
	s.oauth.Now = func() time.Time {
		return time.Now()
	}

	return s
}

// Register registers the server to the web server.
func (s *Server) Register(server *web.Server) {
	group := server.Group.Group("/oauth")
	group.Any("/token", s.tokenHandler)
	group.Any("/authorize", s.authorizationHandler)
	group.Any("/info", s.infoHandler)
}

type tokenRequest struct {
	GrantType   string `json:"grant_type" form:"grant_type"`
	Code        string `json:"code" form:"code"`
	RedirectURI string `json:"redirect_uri" form:"redirect_uri"`
}

func (s *Server) tokenHandler(c echo.Context) error {
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
		// not supported
		ar.Authorized = false
	}

	s.oauth.FinishAccessRequest(resp, req, ar)

	return s.output(c, resp)
}

func (s *Server) authorizationHandler(c echo.Context) error {
	req := c.Request()
	resp := s.oauth.NewResponse()
	defer resp.Close()

	ar := s.oauth.HandleAuthorizeRequest(resp, req)
	if ar == nil {
		return s.output(c, resp)
	}
	client := ar.Client.(store.Client)

	// make sure client supports authorization code
	if !client.GetClient().HasGrant(ttnpb.GRANT_AUTHORIZATION_CODE) {
		resp.SetError(osin.E_INVALID_CLIENT, "")
		s.oauth.FinishAuthorizeRequest(resp, req, ar)
		return s.output(c, resp)
	}

	scope, err := ParseScope(ar.Scope)
	if err != nil {
		return err
	}

	// check if scope contains rights client does not have
	leftover := Subtract(scope, client.GetClient().Rights)
	if len(leftover) > 0 {
		return fmt.Errorf("Client does not have access to rights: %s", leftover)
	}

	// make sure the user is logged in or redirect
	userID, err := s.authorizer.CheckLogin(c)
	if err != nil || c.Response().Committed {
		return err
	}

	// check if the user authorized, or redner the form
	authorized, err := s.authorizer.Authorize(c, client)
	if err != nil || c.Response().Committed {
		return err
	}

	ar.UserData = &UserData{
		UserID: userID,
	}

	ar.Authorized = authorized

	s.oauth.FinishAuthorizeRequest(resp, req, ar)
	return s.output(c, resp)
}

func (s *Server) infoHandler(c echo.Context) error {
	req := c.Request()
	resp := s.oauth.NewResponse()
	defer resp.Close()

	ir := s.oauth.HandleInfoRequest(resp, req)
	if ir == nil {
		return s.output(c, resp)
	}

	resp.Output["user_id"] = getUserID(ir.AccessData.UserData)
	s.oauth.FinishInfoRequest(resp, req, ir)

	return s.output(c, resp)
}

func (s *Server) output(c echo.Context, resp *osin.Response) error {
	if resp.IsError && resp.InternalError != nil {
		s.logger.WithError(resp.InternalError).WithFields(log.Fields(
			"StatusCode", resp.StatusCode,
			"StatusText", resp.StatusText,
			"ErrorStatusCode", resp.ErrorStatusCode,
			"URL", resp.URL,
			"ErrorId", resp.ErrorId,
			"Output", resp.Output,
		)).Error("OAuth provider error when handling a request")

		return echo.NewHTTPError(400, "Something went wrong")
	}

	headers := c.Response().Header()

	// Add headers
	for i, k := range resp.Headers {
		for _, v := range k {
			headers.Add(i, v)
		}
	}

	if resp.Type == osin.REDIRECT {
		// output redirect with parameters
		location, err := resp.GetRedirectUrl()
		if err != nil {
			return err
		}
		headers.Add("Location", location)

		return c.NoContent(http.StatusFound)
	}

	return c.JSON(resp.StatusCode, resp.Output)
}
