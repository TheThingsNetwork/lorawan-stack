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

package gatewayserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/bluele/gcache"
	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
)

// Cluster provides cluster operations.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
	WithClusterAuth() grpc.CallOption
}

// IS exposes Identity Server functions.
type IS struct {
	Cluster
}

// NewIS returns a new IS.
func NewIS(c Cluster) *IS {
	return &IS{
		Cluster: c,
	}
}

// AssertGatewayRights implements EntityRegistry.
func (is IS) AssertGatewayRights(ctx context.Context, ids ttnpb.GatewayIdentifiers, required ...ttnpb.Right) error {
	return rights.RequireGateway(ctx, ids, required...)
}

// GetIdentifiersForEUI implements EntityRegistry.
func (is IS) GetIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error) {
	registry, err := is.newRegistryClient(ctx, &ttnpb.GatewayIdentifiers{Eui: &req.Eui})
	if err != nil {
		return nil, err
	}
	return registry.GetIdentifiersForEUI(ctx, req, is.WithClusterAuth())
}

// Get implements EntityRegistry.
func (is IS) Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, is.AllowInsecureForCredentials())
	if errors.IsUnauthenticated(err) {
		callOpt = is.WithClusterAuth()
	} else if err != nil {
		return nil, err
	}
	registry, err := is.newRegistryClient(ctx, &req.GatewayIdentifiers)
	if err != nil {
		return nil, err
	}
	return registry.Get(ctx, req, callOpt)
}

// UpdateAntennas updates the gateway antennas.
func (is IS) UpdateAntennas(ctx context.Context, ids ttnpb.GatewayIdentifiers, antennas []ttnpb.GatewayAntenna) error {
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, is.AllowInsecureForCredentials())
	if err != nil {
		return err
	}

	registry, err := is.newRegistryClient(ctx, &ids)
	if err != nil {
		return err
	}
	req := &ttnpb.UpdateGatewayRequest{
		Gateway: ttnpb.Gateway{
			GatewayIdentifiers: ids,
			Antennas:           antennas,
		},
		FieldMask: &pbtypes.FieldMask{
			Paths: []string{"antennas"},
		},
	}
	_, err = registry.Update(ctx, req, callOpt)
	return err
}

// ValidateGatewayID implements EntityRegistry.
func (is IS) ValidateGatewayID(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return ids.ValidateContext(ctx)
}

func (is IS) newRegistryClient(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayRegistryClient, error) {
	var (
		cc  *grpc.ClientConn
		err error
	)
	if ids != nil {
		cc, err = is.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, ids)
	} else {
		// Don't pass a (*ttnpb.GatewayIdentifiers)(nil) to GetPeerConn.
		cc, err = is.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	}
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayRegistryClient(cc), nil
}

type CachedEntityRegistry struct {
	EntityRegistry
	identifiers gcache.Cache
	gateways    gcache.Cache
}

type cachedResponse struct {
	result interface{}
	err    error
}

func (is CachedEntityRegistry) GetIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error) {
	if cached, err := is.identifiers.Get(req.Eui); err == nil {
		response := cached.(cachedResponse)
		return response.result.(*ttnpb.GatewayIdentifiers), response.err
	}
	ids, err := is.EntityRegistry.GetIdentifiersForEUI(ctx, req)
	if err := is.identifiers.Set(req.Eui, cachedResponse{ids, err}); err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to cache gateway identifiers")
	}
	return ids, err
}

func (is CachedEntityRegistry) Get(ctx context.Context, in *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	key := fmt.Sprintf("%s:%s", unique.ID(ctx, in.GatewayIdentifiers), strings.Join(in.GetFieldMask().GetPaths(), ","))
	if cached, err := is.gateways.Get(key); err == nil {
		response := cached.(cachedResponse)
		return response.result.(*ttnpb.Gateway), response.err
	}
	gtw, err := is.EntityRegistry.Get(ctx, in)
	if err := is.gateways.Set(key, cachedResponse{gtw, err}); err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to cache gateway")
	}
	return gtw, err
}

func EntityRegistryWithCache(er EntityRegistry, config EntityRegistryCacheConfig) *CachedEntityRegistry {
	buildCache := func() gcache.Cache {
		builder := gcache.New(int(config.Size)).Expiration(config.Timeout)
		if config.Clock != nil {
			builder.Clock(config.Clock)
		}
		return builder.Build()
	}
	return &CachedEntityRegistry{
		EntityRegistry: er,
		identifiers:    buildCache(),
		gateways:       buildCache(),
	}
}
