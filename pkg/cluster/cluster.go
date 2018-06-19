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

// Package cluster defines an interface for clustering network components and provides a simple implementation.
package cluster

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

// Cluster interface that is implemented by all different clustering implementations.
type Cluster interface {
	// Join the cluster.
	Join() error
	// Leave the cluster.
	Leave() error
	// GetPeers returns peers with the given role and the given tags.
	GetPeers(role ttnpb.PeerInfo_Role, tags []string) []Peer
	// GetPeer returns a peer with the given role and the given tags.
	// If the cluster contains more than one peer, the shardKey is used to select the right peer.
	// Tagging and sharding is not part of the reference implementation. The idea of tagging is another layer of filtering
	// peers, which allows network operators to have dedicated instances for premium customers, tenants or separated
	// environments. The idea of sharding is that if multiple peers match the filters, we can still consistently select
	// a single peer. The shardKey is usually the DevAddr or DevEUI to make sure duplicate messages arrive at the same NS,
	// or any other identifier (such as an AppID) that helps achieve external consistency for API calls.
	GetPeer(role ttnpb.PeerInfo_Role, tags []string, shardKey []byte) Peer
	// Auth returns a gRPC CallOption that can be used to identify the component within the cluster.
	Auth() grpc.CallOption
	// VerifySource verifies if the caller providing this context is a component from the cluster.
	VerifySource(context.Context) bool
}

// CustomNew allows you to replace the clustering implementation. New will call CustomNew if not nil.
var CustomNew func(ctx context.Context, config *config.ServiceBase, services ...rpcserver.Registerer) (Cluster, error)

// New instantiates a new clustering implementation.
// The basic clustering implementation allows for a cluster setup with a single-instance deployment of each component
// (GS/NS/AS/JS).
// Network operators can use their own clustering logic, which can be activated by setting the CustomNew variable.
func New(ctx context.Context, config *config.ServiceBase, services ...rpcserver.Registerer) (Cluster, error) {
	if CustomNew != nil {
		return CustomNew(ctx, config, services...)
	}

	c := &cluster{
		ctx:   ctx,
		tls:   config.Cluster.TLS,
		peers: make(map[string]*peer),
	}

	for i, key := range config.Cluster.Keys {
		decodedKey, err := hex.DecodeString(key)
		if err != nil {
			return nil, errors.NewWithCause(err, "Could not decode cluster key")
		}
		switch len(decodedKey) {
		case 16, 24, 32:
		default:
			return nil, fmt.Errorf("Invalid length for cluster key number %d: must be 16, 24 or 32 bytes", i)
		}
		c.keys = append(c.keys, decodedKey)
	}
	if c.keys == nil {
		c.keys = [][]byte{random.Bytes(32)}
		log.FromContext(ctx).WithField("key", hex.EncodeToString(c.keys[0])).Warn("No cluster key configured, generated a random one")
	}

	c.self = &peer{
		name:   config.Cluster.Name,
		target: config.Cluster.Address,
	}
	if c.self.name == "" {
		c.self.name, _ = os.Hostname()
	}
	if c.self.target == "" {
		if c.tls {
			c.self.target = config.GRPC.ListenTLS
		} else {
			c.self.target = config.GRPC.Listen
		}
	}
	for _, service := range services {
		if roles := service.Roles(); len(roles) > 0 {
			c.self.roles = append(c.self.roles, roles...)
		}
	}

	c.peers[c.self.name] = c.self

	tryAddPeer := func(name string, target string, role ttnpb.PeerInfo_Role) {
		if !c.self.HasRole(role) && target != "" {
			c.peers[name] = &peer{
				name:   name,
				target: target,
				roles:  []ttnpb.PeerInfo_Role{role},
			}
		}
	}

	tryAddPeer("is", config.Cluster.IdentityServer, ttnpb.PeerInfo_IDENTITY_SERVER)
	tryAddPeer("gs", config.Cluster.GatewayServer, ttnpb.PeerInfo_GATEWAY_SERVER)
	tryAddPeer("ns", config.Cluster.NetworkServer, ttnpb.PeerInfo_NETWORK_SERVER)
	tryAddPeer("as", config.Cluster.ApplicationServer, ttnpb.PeerInfo_APPLICATION_SERVER)
	tryAddPeer("js", config.Cluster.JoinServer, ttnpb.PeerInfo_JOIN_SERVER)

	for _, join := range config.Cluster.Join {
		c.peers[join] = &peer{
			name:   join,
			target: join,
		}
	}

	return c, nil
}

type cluster struct {
	ctx   context.Context
	tls   bool
	peers map[string]*peer
	self  *peer

	keys [][]byte
}

func (c *cluster) Join() (err error) {
	options := rpcclient.DefaultDialOptions(c.ctx)
	// TODO: Use custom WithBalancer DialOption?
	if c.tls {
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(nil))) // TODO: Get *tls.Config from context
	} else {
		options = append(options, grpc.WithInsecure())
	}
	for _, peer := range c.peers {
		peer.ctx, peer.cancel = context.WithCancel(c.ctx)
		log.FromContext(c.ctx).WithFields(log.Fields(
			"target", peer.target,
			"name", peer.Name(),
			"roles", peer.Roles(),
		)).Debug("Connecting to peer...")
		peer.conn, err = grpc.DialContext(peer.ctx, peer.target, options...)
		if err != nil {
			return errors.NewWithCause(err, "Could not connect to peer")
		}
	}
	return nil
}

func (c *cluster) Leave() error {
	for _, peer := range c.peers {
		if peer.conn != nil {
			if err := peer.conn.Close(); err != nil {
				return err
			}
		}
		if peer.cancel != nil {
			peer.cancel()
		}
	}
	return nil
}

func (c *cluster) GetPeers(role ttnpb.PeerInfo_Role, tags []string) []Peer {
	var matches []Peer
	for _, peer := range c.peers {
		if !peer.HasRole(role) {
			continue
		}
		for _, tag := range tags {
			if !peer.HasTag(tag) {
				continue
			}
		}
		if conn := peer.Conn(); conn != nil && conn.GetState() == connectivity.Ready {
			matches = append(matches, peer)
		}
	}
	return matches
}

func (c *cluster) GetPeer(role ttnpb.PeerInfo_Role, tags []string, shardKey []byte) Peer {
	matches := c.GetPeers(role, tags)
	if len(matches) == 1 {
		// TODO: Select the right SubConn for shardKey?
		return matches[0]
	}
	return nil
}
