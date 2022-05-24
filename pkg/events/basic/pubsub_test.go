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

package basic_test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/basic"
	"go.thethings.network/lorawan-stack/v3/pkg/events/internal/eventstest"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func Example() {
	basicPubSub := basic.NewPubSub()

	// Replace the default pubsub so that we will now publish to this pub sub by default.
	events.SetDefaultPubSub(basicPubSub)
}

var timeout = (1 << 10) * test.Delay

func TestPubSub(t *testing.T) { //nolint:paralleltest
	events.IncludeCaller = true

	test.RunTest(t, test.TestConfig{
		Timeout: timeout,
		Func: func(ctx context.Context, a *assertions.Assertion) {
			pubsub := basic.NewPubSub()
			eventstest.TestBackend(ctx, t, a, pubsub)
		},
	})
}
