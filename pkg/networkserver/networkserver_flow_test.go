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
	"reflect"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

type FlowTestEnvironment struct {
	TestEnvironment
	Config

	*grpc.ClientConn
}

func hasProperStringSubset(x, y []string) bool {
	return test.IsProperSubsetOfElements(test.StringEqual, x, y) || test.IsProperSubsetOfElements(test.StringEqual, y, x)
}

func flowTestEventEqual(x, y events.Event) bool {
	if test.EventEqual(x, y) {
		return true
	}

	if xUp, ok := x.Data().(*ttnpb.UplinkMessage); ok {
		yUp, ok := y.Data().(*ttnpb.UplinkMessage)
		if !ok {
			return false
		}
		xUp = CopyUplinkMessage(xUp)
		yUp = CopyUplinkMessage(yUp)
		if !test.AllTrue(
			hasProperStringSubset(xUp.CorrelationIDs, yUp.CorrelationIDs),
			test.SameElements(reflect.DeepEqual, xUp.RxMetadata, yUp.RxMetadata),
		) {
			return false
		}
		xUp.CorrelationIDs = nil
		yUp.CorrelationIDs = nil
		xUp.RxMetadata = nil
		yUp.RxMetadata = nil
		xUp.ReceivedAt = time.Time{}
		yUp.ReceivedAt = time.Time{}
		if !reflect.DeepEqual(xUp, yUp) {
			return false
		}
	}

	xp, err := events.Proto(x)
	if err != nil {
		return false
	}
	yp, err := events.Proto(y)
	if err != nil {
		return false
	}
	xp.Data = nil
	yp.Data = nil
	xp.Time = time.Time{}
	yp.Time = time.Time{}

	if !hasProperStringSubset(xp.CorrelationIDs, yp.CorrelationIDs) {
		return false
	}
	xp.CorrelationIDs = nil
	yp.CorrelationIDs = nil
	return reflect.DeepEqual(xp, yp)
}

func makeAssertFlowTestEventEqual(t *testing.T) func(x, y events.Event) bool {
	a := assertions.New(t)
	return func(x, y events.Event) bool {
		if test.EventEqual(x, y) {
			return true
		}
		if !a.So(y.Data(), should.HaveSameTypeAs, x.Data()) {
			return false
		}
		if xUp, ok := x.Data().(*ttnpb.UplinkMessage); ok {
			xUp = CopyUplinkMessage(xUp)
			yUp := CopyUplinkMessage(y.Data().(*ttnpb.UplinkMessage))
			if !hasProperStringSubset(xUp.CorrelationIDs, yUp.CorrelationIDs) {
				t.Errorf(`Neither of uplink correlation IDs is a proper subset of the other:
X: %v
Y: %v`,
					xUp.CorrelationIDs, yUp.CorrelationIDs,
				)
				return false
			}
			if !a.So(xUp.RxMetadata, should.HaveSameElementsDeep, yUp.RxMetadata) {
				return false
			}
			xUp.CorrelationIDs = nil
			yUp.CorrelationIDs = nil
			xUp.ReceivedAt = time.Time{}
			yUp.ReceivedAt = time.Time{}
			xUp.RxMetadata = nil
			yUp.RxMetadata = nil
			if !a.So(xUp, should.Resemble, yUp) {
				return false
			}
		}

		xp, err := events.Proto(x)
		if err != nil {
			t.Errorf("Failed to encode x to proto: %s", err)
			return false
		}
		yp, err := events.Proto(y)
		if err != nil {
			t.Errorf("Failed to encode y to proto: %s", err)
			return false
		}
		xp.Data = nil
		yp.Data = nil
		xp.Time = time.Time{}
		yp.Time = time.Time{}

		if !hasProperStringSubset(xp.CorrelationIDs, yp.CorrelationIDs) {
			t.Errorf(`Neither of event correlation IDs is a proper subset of the other:
X: %v
Y: %v`,
				xp.CorrelationIDs, yp.CorrelationIDs,
			)
			return false
		}
		xp.CorrelationIDs = nil
		yp.CorrelationIDs = nil
		return a.So(xp, should.Resemble, yp)
	}
}

