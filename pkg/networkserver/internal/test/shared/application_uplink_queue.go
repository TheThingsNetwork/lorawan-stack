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

package test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// handleApplicationUplinkQueueTest runs a test suite on q.
func handleApplicationUplinkQueueTest(ctx context.Context, q ApplicationUplinkQueue) {
	t, a := test.MustNewTFromContext(ctx)

	appID1 := ttnpb.ApplicationIdentifiers{
		ApplicationID: "application-uplink-queue-app-1",
	}

	appID2 := ttnpb.ApplicationIdentifiers{
		ApplicationID: "application-uplink-queue-app-2",
	}

	pbs := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceID:               "test-dev",
			},
			CorrelationIDs: []string{"correlation-id-1", "correlation-id-2"},
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceID:               "test-dev2",
			},
			CorrelationIDs: []string{"correlation-id-3", "correlation-id-4"},
			Up: &ttnpb.ApplicationUp_LocationSolved{
				LocationSolved: &ttnpb.ApplicationLocation{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID2,
				DeviceID:               "test-dev",
			},
			CorrelationIDs: []string{"correlation-id-5", "correlation-id-6"},
			Up: &ttnpb.ApplicationUp_LocationSolved{
				LocationSolved: &ttnpb.ApplicationLocation{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceID:               "test-dev2",
			},
			CorrelationIDs: []string{"correlation-id-7", "correlation-id-8"},
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID2,
				DeviceID:               "test-dev",
			},
			CorrelationIDs: []string{"correlation-id-9", "correlation-id-10"},
			Up: &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: &ttnpb.ApplicationDownlink{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceID:               "test-dev",
			},
			CorrelationIDs: []string{"correlation-id-11", "correlation-id-12"},
			Up: &ttnpb.ApplicationUp_ServiceData{
				ServiceData: &ttnpb.ApplicationServiceData{},
			},
		},
	}

	assertAdd := func(ctx context.Context, ups ...*ttnpb.ApplicationUp) (error, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()

		errCh := make(chan error, 1)
		go func() {
			errCh <- q.Add(ctx, ups...)
		}()
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for Add to return")
			return nil, false

		case err := <-errCh:
			return err, true
		}
	}

	type subscribeFuncReq struct {
		Context  context.Context
		Uplink   *ttnpb.ApplicationUp
		Response chan<- error
	}
	subscribe := func(ctx context.Context, appID ttnpb.ApplicationIdentifiers) (<-chan subscribeFuncReq, <-chan error, func()) {
		ctx, cancel := context.WithCancel(ctx)
		subscribeFuncCh := make(chan subscribeFuncReq)
		errCh := make(chan error, 1)
		go func() {
			errCh <- q.Subscribe(ctx, appID, func(ctx context.Context, up *ttnpb.ApplicationUp) error {
				respCh := make(chan error, 1)
				subscribeFuncCh <- subscribeFuncReq{
					Context:  ctx,
					Uplink:   up,
					Response: respCh,
				}
				return <-respCh
			})
			close(errCh)
		}()
		return subscribeFuncCh, errCh, func() {
			cancel()
			close(subscribeFuncCh)
		}
	}

	_, app1SubErrCh, app1SubStop := subscribe(ctx, appID1)
	app1SubStop()
	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for Subscribe to return")

	case err := <-app1SubErrCh:
		if !a.So(err, should.Resemble, context.Canceled) {
			t.Fatalf("Received unexpected Subscribe error: %v", err)
		}
	}

	err, ok := assertAdd(ctx, pbs[0:1]...)
	if !a.So(ok, should.BeTrue) || !a.So(err, should.BeNil) {
		t.FailNow()
	}

	app1SubFuncCh, app1SubErrCh, app1SubStop := subscribe(ctx, appID1)
	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for Subscribe callback to be called")

	case err := <-app1SubErrCh:
		t.Fatalf("Received unexpected Subscribe error: %v", err)

	case req := <-app1SubFuncCh:
		if !a.So(req.Context, should.HaveParentContext, ctx) || !a.So(req.Uplink, should.Resemble, pbs[0]) {
			t.FailNow()
		}
		close(req.Response)
	}

	err, ok = assertAdd(ctx, pbs[1:3]...)
	if !a.So(ok, should.BeTrue) {
		t.FailNow()
	}
	a.So(err, should.BeNil)

	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for Subscribe callback to be called")

	case err := <-app1SubErrCh:
		t.Fatalf("Received unexpected Subscribe error: %v", err)

	case req := <-app1SubFuncCh:
		if !a.So(req.Context, should.HaveParentContext, ctx) || !a.So(req.Uplink, should.Resemble, pbs[1]) {
			t.FailNow()
		}
		close(req.Response)
	}

	app2SubFuncCh, app2SubErrCh, app2SubStop := subscribe(ctx, appID2)
	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for Subscribe callback to be called")

	case err := <-app2SubErrCh:
		t.Fatalf("Received unexpected Subscribe error: %v", err)

	case req := <-app2SubFuncCh:
		if !a.So(req.Context, should.HaveParentContext, ctx) || !a.So(req.Uplink, should.Resemble, pbs[2]) {
			t.FailNow()
		}
		close(req.Response)
	}
	app1SubStop()
	app2SubStop()

	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for Subscribe to return")

	case err := <-app1SubErrCh:
		if !a.So(err, should.Resemble, context.Canceled) {
			t.Fatalf("Received unexpected Subscribe error: %v", err)
		}
	}
	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for Subscribe to return")

	case err := <-app2SubErrCh:
		if !a.So(err, should.Resemble, context.Canceled) {
			t.Fatalf("Received unexpected Subscribe error: %v", err)
		}
	}
}

// HandleApplicationUplinkQueueTest runs a ApplicationUplinkQueue test suite on reg.
func HandleApplicationUplinkQueueTest(t *testing.T, q ApplicationUplinkQueue) {
	t.Helper()
	test.RunTest(t, test.TestConfig{
		Parallel: true,
		Func: func(ctx context.Context, a *assertions.Assertion) {
			t.Helper()
			test.RunSubtestFromContext(ctx, test.SubtestConfig{
				Name: "1st run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleApplicationUplinkQueueTest(ctx, q)
				},
			})
			if t.Failed() {
				t.Skip("Skipping 2nd run")
			}
			test.RunSubtestFromContext(ctx, test.SubtestConfig{
				Name: "2st run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleApplicationUplinkQueueTest(ctx, q)
				},
			})
		},
	})
}
