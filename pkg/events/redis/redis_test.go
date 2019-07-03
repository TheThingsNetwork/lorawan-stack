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

package redis_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/events/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

// redisConfig returns a new redis config for testing
func redisConfig() config.Redis {
	var err error
	config := config.Redis{
		Address:   "localhost:6379",
		Database:  1,
		Namespace: []string{"test"},
	}
	if address := os.Getenv("REDIS_ADDRESS"); address != "" {
		config.Address = address
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		config.Database, err = strconv.Atoi(db)
		if err != nil {
			panic(err)
		}
	}
	if prefix := os.Getenv("REDIS_PREFIX"); prefix != "" {
		config.Namespace = []string{prefix}
	}
	return config
}

func Example() {
	// This sends all events received from Redis to the default pubsub.
	redisPubSub := redis.WrapPubSub(events.DefaultPubSub(), config.Redis{
		// Config here...
	})
	// Replace the default pubsub so that we will now publish to Redis.
	events.SetDefaultPubSub(redisPubSub)
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

	pubsub := redis.NewPubSub(redisConfig())
	defer pubsub.Close()

	pubsub.Subscribe("redis.**", handler)

	ctx := events.ContextWithCorrelationID(test.Context(), t.Name())

	appID := &ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}

	pubsub.Publish(events.New(ctx, "redis.test.evt0", appID, nil))
	select {
	case e := <-eventCh:
		a.So(e.Name(), should.Equal, "redis.test.evt0")
		if a.So(e.Identifiers(), should.NotBeNil) && a.So(e.Identifiers(), should.HaveLength, 1) {
			a.So(e.Identifiers()[0].GetApplicationIDs(), should.Resemble, appID)
		}
	case <-time.After(time.Second):
		t.Error("Did not receive expected event")
		t.FailNow()
	}
}
