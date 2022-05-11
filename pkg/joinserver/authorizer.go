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

	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// Authorizer checks whether the request context is authorized.
type Authorizer interface {
	// RequireAuthorized returns an error if the given context is not authorized.
	RequireAuthorized(ctx context.Context) error
}

// EntityAuthorizer authorizes the request context with the entity context.
type EntityAuthorizer interface {
	// RequireEntityContext returns an error if the given entity context is not authorized.
	RequireEntityContext(ctx context.Context) error
}

// ExternalAuthorizer authorizes the request context by the identity that the origin presents.
type ExternalAuthorizer interface {
	Authorizer
	// RequireAddress returns an error if the given address is not authorized in the context.
	RequireAddress(ctx context.Context, addr string) error
	// RequireNetID returns an error if the given NetID is not authorized in the context.
	RequireNetID(ctx context.Context, netID types.NetID) error
	// RequireASID returns an error if the given AS-ID is not authorized in the context.
	RequireASID(ctx context.Context, id string) error
}

// ApplicationAccessAuthorizer authorizes the request context for application access.
type ApplicationAccessAuthorizer interface {
	Authorizer
	RequireApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers, required ...ttnpb.Right) error
}

var (
	// InteropAuthorizer authorizes the caller by proof of identity used with LoRaWAN Backend Interfaces.
	InteropAuthorizer ExternalAuthorizer = new(interop.Authorizer)
)

type clusterAuthorizer struct {
	reqCtx context.Context
}

var _ EntityAuthorizer = (*clusterAuthorizer)(nil)

// ClusterAuthorizer returns an Authorizer that authenticates clusters.
func ClusterAuthorizer(reqCtx context.Context) Authorizer {
	return clusterAuthorizer{reqCtx}
}

// RequireAuthorized implements Authorizer.
func (a clusterAuthorizer) RequireAuthorized(ctx context.Context) error {
	return clusterauth.Authorized(ctx)
}

// RequireEntityContext returns nil.
func (a clusterAuthorizer) RequireEntityContext(ctx context.Context) error {
	return nil
}

type applicationRightsAuthorizer struct {
	reqCtx context.Context
}

var _ EntityAuthorizer = (*applicationRightsAuthorizer)(nil)

// ApplicationRightsAuthorizer returns an Authorizer that authenticates the caller by application rights.
func ApplicationRightsAuthorizer(reqCtx context.Context) Authorizer {
	return &applicationRightsAuthorizer{reqCtx}
}

// RequireEntityContext returns nil.
func (a applicationRightsAuthorizer) RequireEntityContext(ctx context.Context) error {
	return nil
}

// RequireAuthorized implements Authorizer.
func (a applicationRightsAuthorizer) RequireAuthorized(ctx context.Context) error {
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
func (a applicationRightsAuthorizer) RequireApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers, required ...ttnpb.Right) error {
	return rights.RequireApplication(ctx, id, required...)
}
