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
	"testing"
	"time"

	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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

func NewRedisApplicationUplinkQueue(t testing.TB) (ApplicationUplinkQueue, func()) {
	cl, flush := test.NewRedis(t, append(redisNamespace[:], "application-uplinks")...)
	return redis.NewApplicationUplinkQueue(cl, 100, redisConsumerGroup, redisConsumerID),
		func() {
			flush()
			if err := cl.Close(); err != nil {
				t.Errorf("Failed to close Redis uplink queue client: %s", err)
			}
		}
}

func NewRedisDeviceRegistry(t testing.TB) (DeviceRegistry, func()) {
	cl, flush := test.NewRedis(t, append(redisNamespace[:], "devices")...)
	reg := &redis.DeviceRegistry{
		Redis:   cl,
		LockTTL: test.Delay << 10,
	}
	if err := reg.Init(); err != nil {
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

func NewRedisDownlinkTaskQueue(t testing.TB) (DownlinkTaskQueue, func()) {
	a, ctx := test.New(t)

	cl, flush := test.NewRedis(t, append(redisNamespace[:], "downlink-tasks")...)
	q := redis.NewDownlinkTaskQueue(cl, 10000, redisConsumerGroup, redisConsumerID)
	a.So(q.Init(), should.BeNil)

	ctx, cancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	go func() {
		t.Log("Running Redis downlink task queue...")
		err := q.Run(ctx)
		errCh <- err
		close(errCh)
		t.Logf("Stopped Redis downlink task queue with error: %s", err)
	}()
	return q,
		func() {
			cancel()
			err := q.Add(ctx, ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test"},
			}, time.Now(), false)
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to add mock device to task queue: %s", err)
			}

			var runErr error
			select {
			case <-time.After((1 << 6) * test.Delay):
				t.Error("Timed out waiting for redis.DownlinkTaskQueue.Run to return")
			case runErr = <-errCh:
			}

			flush()
			closeErr := cl.Close()
			if closeErr != nil {
				t.Errorf("Failed to close Redis downlink task queue client: %s", closeErr)
			}
			if runErr != nil && runErr != context.Canceled {
				t.Errorf("Failed to run Redis downlink task queue: %s", runErr)
			}
		}
}

func NewRedisUplinkDeduplicator(t testing.TB) (UplinkDeduplicator, func()) {
	cl, flush := test.NewRedis(t, append(redisNamespace[:], "uplink-deduplication")...)
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