func (env FlowTestEnvironment) AssertApplicationUp(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient, assert func(*testing.T, *ttnpb.ApplicationUp) bool, expectedEvs ...events.Event) bool {
	return test.MustTFromContext(ctx).Run("Application uplink", func(t *testing.T) {
		a := assertions.New(t)

		ctx := test.ContextWithT(ctx, t)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var asUp *ttnpb.ApplicationUp
		var err error
		if !a.So(test.WaitContext(ctx, func() {
			asUp, err = link.Recv()
		}), should.BeTrue) {
			t.Error("Timed out while waiting for application uplink to be sent to Application Server")
			return
		}
		if !a.So(err, should.BeNil) {
			t.Errorf("Failed to receive Application Server uplink: %s", err)
			return
		}
		if !a.So(assert(t, asUp), should.BeTrue) {
			t.Errorf("Application uplink assertion failed")
			return
		}
		if !a.So(test.WaitContext(ctx, func() {
			err = link.Send(ttnpb.Empty)
		}), should.BeTrue) {
			t.Error("Timed out while waiting for Network Server to process Application Server response")
			return
		}
		if !a.So(err, should.BeNil) {
			t.Errorf("Failed to send Application Server uplink response: %s", err)
			return
		}
		a.So(env.Events, should.ReceiveEventsFunc, makeAssertFlowTestEventEqual(t), expectedEvs)
	})
}

func (env FlowTestEnvironment) AssertScheduleDownlink(ctx context.Context, assert func(context.Context, *ttnpb.DownlinkMessage) bool, paths []DownlinkPath, expectedEvs ...events.Event) bool {
	return test.MustTFromContext(ctx).Run("Schedule downlink", func(t *testing.T) {
		a := assertions.New(t)

		ctx := test.ContextWithT(ctx, t)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		scheduleDownlinkCh := make(chan NsGsScheduleDownlinkRequest)
		gsPeer := NewGSPeer(ctx, &MockNsGsServer{
			ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlinkCh),
		})
		for _, path := range paths {
			if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer, func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(test.AllTrue(
					a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER),
					a.So(ids, should.Resemble, *path.GatewayIdentifiers),
				), should.BeTrue)
			},
				test.ClusterGetPeerResponse{
					Peer: gsPeer,
				},
			), should.BeTrue) {
				t.Error("Gateway Server peer look-up assertion failed")
				return
			}
		}

		if !a.So(AssertAuthNsGsScheduleDownlinkRequest(ctx, env.Cluster.Auth, scheduleDownlinkCh, assert,
			&grpc.EmptyCallOption{},
			NsGsScheduleDownlinkResponse{
				Response: &ttnpb.ScheduleDownlinkResponse{},
			},
		), should.BeTrue) {
			t.Error("Gateway Server scheduling assertion failed")
			return
		}
		a.So(env.Events, should.ReceiveEventsFunc, makeAssertFlowTestEventEqual(t), expectedEvs)
	})
}

func (env FlowTestEnvironment) AssertSendDeviceUplink(ctx context.Context, expectedEvs []events.Event, ups ...*ttnpb.UplinkMessage) (<-chan error, bool) {
	t := test.MustTFromContext(ctx)
	a := assertions.New(t)
	errCh := make(chan error, len(ups))
	wg := &sync.WaitGroup{}
	wg.Add(len(ups) - 1)
	go func() {
		_, err := ttnpb.NewGsNsClient(env.ClientConn).HandleUplink(ctx, ups[0])
		t.Logf("First HandleUplink returned %v", err)
		errCh <- err
		wg.Wait()
		close(errCh)
	}()
	for _, up := range ups[1:] {
		up := up
		time.AfterFunc((1<<3)*test.Delay, func() {
			_, err := ttnpb.NewGsNsClient(env.ClientConn).HandleUplink(ctx, up)
			t.Logf("Duplicate HandleUplink returned %v", err)
			errCh <- err
			wg.Done()
		})
	}
	if !a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, len(expectedEvs), func(evs ...events.Event) bool {
		if !a.So(evs, should.HaveSameElementsFunc, flowTestEventEqual, expectedEvs) {
			actualEvs := map[events.Event]struct{}{}
			for _, ev := range evs {
				actualEvs[ev] = struct{}{}
			}
		outer:
			for _, expected := range expectedEvs {
				for actual := range actualEvs {
					if flowTestEventEqual(expected, actual) {
						delete(actualEvs, actual)
						continue outer
					}
				}
				t.Logf("Failed to match expected event '%s' with payload '%+v'", expected.Name(), expected.Data())
			}
			for actual := range actualEvs {
				t.Logf("Failed to match actual event '%s' with payload '%+v'", actual.Name(), actual.Data())
			}
			return false
		}
		return true
	}), should.BeTrue) {
		t.Error("Data uplink duplicate event assertion failed")
		return nil, false
	}
	for range ups[1:] {
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for duplicate HandleUplink to return")
			return nil, false

		case err := <-errCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle duplicate uplink: %s", err)
				return nil, false
			}
		}
	}
	return errCh, true
}

