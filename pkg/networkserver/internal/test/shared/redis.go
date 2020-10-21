// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package test

import (
	"context"

	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var redisNamespace = [...]string{
	"networkserver_test",
}

const (
	redisConsumerGroup = "ns"
	redisConsumerID    = "test"
)

func NewRedisApplicationUplinkQueue(ctx context.Context) (ApplicationUplinkQueue, func()) {
	t := test.MustTBFromContext(ctx)
	cl, flush := test.NewRedis(ctx, append(redisNamespace[:], "application-uplinks")...)
	return redis.NewApplicationUplinkQueue(cl, 100, redisConsumerGroup, redisConsumerID),
		func() {
			flush()
			if err := cl.Close(); err != nil {
				t.Errorf("Failed to close Redis uplink queue client: %s", err)
			}
		}
}

func NewRedisDeviceRegistry(ctx context.Context) (DeviceRegistry, func()) {
	t := test.MustTBFromContext(ctx)
	cl, flush := test.NewRedis(ctx, append(redisNamespace[:], "devices")...)
	reg := &redis.DeviceRegistry{
		Redis:   cl,
		LockTTL: test.Delay << 10,
	}
	if err := reg.Init(ctx); err != nil {
		t.Fatalf("Failed to initialize Redis device registry: %s", err)
	}
	return reg,
		func() {
			flush()
			if err := cl.Close(); err != nil {
				t.Errorf("Failed to close Redis device registry client: %s", err)
			}
		}
}

func NewRedisDownlinkTaskQueue(ctx context.Context) (DownlinkTaskQueue, func()) {
	t, a := test.MustNewTBFromContext(ctx)

	cl, flush := test.NewRedis(ctx, append(redisNamespace[:], "downlink-tasks")...)
	q := redis.NewDownlinkTaskQueue(cl, 10000, redisConsumerGroup, redisConsumerID)
	a.So(q.Init(ctx), should.BeNil)

	ctx, cancel := context.WithCancel(ctx)
	return q,
		func() {
			if err := q.Close(ctx); err != nil {
				t.Errorf("Failed to close Redis downlink task queue: %s", err)
			}
			cancel()
			flush()
			if err := cl.Close(); err != nil {
				t.Errorf("Failed to close Redis downlink task queue client: %s", err)
			}
		}
}

func NewRedisUplinkDeduplicator(ctx context.Context) (UplinkDeduplicator, func()) {
	t := test.MustTBFromContext(ctx)

	cl, flush := test.NewRedis(ctx, append(redisNamespace[:], "uplink-deduplication")...)
	return &redis.UplinkDeduplicator{
			Redis: cl,
		},
		func() {
			flush()
			if err := cl.Close(); err != nil {
				t.Errorf("Failed to close Redis uplink deduplicator client: %s", err)
			}
		}
}
