// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package claims

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/auth/oauth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/hooks"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// UnaryHook is a hook specific for unary calls in the Identity Server that
// preloads in the context the claims information.
func UnaryHook(store *sql.Store) hooks.UnaryHandlerMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			c, err := buildClaims(ctx, store)
			if err != nil {
				return nil, err
			}

			return next(NewContext(ctx, c), req)
		}
	}
}

// StreamHook is a hook specific for stream calls in the Identity Server that
// preloads in the context the claims information.
func StreamHook(store *sql.Store) hooks.StreamHandlerMiddleware {
	return func(next grpc.StreamHandler) grpc.StreamHandler {
		return func(srv interface{}, stream grpc.ServerStream) error {
			ctx := stream.Context()

			c, err := buildClaims(ctx, store)
			if err != nil {
				return err
			}

			wrapped := grpc_middleware.WrapServerStream(stream)
			wrapped.WrappedContext = NewContext(ctx, c)

			return next(srv, wrapped)
		}
	}
}

// buildClaims returns the claims based on the authentication metadata contained
// in the request. Returns empty claims if no authentication metadata is found.
func buildClaims(ctx context.Context, store *sql.Store) (*Claims, error) {
	md := rpcmetadata.FromIncomingContext(ctx)

	if md.AuthType == "" && md.AuthValue == "" {
		return new(Claims), nil
	}

	if md.AuthType != "Bearer" {
		return nil, errors.Errorf("Expected authentication type to be `Bearer` but got `%s`", md.AuthType)
	}

	header, payload, err := auth.DecodeTokenOrKey(md.AuthValue)
	if err != nil {
		return nil, err
	}

	var res *Claims
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

		res = &Claims{
			entityID:   data.UserID,
			entityType: User,
			source:     auth.Token,
			rights:     rights,
		}
	case auth.Key:
		var entityID string
		var key *ttnpb.APIKey
		var err error

		res := &Claims{
			source: auth.Key,
		}

		switch payload.Type {
		case auth.ApplicationKey:
			entityID, key, err = store.Applications.GetAPIKey(md.AuthValue)

			res.entityType = Application
		case auth.GatewayKey:
			entityID, key, err = store.Gateways.GetAPIKey(md.AuthValue)

			res.entityType = Gateway
		case auth.UserKey:
			entityID, key, err = store.Users.GetAPIKey(md.AuthValue)

			res.entityType = User
		default:
			return nil, errors.Errorf("Invalid API key type `%s`", payload.Type)
		}

		if err != nil {
			return nil, err
		}

		res.entityID = entityID
		res.rights = key.Rights
	default:
		return nil, errors.New("Invalid authentication value")
	}

	return res, nil
}
