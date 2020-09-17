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

package networkserver_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
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

func TestApplicationUplinkQueues(t *testing.T) {
	for _, tc := range []struct {
		Name string
		New  func(t testing.TB) (q ApplicationUplinkQueue, closeFn func())
	}{
		{
			Name: "Redis",
			New:  NewRedisApplicationUplinkQueue,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     MakeTestCaseName(tc.Name),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				q, closeFn := tc.New(t)
				if closeFn != nil {
					defer closeFn()
				}
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "1st run",
					Func: func(ctx context.Context, _ *testing.T, a *assertions.Assertion) {
						handleApplicationUplinkQueueTest(ctx, q)
					},
				})
				if t.Failed() {
					t.Skip("Skipping 2nd run")
				}
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "2st run",
					Func: func(ctx context.Context, _ *testing.T, a *assertions.Assertion) {
						handleApplicationUplinkQueueTest(ctx, q)
					},
				})
			},
		})
	}
}

func TestDownlinkQueueReplace(t *testing.T) {
	up := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHDR: ttnpb.MHDR{
				MType: ttnpb.MType_UNCONFIRMED_UP,
			},
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{},
			},
		},
		RxMetadata: RxMetadata[:],
	}
	ups := []*ttnpb.UplinkMessage{up}

	for _, tc := range []struct {
		Name           string
		Time           time.Time
		ContextFunc    func(context.Context) context.Context
		AddFunc        func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) error
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
		Request        *ttnpb.DownlinkQueueRequest
		ErrorAssertion func(*testing.T, error) bool
		AddCalls       uint64
		SetByIDCalls   uint64
	}{
		{
			Name: "No link rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				err := errors.New("SetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
		},

		{
			Name: "Non-existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})
				dev, sets, err := f(ctx, nil)
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, ctx, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsNotFound(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/replace/active session/MAC state",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})
				dev, sets, err := f(ctx, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, ctx, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsFailedPrecondition(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/replace/Class A/active session",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})

				getDevice := &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
				}

				dev, sets, err := f(ctx, CopyEndDevice(getDevice))
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}

				setDevice := CopyEndDevice(getDevice)
				setDevice.Session.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				}

				a.So(sets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				a.So(dev, should.ResembleFields, setDevice, sets)
				return dev, ctx, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/replace/Class A/both sessions",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})

				getDevice := &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
					PendingMACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testPendingSession"), FCnt: 1},
							{SessionKeyID: []byte("testPendingSession"), FCnt: 43},
						},
					},
				}

				dev, sets, err := f(ctx, CopyEndDevice(getDevice))
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}

				setDevice := CopyEndDevice(getDevice)
				setDevice.Session.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				}
				setDevice.PendingSession.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testPendingSession"), FCnt: 2},
				}

				a.So(sets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				a.So(dev, should.ResembleFields, setDevice, sets)
				return dev, ctx, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testPendingSession"), FCnt: 2},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/replace/Class C/active session",
			Time: time.Unix(0, 42),
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				})
				a.So(replace, should.BeTrue)
				a.So(startAt, should.Resemble, time.Unix(0, 42).Add(NSScheduleWindow()).UTC())
				return nil
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})

				getDevice := &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
				}

				dev, sets, err := f(ctx, CopyEndDevice(getDevice))
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}

				setDevice := CopyEndDevice(getDevice)
				setDevice.Session.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				}

				a.So(sets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				a.So(dev, should.ResembleFields, setDevice, sets)
				return dev, ctx, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			AddCalls:     1,
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/delete/no active MAC state",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})

				getDevice := &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testPendingSession"), FCnt: 1},
							{SessionKeyID: []byte("testPendingSession"), FCnt: 43},
						},
					},
				}

				dev, sets, err := f(ctx, CopyEndDevice(getDevice))
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}

				setDevice := CopyEndDevice(getDevice)
				setDevice.Session.QueuedApplicationDownlinks = nil
				setDevice.PendingSession.QueuedApplicationDownlinks = nil

				a.So(sets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				a.So(dev, should.ResembleFields, setDevice, sets)
				return dev, ctx, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/delete/Class A",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})

				getDevice := &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testPendingSession"), FCnt: 1},
							{SessionKeyID: []byte("testPendingSession"), FCnt: 43},
						},
					},
				}

				dev, sets, err := f(ctx, CopyEndDevice(getDevice))
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}

				setDevice := CopyEndDevice(getDevice)
				setDevice.Session.QueuedApplicationDownlinks = nil
				setDevice.PendingSession.QueuedApplicationDownlinks = nil

				a.So(sets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				a.So(dev, should.ResembleFields, setDevice, sets)
				return dev, ctx, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/delete/Class C",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})

				getDevice := &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testPendingSession"), FCnt: 1},
							{SessionKeyID: []byte("testPendingSession"), FCnt: 43},
						},
					},
				}

				dev, sets, err := f(ctx, CopyEndDevice(getDevice))
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}

				setDevice := CopyEndDevice(getDevice)
				setDevice.Session.QueuedApplicationDownlinks = nil
				setDevice.PendingSession.QueuedApplicationDownlinks = nil

				a.So(sets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				a.So(dev, should.ResembleFields, setDevice, sets)
				return dev, ctx, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
			},
			SetByIDCalls: 1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				var addCalls, setByIDCalls uint64

				ns, ctx, _, stop := StartTest(t, TestConfig{
					Context: ctx,
					NetworkServer: Config{
						Devices: &MockDeviceRegistry{
							SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
								atomic.AddUint64(&setByIDCalls, 1)
								return tc.SetByIDFunc(ctx, appID, devID, gets, f)
							},
						},
						DownlinkTasks: &MockDownlinkTaskQueue{
							AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
								atomic.AddUint64(&addCalls, 1)
								return tc.AddFunc(ctx, ids, startAt, replace)
							},
						},
						DefaultMACSettings: MACSettingConfig{
							StatusTimePeriodicity:  DurationPtr(0),
							StatusCountPeriodicity: func(v uint32) *uint32 { return &v }(0),
						},
						DownlinkQueueCapacity: 100,
					},
					TaskStarter: StartTaskExclude(
						DownlinkProcessTaskName,
					),
				})
				defer stop()

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				if !tc.Time.IsZero() {
					clock := test.NewMockClock(tc.Time)
					defer SetMockClock(clock)()
				}
				req := deepcopy.Copy(tc.Request).(*ttnpb.DownlinkQueueRequest)
				res, err := ttnpb.NewAsNsClient(ns.LoopbackConn()).DownlinkQueueReplace(ctx, req)
				if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(res, should.BeNil)
				} else if a.So(err, should.BeNil) {
					a.So(res, should.Resemble, ttnpb.Empty)
				}
				a.So(req, should.Resemble, tc.Request)
				a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
				a.So(addCalls, should.Equal, tc.AddCalls)
			},
		})
	}
}

