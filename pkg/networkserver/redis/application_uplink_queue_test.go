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

package redis_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	nsredis "go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var _ networkserver.ApplicationUplinkQueue = &nsredis.ApplicationUplinkQueue{}

var (
	redisNamespace = [...]string{
		"redis_test_uplink_queue",
	}
	readLimit        = 7
	maxLen           = int64(100)
	groupID          = "ns-test"
	minIdle          = (1 << 8) * test.Delay
	streamBlockLimit = (1 << 5) * test.Delay

	appCount       = 5
	devCountPerApp = 3
)

func setupRedusApplicationUplinkQueue(t *testing.T, cl *ttnredis.Client) (*nsredis.ApplicationUplinkQueue, func()) {
	t.Helper()

	_, ctx := test.New(t)

	q := nsredis.NewApplicationUplinkQueue(cl, maxLen, groupID, minIdle, streamBlockLimit)

	return q, func() {
		if err := q.Close(ctx); err != nil {
			t.Errorf("Failed to close Redis application uplink queue: %s", err)
		}
	}
}

func TestApplicationUplinkQueueInitCreatesConsumerGroup(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	cl, redisCloseFn := test.NewRedis(ctx, append(redisNamespace[:], "init")...)
	t.Cleanup(redisCloseFn)

	q, qCloseFn := setupRedusApplicationUplinkQueue(t, cl)
	t.Cleanup(qCloseFn)

	if !a.So(q.Init(ctx), should.BeNil) {
		t.FailNow()
	}

	streamID := cl.Key("uplinks")
	groups, err := cl.XInfoGroups(ctx, streamID).Result()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(groups, should.HaveLength, 1)
	a.So(groups[0].Name, should.Equal, groupID)
	a.So(groups[0].Consumers, should.Equal, 0)
}

func TestApplicationUplinkQueueCloseRemovesAllConsumers(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	cl, redisCloseFn := test.NewRedis(ctx, append(redisNamespace[:], "close")...)
	t.Cleanup(redisCloseFn)

	q, _ := setupRedusApplicationUplinkQueue(t, cl)

	if !a.So(q.Init(ctx), should.BeNil) {
		t.FailNow()
	}

	consumerIDs := []string{"test-consumer-1", "test-consumer-2"}
	up := &ttnpb.ApplicationUp{
		EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
			DeviceId: "test-device",
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: "test-application",
			},
		},
	}
	for _, consumerID := range consumerIDs {
		if err := q.Add(ctx, up); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		if err := q.Pop(ctx, consumerID, 1, func(ctx context.Context, ups []*ttnpb.ApplicationUp) error {
			return nil
		}); !a.So(err, should.BeNil) {
			t.FailNow()
		}
	}

	streamID := cl.Key("uplinks")
	consumers, err := cl.XInfoConsumers(ctx, streamID, groupID).Result()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(consumers, should.HaveLength, 2)

	if !a.So(q.Close(ctx), should.BeNil) {
		t.FailNow()
	}

	consumers, err = cl.XInfoConsumers(ctx, streamID, groupID).Result()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(consumers, should.HaveLength, 0)
}

func generateRandomUplinks(t *testing.T, applicationCount, deviceCount int) []*ttnpb.ApplicationUp {
	t.Helper()

	ups := make([]*ttnpb.ApplicationUp, 0, applicationCount*deviceCount)
	for i := 0; i < applicationCount; i++ {
		applicationID := fmt.Sprintf("test-application-%d", i)
		for j := 0; j < deviceCount; j++ {
			deviceID := fmt.Sprintf("test-device-%d", j)
			ups = append(ups, &ttnpb.ApplicationUp{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: applicationID,
					},
					DeviceId: deviceID,
				},
			})
		}
	}
	return ups
}

func assertAllEqualAppIDs(t *testing.T, ups []*ttnpb.ApplicationUp) {
	t.Helper()

	a := assertions.New(t)
	if !a.So(ups, should.NotBeEmpty) {
		t.FailNow()
	}

	expected := ups[0].EndDeviceIds.ApplicationIds
	for _, up := range ups[1:] {
		actual := up.EndDeviceIds.ApplicationIds
		if !a.So(actual, should.Resemble, expected) {
			t.FailNow()
		}
	}
}

func assertStreamUplinkCount(t *testing.T, cl *ttnredis.Client, expected int) {
	t.Helper()

	a, ctx := test.New(t)
	streamID := cl.Key("uplinks")
	entries, err := cl.XRange(ctx, streamID, "-", "+").Result()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(entries, should.HaveLength, expected)
}

