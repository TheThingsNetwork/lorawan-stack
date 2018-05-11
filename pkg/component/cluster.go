// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (c *Component) initCluster() (err error) {
	c.cluster, err = cluster.New(c.ctx, &c.config.ServiceBase, c.grpcSubsystems...)
	if err != nil {
		return err
	}
	return nil
}

// GetPeers returns cluster peers with the given role and the given tags.
// See package ../cluster for more information.
func (c *Component) GetPeers(role ttnpb.PeerInfo_Role, tags []string) []cluster.Peer {
	return c.cluster.GetPeers(role, tags)
}

// GetPeer returns a cluster peer with the given role and the given tags.
// See package ../cluster for more information.
func (c *Component) GetPeer(role ttnpb.PeerInfo_Role, tags []string, shardKey []byte) cluster.Peer {
	return c.cluster.GetPeer(role, tags, shardKey)
}
