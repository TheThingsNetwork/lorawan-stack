// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package cluster

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Cluster interface
type Cluster interface {
	// Connect to the cluster
	Connect() error
	// Leave the cluster
	Leave() error
	// GetPeer returns a peer with the given role and the given tags.
	// If the cluster contains more than one peer, the shardKey is used to select the right peer.
	GetPeer(role ttnpb.PeerInfo_Role, tags []string, shardKey []byte) Peer
}

// CustomNew allows you to replace clustering logic. New will call CustomNew if not nil.
var CustomNew func(ctx context.Context, config *config.ServiceBase, services ...rpcserver.Registerer) (Cluster, error)

// New cluster
func New(ctx context.Context, config *config.ServiceBase, services ...rpcserver.Registerer) (Cluster, error) {
	if CustomNew != nil {
		return CustomNew(ctx, config, services...)
	}

	return nil, nil // TODO
}