func TestDownlinkQueuePush(t *testing.T) {
	up := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHDR: ttnpb.MHDR{
				MType: ttnpb.MType_UNCONFIRMED_UP,
			},
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{},
			},
		},
		RxMetadata: RxMetadata[:],
	}
	ups := []*ttnpb.UplinkMessage{up}

	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		AddFunc        func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) error
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
		Request        *ttnpb.DownlinkQueueRequest
		ErrorAssertion func(*testing.T, error) bool
		AddCalls       uint64
		SetByIDCalls   uint64
	}{
		{
			Name: "No link rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				err := errors.New("SetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
		},

		{
			Name: "Non-existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})

				dev, sets, err := f(ctx, nil)
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, ctx, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsNotFound(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/push/no MAC state",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})
				dev, sets, err := f(ctx, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, ctx, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsFailedPrecondition(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/push/Class A/active session",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})

				getDevice := &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
				}

				dev, sets, err := f(ctx, CopyEndDevice(getDevice))
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}

				setDevice := CopyEndDevice(getDevice)
				setDevice.Session.QueuedApplicationDownlinks = append(setDevice.Session.QueuedApplicationDownlinks,
					&ttnpb.ApplicationDownlink{SessionKeyID: []byte("testSession"), FCnt: 6},
					&ttnpb.ApplicationDownlink{SessionKeyID: []byte("testSession"), FCnt: 42},
				)

				a.So(sets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				a.So(dev, should.ResembleFields, setDevice, sets)
				return dev, ctx, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 6},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/push/Class A/both sessions",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})

				getDevice := &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
					PendingMACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
					},
				}

				dev, sets, err := f(ctx, CopyEndDevice(getDevice))
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}

				setDevice := CopyEndDevice(getDevice)
				setDevice.Session.QueuedApplicationDownlinks = append(setDevice.Session.QueuedApplicationDownlinks,
					&ttnpb.ApplicationDownlink{SessionKeyID: []byte("testSession"), FCnt: 6},
					&ttnpb.ApplicationDownlink{SessionKeyID: []byte("testSession"), FCnt: 42},
				)
				setDevice.PendingSession.QueuedApplicationDownlinks = append(setDevice.PendingSession.QueuedApplicationDownlinks,
					&ttnpb.ApplicationDownlink{SessionKeyID: []byte("testPendingSession"), FCnt: 2},
				)

				a.So(sets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				a.So(dev, should.ResembleFields, setDevice, sets)
				return dev, ctx, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 6},
					{SessionKeyID: []byte("testPendingSession"), FCnt: 2},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Invalid request/push/Class C/FCnt too low",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})
				dev, sets, err := f(ctx, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession"), FCnt: 1},
							{SessionKeyID: []byte("testSession"), FCnt: 2},
							{SessionKeyID: []byte("testSession"), FCnt: 3},
							{SessionKeyID: []byte("testSession"), FCnt: 5},
						},
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, ctx, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 1},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Invalid request/push/Class C/FCnt lower than NFCntDown",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})
				dev, sets, err := f(ctx, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_0_2,
						RecentUplinks:  ups,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						LastNFCntDown: 10,
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, ctx, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 1},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/push/Class C/No session",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"recent_uplinks",
					"session",
				})
				dev, sets, err := f(ctx, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_0_2,
						RecentUplinks:  ups,
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, ctx, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 1},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsNotFound(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			SetByIDCalls: 1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				var addCalls, setByIDCalls uint64

				ns, ctx, env, stop := StartTest(
					t,
					TestConfig{
						Context: ctx,
						NetworkServer: Config{
							Devices: &MockDeviceRegistry{
								SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
									atomic.AddUint64(&setByIDCalls, 1)
									return tc.SetByIDFunc(ctx, appID, devID, gets, f)
								},
							},
							DownlinkTasks: &MockDownlinkTaskQueue{
								AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
									atomic.AddUint64(&addCalls, 1)
									return tc.AddFunc(ctx, ids, startAt, replace)
								},
							},
							DownlinkQueueCapacity: 100,
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
						),
					},
				)
				defer stop()

				go LogEvents(t, env.Events)

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := deepcopy.Copy(tc.Request).(*ttnpb.DownlinkQueueRequest)
				res, err := ttnpb.NewAsNsClient(ns.LoopbackConn()).DownlinkQueuePush(ctx, req)
				if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(res, should.BeNil)
				} else if a.So(err, should.BeNil) {
					a.So(res, should.Resemble, ttnpb.Empty)
				}
				a.So(req, should.Resemble, tc.Request)
				a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
				a.So(addCalls, should.Equal, tc.AddCalls)
			},
		})
	}
}

