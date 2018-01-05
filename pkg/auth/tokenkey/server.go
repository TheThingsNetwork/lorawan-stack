// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/auth/oauth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var _ ttnpb.IsTokenKeyInfoServer = new(Server)

// ErrNotAuthorized is the error returned when the authorization value in the
// request does not match to the secret.
var ErrNotAuthorized = errors.New("not authorized")

// Server implements ttnpb.IsTokenKeyInfoServer.
type Server struct {
	// secret is the known-secret by the Resource Servers contained in the credentials
	// of every request in order to be authorized to use this gRPC service.
	secret string
	store  *sql.Store
}

// New returns a new Server.
func New(secret string, store *sql.Store) *Server {
	if len(secret) == 0 {
		panic(errors.New("secret cannot be empty"))
	}

	return &Server{
		secret: secret,
		store:  store,
	}
}

// authorize checks if the right secret is contained in the request metadata.
func (s *Server) authorized(ctx context.Context) error {
	md := rpcmetadata.FromIncomingContext(ctx)

	if !(subtle.ConstantTimeEq(int32(len(s.secret)), int32(len(md.AuthValue))) == 1 && subtle.ConstantTimeCompare([]byte(s.secret), []byte(md.AuthValue)) == 1) {
		return ErrNotAuthorized
	}

	return nil
}

// GetTokenInfo returns the information about the requested access token.
// It returns error if the token is expired.
func (s *Server) GetTokenInfo(ctx context.Context, req *ttnpb.GetTokenInfoRequest) (*ttnpb.GetTokenInfoResponse, error) {
	err := s.authorized(ctx)
	if err != nil {
		return nil, err
	}

	data, err := s.store.OAuth.GetAccessToken(req.AccessToken)
	if err != nil {
		return nil, err
	}

	err = data.IsExpired()
	if err != nil {
		return nil, err
	}

	return &ttnpb.GetTokenInfoResponse{
		AccessToken: data.AccessToken,
		TokenType:   oauth.TokenType,
		ClientID:    data.ClientID,
		Scope:       data.Scope,
		ExpiresIn:   int32(data.ExpiresIn.Seconds()),
		UserID:      data.UserID,
	}, nil
}

// GetKeyInfo returns the information about the requested API key.
func (s *Server) GetKeyInfo(ctx context.Context, req *ttnpb.GetKeyInfoRequest) (*ttnpb.GetKeyInfoResponse, error) {
	err := s.authorized(ctx)
	if err != nil {
		return nil, err
	}

	header, payload, err := auth.DecodeTokenOrKey(key)
	if err != nil {
		return nil, err
	}

	resp := new(ttnpb.GetKeyInfoResponse)

	switch req.Type {
	case auth.UserKey:
		resp.EntityID, resp.Key, err = s.store.Users.GetAPIKey(req.Key)
	case auth.ApplicationKey:
		resp.EntityID, resp.Key, err = s.store.Applications.GetAPIKey(req.Key)
	case auth.GatewayKey:
		resp.EntityID, resp.Key, err = s.store.Gateways.GetAPIKey(req.Key)
	default:
		return nil, sql.ErrAPIKeyNotFound.New(nil)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}
