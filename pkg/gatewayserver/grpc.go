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

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// GetGatewayConnectionStats returns statistics about a gateway connection.
func (gs *GatewayServer) GetGatewayConnectionStats(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayConnectionStats, error) {
	if err := rights.RequireGateway(ctx, *ids, ttnpb.RIGHT_GATEWAY_STATUS_READ); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, ids)
	val, ok := gs.connections.Load(uid)
	if !ok {
		return nil, errNotConnected.WithAttributes("gateway_uid", uid)
	}
	conn := val.(*io.Connection)

	stats := &ttnpb.GatewayConnectionStats{}
	ct := conn.ConnectTime()
	stats.ConnectedAt = &ct
	stats.Protocol = conn.Frontend().Protocol()
	if s, t, ok := conn.StatusStats(); ok {
		stats.LastStatusReceivedAt = &t
		stats.LastStatus = s
	}
	if c, t, ok := conn.UpStats(); ok {
		stats.LastUplinkReceivedAt = &t
		stats.UplinkCount = c
	}
	if c, t, ok := conn.DownStats(); ok {
		stats.LastDownlinkReceivedAt = &t
		stats.DownlinkCount = c
	}
	if min, max, median, count := conn.RTTStats(); count > 0 {
		stats.RoundTripTimes = &ttnpb.GatewayConnectionStats_RoundTripTimes{
			Min:    min,
			Max:    max,
			Median: median,
			Count:  uint32(count),
		}
	}
	return stats, nil
}
