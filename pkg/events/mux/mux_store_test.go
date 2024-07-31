// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package mux_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/basic"
	"go.thethings.network/lorawan-stack/v3/pkg/events/internal/eventstest"
	"go.thethings.network/lorawan-stack/v3/pkg/events/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/events/redis"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func redisConfig(extraNamespaces ...string) ttnredis.Config {
	var err error
	conf := ttnredis.Config{
		Address:       "localhost:6379",
		Database:      1,
		RootNamespace: []string{"test"},
	}
	if address := os.Getenv("REDIS_ADDRESS"); address != "" {
		conf.Address = address
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		conf.Database, err = strconv.Atoi(db)
		if err != nil {
			panic(err)
		}
	}
	if prefix := os.Getenv("REDIS_PREFIX"); prefix != "" {
		conf.RootNamespace = []string{prefix}
	}
	conf.RootNamespace = append(conf.RootNamespace, extraNamespaces...)
	return conf
}

func TestRedisPassthrough(t *testing.T) {
	t.Parallel()

	timeout := (1 << 11) * test.Delay
	events.IncludeCaller = true
	taskStarter := task.StartTaskFunc(task.DefaultStartTask)

	test.RunTest(t, test.TestConfig{
		Timeout: timeout,
		Func: func(ctx context.Context, a *assertions.Assertion) {
			conf := config.RedisEvents{
				Config: redisConfig("mux"),
			}
			conf.Store.Enable = true
			batchConf := config.BatchEvents{Enable: true}
			inner := redis.NewPubSub(ctx, mockComponent{taskStarter}, conf, batchConf)
			defer inner.(*redis.PubSubStore).Close(ctx)

			time.Sleep(timeout / 10)

			pubsub := mux.New(mockComponent{taskStarter}, inner)
			eventstest.TestBackend(ctx, t, a, pubsub)
		},
	})
}

func TestMultiplexing(t *testing.T) {
	t.Parallel()

	timeout := (1 << 11) * test.Delay
	events.IncludeCaller = true
	taskStarter := task.StartTaskFunc(task.DefaultStartTask)

	test.RunTest(t, test.TestConfig{
		Timeout: timeout,
		Func: func(ctx context.Context, a *assertions.Assertion) {
			eui := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
			gtwID := ttnpb.GatewayIdentifiers{
				GatewayId: "test-gtw",
				Eui:       eui.Bytes(),
			}

			// A is the main PubSub.
			// A and B are Redis PubSubs which implement events.Store.
			// C and D are basic PubSubs which do not implement events.Store.

			confA := config.RedisEvents{
				Config: redisConfig("mux", "a"),
			}
			confA.Store.Enable = true
			batchConfA := config.BatchEvents{Enable: true}
			innerA := redis.NewPubSub(ctx, mockComponent{taskStarter}, confA, batchConfA)
			defer innerA.(*redis.PubSubStore).Close(ctx)

			confB := config.RedisEvents{
				Config: redisConfig("mux", "b"),
			}
			confB.Store.Enable = true
			batchConfB := config.BatchEvents{Enable: true}
			innerB := redis.NewPubSub(ctx, mockComponent{taskStarter}, confB, batchConfB)
			defer innerB.(*redis.PubSubStore).Close(ctx)

			innerC := basic.NewPubSub()
			innerD := basic.NewPubSub()

			// A and B are matched by all events.
			// C is matched by test.some.evt2.
			// D is ignored.
			pubsub := mux.New(
				mockComponent{taskStarter},
				innerA,
				mux.WithStream(innerB, mux.MatchAll),
				mux.WithStream(innerC, mux.MatchNames("test.some.evt2")),
				mux.WithStream(innerD, mux.MatchNone),
			)
			pubsubStore := pubsub.(events.Store) //nolint:revive

			resCh := make(events.Channel, 10)
			a.So(pubsub.Subscribe(ctx, []string{"test.some.evt1"}, nil, resCh), should.BeNil)
			a.So(pubsub.Subscribe(ctx, []string{"test.some.evt2"}, nil, resCh), should.BeNil)

			// Publish to A, B, C and D; the message from C and D get ignored.
			innerA.Publish(events.New(ctx, "test.some.evt1", "test event 1",
				events.WithData("A"), events.WithIdentifiers(&gtwID)),
			)
			innerB.Publish(events.New(ctx, "test.some.evt1", "test event 1",
				events.WithData("B"), events.WithIdentifiers(&gtwID)),
			)
			innerC.Publish(events.New(ctx, "test.some.evt1", "test event 1",
				events.WithData("C"), events.WithIdentifiers(&gtwID)),
			)
			innerD.Publish(events.New(ctx, "test.some.evt1", "test event 1",
				events.WithData("D"), events.WithIdentifiers(&gtwID)),
			)

			time.Sleep(timeout / 10)

			// Check that the events from A and B were received.
			if a.So(resCh, should.HaveLength, 2) {
				evt1, evt2 := <-resCh, <-resCh
				hasA, hasB, hasC, hasD := false, false, false, false
				for _, evt := range []events.Event{evt1, evt2} {
					hasA = hasA || evt.Data() == "A"
					hasB = hasB || evt.Data() == "B"
					hasC = hasC || evt.Data() == "C"
					hasD = hasD || evt.Data() == "D"
					a.So(evt.Name(), should.Equal, "test.some.evt1")
				}
				a.So(hasA, should.BeTrue)
				a.So(hasB, should.BeTrue)
				a.So(hasC, should.BeFalse)
				a.So(hasD, should.BeFalse)
			}

			// Publish to C.
			innerC.Publish(events.New(ctx, "test.some.evt2", "test event 2"))

			time.Sleep(timeout / 10)

			// Check that this event was matched.
			if a.So(resCh, should.HaveLength, 1) {
				a.So((<-resCh).Name(), should.Equal, "test.some.evt2")
			}

			// Only A and B are stores, so they return historical events.
			evts, err := pubsubStore.FetchHistory(ctx, nil, []*ttnpb.EntityIdentifiers{gtwID.GetEntityIdentifiers()}, nil, 1)
			a.So(err, should.BeNil)
			if a.So(evts, should.HaveLength, 2) {
				a.So(evts[0].Data(), should.Equal, "A")
				a.So(evts[1].Data(), should.Equal, "B")
			}
		},
	})
}
