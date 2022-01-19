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

package gatewayserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// GetGatewayConnectionStats returns statistics about a gateway connection.
func (gs *GatewayServer) GetGatewayConnectionStats(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayConnectionStats, error) {
	if err := gs.entityRegistry.AssertGatewayRights(ctx, *ids, ttnpb.Right_RIGHT_GATEWAY_STATUS_READ); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, ids)
	if gs.statsRegistry != nil {
		stats, err := gs.statsRegistry.Get(ctx, *ids)
		if err != nil || stats == nil {
			if errors.IsNotFound(err) {
				return nil, errNotConnected.WithAttributes("gateway_uid", uid).WithCause(err)
			}
			return nil, err
		}

		return stats, nil
	}

	val, ok := gs.connections.Load(uid)
	if !ok {
		return nil, errNotConnected.WithAttributes("gateway_uid", uid)
	}
	stats, _ := val.(connectionEntry).Stats()
	return stats, nil
}