func downlinkProtoPaths(paths ...DownlinkPath) (pbs []*ttnpb.DownlinkPath) {
	for _, p := range paths {
		pbs = append(pbs, p.DownlinkPath)
	}
	return pbs
}

func (env FlowTestEnvironment) AssertJoin(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient, linkCtx context.Context, ids ttnpb.EndDeviceIdentifiers, fpID string, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, upChIdx uint8, upDRIdx ttnpb.DataRateIndex) (*ttnpb.JoinRequest, bool) {
	t := test.MustTFromContext(ctx)

	fp := FrequencyPlan(fpID)
	phy := Band(fp.BandID, phyVersion)
	upCh := phy.UplinkChannels[upChIdx]
	upDR := phy.DataRates[upDRIdx].Rate

	macSettings := env.Config.DefaultMACSettings.Parse()

	rx1Delay := macSettings.Rx1Delay.GetValue()
	if macSettings.Rx1Delay == nil {
		rx1Delay = ttnpb.RxDelay(phy.ReceiveDelay1.Seconds())
	}
	desiredRx1Delay := macSettings.DesiredRx1Delay.GetValue()
	if macSettings.DesiredRx1Delay == nil {
		desiredRx1Delay = rx1Delay
	}
	rx1DROffset := macSettings.Rx1DataRateOffset.GetValue()
	desiredRx1DROffset := macSettings.Rx1DataRateOffset.GetValue()
	if macSettings.DesiredRx1DataRateOffset == nil {
		desiredRx1DROffset = rx1DROffset
	}
	rx2DataRateIndex := macSettings.Rx2DataRateIndex.GetValue()
	if macSettings.Rx2DataRateIndex == nil {
		rx2DataRateIndex = phy.DefaultRx2Parameters.DataRateIndex
	}
	desiredRx2DataRateIndex := macSettings.DesiredRx2DataRateIndex.GetValue()
	if macSettings.DesiredRx2DataRateIndex == nil {
		if fp.DefaultRx2DataRate != nil {
			desiredRx2DataRateIndex = ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
		} else {
			desiredRx2DataRateIndex = rx2DataRateIndex
		}
	}
	rx2Frequency := macSettings.Rx2Frequency.GetValue()
	if macSettings.Rx2Frequency == nil {
		rx2Frequency = phy.DefaultRx2Parameters.Frequency
	}

	var expectedEvs []events.Event
	var joinReq *ttnpb.JoinRequest
	var joinCIDs []string
	if !t.Run("Join-request", func(t *testing.T) {
		a := assertions.New(t)

		ctx := test.ContextWithT(ctx, t)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handleJoinCh := make(chan NsJsHandleJoinRequest)
		jsPeer := NewJSPeer(ctx, &MockNsJsServer{
			HandleJoinFunc: MakeNsJsHandleJoinChFunc(handleJoinCh),
		})

		makeUplink := func(matched bool, rxMetadata ...*ttnpb.RxMetadata) *ttnpb.UplinkMessage {
			msg := MakeJoinRequest(matched, upDR, upCh.Frequency, rxMetadata...)
			if matched {
				return WithMatchedUplinkSettings(msg, upChIdx, upDRIdx)
			}
			return msg
		}
		var ups []*ttnpb.UplinkMessage
		var firstUp *ttnpb.UplinkMessage
		var preSendEvs []events.Event
		for i, upMD := range [][]*ttnpb.RxMetadata{
			RxMetadata[:2],
			nil,
			RxMetadata[2:],
		} {
			ups = append(ups, makeUplink(false, upMD...))
			if i == 0 {
				firstUp = makeUplink(true, upMD...)
			} else {
				preSendEvs = append(preSendEvs,
					EvtReceiveJoinRequest(ctx, ids, makeUplink(true, upMD...)),
					EvtDropJoinRequest(ctx, ids, ErrDuplicate),
				)
			}
		}

		start := time.Now()
		handleUplinkErrCh, ok := env.AssertSendDeviceUplink(ctx, preSendEvs, ups...)
		if !a.So(ok, should.BeTrue) {
			t.Error("Uplink send assertion failed")
			return
		}

		var getPeerCtx context.Context
		if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
			func(ctx context.Context, role ttnpb.ClusterRole, peerIDs ttnpb.Identifiers) bool {
				for _, cid := range ups[0].CorrelationIDs {
					a.So(events.CorrelationIDsFromContext(ctx), should.Contain, cid)
				}
				getPeerCtx = ctx
				return test.AllTrue(
					a.So(role, should.Equal, ttnpb.ClusterRole_JOIN_SERVER),
					a.So(peerIDs, should.Resemble, ids),
				)
			},
			test.ClusterGetPeerResponse{Peer: jsPeer},
		), should.BeTrue) {
			t.Error("Join Server peer look-up assertion failed")
			return
		}
		joinReq = MakeNsJsJoinRequest(macVersion, phyVersion, fp, nil, desiredRx1Delay, uint8(desiredRx1DROffset), desiredRx2DataRateIndex, events.CorrelationIDsFromContext(getPeerCtx)...)
		joinResp := &ttnpb.JoinResponse{
			RawPayload:     bytes.Repeat([]byte{0x42}, 33),
			SessionKeys:    *MakeSessionKeys(macVersion, true),
			Lifetime:       time.Hour,
			CorrelationIDs: []string{"NsJs-1", "NsJs-2"},
		}
		if !a.So(AssertAuthNsJsHandleJoinRequest(ctx, env.Cluster.Auth, handleJoinCh, func(ctx context.Context, req *ttnpb.JoinRequest) bool {
			joinReq.DevAddr = req.DevAddr
			return test.AllTrue(
				a.So(req.DevAddr, should.NotBeEmpty),
				a.So(req.DevAddr.NwkID(), should.Resemble, env.Config.NetID.ID()),
				a.So(req.DevAddr.NetIDType(), should.Equal, env.Config.NetID.Type()),
				a.So(req, should.Resemble, joinReq),
			)
		},
			&grpc.EmptyCallOption{},
			NsJsHandleJoinResponse{
				Response: joinResp,
			},
		), should.BeTrue) {
			t.Error("Join-request send assertion failed")
			return
		}
		joinCIDs = append(events.CorrelationIDsFromContext(getPeerCtx), joinResp.CorrelationIDs...)

		a.So(env.Events, should.ReceiveEventFunc, makeAssertFlowTestEventEqual(t),
			EvtReceiveJoinRequest(events.ContextWithCorrelationID(test.Context(), ups[0].CorrelationIDs...), ids, firstUp),
		)
		a.So(env.Events, should.ReceiveEventsResembling,
			EvtClusterJoinAttempt(getPeerCtx, ids, joinReq),
			EvtClusterJoinSuccess(getPeerCtx, ids, &ttnpb.JoinResponse{
				RawPayload: joinResp.RawPayload,
				SessionKeys: ttnpb.SessionKeys{
					SessionKeyID: joinResp.SessionKeys.SessionKeyID,
				},
				Lifetime:       joinResp.Lifetime,
				CorrelationIDs: joinResp.CorrelationIDs,
			}),
		)
		a.So(env.Events, should.ReceiveEventFunc, makeAssertFlowTestEventEqual(t),
			EvtProcessJoinRequest(getPeerCtx, ids, makeUplink(true, RxMetadata[:]...)),
		)
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for HandleUplink to return")
			return
		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle uplink: %s", err)
				return
			}
		}

		if !a.So(env.AssertApplicationUp(ctx, link, func(t *testing.T, up *ttnpb.ApplicationUp) bool {
			expectedEvs = append(expectedEvs, EvtForwardJoinAccept(linkCtx, up.EndDeviceIdentifiers, up))

			a := assertions.New(t)
			return a.So(test.AllTrue(
				a.So(up.CorrelationIDs, should.HaveSameElementsDeep, joinCIDs),
				a.So([]time.Time{start, up.GetJoinAccept().GetReceivedAt(), time.Now()}, should.BeChronological),
				a.So(up, should.Resemble, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: *MakeOTAAIdentifiers(&joinReq.DevAddr),
					CorrelationIDs:       up.CorrelationIDs,
					Up: &ttnpb.ApplicationUp_JoinAccept{
						JoinAccept: &ttnpb.ApplicationJoinAccept{
							AppSKey:      joinResp.AppSKey,
							SessionKeyID: joinResp.SessionKeyID,
							ReceivedAt:   up.GetJoinAccept().GetReceivedAt(),
						},
					},
				}),
			), should.BeTrue)
		}), should.BeTrue) {
			t.Error("Failed to send join-accept to Application Server")
			return
		}
	}) {
		t.Error("Join-request assertion failed")
		return nil, false
	}
	return joinReq, t.Run("Join-accept", func(t *testing.T) {
		a := assertions.New(t)

		ctx := test.ContextWithT(ctx, t)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		paths := DownlinkPathsFromMetadata(RxMetadata[:]...)
		txReq := &ttnpb.TxRequest{
			Class:            ttnpb.CLASS_A,
			DownlinkPaths:    downlinkProtoPaths(paths...),
			Rx1Delay:         ttnpb.RxDelay(phy.JoinAcceptDelay1.Seconds()),
			Rx1DataRateIndex: test.Must(phy.Rx1DataRate(upDRIdx, rx1DROffset, fp.DwellTime.GetUplinks())).(ttnpb.DataRateIndex),
			Rx1Frequency:     phy.DownlinkChannels[test.Must(phy.Rx1Channel(upChIdx)).(uint8)].Frequency,
			Rx2DataRateIndex: rx2DataRateIndex,
			Rx2Frequency:     rx2Frequency,
			Priority:         ttnpb.TxSchedulePriority_HIGHEST,
			FrequencyPlanID:  fpID,
		}
		if !a.So(env.AssertScheduleDownlink(ctx, func(ctx context.Context, down *ttnpb.DownlinkMessage) bool {
			return test.AllTrue(
				a.So(events.CorrelationIDsFromContext(ctx), should.NotBeEmpty),
				a.So(down.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, joinCIDs),
				a.So(down, should.Resemble, &ttnpb.DownlinkMessage{
					CorrelationIDs: down.CorrelationIDs,
					RawPayload:     bytes.Repeat([]byte{0x42}, 33),
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: txReq,
					},
				}),
			)
		}, paths), should.BeTrue) {
			t.Error("Join-accept assertion failed")
			return
		}
		a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, 2+len(expectedEvs), func(evs ...events.Event) bool {
			return a.So(evs, should.HaveSameElementsFunc, flowTestEventEqual, append(expectedEvs,
				EvtScheduleJoinAcceptAttempt(ctx, ids, txReq),
				EvtScheduleJoinAcceptSuccess(ctx, ids, &ttnpb.ScheduleDownlinkResponse{}),
			))
		}), should.BeTrue)
	})
}

