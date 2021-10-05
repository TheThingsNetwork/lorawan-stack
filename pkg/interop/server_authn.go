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
	"net"
	"net/http"
	"net/url"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// verifySenderNSID verifies that one of the address patterns matches the NSID.
// The pattern may contain a wildcard (*.host) or port (host:1885).
func verifySenderNSID(patterns []string, nsID string) error {
	if len(patterns) == 0 {
		return errCallerNotAuthorized.WithAttributes("target", nsID)
	}

	host := nsID
	if url, err := url.Parse(nsID); err == nil && url.Host != "" {
		host = url.Host
	}
	if h, _, err := net.SplitHostPort(nsID); err == nil {
		host = h
	}
	if len(host) == 0 {
		return errCallerNotAuthorized.WithAttributes("target", nsID)
	}
	hostParts := strings.Split(host, ".")

nextPattern:
	for _, pattern := range patterns {
		patternParts := strings.Split(pattern, ".")
		if len(patternParts) != len(hostParts) {
			return errCallerNotAuthorized.WithAttributes("target", nsID)
		}
		for i, patternPart := range patternParts {
			if i == 0 && patternPart == "*" {
				continue
			}
			if patternPart != hostParts[i] {
				continue nextPattern
			}
		}
		return nil
	}
	return errCallerNotAuthorized.WithAttributes("target", nsID)
}

type authInfo interface {
	GetAddresses() []string
}

// NetworkServerAuthInfo contains authentication information of the Network Server.
type NetworkServerAuthInfo struct {
	NetID     types.NetID
	Addresses []string
}

func (n NetworkServerAuthInfo) GetAddresses() []string { return n.Addresses }

type nsAuthInfoKeyType struct{}

var nsAuthInfoKey nsAuthInfoKeyType

// NewContextWithNetworkServerAuthInfo returns a derived context with the given authentication information of the
// Network Server.
func NewContextWithNetworkServerAuthInfo(parent context.Context, authInfo NetworkServerAuthInfo) context.Context {
	return context.WithValue(parent, nsAuthInfoKey, authInfo)
}

// NetworkServerAuthInfoFromContext returns the authentication information of the Network Server from context.
func NetworkServerAuthInfoFromContext(ctx context.Context) (NetworkServerAuthInfo, bool) {
	authInfo, ok := ctx.Value(nsAuthInfoKey).(NetworkServerAuthInfo)
	return authInfo, ok
}

// authenticateNS authenticates the client as a Network Server.
//
// If the client presents a TLS client certificate, it is verified against the trusted CAs of the NetID.
// Any DNS names in the X.509 Subject Alternative Names are taken as address patterns used to verify the NSID. If there
// are no DNS names, the Common Name is used as the single address pattern.
// If the TLS client certificate verification fails, this method returns an error.
//
// If the client presents a Bearer token in the Authorization header of the HTTP request, it is verified as token issued
// by Packet Broker.
func (s *Server) authenticateNS(ctx context.Context, r *http.Request, data []byte) (context.Context, error) {
	var header NsMessageHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, err
	}
	if !header.ProtocolVersion.SupportsNSID() && header.SenderNSID != nil {
		return nil, ErrMalformedMessage.New()
	}

	// If the client presents a TLS client certificate, use that for authentication.
	if state := r.TLS; state != nil && len(state.PeerCertificates) > 0 {
		addrs, err := s.verifySenderCertificate(ctx, types.NetID(header.SenderID).String(), state)
		if err != nil {
			return nil, err
		}
		if header.SenderNSID != nil {
			if err := verifySenderNSID(addrs, *header.SenderNSID); err != nil {
				return nil, err
			}
		}
		return NewContextWithNetworkServerAuthInfo(ctx, NetworkServerAuthInfo{
			NetID:     types.NetID(header.SenderID),
			Addresses: addrs,
		}), nil
	}

	// TODO: Verify Packet Broker token (https://github.com/TheThingsNetwork/lorawan-stack/issues/4703).

	return ctx, nil
}

type ApplicationServerAuthInfo struct {
	ASID      string
	Addresses []string
}

func (a ApplicationServerAuthInfo) GetAddresses() []string { return a.Addresses }

type asAuthInfoKeyType struct{}

var asAuthInfoKey asAuthInfoKeyType

// NewContextWithApplicationServerAuthInfo returns a derived context with the given authentication information of the
// Application Server.
func NewContextWithApplicationServerAuthInfo(parent context.Context, authInfo ApplicationServerAuthInfo) context.Context {
	return context.WithValue(parent, asAuthInfoKey, authInfo)
}

// ApplicationServerAuthInfoFromContext returns the authentication information of the Application Server from context.
func ApplicationServerAuthInfoFromContext(ctx context.Context) (ApplicationServerAuthInfo, bool) {
	authInfo, ok := ctx.Value(asAuthInfoKey).(ApplicationServerAuthInfo)
	return authInfo, ok
}

func (s *Server) authenticateAS(ctx context.Context, r *http.Request, data []byte) (context.Context, error) {
	var header AsMessageHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, err
	}

	state := r.TLS
	if state == nil {
		return nil, errUnauthenticated.New()
	}
	addrs, err := s.verifySenderCertificate(ctx, header.SenderID, state)
	if err != nil {
		return nil, err
	}
	return NewContextWithApplicationServerAuthInfo(ctx, ApplicationServerAuthInfo{
		ASID:      header.SenderID,
		Addresses: addrs,
	}), nil
}

type senderAuthenticatorFunc func(ctx context.Context, r *http.Request, data []byte) (context.Context, error)

func (f senderAuthenticatorFunc) Authenticate(ctx context.Context, r *http.Request, data []byte) (context.Context, error) {
	return f(ctx, r, data)
}

type senderAuthenticator interface {
	Authenticate(ctx context.Context, r *http.Request, data []byte) (context.Context, error)
}
