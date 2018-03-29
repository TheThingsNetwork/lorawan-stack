// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/auth/oauth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/hooks"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

// claimsUnaryHook is a hook specific for unary calls in the Identity Server
// that preloads in the context the claims information.
func claimsUnaryHook(store *sql.Store) hooks.UnaryHandlerMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			c, err := buildClaims(ctx, store)
			if err != nil {
				return nil, err
			}

			return next(newContextWithClaims(ctx, c), req)
		}
	}
}

// buildClaims returns the claims based on the authentication metadata contained
// in the request. Returns empty claims if no authentication metadata is found.
func buildClaims(ctx context.Context, store *sql.Store) (*claims, error) {
	md := rpcmetadata.FromIncomingContext(ctx)

	if md.AuthType == "" && md.AuthValue == "" {
		return new(claims), nil
	}

	if md.AuthType != "Bearer" {
		return nil, errors.Errorf("Expected authentication type to be `Bearer` but got `%s`", md.AuthType)
	}

	header, payload, err := auth.DecodeTokenOrKey(md.AuthValue)
	if err != nil {
		return nil, err
	}

	var res *claims
	switch header.Type {
	case auth.Token:
		data, err := store.OAuth.GetAccessToken(md.AuthValue)
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

		res = &claims{
			EntityIdentifiers: ttnpb.UserIdentifiers{UserID: data.UserID},
			Source:            auth.Token,
			Rights:            rights,
		}
	case auth.Key:
		var key ttnpb.APIKey
		var err error

		res := &claims{
			Source: auth.Key,
		}

		switch payload.Type {
		case auth.UserKey:
			res.EntityIdentifiers, key, err = store.Users.GetAPIKey(md.AuthValue)
		case auth.ApplicationKey:
			res.EntityIdentifiers, key, err = store.Applications.GetAPIKey(md.AuthValue)
		case auth.GatewayKey:
			res.EntityIdentifiers, key, err = store.Gateways.GetAPIKey(md.AuthValue)
		case auth.OrganizationKey:
			res.EntityIdentifiers, key, err = store.Organizations.GetAPIKey(md.AuthValue)
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
