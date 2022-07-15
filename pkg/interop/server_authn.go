// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package interop

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"gopkg.in/square/go-jose.v2/jwt"
)

type authInfo interface {
	addressPatterns() []string
}

// NetworkServerAuthInfo contains the authentication information of a Network Server.
type NetworkServerAuthInfo struct {
	NetID     types.NetID
	Addresses []string
}

func (n NetworkServerAuthInfo) addressPatterns() []string { return n.Addresses }

// Require returns an error if the given NetID or NSID does not match.
func (n NetworkServerAuthInfo) Require(netID types.NetID, _ *EUI64) error {
	if !n.NetID.Equal(netID) {
		return errUnauthenticated.New()
	}
	// TODO: Verify NSID (https://github.com/TheThingsNetwork/lorawan-stack/issues/4741).
	return nil
}

type nsAuthInfoKeyType struct{}

var nsAuthInfoKey nsAuthInfoKeyType

// NewContextWithNetworkServerAuthInfo returns a derived context with the given authentication information of the
// Network Server.
func NewContextWithNetworkServerAuthInfo(parent context.Context, authInfo *NetworkServerAuthInfo) context.Context {
	return context.WithValue(parent, nsAuthInfoKey, authInfo)
}

// NetworkServerAuthInfoFromContext returns the authentication information of the Network Server from context.
func NetworkServerAuthInfoFromContext(ctx context.Context) (*NetworkServerAuthInfo, bool) {
	authInfo, ok := ctx.Value(nsAuthInfoKey).(*NetworkServerAuthInfo)
	return authInfo, ok
}

// ApplicationServerAuthInfo contains the authentication information of an Application Server.
type ApplicationServerAuthInfo struct {
	ASID      string
	Addresses []string
}

func (a ApplicationServerAuthInfo) addressPatterns() []string { return a.Addresses }

type asAuthInfoKeyType struct{}

var asAuthInfoKey asAuthInfoKeyType

// NewContextWithApplicationServerAuthInfo returns a derived context with the given authentication information of the
// Application Server.
func NewContextWithApplicationServerAuthInfo(
	parent context.Context, authInfo *ApplicationServerAuthInfo,
) context.Context {
	return context.WithValue(parent, asAuthInfoKey, authInfo)
}

// ApplicationServerAuthInfoFromContext returns the authentication information of the Application Server from context.
func ApplicationServerAuthInfoFromContext(ctx context.Context) (*ApplicationServerAuthInfo, bool) {
	authInfo, ok := ctx.Value(asAuthInfoKey).(*ApplicationServerAuthInfo)
	return authInfo, ok
}

type tokenVerifier interface {
	VerifyNetworkServer(context.Context, *jwt.JSONWebToken) (*NetworkServerAuthInfo, error)
	VerifyApplicationServer(context.Context, *jwt.JSONWebToken) (*ApplicationServerAuthInfo, error)
}

// authenticateNS authenticates the client as a Network Server.
//
// If the client presents a TLS client certificate, it is verified against the trusted CAs of the NetID.
// Any DNS names in the X.509 Subject Alternative Names are taken as address patterns used to verify the component
// address (e.g. NetworkServerAddress and ApplicationServerAddress of an EndDevice should match the pattern).
// If there are no DNS names, the Common Name is used as the single address pattern.
// If the TLS client certificate is presented but its verification fails, this method returns an error.
//
// If the client presents a Bearer token in the Authorization header of the HTTP request, it is verified as token issued
// by Packet Broker.
func (s *Server) authenticateNS(ctx context.Context, r *http.Request, data []byte) (context.Context, error) {
	var header NsMessageHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, ErrMalformedMessage.WithCause(err)
	}
	if header.ProtocolVersion.RequiresNSID() != (header.SenderNSID != nil) {
		return nil, ErrMalformedMessage.New()
	}

	for _, authFunc := range []func(context.Context) (*NetworkServerAuthInfo, error){
		// Verify TLS client certificate.
		func(ctx context.Context) (*NetworkServerAuthInfo, error) {
			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				return nil, nil //nolint:nilnil
			}
			addrs, err := s.verifySenderCertificate(ctx, types.NetID(header.SenderID).String(), r.TLS)
			if err != nil {
				return nil, err
			}
			return &NetworkServerAuthInfo{
				NetID:     types.NetID(header.SenderID),
				Addresses: addrs,
			}, nil
		},
		// Verify token in a best-effort manner.
		func(ctx context.Context) (*NetworkServerAuthInfo, error) {
			logger := log.FromContext(ctx).WithField("authenticator", "packetbroker")
			authz := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(authz) < 2 || strings.ToLower(authz[0]) != "bearer" {
				return nil, nil //nolint:nilnil
			}
			token, err := jwt.ParseSigned(authz[1])
			if err != nil {
				logger.WithError(err).Debug("Failed to parse token")
				return nil, nil //nolint:nilnil
			}
			var claims jwt.Claims
			err = token.UnsafeClaimsWithoutVerification(&claims)
			if err != nil {
				logger.WithError(err).Debug("Failed to parse claims")
				return nil, nil //nolint:nilnil
			}
			if tokenVerifier, ok := s.tokenVerifiers[claims.Issuer]; ok {
				authInfo, err := tokenVerifier.VerifyNetworkServer(ctx, token)
				if err != nil {
					return nil, errUnauthenticated.WithCause(err)
				}
				return authInfo, nil
			}
			logger.WithError(err).WithField("issuer", claims.Issuer).Debug("Unknown token issuer")
			return nil, nil //nolint:nilnil
		},
	} {
		authInfo, err := authFunc(ctx)
		if err != nil {
			return nil, ErrUnknownSender.WithCause(err)
		}
		if authInfo != nil {
			if err := authInfo.Require(types.NetID(header.SenderID), header.SenderNSID); err != nil {
				return nil, ErrUnknownSender.WithCause(err)
			}
			return NewContextWithNetworkServerAuthInfo(ctx, authInfo), nil
		}
	}

	return nil, ErrUnknownSender.New()
}

func (s *Server) authenticateAS(ctx context.Context, r *http.Request, data []byte) (context.Context, error) {
	var header AsMessageHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, ErrMalformedMessage.WithCause(err)
	}

	state := r.TLS
	if state == nil {
		return nil, ErrUnknownSender.New()
	}
	addrs, err := s.verifySenderCertificate(ctx, header.SenderID, state)
	if err != nil {
		return nil, ErrUnknownSender.WithCause(err)
	}
	return NewContextWithApplicationServerAuthInfo(ctx, &ApplicationServerAuthInfo{
		ASID:      header.SenderID,
		Addresses: addrs,
	}), nil
}

type senderAuthenticatorFunc func(ctx context.Context, r *http.Request, data []byte) (context.Context, error)

func (f senderAuthenticatorFunc) Authenticate(
	ctx context.Context, r *http.Request, data []byte,
) (context.Context, error) {
	return f(ctx, r, data)
}

type senderAuthenticator interface {
	Authenticate(ctx context.Context, r *http.Request, data []byte) (context.Context, error)
}