func (env FlowTestEnvironment) AssertSendDataUplink(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient, linkCtx context.Context, ids ttnpb.EndDeviceIdentifiers, makeUplink func(matched bool, rxMetadata ...*ttnpb.RxMetadata) *ttnpb.UplinkMessage, processEvs ...events.Event) bool {
	return test.MustTFromContext(ctx).Run("Data uplink", func(t *testing.T) {
		a := assertions.New(t)

		ctx := test.ContextWithT(ctx, t)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		expectedEvs := processEvs
		var ups []*ttnpb.UplinkMessage
		for i, upMD := range [][]*ttnpb.RxMetadata{
			nil,
			RxMetadata[3:],
			RxMetadata[:3],
		} {
			ups = append(ups, makeUplink(false, upMD...))
			expectedEvs = append(expectedEvs, EvtReceiveDataUplink(ctx, ids, makeUplink(true, upMD...)))
			if i > 0 {
				expectedEvs = append(expectedEvs, EvtDropDataUplink(ctx, ids, ErrDuplicate))
			}
		}

		handleUplinkErrCh, ok := env.AssertSendDeviceUplink(ctx, expectedEvs, ups...)
		if !a.So(ok, should.BeTrue) {
			t.Error("Uplink send assertion failed")
			return
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for HandleUplink to return")
			return
		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle uplink: %s", err)
				return
			}
		}
	})
}

