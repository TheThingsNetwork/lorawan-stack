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
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"google.golang.org/grpc"
)

type clusterPresence struct{}

var (
	clusterPresenceKey = clusterPresence{}
	// HookName is the name of the hook used to verify the identity of incoming calls within a cluster.
	HookName = "cluster-hook"
)

func (c *cluster) verifySource(ctx context.Context) context.Context {
	md := rpcmetadata.FromIncomingContext(ctx)
	if md.AuthType != "Basic" {
		return context.WithValue(ctx, clusterPresenceKey, false)
	}
	key, err := hex.DecodeString(md.AuthValue)
	if err != nil {
		return context.WithValue(ctx, clusterPresenceKey, false)
	}
	for _, acceptedKey := range c.keys {
		if subtle.ConstantTimeCompare(acceptedKey, key) == 1 {
			return context.WithValue(ctx, clusterPresenceKey, true)
		}
	}
	return context.WithValue(ctx, clusterPresenceKey, false)
}

func (c *cluster) Hook() hooks.UnaryHandlerMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			return next(c.verifySource(ctx), req)
		}
	}
}

// Identified returns true if the caller has been identified as a component from the cluster.
//
// Using Identified requires the hook of the cluster to have been registered.
func Identified(ctx context.Context) bool {
	ok, _ := ctx.Value(clusterPresenceKey).(bool)
	return ok
}

func (c *cluster) Auth() grpc.CallOption {
	md := rpcmetadata.MD{
		ID:        c.self.name,
		AuthType:  "Basic",
		AuthValue: hex.EncodeToString(c.keys[0]),
	}
	return grpc.PerRPCCredentials(md)
}
