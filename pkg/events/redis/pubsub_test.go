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

package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/events"
	"github.com/TheThingsNetwork/ttn/pkg/events/redis"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func Example() {
	// This sends all events received from Redis to the default pubsub.
	redisPubSub, err := redis.WrapPubSub(events.DefaultPubSub, config.Redis{
		// Config here...
	})
	if err != nil {
		// Handle error...
	}
	// Replace the default pubsub so that we will now publish to Redis.
	events.DefaultPubSub = redisPubSub
}

func TestRedisPubSub(t *testing.T) {
	a := assertions.New(t)

	events.IncludeCaller = true

	var eventCh = make(chan events.Event)
	handler := events.HandlerFunc(func(e events.Event) {
		t.Logf("Received event %v", e)
		a.So(e.Time().IsZero(), should.BeFalse)
		a.So(e.Context(), should.NotBeNil)
		eventCh <- e
	})

	pubsub, err := redis.NewPubSub(test.RedisConfig())
	a.So(err, should.BeNil)
	defer pubsub.Close()

	pubsub.Subscribe("redis.**", handler)

	ctx := events.ContextWithCorrelationID(context.Background(), t.Name())

	pubsub.Publish(events.New(ctx, "redis.test.evt0", nil, nil))
	select {
	case e := <-eventCh:
		a.So(e.Name(), should.Equal, "redis.test.evt0")
	case <-time.After(time.Second):
		t.Error("Did not receive expected event")
		t.FailNow()
	}
}
