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
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestLinkApplication(t *testing.T) {
	a := assertions.New(t)

	ns, ctx, env, stop := StartTest(t, Config{}, (1<<12)*test.Delay, true)
	defer stop()

	<-env.DownlinkTasks.Pop

	appID1 := ttnpb.ApplicationIdentifiers{
		ApplicationID: "link-application-app-1",
	}

	appID2 := ttnpb.ApplicationIdentifiers{
		ApplicationID: "link-application-app-2",
	}

	link1, link1EndEventClosure, ok := AssertLinkApplication(ctx, ns.LoopbackConn(), env.Cluster.GetPeer, env.Events, appID1)
	if !a.So(ok, should.BeTrue) {
		t.Fatal("Failed to link application 1")
	}
	var link1SubscribeCtx context.Context
	var link1SubscribeFunc func(context.Context, *ttnpb.ApplicationUp) error
	var link1SubscribeRespCh chan<- error
	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for ApplicationUplinks.Subscribe to be called")

	case req := <-env.ApplicationUplinks.Subscribe:
		a.So(req.Identifiers, should.Resemble, appID1)
		if !a.So(req.Func, should.NotBeNil) {
			t.Fatal("Subscribe callback function is nil")
		}
		link1SubscribeCtx = req.Context
		link1SubscribeFunc = req.Func
		link1SubscribeRespCh = req.Response
	}

	for i, link1Up := range []*ttnpb.ApplicationUp{
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
	} {
		t.Run(fmt.Sprintf("uplink %d", i), func(t *testing.T) {
			a := assertions.New(t)

			errCh := make(chan error)
			go func(link1Up *ttnpb.ApplicationUp) {
				errCh <- link1SubscribeFunc(ctx, link1Up)
			}(link1Up)

			var up *ttnpb.ApplicationUp
			var err error
			if !a.So(test.WaitContext(ctx, func() {
				up, err = link1.Recv()
			}), should.BeTrue) {
				t.Fatal("Timed out while waiting for AS link receive to succeed")
			}
			if !a.So(err, should.BeNil) {
				t.Fatal("AS link receive failed")
			}
			a.So(up, should.Resemble, link1Up)

			if !a.So(test.WaitContext(ctx, func() {
				err = link1.Send(ttnpb.Empty)
			}), should.BeTrue) {
				t.Fatal("Timed out while waiting for AS link send to succeed")
			}
			if !a.So(err, should.BeNil) {
				t.Fatal("AS link send failed")
			}

			select {
			case <-ctx.Done():
				t.Fatal("Timed out while waiting for ApplicationUplinks.Subscribe callback to return")

			case err := <-errCh:
				a.So(err, should.BeNil)
			}
		})
	}

	link2, link2EndEventClosure, ok := AssertLinkApplication(ctx, ns.LoopbackConn(), env.Cluster.GetPeer, env.Events, appID2)
	if !a.So(ok, should.BeTrue) {
		t.Fatal("Failed to link application 2")
	}
	var link2SubscribeCtx context.Context
	var link2SubscribeRespCh chan<- error
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for ApplicationUplinks.Subscribe to be called")

	case req := <-env.ApplicationUplinks.Subscribe:
		a.So(req.Identifiers, should.Resemble, appID2)
		if !a.So(req.Func, should.NotBeNil) {
			t.Fatalf("Subscribe callback function is nil")
		}
		link2SubscribeCtx = req.Context
		link2SubscribeRespCh = req.Response
	}

	select {
	case <-link1SubscribeCtx.Done():
		t.Fatal("ApplicationUplinks.Subscribe context is done too early")
	default:
	}

	type linkApplicationResp struct {
		Link            ttnpb.AsNs_LinkApplicationClient
		EndEventClosure func(error) events.Event
		Ok              bool
	}
	linkApplicationRespCh := make(chan linkApplicationResp, 1)
	go func() {
		newLink1, newLink1EndEventClosure, ok := AssertLinkApplication(ctx, ns.LoopbackConn(), env.Cluster.GetPeer, env.Events, appID1, link1EndEventClosure(context.Canceled))
		linkApplicationRespCh <- linkApplicationResp{
			Link:            newLink1,
			EndEventClosure: newLink1EndEventClosure,
			Ok:              ok,
		}
	}()
	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for ApplicationUplinks.Subscribe context to be done")

	case <-link1SubscribeCtx.Done():
		link1SubscribeRespCh <- link1SubscribeCtx.Err()
		close(link1SubscribeRespCh)
	}
	up, err := link1.Recv()
	if !a.So(up, should.BeNil) {
		t.Fatalf("Received uplink on link 1: %v", up)
	}
	a.So(err, should.BeError)

	var newLink1 ttnpb.AsNs_LinkApplicationClient
	var newLink1EndEventClosure func(error) events.Event
	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for application 1 to relink")

	case resp := <-linkApplicationRespCh:
		if !a.So(resp.Ok, should.BeTrue) {
			t.Fatal("Application 1 failed to relink")
		}
		newLink1 = resp.Link
		newLink1EndEventClosure = resp.EndEventClosure
	}

	var newLink1SubscribeCtx context.Context
	var newLink1SubscribeRespCh chan<- error
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for ApplicationUplinks.Subscribe to be called")

	case req := <-env.ApplicationUplinks.Subscribe:
		a.So(req.Identifiers, should.Resemble, appID1)
		if !a.So(req.Func, should.NotBeNil) {
			t.Fatalf("Subscribe callback function is nil")
		}
		newLink1SubscribeCtx = req.Context
		newLink1SubscribeRespCh = req.Response
	}

	select {
	case <-newLink1SubscribeCtx.Done():
		t.Fatal("ApplicationUplinks.Subscribe context is done too early")

	case <-link2SubscribeCtx.Done():
		t.Fatal("ApplicationUplinks.Subscribe context is done too early")

	default:
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		a.So(AssertNetworkServerClose(ctx, ns), should.BeTrue)
		wg.Done()
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for ApplicationUplinks.Subscribe context to be done")

	case <-newLink1SubscribeCtx.Done():
		newLink1SubscribeRespCh <- link1SubscribeCtx.Err()
	}

	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for ApplicationUplinks.Subscribe context to be done")

	case <-link2SubscribeCtx.Done():
		link2SubscribeRespCh <- link1SubscribeCtx.Err()
	}

	if !a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, 2, func(evs ...events.Event) bool {
		return a.So(evs, should.HaveSameElements, []events.Event{
			newLink1EndEventClosure(context.Canceled),
			link2EndEventClosure(context.Canceled),
		}, test.EventEqual)
	}), should.BeTrue) {
		t.Fatal("AS link end events assertion failed")
	}

	up, err = newLink1.Recv()
	if !a.So(up, should.BeNil) {
		t.Fatalf("Received uplink on new link 1: %v", up)
	}
	a.So(err, should.BeError)

	up, err = link2.Recv()
	if !a.So(up, should.BeNil) {
		t.Fatalf("Received uplink on link 2: %v", up)
	}
	a.So(err, should.BeError)

	if !a.So(test.WaitContext(ctx, wg.Wait), should.BeTrue) {
		t.Fatal("Timed out while waiting for Network Server to close")
	}
}

