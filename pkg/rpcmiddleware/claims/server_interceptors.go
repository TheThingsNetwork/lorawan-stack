// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package claims

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/auth/oauth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// TokenInfoProvider is the interface that online validates and introspects
// OAuth access tokens.
type TokenInfoProvider interface {
	// TokenInfo returns the access data of an OAuth access token.
	// It returns error if token is expired.
	TokenInfo(accessToken string) (*types.AccessData, error)
}

// KeyInfoProvider is the interface that online validates and introspects API keys.
type KeyInfoProvider interface {
	// KeyInfo returns the entityID an API key belongs to and its rights.
	// The Resource Server must check that the entityID of the API key matches
	// with the resource that is trying to being access to.
	KeyInfo(key string) (string, *ttnpb.APIKey, error)
}

// claims constructs the claims based on the authentication values in the request
// metadata. Mainly there are three scenarios:
//   - Authentication values are empty: empty claims are returned
//   - A token is provided: it is introspected (and therefore validated) through
//       the TokenInfoProvider and then claims are built based on this.
//   - A key is provided: it is introspected (and therefore validated) through
//       the KeyInfoProvider and then claims are built based on this.
func claims(ctx context.Context, t TokenInfoProvider, k KeyInfoProvider) (*auth.Claims, error) {
	md := rpcmetadata.FromIncomingContext(ctx)

	if md.AuthType == "" && md.AuthValue == "" {
		return new(auth.Claims), nil
	}

	if md.AuthType != "Bearer" {
		return nil, errors.Errorf("Expected authentication type to be Bearer but got `%s", md.AuthType)
	}

	header, payload, err := auth.DecodeTokenOrKey(md.AuthValue)
	if err != nil {
		return nil, err
	}

	var claims *auth.Claims
	switch header.Type {
	case auth.Token:
		data, err := t.TokenInfo(md.AuthValue)
		if err != nil {
			return nil, err
		}

		rights, err := oauth.ParseScope(data.Scope)
		if err != nil {
			return nil, err
		}

		claims = &Claims{
			EntityID:  data.UserID,
			EntityTyp: auth.EntityUser,
			Source:    auth.Token,
			Rights:    rights,
		}
	case auth.Key:
		entityID, key, err := k.KeyInfo(md.AuthValue)
		if err != nil {
			return nil, err
		}

		claims = &auth.Claims{
			EntityID: entityID,
			Source:   auth.Key,
			Rights:   key.Rights,
		}

		switch payload.Type {
		case auth.ApplicationKey:
			claims.EntityTyp = auth.EntityApplication
		case auth.GatewayKey:
			claims.EntityTyp = auth.EntityGateway
		case auth.UserKey:
			claims.EntityTyp = auth.EntityUser
		default:
			return nil, errors.Errorf("Invalid API key type `%s`", payload.Type)
		}
	default:
		return nil, errors.New("Invalid authentication value")
	}

	return claims, nil
}

// UnaryServerInterceptor returns a new unary server interceptor that construct
// the claims based on the authentication value in the request metadata.
// Empty claims are injected if authentication is missing in the request metadata.
func UnaryServerInterceptor(t TokenInfoProvider, k KeyInfoProvider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInterceptor, handler grpc.UnaryHandler) (interface{}, error) {
		c, err := claims(ctx, t, k)
		if err != nil {
			return nil, err
		}

		return handler(NewContext(ctx, c), req)
	}
}

// StreamServerInterceptor returns a new unary server interceptor that construct
// the claims based on the authentication value in the request metadata.
// Empty claims are injected if authentication is missing in the request metadata.
func StreamServerInterceptor(t TokenInfoProvider, k KeyInfoProvider) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		c, err := claims(ctx, t, k)
		if err != nil {
			return nil, err
		}

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = NewContext(stream.Context(), c)

		return handler(srv, wrapped)
	}
}
