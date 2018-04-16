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

package identityserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/oauth"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

const authorizationDataHookName = "authorization-data-fetcher"

// authorizationDataUnaryHook is a hook specific for unary calls that preloads
// in the context the authorization data information based on the provided authorization
// value in the request.
func (is *IdentityServer) authorizationDataUnaryHook() hooks.UnaryHandlerMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ad, err := is.buildAuthorizationData(ctx)
			if err != nil {
				return nil, err
			}

			return next(newContextWithAuthorizationData(ctx, ad), req)
		}
	}
}

// buildAuthorizationData builds an `authorizationData` based on the authorization
// value found in the context, if any. Otherwise returns an empty `authorizationData`.
func (is *IdentityServer) buildAuthorizationData(ctx context.Context) (*authorizationData, error) {
	md := rpcmetadata.FromIncomingContext(ctx)

	if md.AuthType == "" && md.AuthValue == "" {
		return new(authorizationData), nil
	}

	if md.AuthType != "Bearer" {
		return nil, errors.Errorf("Expected authentication type to be `Bearer` but got `%s`", md.AuthType)
	}

	header, payload, err := auth.DecodeTokenOrKey(md.AuthValue)
	if err != nil {
		return nil, err
	}

	var res *authorizationData
	switch header.Type {
	case auth.Token:
		data, err := is.store.OAuth.GetAccessToken(md.AuthValue)
		if err != nil {
			return nil, err
		}

		err = data.IsExpired()
		if err != nil {
			return nil, err
		}

		rights, err := oauth.ParseScope(data.Scope)
		if err != nil {
			return nil, err
		}

		res = &authorizationData{
			EntityIdentifiers: ttnpb.UserIdentifiers{UserID: data.UserID},
			Source:            auth.Token,
			Rights:            rights,
		}
	case auth.Key:
		var key ttnpb.APIKey
		var err error

		res = &authorizationData{
			Source: auth.Key,
		}

		switch payload.Type {
		case auth.UserKey:
			res.EntityIdentifiers, key, err = is.store.Users.GetAPIKey(md.AuthValue)
		case auth.ApplicationKey:
			res.EntityIdentifiers, key, err = is.store.Applications.GetAPIKey(md.AuthValue)
		case auth.GatewayKey:
			res.EntityIdentifiers, key, err = is.store.Gateways.GetAPIKey(md.AuthValue)
		case auth.OrganizationKey:
			res.EntityIdentifiers, key, err = is.store.Organizations.GetAPIKey(md.AuthValue)
		default:
			return nil, errors.Errorf("Invalid API key type `%s`", payload.Type)
		}

		if err != nil {
			return nil, err
		}

		res.Rights = key.Rights
	default:
		return nil, errors.New("Invalid authentication value")
	}

	return res, nil
}
