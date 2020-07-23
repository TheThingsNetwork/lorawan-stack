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
	"fmt"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/oklog/ulid/v2"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

// handleDownlinkTaskQueueTest runs a test suite on q.
func handleDownlinkTaskQueueTest(t *testing.T, q DownlinkTaskQueue) {
	a := assertions.New(t)

	ctx := test.Context()

	pbs := [...]ttnpb.EndDeviceIdentifiers{
		{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test-app",
			},
			DeviceID: "test-dev",
		},
		{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test-app2",
			},
			DeviceID: "test-dev",
		},
		{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test-app2",
			},
			DeviceID: "test-dev2",
		},
	}

	type slot struct {
		ctx   context.Context
		id    ttnpb.EndDeviceIdentifiers
		t     time.Time
		errCh chan<- error
	}

	popCtx := context.WithValue(ctx, &struct{}{}, "pop")
	nextPop := make(chan struct{})
	slotCh := make(chan slot)
	errCh := make(chan error)
	go func() {
		for {
			<-nextPop
			select {
			case <-ctx.Done():
				return
			case errCh <- q.Pop(popCtx, func(ctx context.Context, id ttnpb.EndDeviceIdentifiers, t time.Time) error {
				errCh := make(chan error)
				slotCh <- slot{
					ctx:   ctx,
					id:    id,
					t:     t,
					errCh: errCh,
				}
				return <-errCh
			}):
			}
		}
	}()

	// Ensure the goroutine has started
	nextPop <- struct{}{}

	select {
	case s := <-slotCh:
		t.Fatalf("Pop called f on empty schedule, slot: %+v", s)

	case err := <-errCh:
		a.So(err, should.BeNil)
		t.Fatal("Pop returned on empty schedule")

	case <-time.After(test.Delay):
	}

	err := q.Add(ctx, pbs[0], time.Unix(0, 0), false)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	select {
	case s := <-slotCh:
		a.So(s.id, should.Resemble, pbs[0])
		a.So(s.t, should.Equal, time.Unix(0, 0))
		if !a.So(s.ctx, should.HaveParentContextOrEqual, popCtx) {
			t.Log(s.ctx == popCtx)
			t.Fatal(s.ctx)
		}
		s.errCh <- nil

	case err := <-errCh:
		a.So(err, should.BeNil)
		t.Fatal("Pop returned without calling f on non-empty schedule")

	case <-time.After(10 * Timeout):
		t.Fatal("Timed out waiting for Pop to call f")
	}

	select {
	case err := <-errCh:
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

	case <-time.After(Timeout):
		t.Fatal("Timed out waiting for Pop to return")
	}

	err = q.Add(ctx, pbs[0], time.Unix(0, 42), false)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[1], time.Now().Add(time.Hour), false)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[1], time.Now().Add(2*time.Hour), true)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[1], time.Unix(13, 0), false)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[1], time.Unix(42, 0), true)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[2], time.Now().Add(42*time.Hour), true)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[2], time.Unix(42, 42), true)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	expectSlot := func(t *testing.T, expectedID ttnpb.EndDeviceIdentifiers, expectedAt time.Time) {
		nextPop <- struct{}{}

		t.Helper()

		a := assertions.New(t)

		select {
		case s := <-slotCh:
			a.So(s.id, should.Resemble, expectedID)
			a.So(s.t, should.Equal, expectedAt)
			a.So(s.ctx, should.HaveParentContextOrEqual, popCtx)
			s.errCh <- nil

		case err := <-errCh:
			a.So(err, should.BeNil)
			t.Fatal("Pop returned without calling f on non-empty schedule")

		case <-time.After(Timeout):
			t.Fatal("Timed out waiting for Pop to call f")
		}

		select {
		case err := <-errCh:
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

		case <-time.After(Timeout):
			t.Fatal("Timed out waiting for Pop to return")
		}
	}

	expectSlot(t, pbs[0], time.Unix(0, 42))
	expectSlot(t, pbs[1], time.Unix(42, 0))
	expectSlot(t, pbs[2], time.Unix(42, 42))
}

func TestDownlinkTaskQueues(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string
		New  func(t testing.TB) (q DownlinkTaskQueue, closeFn func())
		N    uint16
	}{
		{
			Name: "Redis",
			New:  NewRedisDownlinkTaskQueue,
			N:    8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			t.Run(fmt.Sprintf("%s/%d", tc.Name, i), func(t *testing.T) {
				t.Parallel()
				q, closeFn := tc.New(t)
				if closeFn != nil {
					defer closeFn()
				}
				t.Run("1st run", func(t *testing.T) { handleDownlinkTaskQueueTest(t, q) })
				if t.Failed() {
					t.Skip("Skipping 2nd run")
				}
				t.Run("2nd run", func(t *testing.T) { handleDownlinkTaskQueueTest(t, q) })
			})
		}
	}
}

