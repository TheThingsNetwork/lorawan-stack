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

package connector

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth/rights"
	"github.com/TheThingsNetwork/ttn/pkg/cluster"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Component that can retrieve a peer given its role.
type Component interface {
	GetPeer(peerInfo ttnpb.PeerInfo_Role, tags []string, shardKey []byte) cluster.Peer
}

type componentConnector struct {
	Component
}

func (c componentConnector) Get(context.Context) *grpc.ClientConn {
	peer := c.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, nil, nil)
	if peer == nil {
		return nil
	}

	return peer.Conn()
}

// FromComponent returns an IdentityServerConnector tied to a component.
// The tags and shard key are then used to select the identity server to use.
func FromComponent(c Component) rights.IdentityServerConnector {
	return componentConnector{
		Component: c,
	}
}
