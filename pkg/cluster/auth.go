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
	"encoding/hex"

	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"google.golang.org/grpc"
)

// HookName is the name of the hook used to verify the identity of incoming calls within a cluster.
var HookName = "cluster-hook"

func (c *cluster) TLS() bool { return c.tls }

func (c *cluster) WithVerifiedSource(ctx context.Context) context.Context {
	return clusterauth.VerifySource(ctx, c.keys)
}

func (c *cluster) Auth() grpc.CallOption {
	md := rpcmetadata.MD{
		ID:            c.self.name,
		AuthType:      clusterauth.AuthType,
		AuthValue:     hex.EncodeToString(c.keys[0]),
		AllowInsecure: !c.tls,
	}
	return grpc.PerRPCCredentials(md)
}
