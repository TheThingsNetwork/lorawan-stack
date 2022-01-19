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
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
	"github.com/openshift/osin"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

var tokenHashSettings auth.HashValidator = pbkdf2.PBKDF2{
	Iterations: 1000,
	KeyLength:  32,
	Algorithm:  pbkdf2.Sha256,
	SaltLength: 16,
}

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

func (s *server) Authorize(authorizePage http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r, session, err := s.session.Get(w, r)
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		oauth2 := s.oauth2(r.Context())
		resp := oauth2.NewResponse()
		defer resp.Close()
		ar := oauth2.HandleAuthorizeRequest(resp, r)
		if ar == nil {
			s.output(w, r, resp)
			return
		}
		ar.UserData = userData{UserSessionIdentifiers: &ttnpb.UserSessionIdentifiers{
			UserIds:   session.GetUserIds(),
			SessionId: session.SessionId,
		}}
		client := ttnpb.Client(ar.Client.(osinClient))
		if !clientHasGrant(&client, ttnpb.GrantType_GRANT_AUTHORIZATION_CODE) {
			resp.InternalError = errClientMissingGrant.WithAttributes("grant", "authorization_code")
			resp.SetError(osin.E_INVALID_GRANT, resp.InternalError.Error())
			oauth2.FinishAuthorizeRequest(resp, r, ar)
			s.output(w, r, resp)
			return
		}
		switch client.State {
		case ttnpb.STATE_REJECTED:
			resp.InternalError = errClientRejected
			resp.SetError(osin.E_INVALID_CLIENT, resp.InternalError.Error())
			oauth2.FinishAuthorizeRequest(resp, r, ar)
			s.output(w, r, resp)
			return
		case ttnpb.STATE_SUSPENDED:
			resp.InternalError = errClientSuspended
			resp.SetError(osin.E_INVALID_CLIENT, resp.InternalError.Error())
			oauth2.FinishAuthorizeRequest(resp, r, ar)
			s.output(w, r, resp)
			return
		case ttnpb.STATE_REQUESTED:
			// TODO: Allow if user is collaborator (https://github.com/TheThingsNetwork/lorawan-stack/issues/49).
			resp.InternalError = errClientNotApproved
			resp.SetError(osin.E_INVALID_CLIENT, resp.InternalError.Error())
			oauth2.FinishAuthorizeRequest(resp, r, ar)
			s.output(w, r, resp)
			return
		}
		ar.Authorized = client.SkipAuthorization
		ar.Scope = rightsToScope(client.Rights...)
		if !ar.Authorized {
			authorization, err := s.store.GetAuthorization(
				r.Context(),
				session.GetUserIds(),
				client.GetIds(),
			)
			if err != nil && !errors.IsNotFound(err) {
				webhandlers.Error(w, r, err)
				return
			}
			if ttnpb.RightsFrom(authorization.GetRights()...).IncludesAll(client.Rights...) {
				ar.Authorized = true
			}
		}
		if !ar.Authorized {
			switch r.Method {
			case http.MethodPost:
				ar.Authorized, _ = strconv.ParseBool(r.PostForm.Get("authorize"))
			case http.MethodGet:
				safeClient := client.PublicSafe()
				clientJSON, _ := jsonpb.TTN().Marshal(safeClient)
				r, user, err := s.session.GetUser(w, r)
				if err != nil {
					webhandlers.Error(w, r, err)
					return
				}
				safeUser := user.PublicSafe()
				userJSON, err := jsonpb.TTN().Marshal(safeUser)
				if err != nil {
					webhandlers.Error(w, r, err)
					return
				}
				r = webui.WithPageData(r, struct {
					Client json.RawMessage `json:"client"`
					User   json.RawMessage `json:"user"`
				}{
					Client: clientJSON,
					User:   userJSON,
				})
				authorizePage.ServeHTTP(w, r)
				return
			}
		}
		if ar.Authorized {
			events.Publish(evtAuthorize.New(r.Context(), events.WithIdentifiers(session.GetUserIds(), client.GetIds())))
		}
		oauth2.FinishAuthorizeRequest(resp, r, ar)
		s.output(w, r, resp)
	}
}

type tokenRequest struct {
	GrantType    string `json:"grant_type" schema:"grant_type"`
	Code         string `json:"code" schema:"code"`
	RefreshToken string `json:"refresh_token" schema:"refresh_token"`
	RedirectURI  string `json:"redirect_uri" schema:"redirect_uri"`
	ClientID     string `json:"client_id" schema:"client_id"`
	ClientSecret string `json:"client_secret" schema:"client_secret"`
}

