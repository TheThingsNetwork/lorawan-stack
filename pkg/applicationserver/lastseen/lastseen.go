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
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LastSeenProvider is an interface for storing device last seen timestamp from uplink.
type LastSeenProvider interface {
	PushLastSeenFromUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, lastSeen *timestamppb.Timestamp) error
}

type noopLastSeenProvider struct{}

func (noopLastSeenProvider) PushLastSeenFromUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, lastSeen *timestamppb.Timestamp) error {
	return nil
}

// NewNoopLastSeenProvider returns a noop LastSeenProvider.
func NewNoopLastSeenProvider() (LastSeenProvider, error) {
	return noopLastSeenProvider{}, nil
}

// BatchLastSeenProvider represents
type BatchLastSeenProvider struct {
	ctx     context.Context
	cluster metadata.ClusterPeerAccess
	ticker  <-chan time.Time

	batchSize   int
	lastSeenMap map[string]time.Time
	lsMu        sync.Mutex
}

func (ls *BatchLastSeenProvider) push(uid string, lastSeen *timestamppb.Timestamp) (map[string]*ttnpb.EndDevice, error) {
	ls.lsMu.Lock()
	defer ls.lsMu.Unlock()
	currentTimestamp := ttnpb.StdTime(lastSeen)
	if lastSeen, ok := ls.lastSeenMap[uid]; ok {
		if lastSeen.Before(*currentTimestamp) {
			ls.lastSeenMap[uid] = *currentTimestamp
		}
		return nil, nil
	}
	if len(ls.lastSeenMap) == ls.batchSize {
		devs, err := ls.clear()
		if err != nil {
			return nil, err
		}
		ls.lastSeenMap[uid] = *currentTimestamp
		return devs, nil
	}
	ls.lastSeenMap[uid] = *currentTimestamp
	return nil, nil
}

// PushLastSeenFromUplink pushes the timestamp of the device uplink to the last seen provider.
// If the data structure for storing last seen timestamps is full, the timestamps
// are updated in batch in Identity Server.
func (ls *BatchLastSeenProvider) PushLastSeenFromUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, lastSeen *timestamppb.Timestamp) error {
	if ls.batchSize == 0 {
		if err := ls.updateEndDeviceLastSeen(ctx, &ttnpb.EndDevice{
			Ids:        ids,
			LastSeenAt: lastSeen,
		}); err != nil {
			return err
		}
	}
	endDevices, err := ls.push(unique.ID(ctx, ids), lastSeen)
	if err != nil {
		return err
	}
	if len(endDevices) == 0 {
		return nil
	}
	return batchUpdateEndDeviceLastSeen(ctx, endDevices, ls.cluster)
}

// clear returns an array of end devices with timestamps from the last seen map.
// This method is not safe for concurrent use.
func (ls *BatchLastSeenProvider) clear() (map[string]*ttnpb.EndDevice, error) {
	if len(ls.lastSeenMap) == 0 {
		return nil, nil
	}
	endDevices := make(map[string]*ttnpb.EndDevice)
	for devUID, timestamp := range ls.lastSeenMap {
		ids, err := unique.ToDeviceID(devUID)
		if err != nil {
			return nil, err
		}
		timestampProto := ttnpb.ProtoTime(&timestamp)
		endDevices[devUID] = &ttnpb.EndDevice{
			Ids:        ids,
			LastSeenAt: timestampProto,
		}
	}
	// Clear map.
	ls.lastSeenMap = make(map[string]time.Time)

	return endDevices, nil
}

func (ls *BatchLastSeenProvider) flush(ctx context.Context) error {
	ls.lsMu.Lock()
	endDevices, err := ls.clear()
	ls.lsMu.Unlock()
	if err != nil {
		return err
	}
	if len(endDevices) == 0 {
		return nil
	}
	return batchUpdateEndDeviceLastSeen(ctx, endDevices, ls.cluster)
}

func (ls *BatchLastSeenProvider) updateEndDeviceLastSeen(ctx context.Context, endDevice *ttnpb.EndDevice) error {
	conn, err := ls.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		log.FromContext(ls.ctx).WithError(err).Warn("Failed to get Identity Server peer")
		return err
	}
	cl := ttnpb.NewEndDeviceRegistryClient(conn)
	_, err = cl.Update(ctx, &ttnpb.UpdateEndDeviceRequest{
		EndDevice: endDevice,
		FieldMask: ttnpb.FieldMask("last_seen_at"),
	}, ls.cluster.WithClusterAuth())
	return err
}

// NewBatchLastSeenProvider returns a new BatchLastSeenProvider struct.
func NewBatchLastSeenProvider(ctx context.Context, batchSize int, ticker <-chan time.Time, cluster metadata.ClusterPeerAccess) *BatchLastSeenProvider {
	lsMap := make(map[string]time.Time)
	return &BatchLastSeenProvider{
		ctx:         ctx,
		ticker:      ticker,
		batchSize:   batchSize,
		lastSeenMap: lsMap,
		cluster:     cluster,
	}
}

// NewBatchLastSeen creates a new BatchLastSeenProvider that manages batch updates of last seen timestamps in Identity Server.
func NewBatchLastSeen(ctx context.Context, batchSize int, ticker <-chan time.Time, cluster metadata.ClusterPeerAccess) (LastSeenProvider, error) {
	lastSeenProv := NewBatchLastSeenProvider(ctx, batchSize, ticker, cluster)
	if batchSize <= 0 {
		return lastSeenProv, nil
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				if err := lastSeenProv.flush(ctx); err != nil {
					log.FromContext(lastSeenProv.ctx).WithError(err).Error("Failed to flush end device last seen timestamps")
				}
			}
		}
	}()
	return lastSeenProv, nil
}
