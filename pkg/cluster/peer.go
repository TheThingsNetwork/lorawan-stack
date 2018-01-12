// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package cluster

import (
	"context"

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

type peer struct {
	name  string
	roles []ttnpb.PeerInfo_Role
	tags  []string

	target string

	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
}

func (p *peer) Name() string                 { return p.name }
func (p *peer) Conn() *grpc.ClientConn       { return p.conn }
func (p *peer) Roles() []ttnpb.PeerInfo_Role { return p.roles }
func (p *peer) Tags() []string               { return p.tags }

func (p *peer) HasRole(wanted ttnpb.PeerInfo_Role) bool {
	roles := p.Roles()
	for _, role := range roles {
		if role == wanted {
			return true
		}
	}
	return false
}

func (p *peer) HasTag(wanted string) bool {
	tags := p.Tags()
	for _, tag := range tags {
		if tag == wanted {
			return true
		}
	}
	return false
}