func TestDownlinkQueueReplace(t *testing.T) {
	start := time.Now()

	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		AddFunc        func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) error
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				err := errors.New("SetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(nil)
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
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
			Name: "Valid request/replace/no MAC state",
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
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
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
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
			Name: "Valid request/replace/Class A",
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"queued_application_downlinks",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 0},
						{SessionKeyID: []byte("testSession"), FCnt: 42},
					},
				})
				return dev, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/replace/Class C",
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				})
				a.So(replace, should.BeTrue)
				a.So([]time.Time{start, at.Add(NSScheduleWindow()), time.Now()}, should.BeChronological)
				return nil
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"queued_application_downlinks",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 0},
						{SessionKeyID: []byte("testSession"), FCnt: 42},
					},
				})
				return dev, nil
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			AddCalls:     1,
			SetByIDCalls: 1,
		},

		{
			Name: "Valid request/delete/no MAC state",
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
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
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"queued_application_downlinks",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
				})
				return dev, nil
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"queued_application_downlinks",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
				})
				return dev, nil
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
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var addCalls, setByIDCalls uint64

			ns := test.Must(New(
				componenttest.NewComponent(t, &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setByIDCalls, 1)
							return tc.SetByIDFunc(ctx, appID, devID, gets, f)
						},
					},
					DownlinkTasks: &MockDownlinkTaskQueue{
						AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
							atomic.AddUint64(&addCalls, 1)
							return tc.AddFunc(ctx, ids, at, replace)
						},
						PopFunc: DownlinkTaskPopBlockFunc,
					},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
				})).(*NetworkServer)
			ns.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)

			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			componenttest.StartComponent(t, ns.Component)
			defer ns.Close()

			req := deepcopy.Copy(tc.Request).(*ttnpb.DownlinkQueueRequest)

			res, err := ttnpb.NewAsNsClient(ns.LoopbackConn()).DownlinkQueueReplace(test.Context(), req)
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(res, should.BeNil)
			} else if a.So(err, should.BeNil) {
				a.So(res, should.Resemble, ttnpb.Empty)
			}
			a.So(req, should.Resemble, tc.Request)
			a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
			a.So(addCalls, should.Equal, tc.AddCalls)
		})
	}
}