func TestApplicationUplinkQueueAdd(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	cl, redisCloseFn := test.NewRedis(ctx, append(redisNamespace[:], "add")...)
	t.Cleanup(redisCloseFn)

	q, qCloseFn := setupRedusApplicationUplinkQueue(t, cl)
	t.Cleanup(qCloseFn)

	if !a.So(q.Init(ctx), should.BeNil) {
		t.FailNow()
	}

	expectedUps := generateRandomUplinks(t, appCount, devCountPerApp)
	expectedUIDs := make([]string, 0, len(expectedUps))
	for _, up := range expectedUps {
		expectedUIDs = append(expectedUIDs, unique.ID(ctx, up.EndDeviceIds))
	}

	if err := q.Add(ctx, expectedUps...); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	streamID := cl.Key("uplinks")
	entries, err := cl.XRange(ctx, streamID, "-", "+").Result()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if !a.So(entries, should.HaveLength, len(expectedUps)) {
		t.FailNow()
	}

	actualUIDs := make([]string, 0, len(entries))
	actualUps := make([]*ttnpb.ApplicationUp, 0, 4)
	for _, entry := range entries {
		a.So(entry.Values, should.HaveLength, 2)
		a.So(entry.Values["payload"], should.NotBeEmpty)
		a.So(entry.Values["uid"], should.NotBeEmpty)
		actualUIDs = append(actualUIDs, entry.Values["uid"].(string))
		up := &ttnpb.ApplicationUp{}
		if err := ttnredis.UnmarshalProto(entry.Values["payload"].(string), up); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		actualUps = append(actualUps, up)
	}
	a.So(actualUIDs, should.HaveSameElementsDeep, expectedUIDs)
	a.So(actualUps, should.HaveSameElementsDeep, expectedUps)
}

func TestApplicationUplinkQueuePopAll(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	cl, redisCloseFn := test.NewRedis(ctx, append(redisNamespace[:], "pop_all")...)
	t.Cleanup(redisCloseFn)

	q, qCloseFn := setupRedusApplicationUplinkQueue(t, cl)
	t.Cleanup(qCloseFn)

	if !a.So(q.Init(ctx), should.BeNil) {
		t.FailNow()
	}

	consumerCount := 3
	uplinkCh := make(chan []*ttnpb.ApplicationUp, appCount)
	errCh := make(chan error, consumerCount)
	wg := sync.WaitGroup{}

	for i := 0; i < consumerCount; i++ {
		consumerID := fmt.Sprintf("test-consumer-%d", i)
		wg.Add(1)
		go func() {
			defer wg.Done()

			errCh <- q.Pop(ctx, consumerID, readLimit, func(ctx context.Context, ups []*ttnpb.ApplicationUp) error {
				assertAllEqualAppIDs(t, ups)
				uplinkCh <- ups
				return nil
			})
		}()
	}

	expected := generateRandomUplinks(t, appCount, devCountPerApp)
	actual := make([]*ttnpb.ApplicationUp, 0, len(expected))
	var err error

	go func() {
		for {
			select {
			case ups := <-uplinkCh:
				actual = append(actual, ups...)
			case <-ctx.Done():
				errCh <- ctx.Err()
			}
		}
	}()

	go func() {
		for {
			select {
			case err = <-errCh:
				return
			case <-ctx.Done():
				errCh <- ctx.Err()
			}
		}
	}()

	if err := q.Add(ctx, expected...); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	wg.Wait()

	a.So(err, should.BeNil)
	a.So(actual, should.HaveLength, len(expected))
	assertStreamUplinkCount(t, cl, 0)
}

func TestApplicationUplinkQueuePopErr(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	cl, redisCloseFn := test.NewRedis(ctx, append(redisNamespace[:], "pop_err")...)
	t.Cleanup(redisCloseFn)

	q, qCloseFn := setupRedusApplicationUplinkQueue(t, cl)
	t.Cleanup(qCloseFn)

	if !a.So(q.Init(ctx), should.BeNil) {
		t.FailNow()
	}

	generateError := func(ups []*ttnpb.ApplicationUp) error {
		appID := ups[0].EndDeviceIds.ApplicationIds.ApplicationId
		if appID == "test-application-1" || appID == "test-application-2" {
			return fmt.Errorf("test error")
		}
		return nil
	}

	consumerCount := 3
	uplinkCh := make(chan []*ttnpb.ApplicationUp, appCount)
	errCh := make(chan error, consumerCount)
	wg := sync.WaitGroup{}

	for i := 0; i < consumerCount; i++ {
		consumerID := fmt.Sprintf("test-consumer-%d", i)
		wg.Add(1)
		go func() {
			defer wg.Done()

			errCh <- q.Pop(ctx, consumerID, readLimit, func(ctx context.Context, ups []*ttnpb.ApplicationUp) error {
				assertAllEqualAppIDs(t, ups)
				uplinkCh <- ups
				return generateError(ups)
			})
		}()
	}

	expected := generateRandomUplinks(t, appCount, devCountPerApp)
	actual := make([]*ttnpb.ApplicationUp, 0, len(expected))
	var err error

	go func() {
		for {
			select {
			case ups := <-uplinkCh:
				actual = append(actual, ups...)
			case <-ctx.Done():
				errCh <- ctx.Err()
			}
		}
	}()

	go func() {
		for {
			select {
			case err = <-errCh:
				return
			case <-ctx.Done():
				errCh <- ctx.Err()
			}
		}
	}()

	if err := q.Add(ctx, expected...); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	wg.Wait()

	expectedFailCount := devCountPerApp * 2

	a.So(err, should.BeNil)
	a.So(actual, should.HaveLength, len(expected))    // All uplinks should have been processed
	assertStreamUplinkCount(t, cl, expectedFailCount) // Only failed uplinks should remain in the stream
}
