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
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// handleApplicationUplinkQueueTest runs a test suite on q.
func handleApplicationUplinkQueueTest(ctx context.Context, q ApplicationUplinkQueue) {
	assertAdd := func(ctx context.Context, ups ...*ttnpb.ApplicationUp) bool {
		t, a := test.MustNewTFromContext(ctx)
		t.Helper()

		errCh := make(chan error, 1)
		go func() {
			errCh <- q.Add(ctx, ups...)
		}()
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for Add to return")
			return false

		case err := <-errCh:
			return a.So(err, should.BeNil)
		}
	}

	type popFuncReq struct {
		Context                context.Context
		ApplicationIdentifiers ttnpb.ApplicationIdentifiers
		Func                   ApplicationUplinkQueueRangeFunc
		Response               chan<- TaskPopFuncResponse
	}
	pop := func(ctx context.Context, reqCh chan<- popFuncReq, errCh chan<- error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()

		go func() {
			errCh <- q.Pop(ctx, func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f ApplicationUplinkQueueRangeFunc) (time.Time, error) {
				respCh := make(chan TaskPopFuncResponse, 1)
				reqCh <- popFuncReq{
					Context:                ctx,
					ApplicationIdentifiers: appID,
					Func:                   f,
					Response:               respCh,
				}
				resp := <-respCh
				return resp.Time, resp.Error
			})
		}()
		return true
	}
	assertPopUplinks := func(ctx context.Context, popReqCh <-chan popFuncReq, popErrCh <-chan error, expected ...*ttnpb.ApplicationUp) bool {
		t, a := test.MustNewTFromContext(ctx)
		t.Helper()

		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for Pop callback to be called")
			return false

		case err := <-popErrCh:
			t.Errorf("Received unexpected Pop error: %v", err)
			return false

		case req := <-popReqCh:
			if !test.AllTrue(
				a.So(req.Context, should.HaveParentContextOrEqual, ctx),
				a.So(req.ApplicationIdentifiers, should.Resemble, expected[0].ApplicationIdentifiers),
			) {
				t.Error("Pop callback assertion failed")
				return false
			}
			var called bool
			ok, err := req.Func(len(expected)+1, func(ups ...*ttnpb.ApplicationUp) error {
				a.So(called, should.BeFalse)
				a.So(ups, should.Resemble, expected)
				called = true
				return nil
			})
			if !test.AllTrue(
				a.So(called, should.BeTrue),
				a.So(ok, should.BeTrue),
				a.So(err, should.BeNil),
			) {
				return false
			}
			close(req.Response)

			select {
			case <-ctx.Done():
				t.Error("Timed out while waiting for Pop to return")
				return false

			case err := <-popErrCh:
				return a.So(err, should.BeNil)
			}
		}
	}

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
				ApplicationIdentifiers: appID2,
				DeviceID:               "test-dev2",
			},
			CorrelationIDs: []string{"correlation-id-7", "correlation-id-8"},
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
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

	popReqCh := make(chan popFuncReq, 1)
	popErrCh := make(chan error, 1)
	_, a := test.MustNewTFromContext(ctx)
	switch {
	case !pop(ctx, popReqCh, popErrCh),
		!a.So(assertAdd(ctx, pbs[0]), should.BeTrue),
		!a.So(assertPopUplinks(ctx, popReqCh, popErrCh, pbs[0]), should.BeTrue),
		!pop(ctx, popReqCh, popErrCh),
		!a.So(assertAdd(ctx, pbs[1]), should.BeTrue),
		!a.So(assertAdd(ctx, pbs[2]), should.BeTrue),
		!a.So(assertPopUplinks(ctx, popReqCh, popErrCh, pbs[1]), should.BeTrue),
		!pop(ctx, popReqCh, popErrCh),
		!a.So(assertPopUplinks(ctx, popReqCh, popErrCh, pbs[2]), should.BeTrue),
		!a.So(assertAdd(ctx, pbs[3]), should.BeTrue),
		!a.So(assertAdd(ctx, pbs[4], pbs[5]), should.BeTrue),
		!pop(ctx, popReqCh, popErrCh),
		!a.So(assertPopUplinks(ctx, popReqCh, popErrCh, pbs[3]), should.BeTrue),
		!pop(ctx, popReqCh, popErrCh),
		!a.So(assertPopUplinks(ctx, popReqCh, popErrCh, pbs[4], pbs[5]), should.BeTrue):
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
