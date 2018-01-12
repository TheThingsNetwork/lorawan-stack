// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package cluster

import (
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Peer interface
type Peer interface {
	// Name of the peer
	Name() string
	// gRPC ClientConn to the peer (if available)
	Conn() *grpc.ClientConn
	// Roles announced by the peer
	Roles() []ttnpb.PeerInfo_Role
	// HasRole returns true iff the peer has the given role
	HasRole(ttnpb.PeerInfo_Role) bool
	// Tags announced by the peer
	Tags() []string
	// HasTag returns true iff the peer has the given tag
	HasTag(string) bool
}
