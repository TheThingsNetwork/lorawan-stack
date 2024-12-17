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

package identityserver

import (
	"context"
	"fmt"
	"runtime/trace"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewaytokens"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	errUnauthenticated          = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")
	errUnsupportedAuthorization = errors.DefineUnauthenticated("unsupported_authorization", "unsupported authorization method")
	errAPIKeyNotFound           = errors.DefineUnauthenticated("api_key_not_found", "API key not found")
	errInvalidAuthorization     = errors.DefineUnauthenticated("invalid_authorization", "invalid authorization")
	errTokenNotFound            = errors.DefineUnauthenticated("token_not_found", "token not found")
	errTokenExpired             = errors.DefineUnauthenticated("token_expired", "token expired")
	errAPIKeyExpired            = errors.DefineUnauthenticated("api_key_expired", "api key expired")
	errUserRejected             = errors.DefinePermissionDenied("user_rejected", "user account was rejected", "description")
	errUserRequested            = errors.DefinePermissionDenied("user_requested", "user account approval is pending", "description")
	errUserSuspended            = errors.DefinePermissionDenied("user_suspended", "user account was suspended", "description")
	errOAuthClientRejected      = errors.DefinePermissionDenied("oauth_client_rejected", "OAuth client was rejected", "description")
	errOAuthClientSuspended     = errors.DefinePermissionDenied("oauth_client_suspended", "OAuth client was suspended", "description")
	errPermissionDenied         = errors.DefinePermissionDenied("permission_denied", "unauthorized request to restricted resource")
)

type requestAccessKeyType struct{}

var requestAccessKey requestAccessKeyType

type requestAccess struct {
	authInfo     *ttnpb.AuthInfoResponse
	entityRights map[*ttnpb.EntityIdentifiers]*ttnpb.Rights
}

func (is *IdentityServer) withRequestAccessCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestAccessKey, new(requestAccess))
}