func TestDownlinkQueuePush(t *testing.T) {
	start := time.Now()

	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		AddFunc        func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) error
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				err := errors.New("SetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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

				dev, sets, err := f(nil)
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
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
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
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
			Name: "Valid request/push/Class A",
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"queued_application_downlinks",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
						{SessionKeyID: []byte("testSession"), FCnt: 6},
						{SessionKeyID: []byte("testSession"), FCnt: 42},
					},
				})
				return dev, nil
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
			Name: "Valid request/push/Class A",
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				err := errors.New("AddFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.Resemble, []string{
					"queued_application_downlinks",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_A,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testPendingSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
						{SessionKeyID: []byte("testSession"), FCnt: 6},
						{SessionKeyID: []byte("testSession"), FCnt: 42},
					},
				})
				return dev, nil
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				})
				a.So(replace, should.BeTrue)
				a.So([]time.Time{start, at, time.Now()}, should.BeChronological)
				return nil
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 1},
						{SessionKeyID: []byte("testSession"), FCnt: 2},
						{SessionKeyID: []byte("testSession"), FCnt: 3},
						{SessionKeyID: []byte("testSession"), FCnt: 5},
					},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				})
				a.So(replace, should.BeTrue)
				a.So([]time.Time{start, at, time.Now()}, should.BeChronological)
				return nil
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_0_2,
					},
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte("testSession"),
						},
						LastNFCntDown: 10,
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				})
				a.So(replace, should.BeTrue)
				a.So([]time.Time{start, at, time.Now()}, should.BeChronological)
				return nil
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				dev, sets, err := f(&ttnpb.EndDevice{
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					MACState: &ttnpb.MACState{
						DeviceClass:    ttnpb.CLASS_C,
						LoRaWANVersion: ttnpb.MAC_V1_0_2,
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
				})
				if !a.So(err, should.BeError) {
					t.Error("Error was expected")
					return nil, errors.New("Error was expected")
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, err
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
					{SessionKeyID: []byte("testSession"), FCnt: 1},
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
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var addCalls, setByIDCalls uint64

			ns := test.Must(New(
				componenttest.NewComponent(t, &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setByIDCalls, 1)
							return tc.SetByIDFunc(ctx, appID, devID, gets, f)
						},
					},
					DownlinkTasks: &MockDownlinkTaskQueue{
						AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, at time.Time, replace bool) error {
							atomic.AddUint64(&addCalls, 1)
							return tc.AddFunc(ctx, ids, at, replace)
						},
						PopFunc: DownlinkTaskPopBlockFunc,
					},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
				})).(*NetworkServer)
			ns.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)

			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			componenttest.StartComponent(t, ns.Component)
			defer ns.Close()

			req := deepcopy.Copy(tc.Request).(*ttnpb.DownlinkQueueRequest)

			res, err := ttnpb.NewAsNsClient(ns.LoopbackConn()).DownlinkQueuePush(test.Context(), req)
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(res, should.BeNil)
			} else if a.So(err, should.BeNil) {
				a.So(res, should.Resemble, ttnpb.Empty)
			}
			a.So(req, should.Resemble, tc.Request)
			a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
			a.So(addCalls, should.Equal, tc.AddCalls)
		})
	}
}

