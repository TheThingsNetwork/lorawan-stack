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

package cluster

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Peer interface
type Peer interface {
	// Name of the peer
	Name() string
	// gRPC ClientConn to the peer (if available)
	Conn() (*grpc.ClientConn, error)
	// Roles announced by the peer
	Roles() []ttnpb.ClusterRole
	// HasRole returns true iff the peer has the given role
	HasRole(ttnpb.ClusterRole) bool
	// Tags announced by the peer
	Tags() map[string]string
}

type peer struct {
	name  string
	roles []ttnpb.ClusterRole
	tags  map[string]string

	target string

	ctx     context.Context
	cancel  context.CancelFunc
	conn    *grpc.ClientConn
	connErr error
}

func (p *peer) Name() string                    { return p.name }
func (p *peer) Conn() (*grpc.ClientConn, error) { return p.conn, p.connErr }
func (p *peer) Roles() []ttnpb.ClusterRole      { return p.roles }
func (p *peer) Tags() map[string]string         { return p.tags }

func (p *peer) HasRole(wanted ttnpb.ClusterRole) bool {
	roles := p.Roles()
	for _, role := range roles {
		if role == wanted {
			return true
		}
	}
	return false
}
