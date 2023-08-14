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

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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

// AssertGatewayBatchRights implements EntityRegistry.
func (is IS) AssertGatewayBatchRights(ctx context.Context, ids []*ttnpb.GatewayIdentifiers, required ...ttnpb.Right) error {
	ctx, err := getAuthenticatedContext(ctx)
	if err != nil {
		return err
	}

	client, err := is.newGatewayBatchAccessClient(ctx)
	if err != nil {
		return err
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, is.AllowInsecureForCredentials())
	if err != nil {
		return err
	}
	_, err = client.AssertRights(ctx, &ttnpb.AssertGatewayRightsRequest{
		GatewayIds: ids,
		Required: &ttnpb.Rights{
			Rights: required,
		},
	}, callOpt)
	return err
}

// AssertGatewayRights implements EntityRegistry.
func (is IS) AssertGatewayRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers, required ...ttnpb.Right) error {
	ctx, err := getAuthenticatedContext(ctx)
	if err != nil {
		return err
	}
	return rights.RequireGateway(ctx, ids, required...)
}

// GetIdentifiersForEUI implements EntityRegistry.
func (is IS) GetIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error) {
	registry, err := is.newRegistryClient(ctx, &ttnpb.GatewayIdentifiers{Eui: req.Eui})
	if err != nil {
		return nil, err
	}
	return registry.GetIdentifiersForEUI(ctx, req, is.WithClusterAuth())
}

// Get implements EntityRegistry.
func (is IS) Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	ctx, err := getAuthenticatedContext(ctx)
	if err != nil {
		return nil, err
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, is.AllowInsecureForCredentials())
	if errors.IsUnauthenticated(err) {
		callOpt = is.WithClusterAuth()
	} else if err != nil {
		return nil, err
	}
	registry, err := is.newRegistryClient(ctx, req.GetGatewayIds())
	if err != nil {
		return nil, err
	}
	return registry.Get(ctx, req, callOpt)
}

// UpdateAntennas updates the gateway antennas.
func (is IS) UpdateAntennas(ctx context.Context, ids *ttnpb.GatewayIdentifiers, antennas []*ttnpb.GatewayAntenna) error {
	ctx, err := getAuthenticatedContext(ctx)
	if err != nil {
		return err
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, is.AllowInsecureForCredentials())
	if err != nil {
		return err
	}

	registry, err := is.newRegistryClient(ctx, ids)
	if err != nil {
		return err
	}

	req := &ttnpb.UpdateGatewayRequest{
		Gateway: &ttnpb.Gateway{
			Ids:      ids,
			Antennas: antennas,
		},
		FieldMask: ttnpb.FieldMask("antennas"),
	}
	_, err = registry.Update(ctx, req, callOpt)
	return err
}

// UpdateAttributes implements EntityRegistry.
func (is IS) UpdateAttributes(ctx context.Context, ids *ttnpb.GatewayIdentifiers, current, new map[string]string) error {
	ctx, err := getAuthenticatedContext(ctx)
	if err != nil {
		return err
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, is.AllowInsecureForCredentials())
	if err != nil {
		return err
	}

	registry, err := is.newRegistryClient(ctx, ids)
	if err != nil {
		return err
	}

	// Take the union of keys with new values overwriting existing entries.
	merged := make(map[string]string)
	for k, v := range current {
		merged[k] = v
	}
	for k, v := range new {
		merged[k] = v
	}

	req := &ttnpb.UpdateGatewayRequest{
		Gateway: &ttnpb.Gateway{
			Ids:        ids,
			Attributes: merged,
		},
		FieldMask: ttnpb.FieldMask("attributes"),
	}
	_, err = registry.Update(ctx, req, callOpt)
	return err
}

// ValidateGatewayID implements EntityRegistry.
func (is IS) ValidateGatewayID(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	return ids.ValidateContext(ctx)
}

func (is IS) newRegistryClient(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayRegistryClient, error) {
	cc, err := is.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayRegistryClient(cc), nil
}

func (is IS) newGatewayBatchAccessClient(ctx context.Context) (ttnpb.GatewayBatchAccessClient, error) {
	cc, err := is.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayBatchAccessClient(cc), nil
}

func getAuthenticatedContext(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
