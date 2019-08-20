// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package component

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (c *Component) initCluster() (err error) {
	clusterOpts := []cluster.Option{
		cluster.WithServices(c.grpcSubsystems...),
		cluster.WithConn(c.LoopbackConn()),
	}
	if tlsConfig, err := c.config.TLS.Config(c.Context()); err == nil {
		clusterOpts = append(clusterOpts, cluster.WithTLSConfig(tlsConfig))
	}
	c.cluster, err = c.clusterNew(c.ctx, &c.config.ServiceBase.Cluster, clusterOpts...)
	if err != nil {
		return err
	}
	return nil
}

// ClusterTLS returns whether the cluster uses TLS for cluster connections.
func (c *Component) ClusterTLS() bool {
	return c.cluster.TLS()
}

// GetPeers returns cluster peers with the given role and the given tags.
// See package ../cluster for more information.
func (c *Component) GetPeers(ctx context.Context, role ttnpb.ClusterRole) []cluster.Peer {
	return c.cluster.GetPeers(ctx, role)
}

// GetPeer returns a cluster peer with the given role and the given tags.
// See package ../cluster for more information.
func (c *Component) GetPeer(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) cluster.Peer {
	return c.cluster.GetPeer(ctx, role, ids)
}

// ClaimIDs claims the identifiers in the cluster.
// See package ../cluster for more information.
func (c *Component) ClaimIDs(ctx context.Context, ids ttnpb.Identifiers) error {
	return c.cluster.ClaimIDs(ctx, ids)
}

// UnclaimIDs unclaims the identifiers in the cluster.
// See package ../cluster for more information.
func (c *Component) UnclaimIDs(ctx context.Context, ids ttnpb.Identifiers) error {
	return c.cluster.UnclaimIDs(ctx, ids)
}
