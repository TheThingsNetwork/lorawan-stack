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
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/basic"
	"go.thethings.network/lorawan-stack/v3/pkg/events/internal/eventstest"
	"go.thethings.network/lorawan-stack/v3/pkg/events/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

type mockComponent struct {
	task.Starter
}

func (mockComponent) FromRequestContext(ctx context.Context) context.Context {
	return ctx
}

// ephemeralPubSub wraps a events.Store but does not implement events.Store.
// This can be used to only test the ephemeral aspects of the PubSub.
type ephemeralPubSub struct {
	ps events.PubSub
}

func (e *ephemeralPubSub) Publish(evts ...events.Event) {
	e.ps.Publish(evts...)
}

func (e *ephemeralPubSub) Subscribe(
	ctx context.Context, names []string, identifiers []*ttnpb.EntityIdentifiers, hdl events.Handler,
) error {
	return e.ps.Subscribe(ctx, names, identifiers, hdl)
}

func TestEphemeralPassthrough(t *testing.T) {
	t.Parallel()

	timeout := (1 << 10) * test.Delay
	events.IncludeCaller = true
	taskStarter := task.StartTaskFunc(task.DefaultStartTask)

	test.RunTest(t, test.TestConfig{
		Timeout: timeout,
		Func: func(ctx context.Context, a *assertions.Assertion) {
			inner := basic.NewPubSub()
			pubsub := mux.New(mockComponent{taskStarter}, inner)
			eventstest.TestBackend(ctx, t, a, &ephemeralPubSub{pubsub})
		},
	})
}
