// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package cloud_test

import (
	"context"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/cloud"
	"go.thethings.network/lorawan-stack/v3/pkg/events/internal/eventstest"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	_ "gocloud.dev/pubsub/mempubsub"
)

func Example() {
	// The task starter is used for automatic re-subscription on failure.
	taskStarter := task.StartTaskFunc(task.DefaultStartTask)

	// Import the desired cloud pub-sub drivers (see godoc.org/gocloud.dev).
	// In this example we use "gocloud.dev/pubsub/mempubsub".

	cloudPubSub, err := cloud.NewPubSub(context.TODO(), taskStarter, "mem://events", "mem://events")
	if err != nil {
		panic(err)
	}

	// Replace the default pubsub so that we will now publish to a Go Cloud pub sub.
	events.SetDefaultPubSub(cloudPubSub)
}

var timeout = (1 << 10) * test.Delay

func TestCloudPubSub(t *testing.T) { //nolint:paralleltest
	events.IncludeCaller = true

	taskStarter := task.StartTaskFunc(task.DefaultStartTask)

	test.RunSubtest(t, test.SubtestConfig{
		Name:    "json",
		Timeout: 10 * timeout,
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()
			pubsub, err := cloud.NewPubSub(ctx, taskStarter, "mem://json_events_test", "mem://json_events_test")
			a.So(err, should.BeNil)
			defer pubsub.Close(ctx)
			cloud.SetContentType(pubsub, "application/json")
			eventstest.TestBackend(ctx, t, a, pubsub)
		},
	})

	test.RunSubtest(t, test.SubtestConfig{
		Name:    "protobuf",
		Timeout: 10 * timeout,
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()
			pubsub, err := cloud.NewPubSub(ctx, taskStarter, "mem://protobuf_events_test", "mem://protobuf_events_test")
			a.So(err, should.BeNil)
			defer pubsub.Close(ctx)
			cloud.SetContentType(pubsub, "application/protobuf")
			eventstest.TestBackend(ctx, t, a, pubsub)
		},
	})
}
