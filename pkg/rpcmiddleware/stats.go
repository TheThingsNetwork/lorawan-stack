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

package rpcmiddleware

import (
	"context"

	"google.golang.org/grpc/stats"
)

// StatsHandlers is a slice of stats.Handler that implements stats.Handler.
// Calls are delegated to all handlers in order.
type StatsHandlers []stats.Handler

// TagRPC implements stats.Handler.
func (s StatsHandlers) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	for _, hdl := range s {
		ctx = hdl.TagRPC(ctx, info)
	}
	return ctx
}

// HandleRPC implements stats.Handler.
func (s StatsHandlers) HandleRPC(ctx context.Context, stats stats.RPCStats) {
	for _, hdl := range s {
		hdl.HandleRPC(ctx, stats)
	}
}

// TagConn implements stats.Handler.
func (s StatsHandlers) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	for _, hdl := range s {
		ctx = hdl.TagConn(ctx, info)
	}
	return ctx
}

// HandleConn implements stats.Handler.
func (s StatsHandlers) HandleConn(ctx context.Context, stats stats.ConnStats) {
	for _, hdl := range s {
		hdl.HandleConn(ctx, stats)
	}
}
