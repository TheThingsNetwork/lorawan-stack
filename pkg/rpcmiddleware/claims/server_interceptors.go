// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package claims

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/apikey"
	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/tokenkey"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func claims(ctx context.Context, tokenkey tokenkey.Provider, keyprovider apikey.Provider) (*auth.Claims, error) {
	md := rpcmetadata.FromIncomingContext(ctx)

	if md.AuthType == "" && md.AuthValue == "" {
		return new(auth.Claims), nil
	}

	if md.AuthType != "Bearer" {
		return nil, errors.Errorf("Expected authentication type to be Bearer but got '%s'", md.AuthType)
	}

	return auth.ClaimsFromTokenOrKey(tokenkey, keyprovider, md.AuthValue)
}

// UnaryServerInterceptor returns a new server interceptor that injects in the
// context the claims of the key in the authorization metadata field.
// If authorization header is empty an empty claims will be set.
func UnaryServerInterceptor(tokenkey tokenkey.Provider, apikey apikey.Provider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		clms, err := claims(ctx, tokenkey, apikey)
		if err != nil {
			return nil, err
		}

		return handler(NewContext(ctx, clms), req)
	}
}

// StreamServerInterceptor returns a new server interceptor that injects in the
// context the claims of the key in the authorization metadata field.
// If authorization header is empty an empty claims will be set.
func StreamServerInterceptor(tokenkey tokenkey.Provider, apikey apikey.Provider) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		clms, err := claims(stream.Context(), tokenkey, apikey)
		if err != nil {
			return err
		}

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = NewContext(stream.Context(), clms)

		return handler(srv, wrapped)
	}
}
