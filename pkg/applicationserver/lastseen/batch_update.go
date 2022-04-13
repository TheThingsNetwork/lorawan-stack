// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package lastseen

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func batchUpdateEndDeviceLastSeen(ctx context.Context, endDevices map[string]*ttnpb.EndDevice, cls metadata.ClusterPeerAccess) error {
	conn, err := cls.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to get Identity Server peer")
		return err
	}
	cl := ttnpb.NewEndDeviceRegistryClient(conn)
	deviceLastSeenList := make([]*ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate, 0, len(endDevices))
	for _, dev := range endDevices {
		deviceLastSeenList = append(deviceLastSeenList, &ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate{
			Ids:        dev.Ids,
			LastSeenAt: dev.LastSeenAt,
		})
	}
	_, err = cl.BatchUpdateLastSeen(ctx, &ttnpb.BatchUpdateEndDeviceLastSeenRequest{
		Updates: deviceLastSeenList,
	}, cls.WithClusterAuth())
	return err
}