func (is *IdentityServer) authInfo(ctx context.Context) (info *ttnpb.AuthInfoResponse, err error) {
	if access, ok := ctx.Value(requestAccessKey).(*requestAccess); ok {
		if access.authInfo != nil {
			return access.authInfo, nil
		}
		defer func() {
			if err == nil {
				access.authInfo = info
			}
		}()
	}

	md := rpcmetadata.FromIncomingContext(ctx)
	if md.AuthType == "" {
		return &ttnpb.AuthInfoResponse{}, nil
	}

	if md.AuthType == gatewaytokens.AuthType {
		token, err := gatewaytokens.DecodeFromString(md.AuthValue)
		if err != nil {
			return nil, err
		}
		rights, err := gatewaytokens.Verify(ctx, token, is.config.Gateways.TokenValidity, is.KeyService())
		if err != nil {
			if errors.IsNotFound(err) {
				log.FromContext(ctx).WithField("id", token.KeyId).Warn("Key not found in key vault")
			}
			return nil, err
		}
		return &ttnpb.AuthInfoResponse{
			AccessMethod: &ttnpb.AuthInfoResponse_GatewayToken_{
				GatewayToken: &ttnpb.AuthInfoResponse_GatewayToken{
					Rights:     ttnpb.RightsFrom(rights.Rights...).Implied().GetRights(),
					GatewayIds: token.Payload.GatewayIds,
				},
			},
		}, nil
	}

	if md.AuthType == clusterauth.AuthType {
		if err := clusterauth.Authorized(ctx); err != nil {
			return nil, err
		}
		return &ttnpb.AuthInfoResponse{
			UniversalRights: ttnpb.AllClusterRights.Implied(),
		}, nil
	}
	if strings.ToLower(md.AuthType) != "bearer" {
		return nil, errUnsupportedAuthorization.New()
	}

	token := md.AuthValue
	tokenType, tokenID, tokenKey, err := auth.SplitToken(token)
	if err != nil {
		return nil, err
	}

	var fetch func(ctx context.Context, st store.Store) error
	res := &ttnpb.AuthInfoResponse{}
	userFieldMask := []string{
		"admin",
		"state",
		"state_description",
		"primary_email_address_validated_at",
		"universal_rights",
	}
	clientFieldMask := []string{"state", "state_description"}
	var user *ttnpb.User
	var userRights *ttnpb.Rights

	switch tokenType {
	case auth.APIKey:
		fetch = func(ctx context.Context, st store.Store) error {
			ids, apiKey, err := st.GetAPIKeyByID(ctx, tokenID)
			if err != nil {
				if errors.IsNotFound(err) {
					return errAPIKeyNotFound.WithCause(err)
				}
				return err
			}
			region := trace.StartRegion(ctx, "validate api key")
			valid, err := auth.Validate(apiKey.GetKey(), tokenKey)
			region.End()
			if err != nil {
				return errInvalidAuthorization.WithCause(err)
			}
			if !valid {
				return errInvalidAuthorization.New()
			}
			if expiresAt := ttnpb.StdTime(apiKey.ExpiresAt); expiresAt != nil && expiresAt.Before(time.Now()) {
				return errAPIKeyExpired.New()
			}
			apiKey.Key = ""
			apiKey.Rights = ttnpb.RightsFrom(apiKey.Rights...).Implied().GetRights()
			res.AccessMethod = &ttnpb.AuthInfoResponse_ApiKey{
				ApiKey: &ttnpb.AuthInfoResponse_APIKeyAccess{
					ApiKey:    apiKey,
					EntityIds: ids.GetEntityIdentifiers(),
				},
			}
			if ids.EntityType() == "user" {
				user, err = st.GetUser(ctx, ids.GetUserIds(), userFieldMask)
				if err != nil {
					if errors.IsNotFound(err) {
						return errAPIKeyNotFound.WithCause(err)
					}
					return err
				}
				userRights = ttnpb.RightsFrom(apiKey.Rights...)
			}
			return nil
		}
	case auth.AccessToken:
		fetch = func(ctx context.Context, st store.Store) error {
			accessToken, err := st.GetAccessToken(ctx, tokenID)
			if err != nil {
				if errors.IsNotFound(err) {
					return errTokenNotFound.WithCause(err)
				}
				return err
			}
			region := trace.StartRegion(ctx, "validate access token")
			valid, err := auth.Validate(accessToken.GetAccessToken(), tokenKey)
			region.End()
			if err != nil {
				return errInvalidAuthorization.WithCause(err)
			}
			if !valid {
				return errInvalidAuthorization.New()
			}
			if expiresAt := ttnpb.StdTime(accessToken.ExpiresAt); expiresAt != nil && expiresAt.Before(time.Now()) {
				return errTokenExpired.New()
			}
			accessToken.AccessToken, accessToken.RefreshToken = "", ""
			accessToken.Rights = ttnpb.RightsFrom(accessToken.Rights...).Implied().GetRights()
			res.AccessMethod = &ttnpb.AuthInfoResponse_OauthAccessToken{
				OauthAccessToken: accessToken,
			}
			user, err = st.GetUser(ctx, accessToken.UserIds, userFieldMask)
			if err != nil {
				if errors.IsNotFound(err) {
					return errTokenNotFound.WithCause(err)
				}
				return err
			}
			client, err := st.GetClient(ctx, accessToken.ClientIds, clientFieldMask)
			if err != nil {
				if errors.IsNotFound(err) {
					return errTokenNotFound.WithCause(err)
				}
				return err
			}
			switch client.State {
			case ttnpb.State_STATE_REQUESTED:
				// OAuth authorization only passes for collaborators, so this is ok.
			case ttnpb.State_STATE_APPROVED:
				// Normal OAuth client.
			case ttnpb.State_STATE_REJECTED:
				if client.StateDescription != "" {
					return errOAuthClientRejected.WithAttributes("description", client.StateDescription)
				}
				return errOAuthClientRejected.New()
			case ttnpb.State_STATE_FLAGGED:
				// Innocent until proven guilty.
			case ttnpb.State_STATE_SUSPENDED:
				if client.StateDescription != "" {
					return errOAuthClientSuspended.WithAttributes("description", client.StateDescription)
				}
				return errOAuthClientSuspended.New()
			default:
				panic(fmt.Sprintf("Unhandled client state: %s", client.State.String()))
			}
			userRights = ttnpb.RightsFrom(accessToken.Rights...)
			return nil
		}
	case auth.SessionToken:
		fetch = func(ctx context.Context, st store.Store) error {
			session, err := st.GetSessionByID(ctx, tokenID)
			if err != nil {
				if errors.IsNotFound(err) {
					return errTokenNotFound.WithCause(err)
				}
				return err
			}
			region := trace.StartRegion(ctx, "validate session token")
			valid, err := auth.Validate(session.GetSessionSecret(), tokenKey)
			region.End()
			if err != nil {
				return errInvalidAuthorization.WithCause(err)
			}
			if !valid {
				return errInvalidAuthorization.New()
			}
			if expiresAt := ttnpb.StdTime(session.ExpiresAt); expiresAt != nil && expiresAt.Before(time.Now()) {
				return errTokenExpired.New()
			}
			session.SessionSecret = ""
			res.AccessMethod = &ttnpb.AuthInfoResponse_UserSession{
				UserSession: session,
			}
			user, err = st.GetUser(ctx, session.GetUserIds(), userFieldMask)
			if err != nil {
				if errors.IsNotFound(err) {
					return errTokenNotFound.WithCause(err)
				}
				return err
			}

			// Warning: A user authorized by session cookie will be granted all
			// current and future rights. When using this auth type, the respective
			// handlers need to ensure thorough CSRF and CORS protection using
			// appropriate middleware.
			userRights = ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL).Implied()
			return nil
		}
	default:
		return nil, errUnsupportedAuthorization.New()
	}

	if err = is.store.Transact(ctx, fetch); err != nil {
		return nil, err
	}

	if user != nil {
		rpclog.AddField(ctx, "auth.user_id", user.GetIds().GetUserId())

		if is.configFromContext(ctx).UserRegistration.ContactInfoValidation.Required && user.PrimaryEmailAddressValidatedAt == nil {
			// Go to profile page, edit basic settings (such as email), delete account.
			restrictRights(res, ttnpb.RightsFrom(ttnpb.Right_RIGHT_USER_INFO, ttnpb.Right_RIGHT_USER_SETTINGS_BASIC, ttnpb.Right_RIGHT_USER_DELETE))
			warning.Add(ctx, "Restricted rights until email address validated")
		}

		switch user.State {
		case ttnpb.State_STATE_REQUESTED:
			// Go to profile page, edit basic settings (such as email), delete account.
			restrictRights(res, ttnpb.RightsFrom(ttnpb.Right_RIGHT_USER_INFO, ttnpb.Right_RIGHT_USER_SETTINGS_BASIC, ttnpb.Right_RIGHT_USER_DELETE))
			warning.Add(ctx, "Restricted rights while account pending")
		case ttnpb.State_STATE_APPROVED:
			// Normal user.
			if user.Admin {
				res.IsAdmin = true
				if is.configFromContext(ctx).AdminRights.All {
					res.UniversalRights = ttnpb.AllRights.Implied().Intersect(userRights)
				} else {
					res.UniversalRights = ttnpb.AllAdminRights.Implied().Intersect(userRights)
				}

				if len(user.UniversalRights) > 0 {
					res.UniversalRights = res.UniversalRights.Intersect(ttnpb.RightsFrom(user.UniversalRights...))
				}
			}
		case ttnpb.State_STATE_REJECTED:
			// Go to profile page, delete account.
			restrictRights(res, ttnpb.RightsFrom(ttnpb.Right_RIGHT_USER_INFO, ttnpb.Right_RIGHT_USER_DELETE))
			if user.StateDescription != "" {
				warning.Add(ctx, fmt.Sprintf("Restricted rights after account rejection: %s", user.StateDescription))
			} else {
				warning.Add(ctx, "Restricted rights after account rejection")
			}
		case ttnpb.State_STATE_FLAGGED:
			// Innocent until proven guilty.
		case ttnpb.State_STATE_SUSPENDED:
			// Go to profile page.
			restrictRights(res, ttnpb.RightsFrom(ttnpb.Right_RIGHT_USER_INFO))
			if user.StateDescription != "" {
				warning.Add(ctx, fmt.Sprintf("Restricted rights after account suspension: %s", user.StateDescription))
			} else {
				warning.Add(ctx, "Restricted rights after account suspension")
			}
		default:
			panic(fmt.Sprintf("Unhandled user state: %s", user.State.String()))
		}
	}

	return res, nil
}

