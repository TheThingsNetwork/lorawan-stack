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
	"net"
	"net/url"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// Authorizer authorizes requests handled by the interop server.
type Authorizer struct{}

// RequireAuthorized returns an error if the given context is not authorized as neither Network Server nor Application Server.
func (a Authorizer) RequireAuthorized(ctx context.Context) error {
	if ctx.Value(nsAuthInfoKey) != nil || ctx.Value(asAuthInfoKey) != nil {
		return nil
	}
	return errUnauthenticated.New()
}

// RequireAddress returns an error if the given address is not authorized in the context.
func (a Authorizer) RequireAddress(ctx context.Context, addr string) error {
	var authInfo authInfo
	if nsAuthInfo, ok := ctx.Value(nsAuthInfoKey).(*nsAuthInfo); ok {
		authInfo = nsAuthInfo
	} else if asAuthInfo, ok := ctx.Value(asAuthInfoKey).(*asAuthInfo); ok {
		authInfo = asAuthInfo
	} else {
		return errUnauthenticated.New()
	}
	// The authenticated addresses are considered patterns; a *. suffix matches any first host part.
	patterns := authInfo.Addresses()
	if len(patterns) == 0 {
		return errCallerNotAuthorized.WithAttributes("target", addr)
	}

	host := addr
	if url, err := url.Parse(addr); err == nil && url.Host != "" {
		host = url.Host
	}
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	if len(host) == 0 {
		return errCallerNotAuthorized.WithAttributes("target", addr)
	}
	hostParts := strings.Split(host, ".")

	for _, pattern := range patterns {
		patternParts := strings.Split(pattern, ".")
		if len(patternParts) != len(hostParts) {
			return errCallerNotAuthorized.WithAttributes("target", addr)
		}
		for i, patternPart := range patternParts {
			if i == 0 && patternPart == "*" {
				continue
			}
			if patternPart != hostParts[i] {
				return errCallerNotAuthorized.WithAttributes("target", addr)
			}
		}
	}
	return nil
}

// RequireID returns an error if the given NetID is not authorized in the context.
func (a Authorizer) RequireNetID(ctx context.Context, netID types.NetID) error {
	nsAuthInfo, ok := ctx.Value(nsAuthInfoKey).(*nsAuthInfo)
	if !ok {
		return errCallerNotAuthorized.WithAttributes("target", netID.String())
	}
	if !nsAuthInfo.netID.Equal(netID) {
		return errCallerNotAuthorized.WithAttributes("target", netID.String())
	}
	return nil
}

// RequireID returns an error if the given AS-ID is not authorized in the context.
func (a Authorizer) RequireASID(ctx context.Context, id string) error {
	asAuthInfo, ok := ctx.Value(asAuthInfoKey).(*asAuthInfo)
	if !ok {
		return errCallerNotAuthorized.WithAttributes("target", id)
	}
	if asAuthInfo.asID != id {
		return errCallerNotAuthorized.WithAttributes("target", id)
	}
	return nil
}