func TestDownlinkQueueList(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		GetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, context.Context, error)
		Request        *ttnpb.EndDeviceIdentifiers
		Downlinks      *ttnpb.ApplicationDownlinks
		ErrorAssertion func(*testing.T, error) bool
		GetByIDCalls   uint64
	}{
		{
			Name: "No link rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				err := errors.New("GetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, ctx, err
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
		},

		{
			Name: "Valid request/empty queues",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
				}, ctx, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			Downlinks:    &ttnpb.ApplicationDownlinks{},
			GetByIDCalls: 1,
		},

		{
			Name: "Valid request/active session queue",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession")},
							{SessionKeyID: []byte("testSession"), FCnt: 42},
						},
					},
				}, ctx, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			Downlinks: &ttnpb.ApplicationDownlinks{
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			GetByIDCalls: 1,
		},

		{
			Name: "Valid request/pending session queue",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testPendingSession"), FCnt: 1},
							{SessionKeyID: []byte("testPendingSession"), FCnt: 43},
							{SessionKeyID: []byte("testPendingSession"), FCnt: 44},
						},
					},
				}, ctx, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			Downlinks: &ttnpb.ApplicationDownlinks{
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testPendingSession"), FCnt: 1},
					{SessionKeyID: []byte("testPendingSession"), FCnt: 43},
					{SessionKeyID: []byte("testPendingSession"), FCnt: 44},
				},
			},
			GetByIDCalls: 1,
		},

		{
			Name: "Valid request/both queues present",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"pending_session.queued_application_downlinks",
					"queued_application_downlinks",
					"session.queued_application_downlinks",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testSession")},
							{SessionKeyID: []byte("testSession"), FCnt: 42},
						},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{SessionKeyID: []byte("testPendingSession"), FCnt: 1},
							{SessionKeyID: []byte("testPendingSession"), FCnt: 43},
							{SessionKeyID: []byte("testPendingSession"), FCnt: 44},
						},
					},
				}, ctx, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			Downlinks: &ttnpb.ApplicationDownlinks{
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession")},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
					{SessionKeyID: []byte("testPendingSession"), FCnt: 1},
					{SessionKeyID: []byte("testPendingSession"), FCnt: 43},
					{SessionKeyID: []byte("testPendingSession"), FCnt: 44},
				},
			},
			GetByIDCalls: 1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				var getByIDCalls uint64

				ns, ctx, env, stop := StartTest(
					t,
					TestConfig{
						Context: ctx,
						NetworkServer: Config{
							Devices: &MockDeviceRegistry{
								GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
									atomic.AddUint64(&getByIDCalls, 1)
									return tc.GetByIDFunc(ctx, appID, devID, gets)
								},
							},
							DownlinkQueueCapacity: 100,
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
						),
					},
				)
				defer stop()

				go LogEvents(t, env.Events)

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := deepcopy.Copy(tc.Request).(*ttnpb.EndDeviceIdentifiers)
				res, err := ttnpb.NewAsNsClient(ns.LoopbackConn()).DownlinkQueueList(ctx, req)
				if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(res, should.BeNil)
				} else if a.So(err, should.BeNil) {
					a.So(res, should.Resemble, tc.Downlinks)
				}
				a.So(req, should.Resemble, tc.Request)
				a.So(getByIDCalls, should.Equal, tc.GetByIDCalls)
			},
		})
	}
}
