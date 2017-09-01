// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
)

// Server is a gRPC server that can serve public keys for JWT validation.
type Server struct {
	manager *auth.Manager
}

// New returns a new Server, ready for use.
func New(manager *auth.Manager) *Server {
	return &Server{
		manager: manager,
	}
}

// GetTokenKey gets the token public key for with the specified kid.
func (s *Server) GetTokenKey(ctx context.Context, in *ttnpb.TokenKeyRequest, opts ...grpc.CallOption) (*ttnpb.TokenKeyResponse, error) {
	kid := in.GetKID()
	key, err := s.manager.GetTokenKey(kid)
	if err != nil {
		return nil, err
	}

	var alg string
	block := &pem.Block{}
	switch key.(type) {
	case *rsa.PublicKey:
		alg = jwt.SigningMethodRS512.Name
		block.Type = "RSA PUBLIC KEY"
	case *ecdsa.PublicKey:
		alg = jwt.SigningMethodES512.Name
		block.Type = "EC PUBLIC KEY"
	default:
		return nil, auth.ErrUnsupportedSigningMethod
	}

	encoded, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	block.Bytes = encoded

	return &ttnpb.TokenKeyResponse{
		KID:       kid,
		Algorithm: alg,
		PublicKey: string(pem.EncodeToMemory(block)),
	}, nil
}