var (
	errMissingGrantType         = errors.DefineInvalidArgument("missing_grant_type", "missing grant type")
	errInvalidGrantType         = errors.DefineInvalidArgument("invalid_grant_type", "invalid grant type `{grant_type}`")
	errMissingAuthorizationCode = errors.DefineInvalidArgument("missing_authorization_code", "missing authorization code")
	errMissingRefreshToken      = errors.DefineInvalidArgument("missing_refresh_token", "missing refresh token")
	errMissingClientID          = errors.DefineInvalidArgument("missing_client_id", "missing client id")
	errMissingClientSecret      = errors.DefineInvalidArgument("missing_client_secret", "missing client secret")
)

// ValidateContext validates the token request.
func (req *tokenRequest) ValidateContext(ctx context.Context) error {
	if strings.TrimSpace(req.GrantType) == "" {
		return errMissingGrantType.New()
	}
	switch req.GrantType {
	case "authorization_code":
		if strings.TrimSpace(req.Code) == "" {
			return errMissingAuthorizationCode.New()
		}
	case "refresh_token":
		if strings.TrimSpace(req.RefreshToken) == "" {
			return errMissingRefreshToken.New()
		}
	default:
		return errInvalidGrantType.WithAttributes("grant_type", req.GrantType)
	}
	if strings.TrimSpace(req.ClientID) == "" {
		return errMissingClientID.New()
	}
	if strings.TrimSpace(req.ClientSecret) == "" &&
		req.ClientID != "cli" { // NOTE: Compatibility: The CLI does not have a client secret.
		return errMissingClientSecret.New()
	}
	if err := (&ttnpb.ClientIdentifiers{
		ClientId: req.ClientID,
	}).ValidateFields("client_id"); err != nil {
		return err
	}
	return nil
}

var errParse = errors.DefineAborted("parse", "request body parsing")

func (s *server) Token(w http.ResponseWriter, r *http.Request) {
	// Convert request through tokenRequest so that we can accept both forms and JSON.
	var tokenRequest tokenRequest
	switch r.Header.Get("Content-Type") {
	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(&tokenRequest); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
	default:
		if err := r.ParseForm(); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
		if err := s.schemaDecoder.Decode(&tokenRequest, r.Form); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
	}
	if username, password, ok := r.BasicAuth(); ok {
		tokenRequest.ClientID, tokenRequest.ClientSecret = username, password
	}
	if err := tokenRequest.ValidateContext(r.Context()); err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	r = r.WithContext(
		log.NewContextWithField(r.Context(), "oauth_client_id", tokenRequest.ClientID),
	)

	values := make(url.Values)
	if err := schema.NewEncoder().Encode(tokenRequest, values); err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	r.Form = values
	r.PostForm = values

	oauth2 := s.oauth2(r.Context())
	resp := oauth2.NewResponse()
	defer resp.Close()
	ar := oauth2.HandleAccessRequest(resp, r)
	if ar == nil {
		s.output(w, r, resp)
		return
	}

	client := ttnpb.Client(ar.Client.(osinClient))
	userIDs := ar.UserData.(userData).UserSessionIdentifiers.GetUserIds()
	ar.GenerateRefresh = clientHasGrant(&client, ttnpb.GrantType_GRANT_REFRESH_TOKEN)
	switch ar.Type {
	case osin.AUTHORIZATION_CODE:
		ar.Authorized = clientHasGrant(&client, ttnpb.GrantType_GRANT_AUTHORIZATION_CODE)
	case osin.REFRESH_TOKEN:
		ar.Authorized = clientHasGrant(&client, ttnpb.GrantType_GRANT_REFRESH_TOKEN)
	case osin.PASSWORD:
		if clientHasGrant(&client, ttnpb.GrantType_GRANT_PASSWORD) {
			if err := s.session.DoLogin(r.Context(), ar.Username, ar.Password); err != nil {
				webhandlers.Error(w, r, err)
				return
			}
			ar.Authorized = true
		}
	}
	if ar.Authorized {
		events.Publish(evtTokenExchange.New(r.Context(), events.WithIdentifiers(userIDs, client.GetIds())))
	}
	oauth2.FinishAccessRequest(resp, r, ar)
	delete(resp.Output, "scope")
	s.output(w, r, resp)
}

func clientHasGrant(cli *ttnpb.Client, wanted ttnpb.GrantType) bool {
	for _, grant := range cli.Grants {
		if grant == wanted {
			return true
		}
	}
	return false
}
