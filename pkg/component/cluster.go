// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"github.com/TheThingsNetwork/ttn/pkg/cluster"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
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