// AuthInfo implements rights.AuthInfoFetcher.
func (is *IdentityServer) AuthInfo(ctx context.Context) (*ttnpb.AuthInfoResponse, error) {
	return is.authInfo(ctx)
}

// RequireAuthenticated checks the request context for authentication presence
// and returns an error if there is none.
func (is *IdentityServer) RequireAuthenticated(ctx context.Context) error {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return err
	}

	if userID := authInfo.GetEntityIdentifiers().GetUserIds(); userID != nil {
		err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
			user, err := st.GetUser(ctx, userID, []string{"state"})
			if err != nil {
				return err
			}

			switch user.State {
			case ttnpb.State_STATE_APPROVED:
				return nil
			case ttnpb.State_STATE_FLAGGED:
				// Flagged users have the same authentication presence as approved users until proven guilty.
				return nil
			case ttnpb.State_STATE_REQUESTED:
				if user.StateDescription != "" {
					return errUserRequested.WithAttributes("description", user.StateDescription)
				}
				return errUserRequested.New()
			case ttnpb.State_STATE_REJECTED:
				if user.StateDescription != "" {
					return errUserRejected.WithAttributes("description", user.StateDescription)
				}
				return errUserRejected.New()
			case ttnpb.State_STATE_SUSPENDED:
				if user.StateDescription != "" {
					return errUserSuspended.WithAttributes("description", user.StateDescription)
				}
				return errUserSuspended.New()
			default:
				panic(fmt.Sprintf("Unhandled user state: %s", user.State.String()))
			}
		})
		if err != nil {
			return err
		}
	}
	if apiKey := authInfo.GetApiKey(); apiKey != nil {
		return nil
	} else if accessToken := authInfo.GetOauthAccessToken(); accessToken != nil {
		return nil
	} else if userSession := authInfo.GetUserSession(); userSession != nil {
		return nil
	} else if gatewaytoken := authInfo.GetGatewayToken(); gatewaytoken != nil {
		return nil
	}

	if len(authInfo.UniversalRights.GetRights()) > 0 {
		return nil
	}
	return errUnauthenticated.New()
}

