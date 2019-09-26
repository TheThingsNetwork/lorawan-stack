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
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	. "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
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

	link1, link1EndEvent, ok := AssertLinkApplication(ctx, ns.LoopbackConn(), env.Cluster.GetPeer, env.Events, appID1)
	if !a.So(ok, should.BeTrue) {
		t.Fatal("Failed to link application 1")
	}

	link2, link2EndEvent, ok := AssertLinkApplication(ctx, ns.LoopbackConn(), env.Cluster.GetPeer, env.Events, appID2)
	if !a.So(ok, should.BeTrue) {
		t.Fatal("Failed to link application 2")
	}

	newLink1, newLink1EndEvent, ok := AssertLinkApplication(ctx, ns.LoopbackConn(), env.Cluster.GetPeer, env.Events, appID1, link1EndEvent(nil))
	if !a.So(ok, should.BeTrue) {
		t.Fatal("Failed to relink application 1")
	}

	up, err := link1.Recv()
	if !a.So(up, should.BeNil) {
		t.Fatalf("Received uplink on link 1: %v", up)
	}
	a.So(err, should.Resemble, io.EOF)

	a.So(AssertNetworkServerClose(ctx, ns), should.BeTrue)

	if !a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, 2, func(evs ...events.Event) bool {
		return a.So(evs, should.HaveSameElements, []events.Event{
			newLink1EndEvent(context.Canceled),
			link2EndEvent(context.Canceled),
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})
				dev, sets, err := f(&ttnpb.EndDevice{
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})
				dev, sets, err := f(&ttnpb.EndDevice{
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
				a.So([]time.Time{start, at, time.Now()}, should.BeChronological)
				return nil
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
				NewComponent(t, &component.Config{}),
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

			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			StartComponent(t, ns.Component)
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})
				dev, sets, err := f(&ttnpb.EndDevice{
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})
				dev, sets, err := f(&ttnpb.EndDevice{
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})
				dev, sets, err := f(&ttnpb.EndDevice{
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})
				dev, sets, err := f(&ttnpb.EndDevice{
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})
				dev, sets, err := f(&ttnpb.EndDevice{
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
					"mac_state",
					"multicast",
					"pending_mac_state",
					"pending_session",
					"queued_application_downlinks",
					"session",
				})
				dev, sets, err := f(&ttnpb.EndDevice{
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
				NewComponent(t, &component.Config{}),
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

			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			StartComponent(t, ns.Component)

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
				NewComponent(t, &component.Config{}),
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
			StartComponent(t, ns.Component)
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
