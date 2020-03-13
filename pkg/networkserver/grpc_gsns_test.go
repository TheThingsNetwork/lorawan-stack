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
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestHandleUplink(t *testing.T) {
	joinGetByEUIPaths := [...]string{
		"frequency_plan_id",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"session.dev_addr",
		"supports_class_b",
		"supports_class_c",
		"supports_join",
	}
	joinSetByIDGetPaths := [...]string{
		"frequency_plan_id",
		"lorawan_phy_version",
		"pending_session.queued_application_downlinks",
		"queued_application_downlinks",
		"recent_uplinks",
		"session.queued_application_downlinks",
	}
	joinSetByIDSetPaths := [...]string{
		"pending_mac_state",
		"recent_uplinks",
	}

	dataSetByIDGetPaths := [...]string{
		"frequency_plan_id",
		"last_dev_status_received_at",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"mac_state",
		"multicast",
		"pending_mac_state",
		"pending_session",
		"queued_application_downlinks", // deprecated
		"recent_adr_uplinks",
		"recent_uplinks",
		"session",
		"supports_class_b",
		"supports_class_c",
	}
	dataRangeByAddrPaths := dataSetByIDGetPaths

	const (
		DeduplicationWindow = 24 * time.Millisecond
		CooldownWindow      = 42 * time.Millisecond
		CollectionWindow    = DeduplicationWindow + CooldownWindow

		Rx1Delay = ttnpb.RX_DELAY_5

		FPort = 0x42
		FCnt  = 42
	)

	NetID := test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)

	const AppIDString = "handle-uplink-test-app-id"
	AppID := ttnpb.ApplicationIdentifiers{ApplicationID: AppIDString}
	const DevID = "handle-uplink-test-dev-id"

	JoinEUI := types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	DevEUI := types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	DevAddr := types.DevAddr{0x42, 0x00, 0x00, 0x00}

	FNwkSIntKey := types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	NwkSEncKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	SNwkSIntKey := types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	AppSKey := types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	makeOTAAIdentifiers := func(devAddr *types.DevAddr) *ttnpb.EndDeviceIdentifiers {
		return &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: AppID,
			DeviceID:               DevID,

			DevEUI:  DevEUI.Copy(&types.EUI64{}),
			JoinEUI: JoinEUI.Copy(&types.EUI64{}),

			DevAddr: DevAddr.Copy(&types.DevAddr{}),
		}
	}
	makeApplicationDownlink := func() *ttnpb.ApplicationDownlink {
		return &ttnpb.ApplicationDownlink{
			SessionKeyID: []byte("app-down-1-session-key-id"),
			FPort:        FPort,
			FCnt:         0x32,
			FRMPayload:   []byte("app-down-1-frm-payload"),
			Confirmed:    true,
			Priority:     ttnpb.TxSchedulePriority_HIGH,
			CorrelationIDs: []string{
				"app-down-1-correlation-id-1",
			},
		}
	}
	makeUplinkSettings := func(dr ttnpb.DataRate, ch band.Channel) ttnpb.TxSettings {
		return ttnpb.TxSettings{
			DataRate:  *deepcopy.Copy(&dr).(*ttnpb.DataRate),
			EnableCRC: true,
			Frequency: ch.Frequency,
			Timestamp: 42,
		}
	}

	filterEndDevice := func(dev *ttnpb.EndDevice, paths ...string) *ttnpb.EndDevice {
		dev, err := ttnpb.FilterGetEndDevice(dev, paths...)
		if err != nil {
			panic(fmt.Errorf("failed to filter device: %w", err))
		}
		return dev
	}

	assertHandleUplinkResponse := func(ctx context.Context, handleUplinkErrCh <-chan error, expectedErr error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for NetworkServer.HandleUplink to return")
			return false

		case resErr := <-handleUplinkErrCh:
			return a.So(resErr, should.EqualErrorOrDefinition, expectedErr)
		}
	}
	assertHandleUplink := func(ctx context.Context, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error, up *ttnpb.UplinkMessage, f func() bool, expectedErr error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		errCh := handle(ctx, up)
		return assertions.New(t).So(AllTrue(
			f(),
			assertHandleUplinkResponse(ctx, errCh, expectedErr),
		), should.BeTrue)
	}
	assertDeduplicateUplink := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedUp *ttnpb.UplinkMessage, ok bool, err error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		return AssertDeduplicateUplink(ctx, env.UplinkDeduplicator.DeduplicateUplink, func(ctx context.Context, up *ttnpb.UplinkMessage, window time.Duration) bool {
			return AllTrue(
				a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
				a.So(up, should.Resemble, expectedUp),
				a.So(window, should.Resemble, CollectionWindow),
			)
		},
			UplinkDeduplicatorDeduplicateUplinkResponse{
				Ok:    ok,
				Error: err,
			},
		)
	}
	assertAccumulatedMetadata := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedUp *ttnpb.UplinkMessage, mds []*ttnpb.RxMetadata, err error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		return AssertAccumulatedMetadata(ctx, env.UplinkDeduplicator.AccumulatedMetadata, func(ctx context.Context, up *ttnpb.UplinkMessage) bool {
			return AllTrue(
				a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
				a.So(up, should.Resemble, expectedUp),
			)
		},
			UplinkDeduplicatorAccumulatedMetadataResponse{
				Metadata: mds,
				Error:    err,
			},
		)
	}
	assertPublishMergeMetadata := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers, mds ...*ttnpb.RxMetadata) bool {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)
		return a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
			return a.So(ev, should.ResembleEvent, EvtMergeMetadata(expectedCtx, expectedIDs, len(mds)))
		}), should.BeTrue)
	}
	assertDownlinkTaskAdd := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers, expectedStartAt time.Time, replace bool, err error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		return AssertDownlinkTaskAddRequest(ctx, env.DownlinkTasks.Add, func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) bool {
			return AllTrue(
				a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
				a.So(ids, should.Resemble, expectedIDs),
				a.So(startAt, should.Resemble, expectedStartAt),
				a.So(replace, should.Equal, replace),
			)
		},
			err,
		)
	}
	assertInteropJoin := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, joinReq *ttnpb.JoinRequest, joinResp *ttnpb.JoinResponse, err error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		return AssertInteropClientHandleJoinRequestRequest(ctx, env.InteropClient.HandleJoinRequest,
			func(ctx context.Context, id types.NetID, req *ttnpb.JoinRequest) bool {
				joinReq.DevAddr = req.DevAddr
				return AllTrue(
					a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
					a.So(id, should.Equal, NetID),
					a.So(req, should.NotBeNil),
					a.So(req.DevAddr, should.NotBeEmpty),
					a.So(req.DevAddr.NwkID(), should.Resemble, NetID.ID()),
					a.So(req.DevAddr.NetIDType(), should.Equal, NetID.Type()),
					a.So(req, should.Resemble, joinReq),
				)
			},
			InteropClientHandleJoinRequestResponse{
				Response: joinResp,
				Error:    err,
			},
		)
	}
	assertClusterLocalJoin := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers, joinReq *ttnpb.JoinRequest, joinResp *ttnpb.JoinResponse, err error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		return AssertNsJsPeerHandleAuthJoinRequest(ctx, env.Cluster.GetPeer, env.Cluster.Auth,
			func(ctx context.Context, ids ttnpb.Identifiers) bool {
				return AllTrue(
					a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
					a.So(ids, should.Resemble, expectedIDs),
				)
			},
			func(ctx context.Context, req *ttnpb.JoinRequest) bool {
				joinReq.DevAddr = req.DevAddr
				return AllTrue(
					a.So(req, should.NotBeNil),
					a.So(req.DevAddr, should.NotBeEmpty),
					a.So(req.DevAddr.NwkID(), should.Resemble, NetID.ID()),
					a.So(req.DevAddr.NetIDType(), should.Equal, NetID.Type()),
					a.So(req, should.Resemble, joinReq),
				)
			},
			&grpc.EmptyCallOption{},
			NsJsHandleJoinResponse{
				Response: joinResp,
				Error:    err,
			},
		)
	}
	assertJoinGetByEUI := func(ctx context.Context, env TestEnvironment, upCIDs []string, getDevice *ttnpb.EndDevice, err error) (context.Context, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		var getCtx context.Context
		return getCtx, AssertDeviceRegistryGetByEUI(ctx, env.DeviceRegistry.GetByEUI, func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) bool {
			getCtx = ctx
			ctxCIDs := events.CorrelationIDsFromContext(ctx)
			for _, id := range upCIDs {
				a.So(ctxCIDs, should.Contain, id)
			}
			return AllTrue(
				a.So(ctxCIDs, should.HaveLength, 2+len(upCIDs)),
				a.So(joinEUI, should.Resemble, JoinEUI),
				a.So(devEUI, should.Resemble, DevEUI),
				a.So(paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:]),
			)
		},
			func(ctx context.Context) DeviceRegistryGetByEUIResponse {
				if getDevice != nil {
					getDevice = filterEndDevice(CopyEndDevice(getDevice), joinGetByEUIPaths[:]...)
				}
				getCtx = context.WithValue(getCtx, struct{}{}, "get")
				return DeviceRegistryGetByEUIResponse{
					Device:  getDevice,
					Context: getCtx,
					Error:   err,
				}
			})
	}
	assertJoinSetByID := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, getDevice, setDevice *ttnpb.EndDevice, expectedErr error, err error) (context.Context, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		var setCtx context.Context
		return setCtx, AssertDeviceRegistrySetByID(ctx, env.DeviceRegistry.SetByID, func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) bool {
			setCtx = ctx
			dev, sets, err := f(ctx, CopyEndDevice(getDevice))
			var expectedSetPaths []string
			if setDevice != nil {
				expectedSetPaths = joinSetByIDSetPaths[:]
			}
			return AllTrue(
				a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
				a.So(appID, should.Resemble, AppID),
				a.So(devID, should.Resemble, DevID),
				a.So(gets, should.HaveSameElementsDeep, joinSetByIDGetPaths),
				a.So(sets, should.HaveSameElementsDeep, expectedSetPaths),
				a.So(dev, should.ResembleFields, setDevice, sets),
				a.So(err, should.EqualErrorOrDefinition, expectedErr),
			)
		},
			func(ctx context.Context) DeviceRegistrySetByIDResponse {
				if setDevice != nil {
					setDevice = filterEndDevice(CopyEndDevice(setDevice), joinSetByIDGetPaths[:]...)
				}
				setCtx = context.WithValue(setCtx, struct{}{}, "set")
				return DeviceRegistrySetByIDResponse{
					Device:  setDevice,
					Context: setCtx,
					Error:   err,
				}
			})
	}
	assertJoinApplicationUp := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, setDevice *ttnpb.EndDevice, recvAt time.Time, err error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		return AssertApplicationUplinkQueueAddRequest(ctx, env.ApplicationUplinks.Add, func(ctx context.Context, ups ...*ttnpb.ApplicationUp) bool {
			ids := *deepcopy.Copy(&setDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers)
			ids.DevAddr = &setDevice.PendingMACState.QueuedJoinAccept.Request.DevAddr
			return AllTrue(
				a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
				a.So(ups, should.Resemble, []*ttnpb.ApplicationUp{
					{
						CorrelationIDs:       events.CorrelationIDsFromContext(expectedCtx),
						EndDeviceIdentifiers: ids,
						Up: &ttnpb.ApplicationUp_JoinAccept{
							JoinAccept: &ttnpb.ApplicationJoinAccept{
								AppSKey:              setDevice.PendingMACState.QueuedJoinAccept.Keys.AppSKey,
								InvalidatedDownlinks: setDevice.GetSession().GetQueuedApplicationDownlinks(),
								ReceivedAt:           recvAt,
								SessionKeyID:         setDevice.PendingMACState.QueuedJoinAccept.Keys.SessionKeyID,
							},
						},
					},
				}),
			)
		}, err)
	}
	assertJoinDeduplicateSequence := func(ctx context.Context, env TestEnvironment, clock *test.MockClock, msg *ttnpb.UplinkMessage, dev *ttnpb.EndDevice, ok bool, err error) (context.Context, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		getCtx, getOk := assertJoinGetByEUI(ctx, env, msg.CorrelationIDs, dev, nil)
		if !a.So(getOk, should.BeTrue) {
			return nil, false
		}
		msg.ReceivedAt = clock.Now()
		msg.CorrelationIDs = events.CorrelationIDsFromContext(getCtx)
		clock.Add(time.Nanosecond)
		return getCtx, assertDeduplicateUplink(ctx, env, getCtx, msg, ok, err)
	}
	assertJoinGetPeerSequence := func(ctx context.Context, env TestEnvironment, clock *test.MockClock, msg *ttnpb.UplinkMessage, dev *ttnpb.EndDevice, peer cluster.Peer, err error) (context.Context, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		getCtx, ok := assertJoinDeduplicateSequence(ctx, env, clock, msg, dev, true, nil)
		if !a.So(ok, should.BeTrue) {
			return nil, false
		}
		return getCtx, test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer, func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
			return AllTrue(
				a.So(ctx, should.HaveParentContextOrEqual, getCtx),
				a.So(role, should.Equal, ttnpb.ClusterRole_JOIN_SERVER),
				a.So(ids, should.Resemble, dev.EndDeviceIdentifiers),
			)
		},
			test.ClusterGetPeerResponse{
				Peer:  peer,
				Error: err,
			},
		)
	}
	assertJoinClusterLocalSequence := func(ctx context.Context, env TestEnvironment, clock *test.MockClock, msg *ttnpb.UplinkMessage, dev *ttnpb.EndDevice, joinReq *ttnpb.JoinRequest, joinResp *ttnpb.JoinResponse, err error) (context.Context, time.Time, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		getCtx, ok := assertJoinDeduplicateSequence(ctx, env, clock, msg, dev, true, nil)
		if !a.So(ok, should.BeTrue) {
			return nil, time.Time{}, false
		}
		joinReq.CorrelationIDs = msg.CorrelationIDs
		return getCtx, clock.Add(time.Nanosecond), assertClusterLocalJoin(ctx, env, getCtx, dev.EndDeviceIdentifiers, joinReq, joinResp, err)
	}
	assertJoinInteropSequence := func(ctx context.Context, env TestEnvironment, clock *test.MockClock, peerNotFound bool, msg *ttnpb.UplinkMessage, dev *ttnpb.EndDevice, joinReq *ttnpb.JoinRequest, joinResp *ttnpb.JoinResponse, err error) (context.Context, time.Time, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		var getCtx context.Context
		var ok bool
		if peerNotFound {
			getCtx, ok = assertJoinGetPeerSequence(ctx, env, clock, msg, dev, nil, ErrTestNotFound)
			joinReq.CorrelationIDs = msg.CorrelationIDs
		} else {
			getCtx, _, ok = assertJoinClusterLocalSequence(ctx, env, clock, msg, dev, joinReq, nil, ErrTestNotFound)
		}
		if !a.So(ok, should.BeTrue) {
			return nil, time.Time{}, false
		}
		return getCtx, clock.Add(time.Nanosecond), assertInteropJoin(ctx, env, getCtx, joinReq, joinResp, err)
	}
	assertPublishForwardJoinRequest := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers) bool {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)
		return a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
			return a.So(ev, should.ResembleEvent, EvtForwardJoinRequest(expectedCtx, expectedIDs, nil))
		}), should.BeTrue)
	}
	assertPublishDropJoinRequestLocalError := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers, expectedErr error) bool {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)
		return a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
			return a.So(ev, should.ResembleEvent, EvtDropJoinRequest(expectedCtx, expectedIDs, expectedErr))
		}), should.BeTrue)
	}
	assertPublishDropJoinRequestRPCError := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers, expectedErr error) bool {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)
		return a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
			err, ok := ev.Data().(errors.Interface)
			return AllTrue(
				a.So(ok, should.BeTrue),
				a.So(err, should.EqualErrorOrDefinition, expectedErr),
				a.So(ev, should.ResembleEvent, EvtDropJoinRequest(expectedCtx, expectedIDs, err)),
			)
		}), should.BeTrue)
	}

	assertDataRangeByAddr := func(ctx context.Context, env TestEnvironment, upCIDs []string, err error, getDevices ...*ttnpb.EndDevice) ([]context.Context, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		var rangeCtx context.Context
		var fCtxs []context.Context
		return fCtxs, AssertDeviceRegistryRangeByAddr(ctx, env.DeviceRegistry.RangeByAddr, func(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) bool {
			rangeCtx = ctx
			ctxCIDs := events.CorrelationIDsFromContext(ctx)
			for _, id := range upCIDs {
				a.So(ctxCIDs, should.Contain, id)
			}
			if !a.So(AllTrue(
				a.So(ctxCIDs, should.HaveLength, 2+len(upCIDs)),
				a.So(devAddr, should.Resemble, DevAddr),
				a.So(paths, should.HaveSameElementsDeep, dataRangeByAddrPaths[:]),
			), should.BeTrue) {
				return false
			}
			for i, getDevice := range getDevices {
				fCtx := context.WithValue(rangeCtx, &struct{}{}, fmt.Sprintf("range:%d", i))
				fCtxs = append(fCtxs, fCtx)
				a.So(f(
					fCtx,
					filterEndDevice(CopyEndDevice(getDevice), dataRangeByAddrPaths[:]...),
				), should.BeTrue)
			}
			return true
		},
			err,
		)
	}
	assertDataDeduplicateSequence := func(ctx context.Context, env TestEnvironment, clock *test.MockClock, msg *ttnpb.UplinkMessage, devs []*ttnpb.EndDevice, idx int, ok bool, err error) (context.Context, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		rangeCtxs, rangeOk := assertDataRangeByAddr(ctx, env, msg.CorrelationIDs, nil, devs...)
		if !a.So(rangeOk, should.BeTrue) || !a.So(len(rangeCtxs), should.BeGreaterThan, idx) {
			return nil, false
		}
		rangeCtx := rangeCtxs[idx]
		msg.ReceivedAt = clock.Now()
		msg.CorrelationIDs = events.CorrelationIDsFromContext(rangeCtx)
		clock.Add(time.Nanosecond)
		return rangeCtx, assertDeduplicateUplink(ctx, env, rangeCtx, msg, ok, err)
	}
	assertDataSetByID := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, getDevice, setDevice *ttnpb.EndDevice, expectedSets []string, expectedErr error, err error) (context.Context, bool) {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		var setCtx context.Context
		return setCtx, AssertDeviceRegistrySetByID(ctx, env.DeviceRegistry.SetByID, func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) bool {
			setCtx = ctx
			dev, sets, err := f(ctx, CopyEndDevice(getDevice))
			return AllTrue(
				a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
				a.So(appID, should.Resemble, AppID),
				a.So(devID, should.Resemble, DevID),
				a.So(gets, should.HaveSameElementsDeep, dataSetByIDGetPaths),
				a.So(sets, should.HaveSameElementsDeep, expectedSets),
				a.So(dev, should.ResembleFields, setDevice, sets),
				a.So(err, should.EqualErrorOrDefinition, expectedErr),
			)
		},
			func(ctx context.Context) DeviceRegistrySetByIDResponse {
				if setDevice != nil {
					setDevice = filterEndDevice(CopyEndDevice(setDevice), dataSetByIDGetPaths[:]...)
				}
				setCtx = context.WithValue(setCtx, struct{}{}, "set")
				return DeviceRegistrySetByIDResponse{
					Device:  setDevice,
					Context: setCtx,
					Error:   err,
				}
			})
	}
	assertDataApplicationUp := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, setDevice *ttnpb.EndDevice, msg *ttnpb.UplinkMessage, err error) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)
		macPayload := msg.Payload.GetMACPayload()
		return AssertApplicationUplinkQueueAddRequest(ctx, env.ApplicationUplinks.Add, func(ctx context.Context, ups ...*ttnpb.ApplicationUp) bool {
			return AllTrue(
				a.So(ctx, should.HaveParentContextOrEqual, expectedCtx),
				a.So(ups, should.Resemble, []*ttnpb.ApplicationUp{
					{
						CorrelationIDs:       events.CorrelationIDsFromContext(expectedCtx),
						EndDeviceIdentifiers: setDevice.EndDeviceIdentifiers,
						Up: &ttnpb.ApplicationUp_UplinkMessage{
							UplinkMessage: &ttnpb.ApplicationUplink{
								FCnt:         macPayload.FCnt,
								FPort:        macPayload.FPort,
								FRMPayload:   macPayload.FRMPayload,
								ReceivedAt:   msg.ReceivedAt,
								RxMetadata:   msg.RxMetadata,
								SessionKeyID: setDevice.Session.SessionKeyID,
								Settings:     msg.Settings,
							},
						},
					},
				}),
			)
		}, err)
	}
	assertPublishDropDataUplink := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers, expectedErr error) bool {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)
		return a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
			return a.So(ev, should.ResembleEvent, EvtDropDataUplink(expectedCtx, expectedIDs, expectedErr))
		}), should.BeTrue)
	}
	assertPublishDefinitionDataClosures := func(ctx context.Context, env TestEnvironment, expectedEvs []events.DefinitionDataClosure, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers) bool {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)
		return a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, len(expectedEvs), func(evs ...events.Event) bool {
			for i, expectedEv := range expectedEvs {
				if !a.So(evs[i], should.ResembleEvent, expectedEv(expectedCtx, expectedIDs)) {
					return false
				}
			}
			return true
		}), should.BeTrue)
	}
	assertPublishForwardDataUplink := func(ctx context.Context, env TestEnvironment, expectedCtx context.Context, expectedIDs ttnpb.EndDeviceIdentifiers) bool {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)
		return a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
			return a.So(ev, should.ResembleEvent, EvtForwardDataUplink(expectedCtx, expectedIDs, nil))
		}), should.BeTrue)
	}

	uplinkValidationErr, ok := errors.From((&ttnpb.UplinkMessage{}).ValidateFields())
	if !ok {
		t.Fatal("Failed to construct uplink validation error")
	}
	invalidUplinkSettingsErr := uplinkValidationErr.WithAttributes("field", "settings")

	makeChDRName := func(chIdx int, drIdx ttnpb.DataRateIndex, parts ...string) string {
		return MakeTestCaseName(append(parts, fmt.Sprintf("Channel:%d", chIdx), fmt.Sprintf("DR:%d", drIdx))...)
	}
	makeSessionKeys := func(macVersion ttnpb.MACVersion, withAppSKey bool) *ttnpb.SessionKeys {
		sk := &ttnpb.SessionKeys{
			FNwkSIntKey: &ttnpb.KeyEnvelope{
				Key: &FNwkSIntKey,
			},
			SessionKeyID: []byte("handle-uplink-test-session-key-id"),
		}
		if withAppSKey {
			sk.AppSKey = &ttnpb.KeyEnvelope{
				Key: &AppSKey,
			}
		}
		switch {
		case macVersion.Compare(ttnpb.MAC_V1_1) < 0:
			sk.NwkSEncKey = sk.FNwkSIntKey
			sk.SNwkSIntKey = sk.FNwkSIntKey
		default:
			sk.NwkSEncKey = &ttnpb.KeyEnvelope{
				Key: &NwkSEncKey,
			}
			sk.SNwkSIntKey = &ttnpb.KeyEnvelope{
				Key: &SNwkSIntKey,
			}
		}
		return CopySessionKeys(sk)
	}
	withMatchedUplinkSettings := func(chIdx int, drIdx ttnpb.DataRateIndex, msg *ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
		msg = CopyUplinkMessage(msg)
		msg.Settings.DataRateIndex = drIdx
		msg.DeviceChannelIndex = uint32(chIdx)
		return msg
	}

	type TestCase struct {
		Name    string
		Handler func(context.Context, TestEnvironment, *test.MockClock, func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool
	}
	tcs := []TestCase{
		{
			Name: "No settings",
			Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				return a.So(AllTrue(
					assertHandleUplinkResponse(ctx, handle(ctx, &ttnpb.UplinkMessage{}), invalidUplinkSettingsErr),
					assertHandleUplinkResponse(ctx, handle(ctx, &ttnpb.UplinkMessage{
						RawPayload: []byte("testpayload"),
						RxMetadata: RxMetadata[:1],
					}), invalidUplinkSettingsErr),
					assertHandleUplinkResponse(ctx, handle(ctx, &ttnpb.UplinkMessage{
						RawPayload: []byte("testpayload"),
					}), invalidUplinkSettingsErr),
				), should.BeTrue)
			},
		},
	}
	for _, uplinkMDs := range [][]*ttnpb.RxMetadata{
		nil,
		RxMetadata[0:1],
		append(RxMetadata[1:4]),
	} {
		uplinkMDs := uplinkMDs
		makeMDName := func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("Metadata length:%d", len(uplinkMDs)))...)
		}
		ForEachBand(t, func(makeLoopName func(...string) string, phy band.Band, phyVersion ttnpb.PHYVersion) {
			switch phyVersion {
			case ttnpb.PHY_V1_0_3_REV_A:
			case ttnpb.PHY_V1_1_REV_B, ttnpb.PHY_V1_0_2_REV_B:
				if testing.Short() {
					return
				}
			default:
				return
			}

			chIdx := len(phy.UplinkChannels) - 1
			ch := phy.UplinkChannels[chIdx]
			drIdx := ch.MaxDataRate
			dr := phy.DataRates[drIdx].Rate

			makeName := func(parts ...string) string {
				return makeMDName(makeChDRName(chIdx, drIdx, makeLoopName(parts...)))
			}
			tcs = append(tcs,
				TestCase{
					Name: makeName("No payload"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						return assertions.New(test.MustTFromContext(ctx)).So(assertHandleUplinkResponse(ctx, handle(ctx, &ttnpb.UplinkMessage{
							RxMetadata: uplinkMDs,
							Settings:   makeUplinkSettings(dr, ch),
						}), ErrDecodePayload.WithCause(lorawan.UnmarshalMessage(nil, nil))), should.BeTrue)
					},
				},
				TestCase{
					Name: makeName("Unknown Major"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						return assertions.New(test.MustTFromContext(ctx)).So(assertHandleUplinkResponse(ctx, handle(ctx, &ttnpb.UplinkMessage{
							RawPayload: []byte{
								/* MHDR */
								0b000_000_01,
								/* Join-request */
								/** JoinEUI **/
								JoinEUI[7], JoinEUI[6], JoinEUI[5], JoinEUI[4], JoinEUI[3], JoinEUI[2], JoinEUI[1], JoinEUI[0],
								/** DevEUI **/
								DevEUI[7], DevEUI[6], DevEUI[5], DevEUI[4], DevEUI[3], DevEUI[2], DevEUI[1], DevEUI[0],
								/** DevNonce **/
								0x01, 0x00,
								/* MIC */
								0x03, 0x02, 0x01, 0x00,
							},
							RxMetadata: uplinkMDs,
							Settings:   makeUplinkSettings(dr, ch),
						}), ErrUnsupportedLoRaWANVersion.WithAttributes("version", uint32(1))), should.BeTrue)
					},
				},
				TestCase{
					Name: makeName("Invalid MType"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						return assertions.New(test.MustTFromContext(ctx)).So(assertHandleUplinkResponse(ctx, handle(ctx, &ttnpb.UplinkMessage{
							RawPayload: bytes.Repeat([]byte{0x20}, 33),
							RxMetadata: uplinkMDs,
							Settings:   makeUplinkSettings(dr, ch),
						}), nil), should.BeTrue)
					},
				},
				TestCase{
					Name: makeName("Proprietary MType"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						phyPayload := []byte{
							/* MHDR */
							0b111_000_00,
						}
						return assertions.New(test.MustTFromContext(ctx)).So(assertHandleUplinkResponse(ctx, handle(ctx, &ttnpb.UplinkMessage{
							RawPayload: phyPayload,
							RxMetadata: uplinkMDs,
							Settings:   makeUplinkSettings(dr, ch),
						}), ErrDecodePayload.WithCause(lorawan.UnmarshalMessage(phyPayload, &ttnpb.Message{}))), should.BeTrue)
					},
				},
			)
		})

		ForEachFrequencyPlanBandMACVersion(t, func(makeLoopName func(...string) string, fpID string, fp *frequencyplans.FrequencyPlan, phy band.Band, phyVersion ttnpb.PHYVersion, macVersion ttnpb.MACVersion) {
			switch fpID {
			case test.EUFrequencyPlanID, test.USFrequencyPlanID:
			default:
				return
			}
			switch phyVersion {
			case ttnpb.PHY_V1_0_3_REV_A:
			case ttnpb.PHY_V1_1_REV_B, ttnpb.PHY_V1_0_2_REV_B:
				if testing.Short() {
					return
				}
			default:
				return
			}
			switch macVersion {
			case ttnpb.MAC_V1_0_4:
			case ttnpb.MAC_V1_1, ttnpb.MAC_V1_0_3:
				if testing.Short() {
					return
				}
			default:
				return
			}

			makeJoinResponse := func(withAppSKey bool) *ttnpb.JoinResponse {
				return &ttnpb.JoinResponse{
					RawPayload:  bytes.Repeat([]byte{0x42}, 17),
					SessionKeys: *makeSessionKeys(macVersion, withAppSKey),
				}
			}
			makeJoinDevice := func(clock *test.MockClock) *ttnpb.EndDevice {
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      fpID,
					LoRaWANPHYVersion:    phyVersion,
					LoRaWANVersion:       macVersion,
					MACSettings: &ttnpb.MACSettings{
						Rx1Delay: &ttnpb.RxDelayValue{
							Value: ttnpb.RX_DELAY_3,
						},
						DesiredRx2DataRateIndex: &ttnpb.DataRateIndexValue{
							Value: ttnpb.DATA_RATE_2,
						},
					},
					SupportsJoin: true,
					CreatedAt:    clock.Now(),
					UpdatedAt:    clock.Now(),
				}
			}

			chIdx := len(phy.UplinkChannels) - 1
			ch := phy.UplinkChannels[chIdx]
			drIdx := ch.MaxDataRate
			dr := phy.DataRates[drIdx].Rate

			makeJoinRequestDevNonce := func() types.DevNonce {
				return types.DevNonce{0x00, 0x01}
			}
			makeJoinRequestMIC := func() [4]byte {
				return [...]byte{0x03, 0x02, 0x01, 0x00}
			}
			makeJoinRequestPHYPayload := func() [23]byte {
				devNonce := makeJoinRequestDevNonce()
				mic := makeJoinRequestMIC()
				return [...]byte{
					/* MHDR */
					0b000_000_00,
					JoinEUI[7], JoinEUI[6], JoinEUI[5], JoinEUI[4], JoinEUI[3], JoinEUI[2], JoinEUI[1], JoinEUI[0],
					DevEUI[7], DevEUI[6], DevEUI[5], DevEUI[4], DevEUI[3], DevEUI[2], DevEUI[1], DevEUI[0],
					/* DevNonce */
					devNonce[1], devNonce[0],
					/* MIC */
					mic[0], mic[1], mic[2], mic[3],
				}
			}
			makeJoinRequestDecodedPayload := func() *ttnpb.Message {
				mic := makeJoinRequestMIC()
				devNonce := makeJoinRequestDevNonce()
				return &ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
						Major: ttnpb.Major_LORAWAN_R1,
					},
					MIC: mic[:],
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEUI:  *JoinEUI.Copy(&types.EUI64{}),
							DevEUI:   *DevEUI.Copy(&types.EUI64{}),
							DevNonce: devNonce,
						},
					},
				}
			}
			makeNsJsJoinRequest := func(devAddr *types.DevAddr, correlationIDs ...string) *ttnpb.JoinRequest {
				phyPayload := makeJoinRequestPHYPayload()
				return &ttnpb.JoinRequest{
					CFList:         frequencyplans.CFList(*fp, phyVersion),
					CorrelationIDs: correlationIDs,
					DevAddr: func() types.DevAddr {
						if devAddr != nil {
							return *devAddr
						} else {
							return types.DevAddr{}
						}
					}(),
					NetID:              *NetID.Copy(&types.NetID{}),
					RawPayload:         phyPayload[:],
					Payload:            makeJoinRequestDecodedPayload(),
					RxDelay:            ttnpb.RX_DELAY_3,
					SelectedMACVersion: macVersion,
					DownlinkSettings: ttnpb.DLSettings{
						OptNeg: macVersion.Compare(ttnpb.MAC_V1_1) >= 0,
						Rx2DR:  ttnpb.DATA_RATE_2,
					},
				}
			}
			makeJoinSetDevice := func(getDevice *ttnpb.EndDevice, decodedMsg *ttnpb.UplinkMessage, joinReq *ttnpb.JoinRequest, joinResp *ttnpb.JoinResponse) *ttnpb.EndDevice {
				macState := test.Must(NewMACState(getDevice, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})).(*ttnpb.MACState)
				macState.RxWindowsAvailable = true
				macState.QueuedJoinAccept = &ttnpb.MACState_JoinAccept{
					Keys:    joinResp.SessionKeys,
					Payload: joinResp.RawPayload,
					Request: *joinReq,
				}
				setDevice := CopyEndDevice(getDevice)
				setDevice.RecentUplinks = AppendRecentUplink(setDevice.RecentUplinks, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), RecentUplinkCount)
				setDevice.PendingMACState = macState
				return setDevice
			}

			makeJoinRequest := func(decodePayload bool) *ttnpb.UplinkMessage {
				phyPayload := makeJoinRequestPHYPayload()
				msg := &ttnpb.UplinkMessage{
					CorrelationIDs: []string{
						"join-request-correlation-id-1",
						"join-request-correlation-id-2",
						"join-request-correlation-id-3",
					},
					RawPayload: phyPayload[:],
					RxMetadata: uplinkMDs,
					Settings:   makeUplinkSettings(dr, ch),
				}
				if decodePayload {
					msg.Payload = makeJoinRequestDecodedPayload()
				}
				return msg
			}

			makeJoinName := func(parts ...string) string {
				return makeMDName(makeChDRName(chIdx, drIdx, makeLoopName(append([]string{"Join-request"}, parts...)...)))
			}
			tcs = append(tcs,
				TestCase{
					Name: makeJoinName("Get fail"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							_, ok := assertJoinGetByEUI(ctx, env, msg.CorrelationIDs, nil, ErrTestInternal)
							if !a.So(ok, should.BeTrue) {
								return false
							}
							return ok
						}, ErrTestInternal), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get ABP device", "Deduplication fail"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						getDevice.SupportsJoin = false
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, ok := assertJoinDeduplicateSequence(ctx, env, clock, decodedMsg, CopyEndDevice(getDevice), false, ErrTestInternal)
							return AllTrue(
								ok,
								assertPublishDropJoinRequestLocalError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, ErrTestInternal),
							)
						}, ErrTestInternal), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "Deduplication fail"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, ok := assertJoinDeduplicateSequence(ctx, env, clock, decodedMsg, CopyEndDevice(getDevice), false, ErrTestInternal)
							return AllTrue(
								ok,
								assertPublishDropJoinRequestLocalError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, ErrTestInternal),
							)
						}, ErrTestInternal), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get ABP device", "Duplicate uplink"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						getDevice.SupportsJoin = false
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							_, ok := assertJoinDeduplicateSequence(ctx, env, clock, decodedMsg, CopyEndDevice(getDevice), false, nil)
							return ok
						}, nil), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "Duplicate uplink"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							_, ok := assertJoinDeduplicateSequence(ctx, env, clock, decodedMsg, CopyEndDevice(getDevice), false, nil)
							return ok
						}, nil), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get ABP device", "First uplink"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						getDevice.SupportsJoin = false
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, ok := assertJoinDeduplicateSequence(ctx, env, clock, decodedMsg, CopyEndDevice(getDevice), true, nil)
							return AllTrue(
								ok,
								assertPublishDropJoinRequestLocalError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, ErrABPJoinRequest),
							)
						}, ErrABPJoinRequest), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "First uplink", "Cluster-local JS fail"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, _, ok := assertJoinClusterLocalSequence(ctx, env, clock, decodedMsg, CopyEndDevice(getDevice), makeNsJsJoinRequest(nil), nil, ErrTestInternal)
							return AllTrue(
								ok,
								assertPublishDropJoinRequestRPCError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, ErrTestInternal),
							)
						}, ErrTestInternal), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "First uplink", "Cluster-local JS not found", "Interop JS fail"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, _, ok := assertJoinInteropSequence(ctx, env, clock, true, decodedMsg, CopyEndDevice(getDevice), makeNsJsJoinRequest(nil), nil, ErrTestInternal)
							return AllTrue(
								ok,
								assertPublishDropJoinRequestRPCError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, ErrTestInternal),
							)
						}, ErrTestInternal), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "First uplink", "Cluster-local JS does not contain device", "Interop JS accept", "Metadata merge fail", "Downlink queue present", "Set fail on read"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						getDevice.Session = &ttnpb.Session{
							DevAddr: DevAddr,
							QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
								makeApplicationDownlink(),
							},
						}
						getDevice.DevAddr = &getDevice.Session.DevAddr
						joinResp := makeJoinResponse(true)
						joinReq := makeNsJsJoinRequest(nil)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, _, ok := assertJoinInteropSequence(ctx, env, clock, false, decodedMsg, CopyEndDevice(getDevice), joinReq, joinResp, nil)
							if !a.So(AllTrue(
								ok,
								assertPublishForwardJoinRequest(ctx, env, getCtx, getDevice.EndDeviceIdentifiers),
							), should.BeTrue) {
								return false
							}
							clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
							if !a.So(assertAccumulatedMetadata(ctx, env, getCtx, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), nil, ErrTestInternal), should.BeTrue) {
								return true
							}
							return a.So(AllTrue(AssertDeviceRegistrySetByID(ctx, env.DeviceRegistry.SetByID, func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) bool {
								return AllTrue(
									a.So(ctx, should.HaveParentContextOrEqual, getCtx),
									a.So(appID, should.Resemble, AppID),
									a.So(devID, should.Resemble, DevID),
									a.So(gets, should.HaveSameElementsDeep, joinSetByIDGetPaths),
								)
							}, func(context.Context) DeviceRegistrySetByIDResponse {
								return DeviceRegistrySetByIDResponse{
									Error: ErrTestInternal,
								}
							}),
								assertPublishDropJoinRequestLocalError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, ErrTestInternal),
							), should.BeTrue)
						}, ErrTestInternal), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "First uplink", "Cluster-local JS does not contain device", "Interop JS accept", "Metadata merge fail", "Downlink queue present", "Device deleted during handling"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						getDevice.Session = &ttnpb.Session{
							DevAddr: DevAddr,
							QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
								makeApplicationDownlink(),
							},
						}
						getDevice.DevAddr = &getDevice.Session.DevAddr
						joinResp := makeJoinResponse(true)
						joinReq := makeNsJsJoinRequest(nil)
						innerErr := ErrOutdatedData
						registryErr := ErrTestInternal.WithCause(innerErr)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, _, ok := assertJoinInteropSequence(ctx, env, clock, false, decodedMsg, CopyEndDevice(getDevice), joinReq, joinResp, nil)
							if !a.So(AllTrue(
								ok,
								assertPublishForwardJoinRequest(ctx, env, getCtx, getDevice.EndDeviceIdentifiers),
							), should.BeTrue) {
								return false
							}
							clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
							if !a.So(assertAccumulatedMetadata(ctx, env, getCtx, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), nil, ErrTestInternal), should.BeTrue) {
								return true
							}
							_, ok = assertJoinSetByID(ctx, env, getCtx, nil, nil, innerErr, registryErr)
							return AllTrue(
								ok,
								assertPublishDropJoinRequestLocalError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, registryErr),
							)
						}, registryErr), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "First uplink", "Cluster-local JS does not contain device", "Interop JS accept", "Metadata merge fail", "Downlink queue present", "Set fail on write"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						getDevice.Session = &ttnpb.Session{
							DevAddr: DevAddr,
							QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
								makeApplicationDownlink(),
							},
						}
						getDevice.DevAddr = &getDevice.Session.DevAddr
						joinResp := makeJoinResponse(true)
						joinReq := makeNsJsJoinRequest(nil)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, _, ok := assertJoinInteropSequence(ctx, env, clock, false, decodedMsg, CopyEndDevice(getDevice), joinReq, joinResp, nil)
							if !a.So(AllTrue(
								ok,
								assertPublishForwardJoinRequest(ctx, env, getCtx, getDevice.EndDeviceIdentifiers),
							), should.BeTrue) {
								return false
							}
							clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
							if !a.So(assertAccumulatedMetadata(ctx, env, getCtx, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), nil, ErrTestInternal), should.BeTrue) {
								return true
							}
							_, ok = assertJoinSetByID(ctx, env, getCtx, getDevice, makeJoinSetDevice(getDevice, decodedMsg, joinReq, joinResp), nil, ErrTestInternal)
							return AllTrue(
								ok,
								assertPublishDropJoinRequestLocalError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, ErrTestInternal),
							)
						}, ErrTestInternal), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "First uplink", "Cluster-local JS does not contain device", "Interop JS accept", "Metadata merge fail", "Downlink queue present", "Set success", "Downlink add success", "AppSKey present", "Application uplink add success"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						getDevice.Session = &ttnpb.Session{
							DevAddr: DevAddr,
							QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
								makeApplicationDownlink(),
							},
						}
						getDevice.DevAddr = &getDevice.Session.DevAddr
						joinResp := makeJoinResponse(true)
						joinReq := makeNsJsJoinRequest(nil)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, joinRespRecvAt, ok := assertJoinInteropSequence(ctx, env, clock, false, decodedMsg, CopyEndDevice(getDevice), joinReq, joinResp, nil)
							if !a.So(AllTrue(
								ok,
								assertPublishForwardJoinRequest(ctx, env, getCtx, getDevice.EndDeviceIdentifiers),
							), should.BeTrue) {
								return false
							}
							clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
							if !a.So(assertAccumulatedMetadata(ctx, env, getCtx, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), nil, ErrTestInternal), should.BeTrue) {
								return true
							}
							setDevice := makeJoinSetDevice(getDevice, decodedMsg, joinReq, joinResp)
							setCtx, ok := assertJoinSetByID(ctx, env, getCtx, getDevice, setDevice, nil, nil)
							return AllTrue(
								ok,
								assertDownlinkTaskAdd(ctx, env, setCtx, setDevice.EndDeviceIdentifiers, decodedMsg.ReceivedAt.Add(-InfrastructureDelay/2+phy.JoinAcceptDelay1-joinReq.RxDelay.Duration()/2-NSScheduleWindow()), true, nil),
								assertJoinApplicationUp(ctx, env, setCtx, setDevice, joinRespRecvAt, nil),
							)
						}, nil), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "First uplink", "Cluster-local JS accept", "Metadata merge success", "No downlink queue", "Set success", "Downlink add fail", "No AppSKey", "Application uplink add fail"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						joinResp := makeJoinResponse(false)
						joinReq := makeNsJsJoinRequest(nil)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, joinRespRecvAt, ok := assertJoinClusterLocalSequence(ctx, env, clock, decodedMsg, CopyEndDevice(getDevice), joinReq, joinResp, nil)
							if !a.So(AllTrue(
								ok,
								assertPublishForwardJoinRequest(ctx, env, getCtx, getDevice.EndDeviceIdentifiers),
							), should.BeTrue) {
								return false
							}
							clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
							if !a.So(assertAccumulatedMetadata(ctx, env, getCtx, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), RxMetadata[:], nil), should.BeTrue) {
								return true
							}
							decodedMsg.RxMetadata = RxMetadata[:]
							setDevice := makeJoinSetDevice(getDevice, decodedMsg, joinReq, joinResp)
							setCtx, ok := assertJoinSetByID(ctx, env, getCtx, getDevice, setDevice, nil, nil)
							return a.So(AllTrue(
								ok,
								assertDownlinkTaskAdd(ctx, env, setCtx, setDevice.EndDeviceIdentifiers, decodedMsg.ReceivedAt.Add(-InfrastructureDelay/2+phy.JoinAcceptDelay1-joinReq.RxDelay.Duration()/2-NSScheduleWindow()), true, ErrTestInternal),
								assertJoinApplicationUp(ctx, env, setCtx, setDevice, joinRespRecvAt, ErrTestInternal),
								assertPublishMergeMetadata(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, RxMetadata[:]...),
							), should.BeTrue)
						}, nil), should.BeTrue)
					},
				},
				TestCase{
					Name: makeJoinName("Get OTAA device", "First uplink", "Cluster-local JS accept", "Metadata merge success", "No downlink queue", "Set fail on write"),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						t := test.MustTFromContext(ctx)
						a := assertions.New(t)
						msg := makeJoinRequest(false)
						decodedMsg := makeJoinRequest(true)
						getDevice := makeJoinDevice(clock)
						joinResp := makeJoinResponse(false)
						joinReq := makeNsJsJoinRequest(nil)
						return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
							getCtx, _, ok := assertJoinClusterLocalSequence(ctx, env, clock, decodedMsg, CopyEndDevice(getDevice), joinReq, joinResp, nil)
							if !a.So(AllTrue(
								ok,
								assertPublishForwardJoinRequest(ctx, env, getCtx, getDevice.EndDeviceIdentifiers),
							), should.BeTrue) {
								return false
							}
							clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
							if !a.So(assertAccumulatedMetadata(ctx, env, getCtx, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), RxMetadata[:], nil), should.BeTrue) {
								return true
							}
							decodedMsg.RxMetadata = RxMetadata[:]
							setDevice := makeJoinSetDevice(getDevice, decodedMsg, joinReq, joinResp)
							_, ok = assertJoinSetByID(ctx, env, getCtx, getDevice, setDevice, nil, ErrTestInternal)
							return a.So(AllTrue(
								ok,
								assertPublishDropJoinRequestLocalError(ctx, env, getCtx, getDevice.EndDeviceIdentifiers, ErrTestInternal),
							), should.BeTrue)
						}, ErrTestInternal), should.BeTrue)
					},
				},
			)

			for _, typ := range []ttnpb.RejoinType{
				ttnpb.RejoinType_CONTEXT,
				ttnpb.RejoinType_KEYS,
				ttnpb.RejoinType_SESSION,
			} {
				typ := typ
				if macVersion.Compare(ttnpb.MAC_V1_1) < 0 {
					continue
				}
				makeRejoinRequest := func(decodePayload bool) *ttnpb.UplinkMessage {
					var phyPayload []byte
					switch typ {
					case ttnpb.RejoinType_CONTEXT, ttnpb.RejoinType_KEYS:
						phyPayload = []byte{
							/* MHDR */
							0b110_000_00,
							byte(typ),
							NetID[2], NetID[1], NetID[0],
							DevEUI[7], DevEUI[6], DevEUI[5], DevEUI[4], DevEUI[3], DevEUI[2], DevEUI[1], DevEUI[0],
							/* RejoinCnt0 */
							0x01, 0x00,
							/* MIC */
							0x03, 0x02, 0x01, 0x00,
						}
					case ttnpb.RejoinType_SESSION:
						phyPayload = []byte{
							/* MHDR */
							0b110_000_00,
							byte(typ),
							JoinEUI[7], JoinEUI[6], JoinEUI[5], JoinEUI[4], JoinEUI[3], JoinEUI[2], JoinEUI[1], JoinEUI[0],
							DevEUI[7], DevEUI[6], DevEUI[5], DevEUI[4], DevEUI[3], DevEUI[2], DevEUI[1], DevEUI[0],
							/* RejoinCnt1 */
							0x01, 0x00,
							/* MIC */
							0x03, 0x02, 0x01, 0x00,
						}
					default:
						panic(fmt.Sprintf("unknown rejoin type `%d`", typ))
					}
					msg := &ttnpb.UplinkMessage{
						CorrelationIDs: []string{
							"rejoin-request-correlation-id-1",
							"rejoin-request-correlation-id-2",
							"rejoin-request-correlation-id-3",
						},
						RawPayload: phyPayload,
						RxMetadata: uplinkMDs,
						Settings:   makeUplinkSettings(dr, ch),
					}
					if decodePayload {
						var pld *ttnpb.RejoinRequestPayload
						switch typ {
						case ttnpb.RejoinType_CONTEXT, ttnpb.RejoinType_KEYS:
							pld = &ttnpb.RejoinRequestPayload{
								DevEUI:     *DevEUI.Copy(&types.EUI64{}),
								NetID:      *NetID.Copy(&types.NetID{}),
								RejoinCnt:  uint32(binary.LittleEndian.Uint16(phyPayload[13:14])),
								RejoinType: typ,
							}
						case ttnpb.RejoinType_SESSION:
							pld = &ttnpb.RejoinRequestPayload{
								JoinEUI:    *JoinEUI.Copy(&types.EUI64{}),
								DevEUI:     *DevEUI.Copy(&types.EUI64{}),
								RejoinCnt:  uint32(binary.LittleEndian.Uint16(phyPayload[18:19])),
								RejoinType: typ,
							}
						}
						msg.Payload = &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_REJOIN_REQUEST,
								Major: ttnpb.Major_LORAWAN_R1,
							},
							MIC: phyPayload[len(phyPayload)-4:],
							Payload: &ttnpb.Message_RejoinRequestPayload{
								RejoinRequestPayload: pld,
							},
						}
					}
					return msg
				}

				tcs = append(tcs, TestCase{
					Name: makeMDName(makeChDRName(chIdx, drIdx, makeLoopName(append([]string{fmt.Sprintf("Rejoin-request Type %d", typ)})...))),
					Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
						return assertions.New(test.MustTFromContext(ctx)).So(assertHandleUplinkResponse(ctx, handle(ctx, makeRejoinRequest(false)), ErrRejoinRequest), should.BeTrue)
					},
				})
			}
		})

		ForEachMACVersion(func(makeLoopName func(...string) string, macVersion ttnpb.MACVersion) {
			switch macVersion {
			case ttnpb.MAC_V1_0_4:
			case ttnpb.MAC_V1_1, ttnpb.MAC_V1_0_3:
				if testing.Short() {
					return
				}
			default:
				return
			}

			fpID := test.EUFrequencyPlanID
			phyVersion := ttnpb.PHY_V1_1_REV_B
			fp := test.Must(frequencyplans.NewStore(test.FrequencyPlansFetcher).GetByID(fpID)).(*frequencyplans.FrequencyPlan)
			phy := test.Must(test.Must(band.GetByID(fp.BandID)).(band.Band).Version(phyVersion)).(band.Band)
			chIdx := len(phy.UplinkChannels) - 1
			ch := phy.UplinkChannels[chIdx]
			drIdx := ch.MaxDataRate
			dr := phy.DataRates[drIdx].Rate

			makeDataRangeDevice := func(clock *test.MockClock, useADR bool) *ttnpb.EndDevice {
				sk := *makeSessionKeys(macVersion, false)
				dev := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(&DevAddr),
					FrequencyPlanID:      fpID,
					LoRaWANPHYVersion:    phyVersion,
					LoRaWANVersion:       macVersion,
					MACSettings: &ttnpb.MACSettings{
						UseADR: &pbtypes.BoolValue{
							Value: useADR,
						},
						Rx1Delay: &ttnpb.RxDelayValue{
							Value: Rx1Delay,
						},
					},
					Session: &ttnpb.Session{
						DevAddr:     *DevAddr.Copy(&types.DevAddr{}),
						SessionKeys: sk,
						LastFCntUp:  FCnt - 1,
						StartedAt:   clock.Now(),
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								SessionKeyID: sk.SessionKeyID,
								FPort:        FPort,
								FCnt:         42,
							},
						},
					},
					CreatedAt: clock.Now(),
					UpdatedAt: clock.Now(),
				}
				dev.MACState = test.Must(NewMACState(dev, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})).(*ttnpb.MACState)
				dev.MACState.CurrentParameters.ADRNbTrans = 2
				return dev
			}

			const matchIdx = 2
			makeDataRangeDevices := func(clock *test.MockClock, withMatch bool, useADR bool) []*ttnpb.EndDevice {
				withLastFCntUp := func(dev *ttnpb.EndDevice, fCnt uint32) *ttnpb.EndDevice {
					ret := CopyEndDevice(dev)
					ret.Session.LastFCntUp = fCnt
					return ret
				}
				match := makeDataRangeDevice(clock, useADR)
				rets := []*ttnpb.EndDevice{
					withLastFCntUp(match, FCnt+2),
					withLastFCntUp(match, FCnt+25),
				}
				if withMatch {
					rets = append(rets, match)
				}
				return append(rets,
					withLastFCntUp(match, FCnt+42),
				)
			}

			for _, confirmed := range [2]bool{true, false} {
				confirmed := confirmed
				makeDataUplink := func(decodePayload bool, adr bool, frmPayload []byte) *ttnpb.UplinkMessage {
					mType := ttnpb.MType_UNCONFIRMED_UP
					if confirmed {
						mType = ttnpb.MType_CONFIRMED_UP
					}
					mhdr := ttnpb.MHDR{
						MType: mType,
						Major: ttnpb.Major_LORAWAN_R1,
					}
					fOpts := []byte{0x02}
					if macVersion.EncryptFOpts() {
						fOpts = MustEncryptUplink(NwkSEncKey, DevAddr, FCnt, fOpts...)
					}
					fhdr := ttnpb.FHDR{
						DevAddr: *DevAddr.Copy(&types.DevAddr{}),
						FCtrl: ttnpb.FCtrl{
							ADR: adr,
						},
						FCnt:  FCnt,
						FOpts: fOpts,
					}
					phyPayload := append(
						append(
							test.Must(lorawan.AppendFHDR(
								test.Must(lorawan.AppendMHDR(nil, mhdr)).([]byte), fhdr, true),
							).([]byte),
							FPort),
						frmPayload...)
					switch {
					case macVersion.Compare(ttnpb.MAC_V1_1) < 0:
						phyPayload = MustAppendLegacyUplinkMIC(
							FNwkSIntKey,
							DevAddr,
							FCnt,
							phyPayload...,
						)
					default:
						phyPayload = MustAppendUplinkMIC(
							SNwkSIntKey,
							FNwkSIntKey,
							0,
							uint8(drIdx),
							uint8(chIdx),
							DevAddr,
							FCnt,
							phyPayload...,
						)
					}
					msg := &ttnpb.UplinkMessage{
						CorrelationIDs: []string{
							"data-uplink-correlation-id-1",
							"data-uplink-correlation-id-2",
							"data-uplink-correlation-id-3",
						},
						RawPayload: phyPayload,
						RxMetadata: uplinkMDs,
						Settings:   makeUplinkSettings(dr, ch),
					}
					if decodePayload {
						if frmPayload == nil {
							frmPayload = []byte{}
						}
						msg.Payload = &ttnpb.Message{
							MHDR: mhdr,
							MIC:  phyPayload[len(phyPayload)-4:],
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR:       fhdr,
									FPort:      FPort,
									FRMPayload: frmPayload,
								},
							},
						}
					}
					return msg
				}
				dataSetByIDSetPaths := [...]string{
					"mac_state",
					"pending_mac_state",
					"pending_session",
					"recent_adr_uplinks",
					"recent_uplinks",
					"session",
				}
				makeDataSetDevice := func(getDevice *ttnpb.EndDevice, decodedMsg *ttnpb.UplinkMessage) (*ttnpb.EndDevice, []events.DefinitionDataClosure) {
					setDevice := CopyEndDevice(getDevice)
					setDevice.MACState.QueuedResponses = nil
					evs := test.Must(HandleLinkCheckReq(test.Context(), setDevice, decodedMsg)).([]events.DefinitionDataClosure)
					setDevice.MACState.RecentUplinks = AppendRecentUplink(setDevice.MACState.RecentUplinks, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), RecentUplinkCount)
					setDevice.MACState.RxWindowsAvailable = true
					setDevice.RecentUplinks = AppendRecentUplink(setDevice.RecentUplinks, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), RecentUplinkCount)
					setDevice.Session.LastFCntUp = FCnt
					if decodedMsg.Payload.GetMACPayload().ADR {
						setDevice.RecentADRUplinks = AppendRecentUplink(setDevice.RecentADRUplinks, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), OptimalADRUplinkCount)
						test.Must(nil, AdaptDataRate(setDevice, phy, ttnpb.MACSettings{}))
					}
					return setDevice, evs
				}

				makeName := func(parts ...string) string {
					mTypeStr := "Unconfirmed Data"
					if confirmed {
						mTypeStr = "Confirmed Data"
					}
					return makeMDName(makeChDRName(chIdx, drIdx, makeLoopName(append([]string{mTypeStr}, parts...)...)))
				}
				for _, conf := range []struct {
					MakeName       func(...string) string
					MakeDataUplink func(bool) *ttnpb.UplinkMessage
				}{
					{
						MakeName: func(parts ...string) string {
							return makeName(append(parts, "No ADR", "No FRMPayload")...)
						},
						MakeDataUplink: func(decoded bool) *ttnpb.UplinkMessage {
							return makeDataUplink(decoded, false, nil)
						},
					},
					{
						MakeName: func(parts ...string) string {
							return makeName(append(parts, "No ADR", "FRMPayload")...)
						},
						MakeDataUplink: func(decoded bool) *ttnpb.UplinkMessage {
							return makeDataUplink(decoded, false, []byte("test"))
						},
					},
					{
						MakeName: func(parts ...string) string {
							return makeName(append(parts, "ADR", "No FRMPayload")...)
						},
						MakeDataUplink: func(decoded bool) *ttnpb.UplinkMessage {
							return makeDataUplink(decoded, true, nil)
						},
					},
					{
						MakeName: func(parts ...string) string {
							return makeName(append(parts, "ADR", "FRMPayload")...)
						},
						MakeDataUplink: func(decoded bool) *ttnpb.UplinkMessage {
							return makeDataUplink(decoded, true, []byte("test"))
						},
					},
				} {
					conf := conf
					tcs = append(tcs,
						TestCase{
							Name: conf.MakeName("Range fail"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									_, ok := assertDataRangeByAddr(ctx, env, msg.CorrelationIDs, ErrTestInternal)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									return ok
								}, ErrTestInternal), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "No devices"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									_, ok := assertDataRangeByAddr(ctx, env, msg.CorrelationIDs, nil)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									return ok
								}, ErrDeviceNotFound), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "No match"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									_, ok := assertDataRangeByAddr(ctx, env, msg.CorrelationIDs, nil, makeDataRangeDevices(clock, false, true)...)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									return ok
								}, ErrDeviceNotFound), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "Deduplication fail"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								decodedMsg := conf.MakeDataUplink(true)
								rangeDevices := makeDataRangeDevices(clock, true, true)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									rangeCtx, ok := assertDataDeduplicateSequence(ctx, env, clock, decodedMsg, rangeDevices, matchIdx, false, ErrTestInternal)
									return a.So(AllTrue(
										ok,
										assertPublishDropDataUplink(ctx, env, rangeCtx, rangeDevices[matchIdx].EndDeviceIdentifiers, ErrTestInternal),
									), should.BeTrue)
								}, ErrTestInternal), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "Duplicate uplink"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								decodedMsg := conf.MakeDataUplink(true)
								rangeDevices := makeDataRangeDevices(clock, true, true)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									_, ok := assertDataDeduplicateSequence(ctx, env, clock, decodedMsg, rangeDevices, matchIdx, false, nil)
									return a.So(ok, should.BeTrue)
								}, nil), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "First uplink", "Metadata merge fail", "Set fail on read"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								decodedMsg := conf.MakeDataUplink(true)
								rangeDevices := makeDataRangeDevices(clock, true, true)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									rangeCtx, ok := assertDataDeduplicateSequence(ctx, env, clock, decodedMsg, rangeDevices, matchIdx, true, nil)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
									return a.So(AllTrue(
										assertAccumulatedMetadata(ctx, env, rangeCtx, decodedMsg, nil, ErrTestInternal),
										AssertDeviceRegistrySetByID(ctx, env.DeviceRegistry.SetByID, func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) bool {
											return AllTrue(
												a.So(ctx, should.HaveParentContextOrEqual, rangeCtx),
												a.So(appID, should.Resemble, AppID),
												a.So(devID, should.Resemble, DevID),
												a.So(gets, should.HaveSameElementsDeep, dataSetByIDGetPaths),
											)
										}, func(context.Context) DeviceRegistrySetByIDResponse {
											return DeviceRegistrySetByIDResponse{
												Error: ErrTestInternal,
											}
										}),
										assertPublishDropDataUplink(ctx, env, rangeCtx, rangeDevices[matchIdx].EndDeviceIdentifiers, ErrTestInternal),
									), should.BeTrue)
								}, ErrTestInternal), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "First uplink", "Metadata merge success", "Device deleted during handling"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								decodedMsg := conf.MakeDataUplink(true)
								rangeDevices := makeDataRangeDevices(clock, true, true)
								innerErr := ErrOutdatedData
								registryErr := ErrTestInternal.WithCause(innerErr)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									rangeCtx, ok := assertDataDeduplicateSequence(ctx, env, clock, decodedMsg, rangeDevices, matchIdx, true, nil)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
									if !a.So(assertAccumulatedMetadata(ctx, env, rangeCtx, decodedMsg, RxMetadata[:], nil), should.BeTrue) {
										return false
									}
									_, ok = assertDataSetByID(ctx, env, rangeCtx, nil, nil, nil, innerErr, registryErr)
									return a.So(AllTrue(
										ok,
										assertPublishDropDataUplink(ctx, env, rangeCtx, rangeDevices[matchIdx].EndDeviceIdentifiers, registryErr),
									), should.BeTrue)
								}, registryErr), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "First uplink", "Metadata merge success", "Rematch fail"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								decodedMsg := conf.MakeDataUplink(true)
								rangeDevices := makeDataRangeDevices(clock, true, true)
								innerErr := ErrOutdatedData.WithCause(ErrDeviceNotFound)
								registryErr := ErrTestInternal.WithCause(innerErr)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									rangeCtx, ok := assertDataDeduplicateSequence(ctx, env, clock, decodedMsg, rangeDevices, matchIdx, true, nil)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
									if !a.So(assertAccumulatedMetadata(ctx, env, rangeCtx, decodedMsg, RxMetadata[:], nil), should.BeTrue) {
										return false
									}
									decodedMsg.RxMetadata = RxMetadata[:]
									getDevice := CopyEndDevice(rangeDevices[matchIdx])
									getDevice.UpdatedAt = clock.Now()
									getDevice.MACState = nil
									getDevice.Session = nil
									_, ok = assertDataSetByID(ctx, env, rangeCtx, getDevice, nil, nil, innerErr, registryErr)
									return a.So(AllTrue(
										ok,
										assertPublishDropDataUplink(ctx, env, rangeCtx, rangeDevices[matchIdx].EndDeviceIdentifiers, registryErr),
									), should.BeTrue)
								}, registryErr), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "First uplink", "Metadata merge fail", "No rematch", "Set success", "NbTrans=1", "Downlink add fail", "Application uplink add fail"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								decodedMsg := conf.MakeDataUplink(true)
								rangeDevices := makeDataRangeDevices(clock, true, true)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									rangeCtx, ok := assertDataDeduplicateSequence(ctx, env, clock, decodedMsg, rangeDevices, matchIdx, true, nil)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
									if !a.So(assertAccumulatedMetadata(ctx, env, rangeCtx, decodedMsg, nil, ErrTestInternal), should.BeTrue) {
										return false
									}
									getDevice := CopyEndDevice(rangeDevices[matchIdx])
									setDevice, macEvs := makeDataSetDevice(getDevice, decodedMsg)
									setCtx, ok := assertDataSetByID(ctx, env, rangeCtx, getDevice, setDevice, dataSetByIDSetPaths[:], nil, nil)
									return a.So(AllTrue(
										ok,
										assertDownlinkTaskAdd(ctx, env, setCtx, setDevice.EndDeviceIdentifiers, decodedMsg.ReceivedAt.Add(-InfrastructureDelay/2+Rx1Delay.Duration()/2-NSScheduleWindow()), true, ErrTestInternal),
										assertDataApplicationUp(ctx, env, setCtx, setDevice, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), ErrTestInternal),
										assertPublishDefinitionDataClosures(ctx, env, macEvs, setCtx, setDevice.EndDeviceIdentifiers),
									), should.BeTrue)
								}, nil), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "First uplink", "Metadata merge success", "Rematch success", "Set success", "NbTrans=1", "Downlink add success", "Application uplink add success"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								decodedMsg := conf.MakeDataUplink(true)
								rangeDevices := makeDataRangeDevices(clock, true, true)
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									rangeCtx, ok := assertDataDeduplicateSequence(ctx, env, clock, decodedMsg, rangeDevices, matchIdx, true, nil)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
									if !a.So(assertAccumulatedMetadata(ctx, env, rangeCtx, decodedMsg, RxMetadata[:], nil), should.BeTrue) {
										return false
									}
									decodedMsg.RxMetadata = RxMetadata[:]
									getDevice := CopyEndDevice(rangeDevices[matchIdx])
									getDevice.UpdatedAt = clock.Now()
									setDevice, macEvs := makeDataSetDevice(getDevice, decodedMsg)
									setCtx, ok := assertDataSetByID(ctx, env, rangeCtx, getDevice, setDevice, dataSetByIDSetPaths[:], nil, nil)
									return a.So(AllTrue(
										ok,
										assertDownlinkTaskAdd(ctx, env, setCtx, setDevice.EndDeviceIdentifiers, decodedMsg.ReceivedAt.Add(-InfrastructureDelay/2+Rx1Delay.Duration()/2-NSScheduleWindow()), true, nil),
										assertDataApplicationUp(ctx, env, setCtx, setDevice, withMatchedUplinkSettings(chIdx, drIdx, decodedMsg), nil),
										assertPublishMergeMetadata(ctx, env, setCtx, setDevice.EndDeviceIdentifiers, RxMetadata[:]...),
										assertPublishDefinitionDataClosures(ctx, env, macEvs, setCtx, setDevice.EndDeviceIdentifiers),
										assertPublishForwardDataUplink(ctx, env, setCtx, setDevice.EndDeviceIdentifiers),
									), should.BeTrue)
								}, nil), should.BeTrue)
							},
						},
						TestCase{
							Name: conf.MakeName("Range success", "First uplink", "Metadata merge success", "No rematch", "Set success", "NbTrans=2", "Downlink add success"),
							Handler: func(ctx context.Context, env TestEnvironment, clock *test.MockClock, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)
								msg := conf.MakeDataUplink(false)
								decodedMsg := conf.MakeDataUplink(true)
								rangeDevices := makeDataRangeDevices(clock, true, true)
								prevMsg := CopyUplinkMessage(decodedMsg)
								prevMsg.ReceivedAt = clock.Now()
								rangeDevice, _ := makeDataSetDevice(rangeDevices[matchIdx], prevMsg)
								rangeDevices[matchIdx] = rangeDevice
								return a.So(assertHandleUplink(ctx, handle, msg, func() bool {
									rangeCtx, ok := assertDataDeduplicateSequence(ctx, env, clock, decodedMsg, rangeDevices, matchIdx, true, nil)
									if !a.So(ok, should.BeTrue) {
										return false
									}
									clock.Set(decodedMsg.ReceivedAt.Add(DeduplicationWindow))
									if !a.So(assertAccumulatedMetadata(ctx, env, rangeCtx, decodedMsg, RxMetadata[:], nil), should.BeTrue) {
										return false
									}
									decodedMsg.RxMetadata = RxMetadata[:]
									setDevice, macEvs := makeDataSetDevice(rangeDevice, decodedMsg)
									setCtx, ok := assertDataSetByID(ctx, env, rangeCtx, rangeDevice, setDevice, dataSetByIDSetPaths[:], nil, nil)
									return a.So(AllTrue(
										ok,
										assertDownlinkTaskAdd(ctx, env, setCtx, setDevice.EndDeviceIdentifiers, decodedMsg.ReceivedAt.Add(-InfrastructureDelay/2+Rx1Delay.Duration()/2-NSScheduleWindow()), true, ErrTestInternal),
										assertPublishMergeMetadata(ctx, env, setCtx, setDevice.EndDeviceIdentifiers, RxMetadata[:]...),
										assertPublishDefinitionDataClosures(ctx, env, macEvs, setCtx, setDevice.EndDeviceIdentifiers),
									), should.BeTrue)
								}, nil), should.BeTrue)
							},
						},
					)
				}
			}
		})
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			ns, ctx, env, stop := StartTest(
				t,
				component.Config{},
				Config{
					NetID:               *NetID.Copy(&types.NetID{}),
					DefaultMACSettings:  MACSettingConfig{},
					DeduplicationWindow: DeduplicationWindow,
					CooldownWindow:      CooldownWindow,
				},
				(1<<10)*test.Delay,
			)
			defer stop()

			<-env.DownlinkTasks.Pop

			clock := test.NewMockClock(time.Now().UTC())
			defer SetMockClock(clock)()

			if !tc.Handler(ctx, env, clock, func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan error {
				ch := make(chan error)
				go func() {
					_, err := ttnpb.NewGsNsClient(ns.LoopbackConn()).HandleUplink(ctx, CopyUplinkMessage(msg))
					ttnErr, ok := errors.From(err)
					if ok {
						ch <- ttnErr
					} else {
						ch <- err
					}
					close(ch)
				}()
				return ch
			}) {
				t.Error("Test handler failed")
			}
			assertions.New(t).So(AssertNetworkServerClose(ctx, ns), should.BeTrue)
		})
	}
}
