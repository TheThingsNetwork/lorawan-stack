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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

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
		RxMetadata: DefaultRxMetadata[:],
	}
	ups := []*ttnpb.UplinkMessage{up}

	for _, tc := range []struct {
		Name           string
		Time           time.Time
		AddFunc        func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) error
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
		Request        *ttnpb.DownlinkQueueRequest
		ErrorAssertion func(*testing.T, error) bool
		AddCalls       uint64
		SetByIDCalls   uint64
	}{
		{
			Name: "Non-existing device",
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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

				ns, ctx, _, stop := StartTest(
					ctx,
					TestConfig{
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
					},
				)
				defer stop()

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
		RxMetadata: DefaultRxMetadata[:],
	}
	ups := []*ttnpb.UplinkMessage{up}

	for _, tc := range []struct {
		Name           string
		AddFunc        func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) error
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
		Request        *ttnpb.DownlinkQueueRequest
		ErrorAssertion func(*testing.T, error) bool
		AddCalls       uint64
		SetByIDCalls   uint64
	}{
		{
			Name: "Non-existing device",
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
			AddFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
				test.MustTFromContext(ctx).Errorf("Add called with %v %v %v", ids, startAt, replace)
				return errors.New("AddFunc must not be called")
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
					ctx,
					TestConfig{
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
		GetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, context.Context, error)
		Request        *ttnpb.EndDeviceIdentifiers
		Downlinks      *ttnpb.ApplicationDownlinks
		ErrorAssertion func(*testing.T, error) bool
		GetByIDCalls   uint64
	}{
		{
			Name: "Valid request/empty queues",
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
					ctx,
					TestConfig{
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
