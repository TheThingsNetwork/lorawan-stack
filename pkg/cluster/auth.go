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

package cluster

import (
	"context"
	"crypto/subtle"
	"encoding/hex"

	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"google.golang.org/grpc"
)

var (
	// HookName is the name of the hook used to verify the identity of incoming calls within a cluster.
	HookName = "cluster-hook"
	// AuthType used to identify components.
	AuthType = "ClusterKey"
)

func (c *cluster) VerifySource(ctx context.Context) bool {
	md := rpcmetadata.FromIncomingContext(ctx)
	if md.AuthType != AuthType {
		return false
	}
	key, err := hex.DecodeString(md.AuthValue)
	if err != nil {
		return false
	}
	for _, acceptedKey := range c.keys {
		if subtle.ConstantTimeCompare(acceptedKey, key) == 1 {
			return true
		}
	}
	return false
}

func (c *cluster) Auth() grpc.CallOption {
	md := rpcmetadata.MD{
		ID:            c.self.name,
		AuthType:      AuthType,
		AuthValue:     hex.EncodeToString(c.keys[0]),
		AllowInsecure: !c.tls,
	}
	return grpc.PerRPCCredentials(md)
}
