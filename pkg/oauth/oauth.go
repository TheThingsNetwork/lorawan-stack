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
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	echo "github.com/labstack/echo/v4"
	"github.com/openshift/osin"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// rightsToScope transforms the list of rights into a string "scope".
// This function is only used for compatibility with osin.
func rightsToScope(rights ...ttnpb.Right) string {
	rights = ttnpb.RightsFrom(rights...).Sorted().GetRights()
	rightStrings := make([]string, len(rights))
	for i, right := range rights {
		rightStrings[i] = right.String()
	}
	return strings.Join(rightStrings, " ")
}

// rightsFromScope is the opposite of RightsToScope. It transforms the string "scope" back into a list of rights.
// This function is only used for compatibility with osin.
func rightsFromScope(scope string) []ttnpb.Right {
	scopes := strings.Split(scope, " ")
	rights := make([]ttnpb.Right, 0, len(scopes))
	for _, scope := range scopes {
		if right, ok := ttnpb.Right_value[scope]; ok {
			rights = append(rights, ttnpb.Right(right))
		}
	}
	return ttnpb.RightsFrom(rights...).Sorted().GetRights()
}

var (
	errClientMissingGrant = errors.DefinePermissionDenied("client_missing_grant", "OAuth client does not have {grant} grant")
	errClientNotApproved  = errors.DefinePermissionDenied("client_not_approved", "OAuth client was not approved")
	errClientRejected     = errors.DefinePermissionDenied("client_rejected", "OAuth client was rejected")
	errClientSuspended    = errors.DefinePermissionDenied("client_suspended", "OAuth client was suspended")
)

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
			return errClientMissingGrant.WithAttributes("grant", "authorization_code")
		}
		switch client.State {
		case ttnpb.STATE_REJECTED:
			return errClientRejected
		case ttnpb.STATE_SUSPENDED:
			return errClientSuspended
		case ttnpb.STATE_REQUESTED:
			// TODO: Allow if user is collaborator (https://github.com/TheThingsNetwork/lorawan-stack/issues/49).
			return errClientNotApproved
		}
		ar.Authorized = client.SkipAuthorization
		ar.Scope = rightsToScope(client.Rights...)
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
				ar.Authorized, _ = strconv.ParseBool(req.PostForm.Get("authorize"))
			case http.MethodGet:
				safeClient := client.PublicSafe()
				clientJSON, _ := jsonpb.TTN().Marshal(safeClient)
				user, err := s.getUser(c)
				if err != nil {
					return err
				}
				safeUser := user.PublicSafe()
				userJSON, _ := jsonpb.TTN().Marshal(safeUser)
				c.Set("page_data", struct {
					Client json.RawMessage `json:"client"`
					User   json.RawMessage `json:"user"`
				}{
					Client: clientJSON,
					User:   userJSON,
				})
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

type tokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type"`
	Code         string `json:"code" form:"code"`
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri"`
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
}

func (r tokenRequest) Values() (values url.Values) {
	values = make(url.Values)
	if r.GrantType != "" {
		values.Set("grant_type", r.GrantType)
	}
	if r.Code != "" {
		values.Set("code", r.Code)
	}
	if r.RefreshToken != "" {
		values.Set("refresh_token", r.RefreshToken)
	}
	if r.RedirectURI != "" {
		values.Set("redirect_uri", r.RedirectURI)
	}
	values.Set("client_id", r.ClientID)
	values.Set("client_secret", r.ClientSecret)
	return
}

func (s *server) Token(c echo.Context) error {
	req := c.Request()

	// Convert request through tokenRequest so that we can accept both forms and JSON.
	var tokenRequest tokenRequest
	if err := c.Bind(&tokenRequest); err != nil {
		return err
	}
	req.Form = tokenRequest.Values()
	req.PostForm = req.Form

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
			if err := s.doLogin(req.Context(), ar.Username, ar.Password); err != nil {
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