func TestProcessDownlinkTask(t *testing.T) {
	// TODO: Refactor. (https://github.com/TheThingsNetwork/lorawan-stack/issues/2475)

	const appIDString = "process-downlink-test-app-id"
	appID := ttnpb.ApplicationIdentifiers{ApplicationID: appIDString}
	const devID = "process-downlink-test-dev-id"

	devAddr := types.DevAddr{0x42, 0xff, 0xff, 0xff}

	fNwkSIntKey := types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nwkSEncKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	sNwkSIntKey := types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	sessionKeys := &ttnpb.SessionKeys{
		FNwkSIntKey: &ttnpb.KeyEnvelope{
			Key: &fNwkSIntKey,
		},
		NwkSEncKey: &ttnpb.KeyEnvelope{
			Key: &nwkSEncKey,
		},
		SNwkSIntKey: &ttnpb.KeyEnvelope{
			Key: &sNwkSIntKey,
		},
		SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
	}

	const rx1Delay = ttnpb.RX_DELAY_1
	makeEU868macParameters := func(ver ttnpb.PHYVersion) ttnpb.MACParameters {
		params := MakeDefaultEU868CurrentMACParameters(ver)
		params.Rx1Delay = rx1Delay
		params.Channels = append(params.Channels, &ttnpb.MACParameters_Channel{
			UplinkFrequency:   430000000,
			DownlinkFrequency: 431000000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_3,
		})
		return params
	}

	assertGetGatewayPeers := func(ctx context.Context, getPeerCh <-chan test.ClusterGetPeerRequest, peer124, peer3 cluster.Peer) bool {
		t, a := test.MustNewTFromContext(ctx)
		t.Helper()
		return test.AssertClusterGetPeerRequestSequence(ctx, getPeerCh,
			[]test.ClusterGetPeerResponse{
				{Error: errors.New("peer not found")},
				{Peer: peer124},
				{Peer: peer124},
				{Peer: peer3},
				{Peer: peer124},
			},
			func(reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(reqCtx, should.HaveParentContextOrEqual, ctx) &&
					a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "gateway-test-0",
					})
			},
			func(reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(reqCtx, should.HaveParentContextOrEqual, ctx) &&
					a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "gateway-test-1",
					})
			},
			func(reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(reqCtx, should.HaveParentContextOrEqual, ctx) &&
					a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "gateway-test-2",
					})
			},
			func(reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(reqCtx, should.HaveParentContextOrEqual, ctx) &&
					a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "gateway-test-3",
					})
			},
			func(reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(reqCtx, should.HaveParentContextOrEqual, ctx) &&
					a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "gateway-test-4",
					})
			},
		)
	}

	assertScheduleGateways := func(ctx context.Context, authCh <-chan test.ClusterAuthRequest, scheduleDownlink124Ch, scheduleDownlink3Ch <-chan NsGsScheduleDownlinkRequest, payload []byte, makeTxRequest func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest, fixedPaths bool, resps ...NsGsScheduleDownlinkResponse) (*ttnpb.DownlinkMessage, bool) {
		if len(resps) < 1 || len(resps) > 3 {
			panic("invalid response count specified")
		}

		t, a := test.MustNewTFromContext(ctx)
		t.Helper()
		makePath := func(i int) *ttnpb.DownlinkPath {
			if fixedPaths {
				return &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_Fixed{
						Fixed: &GatewayAntennaIdentifiers[i],
					},
				}
			}
			return &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: func() *ttnpb.RxMetadata {
						switch i {
						case 1:
							return RxMetadata[0]
						case 2:
							return RxMetadata[4]
						case 3:
							return RxMetadata[1]
						case 4:
							return RxMetadata[5]
						default:
							panic(fmt.Sprintf("Invalid index requested: %d", i))
						}
					}().UplinkToken,
				},
			}
		}

		var lastDown *ttnpb.DownlinkMessage
		var correlationIDs []string
		if !a.So(AssertAuthNsGsScheduleDownlinkRequest(ctx, authCh, scheduleDownlink124Ch,
			func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool {
				correlationIDs = msg.CorrelationIDs
				lastDown = &ttnpb.DownlinkMessage{
					CorrelationIDs: correlationIDs,
					RawPayload:     payload,
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: makeTxRequest(
							makePath(1),
							makePath(2),
						),
					},
				}
				return a.So(msg, should.Resemble, lastDown)
			},
			grpc.EmptyCallOption{},
			resps[0],
		), should.BeTrue) {
			t.Error("Downlink assertion failed for gateways 1 and 2")
			return nil, false
		}
		t.Logf("Downlink correlation IDs: %v", correlationIDs)
		if len(resps) == 1 {
			return lastDown, true
		}

		lastDown = &ttnpb.DownlinkMessage{
			CorrelationIDs: correlationIDs,
			RawPayload:     payload,
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: makeTxRequest(
					makePath(3),
				),
			},
		}
		if !a.So(AssertAuthNsGsScheduleDownlinkRequest(ctx, authCh, scheduleDownlink3Ch,
			func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool {
				return a.So(msg, should.Resemble, lastDown)
			},
			grpc.EmptyCallOption{},
			resps[1],
		), should.BeTrue) {
			t.Error("Downlink assertion failed for gateway 3")
			return nil, false
		}
		if len(resps) == 2 {
			return lastDown, true
		}

		lastDown = &ttnpb.DownlinkMessage{
			CorrelationIDs: correlationIDs,
			RawPayload:     payload,
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: makeTxRequest(
					makePath(4),
				),
			},
		}
		if !a.So(AssertAuthNsGsScheduleDownlinkRequest(ctx, authCh, scheduleDownlink124Ch,
			func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool {
				return a.So(msg, should.Resemble, lastDown)
			},
			grpc.EmptyCallOption{},
			resps[2],
		), should.BeTrue) {
			t.Error("Downlink assertion failed for gateway 4")
			return nil, false
		}
		return lastDown, true
	}

	assertScheduleRxMetadataGateways := func(ctx context.Context, authCh <-chan test.ClusterAuthRequest, scheduleDownlink124Ch, scheduleDownlink3Ch <-chan NsGsScheduleDownlinkRequest, payload []byte, makeTxRequest func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest, resps ...NsGsScheduleDownlinkResponse) (*ttnpb.DownlinkMessage, bool) {
		return assertScheduleGateways(ctx, authCh, scheduleDownlink124Ch, scheduleDownlink3Ch, payload, makeTxRequest, false, resps...)
	}

	assertScheduleClassBCGateways := func(ctx context.Context, authCh <-chan test.ClusterAuthRequest, scheduleDownlink124Ch, scheduleDownlink3Ch <-chan NsGsScheduleDownlinkRequest, payload []byte, makeTxRequest func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest, resps ...NsGsScheduleDownlinkResponse) (*ttnpb.DownlinkMessage, bool) {
		return assertScheduleGateways(ctx, authCh, scheduleDownlink124Ch, scheduleDownlink3Ch, payload, makeTxRequest, true, resps...)
	}

	start := gpstime.Parse(10000 * BeaconPeriod).Add(time.Second + 200*time.Millisecond)
	clock := test.NewMockClock(start)
	defer SetMockClock(clock)()

	type DeviceDiffFunc func(ctx context.Context, expected, created, updated *ttnpb.EndDevice, down *ttnpb.DownlinkMessage, downAt time.Time) bool
	makeRemoveDownlinksDiff := func(n uint) DeviceDiffFunc {
		return func(ctx context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, _ time.Time) bool {
			t := test.MustTFromContext(ctx)
			t.Helper()
			a := assertions.New(t)
			if !a.So(expected.Session.QueuedApplicationDownlinks, should.BeGreaterThanOrEqualTo, n) {
				return false
			}
			expected.Session.QueuedApplicationDownlinks = expected.Session.QueuedApplicationDownlinks[:uint(len(expected.Session.QueuedApplicationDownlinks))-n]
			return true
		}
	}
	makePendingMACCommandsDiff := func(cmds ...*ttnpb.MACCommand) DeviceDiffFunc {
		return func(_ context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, _ time.Time) bool {
			expected.MACState.PendingRequests = cmds
			return true
		}
	}
	makeSetLastNFCntDownDiff := func(v uint32) DeviceDiffFunc {
		return func(_ context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, _ time.Time) bool {
			expected.Session.LastNFCntDown = v
			return true
		}
	}
	makeSetLastConfFCntDownDiff := func(v uint32) DeviceDiffFunc {
		return func(_ context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, _ time.Time) bool {
			expected.Session.LastConfFCntDown = v
			return true
		}
	}
	removeMACQueueDiff := DeviceDiffFunc(func(_ context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, _ time.Time) bool {
		expected.MACState.QueuedResponses = nil
		return true
	})
	appendRecentMACStateDownlinkDiff := DeviceDiffFunc(func(_ context.Context, expected, _, _ *ttnpb.EndDevice, down *ttnpb.DownlinkMessage, _ time.Time) bool {
		expected.MACState.RecentDownlinks = AppendRecentDownlink(expected.MACState.RecentDownlinks, down, RecentDownlinkCount)
		return true
	})
	setRxWindowsUnavailableDiff := DeviceDiffFunc(func(_ context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, _ time.Time) bool {
		expected.MACState.RxWindowsAvailable = false
		return true
	})
	setLastConfirmedDownlinkAtDiff := DeviceDiffFunc(func(_ context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, downAt time.Time) bool {
		expected.MACState.LastConfirmedDownlinkAt = &downAt
		return true
	})
	setLastDownlinkAtDiff := DeviceDiffFunc(func(_ context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, downAt time.Time) bool {
		expected.MACState.LastDownlinkAt = &downAt
		return true
	})
	setLastNetworkInitiatedDownlinkAtDiff := DeviceDiffFunc(func(_ context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, downAt time.Time) bool {
		expected.MACState.LastNetworkInitiatedDownlinkAt = &downAt
		return true
	})
	setPendingApplicationDownlinkDiff := DeviceDiffFunc(func(_ context.Context, expected, created, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, downAt time.Time) bool {
		expected.MACState.PendingApplicationDownlink = created.QueuedApplicationDownlinks[len(created.QueuedApplicationDownlinks)-1]
		return true
	})

	for _, tc := range []struct {
		Name                        string
		CreateDevice                SetDeviceRequest
		DeviceDiffs                 []DeviceDiffFunc
		ApplicationUplinkAssertions []func(context.Context, *ttnpb.EndDevice, *ttnpb.ApplicationUp) bool
		DownlinkAssertion           func(context.Context, TestClusterEnvironment, *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool)
		ErrorAssertion              func(*testing.T, error) bool
	}{
		{
			Name: "no device",
		},

		{
			Name: "no MAC state",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"session",
				},
			},
		},

		{
			Name: "Class A/windows closed",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: time.Now().Add(-time.Second),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"session",
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/no uplink",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters:  makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters:  makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:        ttnpb.CLASS_A,
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"session",
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/no session",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-time.Second),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/RX1,RX2 expired",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters:  makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters:  makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:        ttnpb.CLASS_A,
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-2*time.Second - time.Nanosecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
					},
					MACSettings: &ttnpb.MACSettings{
						StatusTimePeriodicity:  DurationPtr(0),
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"mac_settings",
					"session",
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/RX1,RX2 available/no MAC/no application downlink",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-100 * time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					MACSettings: &ttnpb.MACSettings{
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
						StatusTimePeriodicity:  DurationPtr(0),
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"mac_settings",
					"session",
				},
			},
		},

		{
			Name: "Class A/windows open/1.0.3/RX1,RX2 available/no MAC/generic application downlink/FCnt too low",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_3_REV_A,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_0_3_REV_A),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_0_3_REV_A),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_0_3,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-500 * time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					MACSettings: &ttnpb.MACSettings{
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
						StatusTimePeriodicity:  DurationPtr(0),
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x22,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
							{
								CorrelationIDs: []string{"correlation-app-down-3", "correlation-app-down-4"},
								FCnt:           0x23,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"mac_settings",
					"session",
				},
			},
			DeviceDiffs: []DeviceDiffFunc{
				makeRemoveDownlinksDiff(1),
			},
			ApplicationUplinkAssertions: []func(context.Context, *ttnpb.EndDevice, *ttnpb.ApplicationUp) bool{
				func(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.ApplicationUp) bool {
					return assertions.New(test.MustTFromContext(ctx)).So(up, should.Resemble, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
						CorrelationIDs:       LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs,
						Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
							DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
								LastFCntDown: dev.Session.LastNFCntDown,
								Downlinks:    dev.Session.QueuedApplicationDownlinks,
							},
						},
					})
				},
			},
		},
		{
			Name: "Class A/windows open/1.0.3/RX1,RX2 available/no MAC/generic application downlink/application downlink exceeds length limit",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_3_REV_A,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_0_3_REV_A),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_0_3_REV_A),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_0_3,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-500 * time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					MACSettings: &ttnpb.MACSettings{
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
						StatusTimePeriodicity:  DurationPtr(0),
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     bytes.Repeat([]byte("x"), 250),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"mac_settings",
					"session",
				},
			},
			DeviceDiffs: []DeviceDiffFunc{
				makeRemoveDownlinksDiff(1),
			},
			ApplicationUplinkAssertions: []func(context.Context, *ttnpb.EndDevice, *ttnpb.ApplicationUp) bool{
				func(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.ApplicationUp) bool {
					return assertions.New(test.MustTFromContext(ctx)).So(up, should.Resemble, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
						CorrelationIDs:       append(LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
						Up: &ttnpb.ApplicationUp_DownlinkFailed{
							DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
								ApplicationDownlink: *dev.Session.QueuedApplicationDownlinks[0],
								Error:               *ttnpb.ErrorDetailsToProto(ErrApplicationDownlinkTooLong.WithAttributes("length", 250, "max", uint16(51))),
							},
						},
					})
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/RX1,RX2 available/MAC answers/MAC requests/generic application downlink/data+MAC/RX1,RX2/EU868/scheduling fail",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						QueuedResponses: []*ttnpb.MACCommand{
							(&ttnpb.MACCommand_ResetConf{
								MinorVersion: 1,
							}).MACCommand(),
							(&ttnpb.MACCommand_LinkCheckAns{
								Margin:       2,
								GatewayCount: 5,
							}).MACCommand(),
						},
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"session",
				},
			},
			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0110,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x24,
							[]byte{
								/* ResetConf */
								0x01, 0b0000_0001,
								/* LinkCheckAns */
								0x02, 0x02, 0x05,
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/** FPort **/
						b = append(b, 0x1)

						/** FRMPayload **/
						b = append(b, []byte("testPayload")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx1DataRateIndex: ttnpb.DATA_RATE_0,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     431000000,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					append(LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				removeMACQueueDiff,
				setRxWindowsUnavailableDiff,
			},
		},

		{
			Name: "Class A/windows open/1.1/RX1,RX2 available/MAC answers/MAC requests/generic application downlink/data+MAC/RX1,RX2/EU868",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						QueuedResponses: []*ttnpb.MACCommand{
							(&ttnpb.MACCommand_ResetConf{
								MinorVersion: 1,
							}).MACCommand(),
							(&ttnpb.MACCommand_LinkCheckAns{
								Margin:       2,
								GatewayCount: 5,
							}).MACCommand(),
						},
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-500 * time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"session",
				},
			},
			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0110,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x24,
							[]byte{
								/* ResetConf */
								0x01, 0b0000_0001,
								/* LinkCheckAns */
								0x02, 0x02, 0x05,
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/** FPort **/
						b = append(b, 0x1)

						/** FRMPayload **/
						b = append(b, []byte("testPayload")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx1DataRateIndex: ttnpb.DATA_RATE_0,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     431000000,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{
							Delay: time.Second,
						},
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					append(LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				makeRemoveDownlinksDiff(1),
				setLastConfirmedDownlinkAtDiff,
				setLastDownlinkAtDiff,
				removeMACQueueDiff,
				setRxWindowsUnavailableDiff,
			},
		},

		{
			Name: "Class A/windows open/1.1/RX1,RX2 available/MAC answers/MAC requests/generic application downlink/application downlink does not fit due to FOpts/MAC/RX1,RX2/EU868",
			// NOTE: Maximum MACPayload length in both RX1(DR0) and RX2(DR1) is 59. There are 6 bytes of FOpts, hence maximum fitting application downlink length is 59-8-6 == 45.
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						QueuedResponses: []*ttnpb.MACCommand{
							(&ttnpb.MACCommand_ResetConf{
								MinorVersion: 1,
							}).MACCommand(),
							(&ttnpb.MACCommand_LinkCheckAns{
								Margin:       2,
								GatewayCount: 5,
							}).MACCommand(),
						},
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-500 * time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x15,
								FRMPayload:     bytes.Repeat([]byte{0x42}, 46),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"session",
				},
			},
			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_1_0110,
							/*** FCnt ***/
							0x25, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x25,
							[]byte{
								/* ResetConf */
								0x01, 0b0000_0001,
								/* LinkCheckAns */
								0x02, 0x02, 0x05,
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x25,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx1DataRateIndex: ttnpb.DATA_RATE_0,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     431000000,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{
							Delay: time.Second,
						},
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs,
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				makeRemoveDownlinksDiff(1),
				makeSetLastNFCntDownDiff(0x25),
				setLastConfirmedDownlinkAtDiff,
				setLastDownlinkAtDiff,
				removeMACQueueDiff,
				setRxWindowsUnavailableDiff,
			},
		},

		// Adapted from https://github.com/TheThingsNetwork/lorawan-stack/issues/866#issue-461484955.
		{
			Name: "Class A/windows open/1.0.2/RX1,RX2 available/MAC answers/MAC requests/generic application downlink/data+MAC/RX2 does not fit/RX1/EU868",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						QueuedResponses: []*ttnpb.MACCommand{
							(&ttnpb.MACCommand_ResetConf{
								MinorVersion: 1,
							}).MACCommand(),
							(&ttnpb.MACCommand_LinkCheckAns{
								Margin:       2,
								GatewayCount: 5,
							}).MACCommand(),
						},
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-500 * time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_6,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x15,
								FRMPayload:     []byte("AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4OTo7PD0+P0BBQkNERUU="),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"session",
				},
			},
			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0110,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x24,
							[]byte{
								/* ResetConf */
								0x01, 0b0000_0001,
								/* LinkCheckAns */
								0x02, 0x02, 0x05,
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/** FPort **/
						b = append(b, 0x15)

						/** FRMPayload **/
						b = append(b, []byte("AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4OTo7PD0+P0BBQkNERUU=")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx1DataRateIndex: ttnpb.DATA_RATE_6,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     431000000,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{
							Delay: time.Second,
						},
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					append(LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				setRxWindowsUnavailableDiff,
				removeMACQueueDiff,
				setLastDownlinkAtDiff,
				setLastConfirmedDownlinkAtDiff,
				appendRecentMACStateDownlinkDiff,
				makeRemoveDownlinksDiff(1),
			},
		},

		{
			Name: "Class B/windows closed/ping slot",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACSettings: &ttnpb.MACSettings{
						ClassBTimeout: DurationPtr(42 * time.Second),
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_B,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_8S,
						},
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
										FHDR: ttnpb.FHDR{
											FCtrl: ttnpb.FCtrl{
												ClassB: true,
											},
										},
									}},
								},
								ReceivedAt: start.Add(-500 * time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								Confirmed:      true,
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
							&ttnpb.ApplicationDownlink{
								CorrelationIDs: []string{"correlation-app-down-3", "correlation-app-down-4"},
								FCnt:           0x43,
								FPort:          0x2,
								FRMPayload:     []byte("nextTestPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGH,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"session",
				},
			},

			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				delay := dev.MACState.DesiredParameters.Rx1Delay.Duration() / 2

				pingAt, ok := NextPingSlotAt(ctx, dev, start.Add(delay))
				if !ok {
					t.Errorf("Failed to determine ping slot time")
					return nil, time.Time{}, false
				}
				clock.Set(pingAt.Add(-NSScheduleWindow() - delay))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b101_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_1_0001,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x24,
							[]byte{
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/** FPort **/
						b = append(b, 0x1)

						/** FRMPayload **/
						b = append(b, []byte("testPayload")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_B,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx2DataRateIndex: ttnpb.DATA_RATE_3,
							Rx2Frequency:     869525000,
							AbsoluteTime:     TimePtr(pingAt.UTC()),
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{
							Delay: time.Second,
						},
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				setPendingApplicationDownlinkDiff,
				setLastNetworkInitiatedDownlinkAtDiff,
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				setLastDownlinkAtDiff,
				setLastConfirmedDownlinkAtDiff,
				appendRecentMACStateDownlinkDiff,
				makeSetLastConfFCntDownDiff(0x42),
			},
		},

		{
			Name: "Class C/windows open/1.1/RX1,RX2 available/MAC answers/MAC requests/generic application downlink/data+MAC/RX1,RX2/EU868",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACSettings: &ttnpb.MACSettings{
						ClassCTimeout: DurationPtr(42 * time.Second),
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_C,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						QueuedResponses: []*ttnpb.MACCommand{
							(&ttnpb.MACCommand_ResetConf{
								MinorVersion: 1,
							}).MACCommand(),
							(&ttnpb.MACCommand_LinkCheckAns{
								Margin:       2,
								GatewayCount: 5,
							}).MACCommand(),
						},
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-500 * time.Millisecond),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"session",
				},
			},

			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0110,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x24,
							[]byte{
								/* ResetConf */
								0x01, 0b0000_0001,
								/* LinkCheckAns */
								0x02, 0x02, 0x05,
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/** FPort **/
						b = append(b, 0x1)

						/** FRMPayload **/
						b = append(b, []byte("testPayload")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx1DataRateIndex: ttnpb.DATA_RATE_0,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     431000000,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{
							Delay: time.Second,
						},
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}

				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					append(LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				setLastDownlinkAtDiff,
				makeSetLastNFCntDownDiff(0x24),
				makeRemoveDownlinksDiff(1),
				setLastConfirmedDownlinkAtDiff,
				appendRecentMACStateDownlinkDiff,
			},
		},

		{
			Name: "Class C/windows open/1.1/RX1,RX2 expired/MAC answers/MAC requests/generic application downlink/data+MAC/RXC/EU868",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACSettings: &ttnpb.MACSettings{
						ClassCTimeout: DurationPtr(42 * time.Second),
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_C,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						QueuedResponses: []*ttnpb.MACCommand{
							(&ttnpb.MACCommand_ResetConf{
								MinorVersion: 1,
							}).MACCommand(),
							(&ttnpb.MACCommand_LinkCheckAns{
								Margin:       2,
								GatewayCount: 5,
							}).MACCommand(),
						},
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-rx1Delay.Duration() - time.Second),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"session",
				},
			},

			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0001,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x24,
							[]byte{
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/** FPort **/
						b = append(b, 0x1)

						/** FRMPayload **/
						b = append(b, []byte("testPayload")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_C,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{
							Delay: time.Second,
						},
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				setLastNetworkInitiatedDownlinkAtDiff,
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				setRxWindowsUnavailableDiff,
				removeMACQueueDiff,
				setLastDownlinkAtDiff,
				makeRemoveDownlinksDiff(1),
				setLastConfirmedDownlinkAtDiff,
				appendRecentMACStateDownlinkDiff,
			},
		},

		{
			Name: "Class C/windows open/1.1/RX1,RX2 expired/no MAC answers/MAC requests/classBC application downlink/absolute time within window/no forced gateways/data+MAC/RXC/EU868",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:         test.EUFrequencyPlanID,
					LastDevStatusReceivedAt: TimePtr(start),
					LoRaWANPHYVersion:       ttnpb.PHY_V1_1_REV_B,
					MACSettings: &ttnpb.MACSettings{
						ClassCTimeout:         DurationPtr(42 * time.Second),
						StatusTimePeriodicity: DurationPtr(time.Hour),
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_C,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-2 * time.Second),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
								ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
									AbsoluteTime: TimePtr(start.Add(InfrastructureDelay)),
								},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"session",
				},
			},

			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0000,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FPort **/
						b = append(b, 0x1)

						/** FRMPayload **/
						b = append(b, []byte("testPayload")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_C,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_NORMAL,
							Rx2DataRateIndex: ttnpb.DATA_RATE_0,
							Rx2Frequency:     869525000,
							AbsoluteTime:     TimePtr(start.Add(InfrastructureDelay)),
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{
							Delay: time.Second,
						},
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				setLastNetworkInitiatedDownlinkAtDiff,
				setRxWindowsUnavailableDiff,
				setLastDownlinkAtDiff,
				makeRemoveDownlinksDiff(1),
				appendRecentMACStateDownlinkDiff,
			},
		},

		{
			Name: "Class C/windows closed/1.1/no MAC answers/MAC requests/classBC application downlink with absolute time/no forced gateways/MAC/RXC/EU868/non-retryable errors",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACSettings: &ttnpb.MACSettings{
						ClassCTimeout: DurationPtr(42 * time.Second),
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_C,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-rx1Delay.Duration()),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
								ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
									AbsoluteTime: TimePtr(start.Add(rx1Delay.Duration()).UTC()),
								},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"session",
				},
			},

			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0001,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x24,
							[]byte{
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/** FPort **/
						b = append(b, 0x1)

						/** FRMPayload **/
						b = append(b, []byte("testPayload")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_C,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							AbsoluteTime:     TimePtr(start.Add(rx1Delay.Duration()).UTC()),
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test").WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineAborted(ulid.MustNew(0, test.Randy).String(), "aborted")),
								ttnpb.ErrorDetailsToProto(errors.DefineResourceExhausted(ulid.MustNew(0, test.Randy).String(), "resource exhausted")),
							},
						}),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test").WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineFailedPrecondition(ulid.MustNew(0, test.Randy).String(), "failed precondition")),
							},
						}),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test").WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineResourceExhausted(ulid.MustNew(0, test.Randy).String(), "resource exhausted")),
							},
						}),
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				setRxWindowsUnavailableDiff,
				removeMACQueueDiff,
				makeRemoveDownlinksDiff(1),
			},
			ApplicationUplinkAssertions: []func(context.Context, *ttnpb.EndDevice, *ttnpb.ApplicationUp) bool{
				func(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.ApplicationUp) bool {
					return assertions.New(test.MustTFromContext(ctx)).So(up, should.Resemble, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
						CorrelationIDs:       dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
						Up: &ttnpb.ApplicationUp_DownlinkFailed{
							DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
								ApplicationDownlink: *dev.Session.QueuedApplicationDownlinks[0],
								Error:               *ttnpb.ErrorDetailsToProto(ErrInvalidAbsoluteTime),
							},
						},
					})
				},
			},
		},

		{
			Name: "Class C/windows closed/1.1/no MAC answers/MAC requests/classBC application downlink/forced gateways/MAC/RXC/EU868/retryable error",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACSettings: &ttnpb.MACSettings{
						ClassCTimeout: DurationPtr(42 * time.Second),
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_C,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-2 * rx1Delay.Duration()),
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
								ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
									Gateways: GatewayAntennaIdentifiers[:],
								},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"session",
				},
			},

			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleClassBCGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b011_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0001,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							nwkSEncKey,
							devAddr,
							0x24,
							[]byte{
								/* DevStatusReq */
								0x06,
							},
							true,
						)).([]byte)...)

						/** FPort **/
						b = append(b, 0x1)

						/** FRMPayload **/
						b = append(b, []byte("testPayload")...)

						/* MIC */
						mic := test.Must(crypto.ComputeDownlinkMIC(
							sNwkSIntKey,
							devAddr,
							0,
							0x42,
							b,
						)).([4]byte)
						return append(b, mic[:]...)
					}(),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_C,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test").WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineAborted(ulid.MustNew(0, test.Randy).String(), "aborted")),
								ttnpb.ErrorDetailsToProto(errors.DefineResourceExhausted(ulid.MustNew(0, test.Randy).String(), "resource exhausted")),
							},
						}),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test").WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineCorruption(ulid.MustNew(0, test.Randy).String(), "corruption")), // retryable
							},
						}),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test").WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineResourceExhausted(ulid.MustNew(0, test.Randy).String(), "resource exhausted")),
							},
						}),
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return nil, time.Time{}, a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
				)
			},
		},

		{
			Name: "Class C/windows open/1.1/RX1,RX2 available/no MAC/classBC application downlink/absolute time outside window",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACSettings: &ttnpb.MACSettings{
						ClassCTimeout:          DurationPtr(42 * time.Second),
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
						StatusTimePeriodicity:  DurationPtr(0),
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_C,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: clock.Now().Add(-time.Second),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
								ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
									AbsoluteTime: TimePtr(clock.Now().Add(42 * time.Hour).UTC()),
								},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"session",
				},
			},
		},

		{
			Name: "Class C/windows open/1.1/RX1,RX2 available/no MAC/expired application downlinks",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACSettings: &ttnpb.MACSettings{
						ClassCTimeout:          DurationPtr(42 * time.Second),
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
						StatusTimePeriodicity:  DurationPtr(0),
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_C,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{
							{
								CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
								DeviceChannelIndex: 3,
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_UNCONFIRMED_UP,
									},
									Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
								},
								ReceivedAt: start.Add(-time.Second),
								RxMetadata: RxMetadata[:],
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_0,
									Frequency:     430000000,
								},
							},
						},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
								ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
									AbsoluteTime: TimePtr(start.Add(-2).UTC()),
								},
							},
							{
								CorrelationIDs: []string{"correlation-app-down-3", "correlation-app-down-4"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
								ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
									AbsoluteTime: TimePtr(start.Add(-1).UTC()),
								},
							},
						},
					},
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"mac_state",
					"mac_settings",
					"session",
				},
			},
			DeviceDiffs: []DeviceDiffFunc{
				makeRemoveDownlinksDiff(1),
			},
		},

		{
			Name: "join-accept/windows open/RX1,RX2 available/no active MAC state/EU868",
			CreateDevice: SetDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &types.DevAddr{0x42, 0xff, 0xff, 0xff},
						JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					PendingMACState: &ttnpb.MACState{
						CurrentParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters: makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:       ttnpb.CLASS_A,
						LoRaWANVersion:    ttnpb.MAC_V1_1,
						QueuedJoinAccept: &ttnpb.MACState_JoinAccept{
							Keys:    *sessionKeys,
							Payload: bytes.Repeat([]byte{0x42}, 33),
							Request: ttnpb.JoinRequest{
								DevAddr: devAddr,
							},
						},
						RxWindowsAvailable: true,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						{
							CorrelationIDs:     []string{"correlation-up-1", "correlation-up-2"},
							DeviceChannelIndex: 3,
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_JOIN_REQUEST,
								},
								Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
									JoinEUI:  types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
									DevEUI:   types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
									DevNonce: types.DevNonce{0x00, 0x42},
								}},
							},
							ReceivedAt: start.Add(-time.Second),
							RxMetadata: RxMetadata[:],
							Settings: ttnpb.TxSettings{
								DataRateIndex: ttnpb.DATA_RATE_0,
								Frequency:     430000000,
							},
						},
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 0x24,
						SessionKeys:   *sessionKeys,
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							{
								CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
								FCnt:           0x42,
								FPort:          0x1,
								FRMPayload:     []byte("testPayload"),
								Priority:       ttnpb.TxSchedulePriority_HIGHEST,
								SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					},
					SupportsJoin: true,
				},
				Paths: []string{
					"frequency_plan_id",
					"ids",
					"lorawan_phy_version",
					"pending_mac_state",
					"recent_uplinks",
					"session",
					"supports_join",
				},
			},
			DownlinkAssertion: func(ctx context.Context, env TestClusterEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				scheduleDownlink124Ch := make(chan NsGsScheduleDownlinkRequest)
				peer124 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink124Ch),
				})

				scheduleDownlink3Ch := make(chan NsGsScheduleDownlinkRequest)
				peer3 := NewGSPeer(ctx, &MockNsGsServer{
					ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlink3Ch),
				})

				if !a.So(assertGetGatewayPeers(ctx, env.GetPeer, peer124, peer3), should.BeTrue) {
					return nil, time.Time{}, false
				}

				lastDown, ok := assertScheduleRxMetadataGateways(
					ctx,
					env.Auth,
					scheduleDownlink124Ch,
					scheduleDownlink3Ch,
					bytes.Repeat([]byte{0x42}, 33),
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							Rx1DataRateIndex: ttnpb.DATA_RATE_0,
							Rx1Delay:         ttnpb.RX_DELAY_5,
							Rx1Frequency:     431000000,
							Rx2DataRateIndex: dev.PendingMACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.PendingMACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Error: errors.New("test"),
					},
					NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{
							Delay: time.Second,
						},
					},
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}

				return lastDown, start.Add(time.Second), a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
					LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs,
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				makeRemoveDownlinksDiff(1),
				func(ctx context.Context, expected, created, updated *ttnpb.EndDevice, down *ttnpb.DownlinkMessage, downAt time.Time) bool {
					expected.PendingMACState.QueuedJoinAccept = nil
					expected.PendingMACState.PendingJoinRequest = &ttnpb.JoinRequest{
						DevAddr: devAddr,
					}
					expected.PendingMACState.RxWindowsAvailable = false
					expected.PendingSession = &ttnpb.Session{
						DevAddr:     devAddr,
						SessionKeys: *sessionKeys,
					}
					return true
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			errCh := make(chan error, 1)
			_, ctx, env, stop := StartTest(t, TestConfig{
				TaskStarter: component.StartTaskFunc(func(ctx context.Context, id string, fn component.TaskFunc, restart component.TaskRestart, jitter float64, backoff ...time.Duration) {
					if id != DownlinkProcessTaskName {
						component.DefaultStartTask(ctx, id, fn, restart, jitter, backoff...)
						return
					}
					go func() {
						errCh <- fn(ctx)
						select {
						case <-ctx.Done():
							return
						default:
						}
					}()
				}),
				Timeout: (1 << 10) * test.Delay,
			})
			defer stop()
			go LogEvents(t, env.Events)

			var created *ttnpb.EndDevice
			if tc.CreateDevice.EndDevice != nil {
				created, ctx = MustCreateDevice(ctx, env.Devices, tc.CreateDevice.EndDevice, tc.CreateDevice.Paths...)
			}
			test.Must(nil, env.DownlinkTasks.Add(ctx, ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID,
				DeviceID:               devID,
			}, time.Now(), true))

			select {
			case <-ctx.Done():
				t.Error("Timed out while waiting for processDownlinkTask to return")

			case err := <-errCh:
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
			}

			var (
				down   *ttnpb.DownlinkMessage
				downAt time.Time
			)
			if tc.DownlinkAssertion != nil {
				var ok bool
				down, downAt, ok = tc.DownlinkAssertion(ctx, env.Cluster, created)
				if !a.So(ok, should.BeTrue) {
					t.FailNow()
				}
			}

			updated, ctx, err := env.Devices.GetByID(ctx, appID, devID, ttnpb.EndDeviceFieldPathsTopLevel)
			if tc.CreateDevice.EndDevice != nil {
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
			} else {
				if !test.AllTrue(
					a.So(err, should.NotBeNil),
					a.So(errors.IsNotFound(err), should.BeTrue),
				) {
					t.FailNow()
				}
			}
			if len(tc.DeviceDiffs) == 0 {
				a.So(updated, should.Resemble, created)
			} else {
				expected := CopyEndDevice(created)
				expected.UpdatedAt = updated.UpdatedAt
				if !a.So(updated.UpdatedAt, should.HappenAfter, created.UpdatedAt) {
					t.FailNow()
				}
				if down != nil {
					expected.RecentDownlinks = AppendRecentDownlink(expected.RecentDownlinks, down, RecentDownlinkCount)
				}
				for _, diff := range tc.DeviceDiffs {
					if !a.So(diff(ctx, created, updated, expected, down, downAt), should.BeTrue) {
						t.FailNow()
					}
				}
				if !test.AllTrue(
					a.So(updated, should.Resemble, expected),
					a.So(err, should.BeNil),
				) {
					t.FailNow()
				}
			}

			a.So(env.AssertWithApplicationLink(ctx, appID, func(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient) bool {
				return a.So(AssertReceiveApplicationUplinks(ctx, link, func() []func(context.Context, *ttnpb.ApplicationUp, error) bool {
					var upAsserts []func(context.Context, *ttnpb.ApplicationUp, error) bool
					for _, assert := range tc.ApplicationUplinkAssertions {
						upAsserts = append(upAsserts, func(ctx context.Context, up *ttnpb.ApplicationUp, err error) bool {
							return test.AllTrue(
								a.So(err, should.BeNil),
								a.So(assert(ctx, updated, up), should.BeTrue),
							)
						})
					}
					return upAsserts
				}()...), should.BeTrue)
			}), should.BeTrue)
		})
	}
}
