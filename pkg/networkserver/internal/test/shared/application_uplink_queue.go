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

	assertDrainApplication := func(ctx context.Context, withError bool, expected ...*ttnpb.ApplicationUp) bool {
		t, a := test.MustNewTFromContext(ctx)
		t.Helper()

		type popFuncReq struct {
			Context                context.Context
			ApplicationIdentifiers ttnpb.ApplicationIdentifiers
			Func                   ApplicationUplinkQueueDrainFunc
			Response               chan<- TaskPopFuncResponse
		}
		reqCh := make(chan popFuncReq, 1)
		errCh := make(chan error, 1)
		go func() {
			errCh <- q.Pop(ctx, func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f ApplicationUplinkQueueDrainFunc) (time.Time, error) {
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

		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for Pop callback to be called")
			return false

		case err := <-errCh:
			t.Errorf("Received unexpected Pop error: %v", err)
			return false

		case req := <-reqCh:
			if !test.AllTrue(
				a.So(req.Context, should.HaveParentContextOrEqual, ctx),
				a.So(req.ApplicationIdentifiers, should.Resemble, expected[0].ApplicationIdentifiers),
			) {
				t.Error("Pop callback assertion failed")
				return false
			}

			if withError {
				if !a.So(req.Func(2, func(ups ...*ttnpb.ApplicationUp) error {
					a.So(ups, should.NotBeEmpty)
					return test.ErrInternal
				}), should.HaveSameErrorDefinitionAs, test.ErrInternal) {
					return false
				}
			}

			var collected []*ttnpb.ApplicationUp
			if !a.So(req.Func(2, func(ups ...*ttnpb.ApplicationUp) error {
				a.So(ups, should.NotBeEmpty)
				collected = append(collected, ups...)
				return nil
			}), should.BeNil) {
				return false
			}
			if a.So(len(collected), should.Equal, len(expected)) {
				for i := range collected {
					a.So(collected[i], should.Resemble, expected[i])
				}
			}
			close(req.Response)

			select {
			case <-ctx.Done():
				t.Error("Timed out while waiting for Pop to return")
				return false

			case err := <-errCh:
				return a.So(err, should.BeNil)
			}
		}
	}

	appID1 := ttnpb.ApplicationIdentifiers{
		ApplicationId: "application-uplink-queue-app-1",
	}

	appID2 := ttnpb.ApplicationIdentifiers{
		ApplicationId: "application-uplink-queue-app-2",
	}

	invalidations := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev2",
			},
			CorrelationIDs: []string{"invalidations[0]"},
			Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
				DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev2",
			},
			CorrelationIDs: []string{"invalidations[1]"},
			Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
				DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{},
			},
		},
	}
	joinAccepts := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev",
			},
			CorrelationIDs: []string{"joinAccepts[0]"},
			Up: &ttnpb.ApplicationUp_JoinAccept{
				JoinAccept: &ttnpb.ApplicationJoinAccept{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev2",
			},
			CorrelationIDs: []string{"joinAccepts[1]"},
			Up: &ttnpb.ApplicationUp_JoinAccept{
				JoinAccept: &ttnpb.ApplicationJoinAccept{},
			},
		},
	}
	genericApp1Ups := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev",
			},
			CorrelationIDs: []string{"genericApp1Ups[0]"},
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev",
			},
			CorrelationIDs: []string{"genericApp1Ups[1]"},
			Up: &ttnpb.ApplicationUp_LocationSolved{
				LocationSolved: &ttnpb.ApplicationLocation{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev2",
			},
			CorrelationIDs: []string{"genericApp1Ups[2]"},
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev",
			},
			CorrelationIDs: []string{"genericApp1Ups[3]"},
			Up: &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: &ttnpb.ApplicationDownlink{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID1,
				DeviceId:               "test-dev2",
			},
			CorrelationIDs: []string{"genericApp1Ups[4]"},
			Up: &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: &ttnpb.ApplicationDownlink{},
			},
		},
	}

	genericApp2Ups := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID2,
				DeviceId:               "test-dev2",
			},
			CorrelationIDs: []string{"genericApp2Ups[0]"},
			Up: &ttnpb.ApplicationUp_LocationSolved{
				LocationSolved: &ttnpb.ApplicationLocation{},
			},
		},
		{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID2,
				DeviceId:               "test-dev",
			},
			CorrelationIDs: []string{"genericApp2Ups[1]"},
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{},
			},
		},
	}

	_, a := test.MustNewTFromContext(ctx)
	switch {
	case
		!a.So(assertAdd(ctx,
			genericApp1Ups[0],
		), should.BeTrue),
		!a.So(assertDrainApplication(ctx, false,
			genericApp1Ups[0],
		), should.BeTrue),

		!a.So(assertAdd(ctx,
			genericApp2Ups[0],
		), should.BeTrue),
		!a.So(assertAdd(ctx,
			genericApp1Ups[1],
			genericApp2Ups[1],
		), should.BeTrue),
		!a.So(assertDrainApplication(ctx, true,
			genericApp2Ups[0],
			genericApp2Ups[1],
		), should.BeTrue),

		!a.So(assertAdd(ctx,
			genericApp1Ups[2],
			invalidations[0],
			genericApp1Ups[3],
			invalidations[1],
			joinAccepts[0],
		), should.BeTrue),
		!a.So(assertAdd(ctx,
			genericApp1Ups[4],
			joinAccepts[1],
		), should.BeTrue),
		!a.So(assertDrainApplication(ctx, false,
			joinAccepts[0],
			joinAccepts[1],
			invalidations[0],
			invalidations[1],
			genericApp1Ups[1],
			genericApp1Ups[2],
			genericApp1Ups[3],
			genericApp1Ups[4],
		), should.BeTrue):
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
				Name: "2nd run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleApplicationUplinkQueueTest(ctx, q)
				},
			})
		},
	})
}