func (env FlowTestEnvironment) AssertSetDevice(ctx context.Context, create bool, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, bool) {
	t := test.MustTFromContext(ctx)
	t.Helper()

	a := assertions.New(t)

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer func() {
		close(listRightsCh)
	}()

	var dev *ttnpb.EndDevice
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		dev, err = ttnpb.NewNsEndDeviceRegistryClient(env.ClientConn).Set(
			ctx,
			req,
			grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     "set-key",
				AllowInsecure: true,
			}),
		)
		wg.Done()
	}()

	var reqCIDs []string
	if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
		func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
			reqCIDs = events.CorrelationIDsFromContext(ctx)
			return a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS) && a.So(ids, should.BeNil)
		},
		test.ClusterGetPeerResponse{
			Peer: NewISPeer(ctx, &test.MockApplicationAccessServer{
				ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
			}),
		},
	), should.BeTrue) {
		return nil, false
	}

	a.So(reqCIDs, should.HaveLength, 1)

	if !a.So(test.AssertListRightsRequest(ctx, listRightsCh,
		func(ctx context.Context, ids ttnpb.Identifiers) bool {
			md := rpcmetadata.FromIncomingContext(ctx)
			return a.So(md.AuthType, should.Equal, "Bearer") &&
				a.So(md.AuthValue, should.Equal, "set-key") &&
				a.So(ids, should.Resemble, &req.EndDevice.ApplicationIdentifiers)
		}, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	), should.BeTrue) {
		return nil, false
	}

	ev := EvtCreateEndDevice.BindData(nil)
	if !create {
		ev = EvtUpdateEndDevice.BindData(nil)
	}
	if !a.So(env.Events, should.ReceiveEventResembling, ev(events.ContextWithCorrelationID(test.Context(), reqCIDs...), req.EndDevice.EndDeviceIdentifiers)) {
		if create {
			t.Error("Failed to assert end device create event")
		} else {
			t.Error("Failed to assert end device update event")
		}
		return nil, false
	}

	if !a.So(test.WaitContext(ctx, wg.Wait), should.BeTrue) {
		t.Error("Timed out while waiting for device to be set")
		return nil, false
	}
	return dev, a.So(err, should.BeNil)
}

