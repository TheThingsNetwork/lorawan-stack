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
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/proto"
)

// handleApplicationUplinkQueueTest runs a test suite on q.
func handleApplicationUplinkQueueTest(ctx context.Context, q ApplicationUplinkQueue, consumerIDs []string) {
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

		dispatchErrCh := make(chan error, len(consumerIDs))

		dispatchCtx, cancelDispatchCtx := context.WithCancel(ctx)
		defer cancelDispatchCtx()
		for _, consumerID := range consumerIDs {
			go func(consumerID string) {
				select {
				case <-ctx.Done():
				case dispatchErrCh <- q.Dispatch(dispatchCtx, consumerID):
				}
			}(consumerID)
		}
		defer func() {
			cancelDispatchCtx()

			for range consumerIDs {
				select {
				case <-ctx.Done():
				case err := <-dispatchErrCh:
					a.So(errors.IsCanceled(err), should.BeTrue)
				}
			}
		}()

		type popFuncReq struct {
			Context                context.Context
			ApplicationIdentifiers *ttnpb.ApplicationIdentifiers
			Func                   ApplicationUplinkQueueDrainFunc
			Response               chan<- TaskPopFuncResponse
		}
		reqCh := make(chan popFuncReq, 1)
		errCh := make(chan error, 1)
		popCtx, cancelPopCtx := context.WithCancel(ctx)
		defer cancelPopCtx()
		for _, consumerID := range consumerIDs {
			go func(consumerID string) {
				errCh <- q.Pop(popCtx, consumerID, func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, f ApplicationUplinkQueueDrainFunc) (time.Time, error) {
					respCh := make(chan TaskPopFuncResponse, 1)
					select {
					case <-popCtx.Done():
						return time.Time{}, popCtx.Err()
					case reqCh <- popFuncReq{
						Context:                ctx,
						ApplicationIdentifiers: appID,
						Func:                   f,
						Response:               respCh,
					}:
					}
					select {
					case <-popCtx.Done():
						return time.Time{}, popCtx.Err()
					case resp := <-respCh:
						return resp.Time, resp.Error
					}
				})
			}(consumerID)
		}

		var collected []*ttnpb.ApplicationUp
		var requests int
		for ; len(collected) < len(expected); requests++ {
			select {
			case <-ctx.Done():
				t.Error("Timed out while waiting for Pop callback to be called")
				return false

			case req := <-reqCh:
				if !test.AllTrue(
					a.So(req.Context, should.HaveParentContextOrEqual, ctx),
					a.So(req.ApplicationIdentifiers, should.Resemble, expected[0].EndDeviceIds.ApplicationIds),
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

				if !a.So(req.Func(2, func(ups ...*ttnpb.ApplicationUp) error {
					a.So(ups, should.NotBeEmpty)
					collected = append(collected, ups...)
					return nil
				}), should.BeNil) {
					return false
				}
				close(req.Response)
			}
		}

		for i := 0; i < requests; i++ {
			select {
			case <-ctx.Done():
				t.Error("Timed out while waiting for Pop to return")
				return false
			case err := <-errCh:
				if !a.So(err, should.BeNil) {
					return false
				}
			}
		}

		cancelPopCtx()

		for i := requests; i < len(consumerIDs); i++ {
			select {
			case <-ctx.Done():
				t.Error("Timed out while waiting for Pop to return")
				return false
			case err := <-errCh:
				if !a.So(errors.IsCanceled(err), should.BeTrue) {
					return false
				}
			}
		}

		if a.So(len(collected), should.Equal, len(expected)) {
			for _, ex := range expected {
				expectedB, err := proto.Marshal(ex)
				if !a.So(err, should.BeNil) {
					return false
				}

				var found bool
				for _, coll := range collected {
					collectedB, err := proto.Marshal(coll)
					if !a.So(err, should.BeNil) {
						return false
					}

					if bytes.Equal(expectedB, collectedB) {
						found = true
					}
				}

				if !a.So(found, should.BeTrue) {
					return false
				}
			}
		}

		return true
	}

	appID1 := &ttnpb.ApplicationIdentifiers{
		ApplicationId: "application-uplink-queue-app-1",
	}

	appID2 := &ttnpb.ApplicationIdentifiers{
		ApplicationId: "application-uplink-queue-app-2",
	}

	invalidations := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev2",
			},
			CorrelationIds: []string{"invalidations[0]"},
			Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
				DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{},
			},
		},
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev2",
			},
			CorrelationIds: []string{"invalidations[1]"},
			Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
				DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{},
			},
		},
	}
	joinAccepts := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev",
			},
			CorrelationIds: []string{"joinAccepts[0]"},
			Up: &ttnpb.ApplicationUp_JoinAccept{
				JoinAccept: &ttnpb.ApplicationJoinAccept{},
			},
		},
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev2",
			},
			CorrelationIds: []string{"joinAccepts[1]"},
			Up: &ttnpb.ApplicationUp_JoinAccept{
				JoinAccept: &ttnpb.ApplicationJoinAccept{},
			},
		},
	}
	genericApp1Ups := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev",
			},
			CorrelationIds: []string{"genericApp1Ups[0]"},
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{},
			},
		},
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev",
			},
			CorrelationIds: []string{"genericApp1Ups[1]"},
			Up: &ttnpb.ApplicationUp_LocationSolved{
				LocationSolved: &ttnpb.ApplicationLocation{},
			},
		},
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev2",
			},
			CorrelationIds: []string{"genericApp1Ups[2]"},
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{},
			},
		},
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev",
			},
			CorrelationIds: []string{"genericApp1Ups[3]"},
			Up: &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: &ttnpb.ApplicationDownlink{},
			},
		},
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID1,
				DeviceId:       "test-dev2",
			},
			CorrelationIds: []string{"genericApp1Ups[4]"},
			Up: &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: &ttnpb.ApplicationDownlink{},
			},
		},
	}

	genericApp2Ups := [...]*ttnpb.ApplicationUp{
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID2,
				DeviceId:       "test-dev2",
			},
			CorrelationIds: []string{"genericApp2Ups[0]"},
			Up: &ttnpb.ApplicationUp_LocationSolved{
				LocationSolved: &ttnpb.ApplicationLocation{},
			},
		},
		{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: appID2,
				DeviceId:       "test-dev",
			},
			CorrelationIds: []string{"genericApp2Ups[1]"},
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
			genericApp2Ups[1],
		), should.BeTrue),
		!a.So(assertDrainApplication(ctx, true,
			genericApp2Ups[0],
			genericApp2Ups[1],
		), should.BeTrue),

		!a.So(assertAdd(ctx,
			genericApp1Ups[1],
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
func HandleApplicationUplinkQueueTest(t *testing.T, q ApplicationUplinkQueue, consumerIDs []string) {
	t.Helper()
	test.RunTest(t, test.TestConfig{
		Func: func(ctx context.Context, a *assertions.Assertion) {
			t.Helper()
			test.RunSubtestFromContext(ctx, test.SubtestConfig{
				Name: "1st run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleApplicationUplinkQueueTest(ctx, q, consumerIDs)
				},
			})
			if t.Failed() {
				t.Skip("Skipping 2nd run")
			}
			test.RunSubtestFromContext(ctx, test.SubtestConfig{
				Name: "2nd run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleApplicationUplinkQueueTest(ctx, q, consumerIDs)
				},
			})
		},
	})
}
