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

package redis

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/proto"
)

// UplinkDeduplicator is an implementation of networkserver.UplinkDeduplicator.
type UplinkDeduplicator struct {
	Redis *ttnredis.Client
}

// NewUplinkDeduplicator returns a new uplink deduplicator.
func NewUplinkDeduplicator(cl *ttnredis.Client) *UplinkDeduplicator {
	return &UplinkDeduplicator{
		Redis: cl,
	}
}

func uplinkHash(ctx context.Context, up *ttnpb.UplinkMessage, round uint64) (string, error) {
	drBytes, err := proto.Marshal(up.Settings.DataRate)
	if err != nil {
		return "", err
	}
	return ttnredis.Key(
		uplinkPayloadHash(up.RawPayload),
		// NOTE: Data rate and frequency are included in the key to support retransmissions.
		strconv.FormatUint(up.Settings.Frequency, 32),
		keyEncoding.EncodeToString(drBytes),
		strconv.FormatUint(round, 32),
	), nil
}

// DeduplicateUplink deduplicates up for window. Since highest precision allowed by Redis is milliseconds, window is truncated to milliseconds.
func (d *UplinkDeduplicator) DeduplicateUplink(
	ctx context.Context, up *ttnpb.UplinkMessage, window time.Duration, limit int, round uint64,
) (bool, error) {
	h, err := uplinkHash(ctx, up, round)
	if err != nil {
		return false, err
	}
	msgs := make([]proto.Message, 0, len(up.RxMetadata))
	for _, md := range up.RxMetadata {
		msgs = append(msgs, md)
	}
	return ttnredis.DeduplicateProtos(ctx, d.Redis, d.Redis.Key(h), window, limit, msgs...)
}

// AccumulatedMetadata returns accumulated metadata for up.
func (d *UplinkDeduplicator) AccumulatedMetadata(ctx context.Context, up *ttnpb.UplinkMessage, round uint64) ([]*ttnpb.RxMetadata, error) {
	h, err := uplinkHash(ctx, up, round)
	if err != nil {
		return nil, err
	}
	var cmds []ttnredis.ProtosCmd
	if _, err := d.Redis.Pipelined(ctx, func(p redis.Pipeliner) error {
		cmds = []ttnredis.ProtosCmd{
			ttnredis.ListProtos(ctx, p, d.Redis.Key(ttnredis.ListKey(h))),
		}
		return nil
	}); err != nil {
		return nil, err
	}
	var mds []*ttnpb.RxMetadata
	rangeF := func() (proto.Message, func() (bool, error)) {
		md := &ttnpb.RxMetadata{}
		return md, func() (bool, error) {
			mds = append(mds, md)
			return true, nil
		}
	}
	for _, cmd := range cmds {
		if err := cmd.Range(rangeF); err != nil {
			return nil, err
		}
	}
	return mds, nil
}