func makeClassCOTAAFlowTest(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string, linkADRReqs ...*ttnpb.MACCommand_LinkADRReq) func(context.Context, FlowTestEnvironment) {
	return func(ctx context.Context, env FlowTestEnvironment) {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)

		start := time.Now()

		linkCtx, closeLink := context.WithCancel(ctx)
		link, linkEndEvent, ok := AssertLinkApplication(linkCtx, env.ClientConn, env.Cluster.GetPeer, env.Events, AppID)
		if !a.So(ok, should.BeTrue) || !a.So(link, should.NotBeNil) {
			t.Error("AS link assertion failed")
			closeLink()
			return
		}
		defer func() {
			closeLink()
			if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
				return a.So(ev.Data(), should.BeError) &&
					a.So(errors.IsCanceled(ev.Data().(error)), should.BeTrue) &&
					a.So(ev, should.ResembleEvent, linkEndEvent(ev.Data().(error)))
			}), should.BeTrue) {
				t.Error("AS link end event assertion failed")
			}
		}()

		ids := *MakeOTAAIdentifiers(nil)

		setDevice := &ttnpb.EndDevice{
			EndDeviceIdentifiers: ids,
			FrequencyPlanID:      fpID,
			LoRaWANVersion:       macVersion,
			LoRaWANPHYVersion:    phyVersion,
			SupportsClassC:       true,
			SupportsJoin:         true,
		}
		dev, ok := env.AssertSetDevice(ctx, true, &ttnpb.SetEndDeviceRequest{
			EndDevice: *setDevice,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"supports_class_c",
					"supports_join",
				},
			},
		})
		if !a.So(ok, should.BeTrue) || !a.So(dev, should.NotBeNil) {
			t.Error("Failed to create device")
			return
		}
		t.Log("Device created")
		a.So(dev.CreatedAt, should.HappenAfter, start)
		a.So(dev.UpdatedAt, should.Equal, dev.CreatedAt)
		a.So([]time.Time{start, dev.CreatedAt, time.Now()}, should.BeChronological)
		setDevice.CreatedAt = dev.CreatedAt
		setDevice.UpdatedAt = dev.UpdatedAt
		a.So(dev, should.Resemble, setDevice)

		joinReq, ok := env.AssertJoin(ctx, link, linkCtx, ids, fpID, macVersion, phyVersion, 1, ttnpb.DATA_RATE_2)
		if !a.So(ok, should.BeTrue) {
			t.Error("Device failed to join")
			return
		}
		t.Logf("Device successfully joined. DevAddr: %s", joinReq.DevAddr)

		ids = *MakeOTAAIdentifiers(&joinReq.DevAddr)
		fp := FrequencyPlan(fpID)
		phy := Band(fp.BandID, phyVersion)

		upChs := FrequencyPlanChannels(phy, fp.UplinkChannels, fp.DownlinkChannels...)
		upChIdx := uint8(2)
		upDRIdx := ttnpb.DATA_RATE_1

		var upCmders []MACCommander
		var expectedUpEvs []events.Event
		var downCmders []MACCommander
		if macVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			rekeyInd := &ttnpb.MACCommand_RekeyInd{
				MinorVersion: ttnpb.MINOR_1,
			}
			deviceModeInd := &ttnpb.MACCommand_DeviceModeInd{
				Class: ttnpb.CLASS_C,
			}
			upCmders = append(upCmders,
				rekeyInd,
				deviceModeInd,
			)

			rekeyConf := &ttnpb.MACCommand_RekeyConf{
				MinorVersion: ttnpb.MINOR_1,
			}
			deviceModeConf := &ttnpb.MACCommand_DeviceModeConf{
				Class: ttnpb.CLASS_C,
			}
			expectedUpEvs = append(expectedUpEvs,
				EvtReceiveRekeyIndication(ctx, ids, rekeyInd),
				EvtEnqueueRekeyConfirmation(ctx, ids, rekeyConf),
				EvtReceiveDeviceModeIndication(ctx, ids, deviceModeInd),
				EvtClassCSwitch(ctx, ids, ttnpb.CLASS_A),
				EvtEnqueueDeviceModeConfirmation(ctx, ids, deviceModeConf),
			)
			downCmders = append(downCmders,
				rekeyConf,
				deviceModeConf,
			)
		}
		makeUplink := func(matched bool, rxMetadata ...*ttnpb.RxMetadata) *ttnpb.UplinkMessage {
			msg := MakeDataUplink(macVersion, matched, true, joinReq.DevAddr, ttnpb.FCtrl{
				ADR: true,
			}, 0x00, 0x00, 0x42, []byte("test"), MakeUplinkMACBuffer(phy, upCmders...), phy.DataRates[upDRIdx].Rate, upDRIdx, upChs[upChIdx].UplinkFrequency, upChIdx, rxMetadata...)
			if matched {
				return WithMatchedUplinkSettings(msg, upChIdx, upDRIdx)
			}
			return msg
		}
		expectedUp := makeUplink(true, RxMetadata[:]...)
		start = time.Now()
		if !a.So(env.AssertSendDataUplink(ctx, link, linkCtx, ids, makeUplink, append(expectedUpEvs,
			EvtProcessDataUplink(ctx, ids, expectedUp),
		)...), should.BeTrue) {
			t.Error("Failed to process data uplink")
			return
		}

		var expectedEvs []events.Event
		if !a.So(env.AssertApplicationUp(ctx, link, func(t *testing.T, up *ttnpb.ApplicationUp) bool {
			expectedEvs = append(expectedEvs, EvtForwardDataUplink(linkCtx, up.EndDeviceIdentifiers, up))

			a := assertions.New(t)
			return a.So(test.AllTrue(
				a.So(up.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, DataUplinkCorrelationIDs),
				a.So(up.GetUplinkMessage().GetRxMetadata(), should.HaveSameElementsDeep, expectedUp.RxMetadata),
				a.So([]time.Time{start, up.GetUplinkMessage().GetReceivedAt(), time.Now()}, should.BeChronological),
				a.So(up, should.Resemble, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: ids,
					CorrelationIDs:       up.CorrelationIDs,
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{
							Confirmed:    expectedUp.Payload.MHDR.MType == ttnpb.MType_CONFIRMED_UP,
							FPort:        expectedUp.Payload.GetMACPayload().FPort,
							FRMPayload:   expectedUp.Payload.GetMACPayload().FRMPayload,
							ReceivedAt:   up.GetUplinkMessage().GetReceivedAt(),
							RxMetadata:   up.GetUplinkMessage().GetRxMetadata(),
							Settings:     expectedUp.Settings,
							SessionKeyID: MakeSessionKeys(macVersion, false).SessionKeyID,
						},
					},
				}),
			), should.BeTrue)
		}), should.BeTrue) {
			t.Error("Failed to send data uplink to Application Server")
			return
		}

		downCmders = append(downCmders, ttnpb.CID_DEV_STATUS)
		expectedEvs = append(expectedEvs, EvtEnqueueDevStatusRequest(ctx, ids, nil))
		for _, cmd := range linkADRReqs {
			cmd := cmd
			downCmders = append(downCmders, cmd)
			expectedEvs = append(expectedEvs, EvtEnqueueLinkADRRequest(ctx, ids, cmd))
		}

		paths := DownlinkPathsFromMetadata(RxMetadata[:]...)
		txReq := &ttnpb.TxRequest{
			Class:            ttnpb.CLASS_A,
			DownlinkPaths:    downlinkProtoPaths(paths...),
			Rx1Delay:         joinReq.RxDelay,
			Rx1DataRateIndex: test.Must(phy.Rx1DataRate(upDRIdx, joinReq.DownlinkSettings.Rx1DROffset, fp.DwellTime.GetUplinks())).(ttnpb.DataRateIndex),
			Rx1Frequency:     phy.DownlinkChannels[test.Must(phy.Rx1Channel(upChIdx)).(uint8)].Frequency,
			Rx2DataRateIndex: joinReq.DownlinkSettings.Rx2DR,
			Rx2Frequency:     phy.DefaultRx2Parameters.Frequency,
			Priority:         ttnpb.TxSchedulePriority_HIGHEST,
			FrequencyPlanID:  fpID,
		}
		if !a.So(env.AssertScheduleDownlink(ctx, func(ctx context.Context, down *ttnpb.DownlinkMessage) bool {
			return test.AllTrue(
				a.So(events.CorrelationIDsFromContext(ctx), should.NotBeEmpty),
				a.So(down.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, expectedUp.CorrelationIDs),
				a.So(down, should.Resemble, MakeDataDownlink(macVersion, false, joinReq.DevAddr, ttnpb.FCtrl{
					ADR: true,
					Ack: true,
				}, 0x00, 0x00, 0x00, nil, MakeDownlinkMACBuffer(phy, downCmders...), txReq, down.CorrelationIDs...)),
			)
		}, paths,
		), should.BeTrue) {
			t.Error("Failed to schedule downlink on Gateway Server")
			return
		}
		a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, 2+len(expectedEvs), func(evs ...events.Event) bool {
			return a.So(evs, should.HaveSameElementsFunc, flowTestEventEqual, append(
				expectedEvs,
				EvtScheduleDataDownlinkAttempt(ctx, ids, txReq),
				EvtScheduleDataDownlinkSuccess(ctx, ids, &ttnpb.ScheduleDownlinkResponse{}),
			))
		}), should.BeTrue)
	}
}