func TestDownlinkQueueList(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		GetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, error)
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
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, error) {
				err := errors.New("GetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
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
			Name: "Valid request/empty queue",
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
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"queued_application_downlinks",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
				}, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			Downlinks:    &ttnpb.ApplicationDownlinks{},
			GetByIDCalls: 1,
		},

		{
			Name: "Valid request/non-empty queue",
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
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"queued_application_downlinks",
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
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{SessionKeyID: []byte("testSession"), FCnt: 0},
						{SessionKeyID: []byte("testSession"), FCnt: 42},
					},
				}, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			Downlinks: &ttnpb.ApplicationDownlinks{
				Downlinks: []*ttnpb.ApplicationDownlink{
					{SessionKeyID: []byte("testSession"), FCnt: 0},
					{SessionKeyID: []byte("testSession"), FCnt: 42},
				},
			},
			GetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var getByIDCalls uint64

			ns := test.Must(New(
				componenttest.NewComponent(t, &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&getByIDCalls, 1)
							return tc.GetByIDFunc(ctx, appID, devID, gets)
						},
					},
					DownlinkTasks: &MockDownlinkTaskQueue{
						PopFunc: DownlinkTaskPopBlockFunc,
					},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
				})).(*NetworkServer)

			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			componenttest.StartComponent(t, ns.Component)
			defer ns.Close()

			req := deepcopy.Copy(tc.Request).(*ttnpb.EndDeviceIdentifiers)

			res, err := ttnpb.NewAsNsClient(ns.LoopbackConn()).DownlinkQueueList(test.Context(), req)
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(res, should.BeNil)
			} else if a.So(err, should.BeNil) {
				a.So(res, should.Resemble, tc.Downlinks)
			}
			a.So(req, should.Resemble, tc.Request)
			a.So(getByIDCalls, should.Equal, tc.GetByIDCalls)
		})
	}
}

// handleApplicationUplinkQueueTest runs a test suite on q.
func handleApplicationUplinkQueueTest(t *testing.T, q ApplicationUplinkQueue) {
	a := assertions.New(t)

	ctx := test.ContextWithT(test.Context(), t)
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
	defer cancel()

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
	t.Parallel()

	for _, tc := range []struct {
		Name string
		New  func(t testing.TB) (q ApplicationUplinkQueue, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New:  NewRedisApplicationUplinkQueue,
			N:    8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			t.Run(fmt.Sprintf("%s/%d", tc.Name, i), func(t *testing.T) {
				t.Parallel()
				q, closeFn := tc.New(t)
				if closeFn != nil {
					defer func() {
						if err := closeFn(); err != nil {
							t.Errorf("Failed to close application uplink schedule: %s", err)
						}
					}()
				}
				t.Run("1st run", func(t *testing.T) { handleApplicationUplinkQueueTest(t, q) })
				if t.Failed() {
					t.Skip("Skipping 2nd run")
				}
				t.Run("2nd run", func(t *testing.T) { handleApplicationUplinkQueueTest(t, q) })
			})
		}
	}
}
