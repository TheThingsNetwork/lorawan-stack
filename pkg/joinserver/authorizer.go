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

package joinserver

import (
	"context"
	"net"
	"net/url"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Authorizer checks whether the request context is authorized.
type Authorizer interface {
	Authorized(ctx context.Context) error
}

// TrustedOriginAuthorizer authorizes the request context by the trusted address or ID that the origin presents.
// This is typically used in TLS client authentication where the trusted address or ID are presented in the X.509 DN or SANs.
type TrustedOriginAuthorizer interface {
	Authorizer
	RequireAddress(ctx context.Context, addr string) error
	RequireID(ctx context.Context, id string) error
}

// ApplicationAccessAuthorizer authorizes the request context for application access.
type ApplicationAccessAuthorizer interface {
	Authorizer
	RequireApplication(ctx context.Context, id ttnpb.ApplicationIdentifiers, required ...ttnpb.Right) error
}

var (
	// X509DNAuthorizer authorizes the caller by the X.509 Distinguished Name of the presented client certificate.
	X509DNAuthorizer Authorizer = new(x509DNAuthorizer)

	// ClusterAuthorizes authorizes clusters.
	ClusterAuthorizer Authorizer = new(clusterAuthorizer)

	// ApplicationRightsAuthorizes authorizes the caller by application rights.
	ApplicationRightsAuthorizer Authorizer = new(applicationRightsAuthorizer)
)

type x509DNAuthorizer struct {
}

var _ TrustedOriginAuthorizer = (*x509DNAuthorizer)(nil)

// Authorized implements Authorizer.
func (a x509DNAuthorizer) Authorized(ctx context.Context) error {
	if _, ok := auth.X509DNFromContext(ctx); !ok {
		return errUnauthenticated.New()
	}
	return nil
}

// RequireAddress implements TrustedOriginAuthorizer.
func (a x509DNAuthorizer) RequireAddress(ctx context.Context, addr string) error {
	dn, ok := auth.X509DNFromContext(ctx)
	if !ok {
		return errUnauthenticated.New()
	}

	host := addr
	if url, err := url.Parse(addr); err == nil && url.Host != "" {
		host = url.Host
	}
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}

	host = strings.TrimSuffix(host, ".")
	pattern := strings.TrimSuffix(dn.CommonName, ".")
	if len(pattern) == 0 || len(host) == 0 {
		return errCallerNotAuthorized.WithAttributes("name", dn.CommonName)
	}

	patternParts := strings.Split(pattern, ".")
	hostParts := strings.Split(host, ".")
	if len(patternParts) != len(hostParts) {
		return errCallerNotAuthorized.WithAttributes("name", dn.CommonName)
	}
	for i, patternPart := range patternParts {
		if i == 0 && patternPart == "*" {
			continue
		}
		if patternPart != hostParts[i] {
			return errCallerNotAuthorized.WithAttributes("name", dn.CommonName)
		}
	}

	return nil
}

// RequireID implements TrustedOriginAuthorizer.
func (a x509DNAuthorizer) RequireID(ctx context.Context, id string) error {
	dn, ok := auth.X509DNFromContext(ctx)
	if !ok {
		return errUnauthenticated.New()
	}
	if !strings.EqualFold(id, dn.CommonName) {
		return errCallerNotAuthorized.WithAttributes("name", dn.CommonName)
	}
	return nil
}

type clusterAuthorizer struct {
}

// Authorized implements Authorizer.
func (a clusterAuthorizer) Authorized(ctx context.Context) error {
	return clusterauth.Authorized(ctx)
}

type applicationRightsAuthorizer struct {
}

var _ ApplicationAccessAuthorizer = (*applicationRightsAuthorizer)(nil)

// Authorized implements Authorizer.
func (a applicationRightsAuthorizer) Authorized(ctx context.Context) error {
	authInfo, err := rights.AuthInfo(ctx)
	if err != nil {
		return err
	}
	if authInfo == nil {
		return errUnauthenticated.New()
	}
	return nil
}

// RequireApplication implements ApplicationAccessAuthorizer.
func (a applicationRightsAuthorizer) RequireApplication(ctx context.Context, id ttnpb.ApplicationIdentifiers, required ...ttnpb.Right) error {
	return rights.RequireApplication(ctx, id, required...)
}