func TestFlow(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name                      string
		NewDeviceRegistry         func(t testing.TB) (dr DeviceRegistry, closeFn func() error)
		NewApplicationUplinkQueue func(t testing.TB) (uq ApplicationUplinkQueue, closeFn func() error)
		NewDownlinkTaskQueue      func(t testing.TB) (tq DownlinkTaskQueue, closeFn func() error)
		NewUplinkDeduplicator     func(t testing.TB) (ud UplinkDeduplicator, closeFn func() error)
	}{
		{
			Name:                      "Redis application uplink queue/Redis registry/Redis downlink task queue",
			NewApplicationUplinkQueue: NewRedisApplicationUplinkQueue,
			NewDeviceRegistry:         NewRedisDeviceRegistry,
			NewDownlinkTaskQueue:      NewRedisDownlinkTaskQueue,
			NewUplinkDeduplicator:     NewRedisUplinkDeduplicator,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			eu868LinkADRReqs := []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask:   []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false},
					DataRateIndex: ttnpb.DATA_RATE_4,
					TxPowerIndex:  1,
					NbTrans:       1,
				},
			}
			us915LinkADRReqs := []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask:   []bool{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
					DataRateIndex: ttnpb.DATA_RATE_2,
					TxPowerIndex:  1,
					NbTrans:       1,
				},
			}
			for flow, handleFlowTest := range map[string]func(context.Context, FlowTestEnvironment){
				"Class C/OTAA/MAC:1.0.3/PHY:1.0.3-a/FP:EU868": makeClassCOTAAFlowTest(ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A, test.EUFrequencyPlanID, eu868LinkADRReqs...),
				"Class C/OTAA/MAC:1.0.4/PHY:1.0.3-a/FP:US915": makeClassCOTAAFlowTest(ttnpb.MAC_V1_0_4, ttnpb.PHY_V1_0_3_REV_A, test.USFrequencyPlanID, us915LinkADRReqs...),
				"Class C/OTAA/MAC:1.1/PHY:1.1-b/FP:EU868":     makeClassCOTAAFlowTest(ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B, test.EUFrequencyPlanID, eu868LinkADRReqs...),
			} {
				t.Run(flow, func(t *testing.T) {
					uq, uqClose := tc.NewApplicationUplinkQueue(t)
					if uqClose != nil {
						defer func() {
							if err := uqClose(); err != nil {
								t.Errorf("Failed to close application uplink queue: %s", err)
							}
						}()
					}
					dr, drClose := tc.NewDeviceRegistry(t)
					if drClose != nil {
						defer func() {
							if err := drClose(); err != nil {
								t.Errorf("Failed to close device registry: %s", err)
							}
						}()
					}
					tq, tqClose := tc.NewDownlinkTaskQueue(t)
					if tqClose != nil {
						defer func() {
							if err := tqClose(); err != nil {
								t.Errorf("Failed to close downlink task queue: %s", err)
							}
						}()
					}
					ud, udClose := tc.NewUplinkDeduplicator(t)
					if udClose != nil {
						defer func() {
							if err := udClose(); err != nil {
								t.Errorf("Failed to close downlink task queue: %s", err)
							}
						}()
					}

					conf := DefaultConfig
					conf.NetID = test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)
					conf.ApplicationUplinks = uq
					conf.Devices = dr
					conf.DownlinkTasks = tq
					conf.UplinkDeduplicator = ud
					conf.DeduplicationWindow = (1 << 4) * test.Delay
					conf.CooldownWindow = (1 << 9) * test.Delay

					ns, ctx, env, stop := StartTest(t, component.Config{}, conf, (1<<13)*test.Delay)
					defer stop()

					handleFlowTest(ctx, FlowTestEnvironment{
						TestEnvironment: env,
						Config:          conf,
						ClientConn:      ns.LoopbackConn(),
					})
				})
			}
		})
	}
}