// UniversalRights returns the universal rights (that apply to any entity or
// outside entity scope) contained in the request context. This is used to determine
// admin rights.
func (is *IdentityServer) UniversalRights(ctx context.Context) *ttnpb.Rights {
	info, err := is.authInfo(ctx)
	if err == nil {
		return info.GetUniversalRights()
	}
	return nil
}

// IsAdmin returns whether the caller is an admin.
func (is *IdentityServer) IsAdmin(ctx context.Context) bool {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return false
	}
	return authInfo.IsAdmin
}

// RequireAdmin returns an error when the caller is not an admin.
func (is *IdentityServer) RequireAdmin(ctx context.Context) error {
	if !is.IsAdmin(ctx) {
		return errPermissionDenied.New()
	}
	return nil
}

var errUpdateAdminField = errors.DefinePermissionDenied("update_admin_field", "only admins can update the `{field}` field")

// RequireAdminForFieldUpdate returns an error when the caller tries to update an admin-only field.
func (is *IdentityServer) RequireAdminForFieldUpdate(ctx context.Context, fields, adminFields []string) error {
	if is.IsAdmin(ctx) {
		return nil
	}
	isAdminField := func(field string) bool {
		for _, adminField := range adminFields {
			if field == adminField {
				return true
			}
		}
		return false
	}
	for _, field := range fields {
		if isAdminField(field) {
			return errUpdateAdminField.WithAttributes("field", field)
		}
	}
	return nil
}

func restrictRights(info *ttnpb.AuthInfoResponse, rights *ttnpb.Rights) {
	if apiKey := info.GetApiKey().GetApiKey(); apiKey != nil {
		apiKey.Rights = ttnpb.RightsFrom(apiKey.Rights...).Intersect(rights).GetRights()
	} else if token := info.GetOauthAccessToken(); token != nil {
		token.Rights = ttnpb.RightsFrom(token.Rights...).Intersect(rights).GetRights()
	}
	info.UniversalRights = info.UniversalRights.Intersect(rights)
}

type entityAccess struct {
	ttnpb.UnimplementedEntityAccessServer

	*IdentityServer
}

func (ea *entityAccess) AuthInfo(ctx context.Context, _ *emptypb.Empty) (*ttnpb.AuthInfoResponse, error) {
	return ea.authInfo(ctx)
}
