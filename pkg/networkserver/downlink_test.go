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
	"github.com/mohae/deepcopy"
	"github.com/oklog/ulid/v2"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// handleDownlinkTaskQueueTest runs a test suite on q.
func handleDownlinkTaskQueueTest(ctx context.Context, q DownlinkTaskQueue) {
	t, a := test.MustNewTFromContext(ctx)

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

	// Ensure Pop is blocking on empty queue.
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
			t.Fatal(s.ctx)
		}
		s.errCh <- nil

	case err := <-errCh:
		a.So(err, should.BeNil)
		t.Fatal("Pop returned without calling f on non-empty schedule")

	case <-ctx.Done():
		t.Fatal("Timed out waiting for Pop to call f")
	}

	select {
	case err := <-errCh:
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

	case <-ctx.Done():
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

		case <-ctx.Done():
			t.Fatal("Timed out waiting for Pop to call f")
		}

		select {
		case err := <-errCh:
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

		case <-ctx.Done():
			t.Fatal("Timed out waiting for Pop to return")
		}
	}

	expectSlot(t, pbs[0], time.Unix(0, 42))
	expectSlot(t, pbs[1], time.Unix(42, 0))
	expectSlot(t, pbs[2], time.Unix(42, 42))
}

func TestDownlinkTaskQueues(t *testing.T) {
	test.RunTest(t, test.TestConfig{
		Func: func(ctx context.Context, _ *assertions.Assertion) {
			for _, tc := range []struct {
				Name string
				New  func(t testing.TB) (q DownlinkTaskQueue, closeFn func())
			}{
				{
					Name: "Redis",
					New:  NewRedisDownlinkTaskQueue,
				},
			} {
				tc := tc
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name:     tc.Name,
					Parallel: true,
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						q, closeFn := tc.New(t)
						if closeFn != nil {
							defer closeFn()
						}
						test.RunSubtestFromContext(ctx, test.SubtestConfig{
							Name: "1st run",
							Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
								handleDownlinkTaskQueueTest(ctx, q)
							},
						})
						if t.Failed() {
							t.Skip("Skipping 2nd run")
						}
						test.RunSubtestFromContext(ctx, test.SubtestConfig{
							Name: "2st run",
							Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
								handleDownlinkTaskQueueTest(ctx, q)
							},
						})
					},
				})
			}
		},
	})
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

	joinAcceptBytes := append([]byte{0b001_000_00}, bytes.Repeat([]byte{0x42}, 32)...)

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

	customCh := &ttnpb.MACParameters_Channel{
		UplinkFrequency:   430000000,
		DownlinkFrequency: 431000000,
		MinDataRateIndex:  ttnpb.DATA_RATE_0,
		MaxDataRateIndex:  ttnpb.DATA_RATE_6,
	}
	const customChIdx = 3
	makeEU868macParameters := func(ver ttnpb.PHYVersion) ttnpb.MACParameters {
		params := MakeDefaultEU868CurrentMACParameters(ver)
		if len(params.Channels) != customChIdx {
			panic(fmt.Sprintf("invalid EU868 default channel count, expected %d, got %d", customChIdx, len(params.Channels)))
		}
		params.Channels = append(params.Channels, deepcopy.Copy(customCh).(*ttnpb.MACParameters_Channel))
		return params
	}

	assertScheduleGateways := func(ctx context.Context, env TestEnvironment, fixedPaths bool, payload []byte, makeTxRequest func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest, resps ...NsGsScheduleDownlinkResponse) (*ttnpb.DownlinkMessage, bool) {
		if len(resps) < 1 || len(resps) > 3 {
			panic("invalid response count specified")
		}

		_, a := test.MustNewTFromContext(ctx)

		var downlinkPaths []DownlinkPath
		if !fixedPaths {
			downlinkPaths = DownlinkPathsFromMetadata(RxMetadata[:]...)
		} else {
			for i, ids := range GatewayAntennaIdentifiers {
				ids := ids
				downlinkPaths = append(downlinkPaths, DownlinkPath{
					GatewayIdentifiers: &ids.GatewayIdentifiers,
					DownlinkPath: &ttnpb.DownlinkPath{
						Path: &ttnpb.DownlinkPath_Fixed{
							Fixed: &GatewayAntennaIdentifiers[i],
						},
					},
				})
			}
		}

		var lastDown *ttnpb.DownlinkMessage
		var asserts []func(ctx, reqCtx context.Context, msg *ttnpb.DownlinkMessage) (NsGsScheduleDownlinkResponse, bool)
		for i, resp := range resps {
			i := i
			resp := resp
			asserts = append(asserts, func(ctx, reqCtx context.Context, msg *ttnpb.DownlinkMessage) (NsGsScheduleDownlinkResponse, bool) {
				lastDown = msg
				_, a := test.MustNewTFromContext(ctx)
				return resp, test.AllTrue(
					a.So(events.CorrelationIDsFromContext(reqCtx), should.NotBeEmpty),
					a.So(msg.CorrelationIDs, should.NotBeEmpty),
					a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
						CorrelationIDs: msg.CorrelationIDs,
						RawPayload:     payload,
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: makeTxRequest(func() []*ttnpb.DownlinkPath {
								switch i {
								case 0:
									return []*ttnpb.DownlinkPath{
										downlinkPaths[1].DownlinkPath,
										downlinkPaths[2].DownlinkPath,
									}
								case 1:
									return []*ttnpb.DownlinkPath{
										downlinkPaths[3].DownlinkPath,
									}
								case 2:
									return []*ttnpb.DownlinkPath{
										downlinkPaths[4].DownlinkPath,
									}
								default:
									panic("invalid response count")
								}
							}()...),
						},
					}),
				)
			})
		}
		if !a.So(env.AssertLegacyScheduleDownlink(
			ctx,
			MakeDownlinkPathsWithPeerIndex(downlinkPaths, []uint{0, 1, 1, 2, 1}...),
			asserts...,
		), should.BeTrue) {
			t.Error("NsGs.ScheduleDownlink assertion failed")
			return nil, false
		}
		return lastDown, true
	}

	makeFailEventEqual := func(t *testing.T) func(x, y events.Event) bool {
		t.Helper()
		a := assertions.New(t)
		return func(x, y events.Event) bool {
			xProto := test.Must(events.Proto(x)).(*ttnpb.Event)
			yProto := test.Must(events.Proto(y)).(*ttnpb.Event)
			xProto.Time = time.Time{}
			yProto.Time = time.Time{}
			xProto.Data = nil
			yProto.Data = nil
			xProto.UniqueID = ""
			yProto.UniqueID = ""
			return test.AllTrue(
				a.So(x.Data(), should.BeError) && a.So(y.Data(), should.BeError) && a.So(x.Data(), should.HaveSameErrorDefinitionAs, y.Data()),
				a.So(xProto, should.Resemble, yProto),
			)
		}
	}

	makeAssertReceiveScheduleFailAttemptEvents := func(attempt, fail events.Builder) func(context.Context, TestEnvironment, *ttnpb.DownlinkMessage, ttnpb.EndDeviceIdentifiers, error, uint) bool {
		return func(ctx context.Context, env TestEnvironment, down *ttnpb.DownlinkMessage, ids ttnpb.EndDeviceIdentifiers, err error, n uint) bool {
			_, a := test.MustNewTFromContext(ctx)
			ctx = events.ContextWithCorrelationID(ctx, down.CorrelationIDs...)
			evIDOpt := events.WithIdentifiers(ids)
			for i := uint(0); i < n; i++ {
				if !test.AllTrue(
					a.So(env.Events, should.ReceiveEventFunc, test.MakeEventEqual(test.EventEqualConfig{
						Identifiers:    true,
						CorrelationIDs: true,
						Origin:         true,
						Context:        true,
						Visibility:     true,
						Authentication: true,
						RemoteIP:       true,
						UserAgent:      true,
					}),
						attempt.With(events.WithData(down)).New(ctx, evIDOpt),
					),
					a.So(env.Events, should.ReceiveEventFunc, makeFailEventEqual(t),
						fail.With(events.WithData(err)).New(ctx, evIDOpt)),
				) {
					return false
				}
			}
			return true
		}
	}
	makeAssertReceiveScheduleSuccessAttemptEvents := func(attempt, success events.Builder) func(context.Context, TestEnvironment, *ttnpb.DownlinkMessage, ttnpb.EndDeviceIdentifiers, *ttnpb.ScheduleDownlinkResponse, ...events.Builder) bool {
		return func(ctx context.Context, env TestEnvironment, down *ttnpb.DownlinkMessage, ids ttnpb.EndDeviceIdentifiers, resp *ttnpb.ScheduleDownlinkResponse, evs ...events.Builder) bool {
			_, a := test.MustNewTFromContext(ctx)
			ctx = events.ContextWithCorrelationID(ctx, down.CorrelationIDs...)
			evIDOpt := events.WithIdentifiers(ids)
			return test.AllTrue(
				a.So(env.Events, should.ReceiveEventFunc, test.MakeEventEqual(test.EventEqualConfig{
					Identifiers:    true,
					CorrelationIDs: true,
					Origin:         true,
					Context:        true,
					Visibility:     true,
					Authentication: true,
					RemoteIP:       true,
					UserAgent:      true,
				}),
					attempt.With(events.WithData(down)).New(ctx, evIDOpt),
				),
				a.So(env.Events, should.ReceiveEventsResembling, events.Builders(append([]events.Builder{
					success.With(events.WithData(resp)),
				}, evs...)).New(ctx, evIDOpt)),
			)
		}
	}

	assertReceiveScheduleDataFailAttemptEvents := makeAssertReceiveScheduleFailAttemptEvents(
		EvtScheduleDataDownlinkAttempt,
		EvtScheduleDataDownlinkFail,
	)
	assertReceiveScheduleJoinFailAttemptEvents := makeAssertReceiveScheduleFailAttemptEvents(
		EvtScheduleJoinAcceptAttempt,
		EvtScheduleJoinAcceptFail,
	)
	assertReceiveScheduleDataSuccessAttemptEvents := makeAssertReceiveScheduleSuccessAttemptEvents(
		EvtScheduleDataDownlinkAttempt,
		EvtScheduleDataDownlinkSuccess,
	)
	assertReceiveScheduleJoinSuccessAttemptEvents := makeAssertReceiveScheduleSuccessAttemptEvents(
		EvtScheduleJoinAcceptAttempt,
		EvtScheduleJoinAcceptSuccess,
	)

	pingSlotPeriodicity := &ttnpb.PingSlotPeriodValue{
		Value: ttnpb.PING_EVERY_8S,
	}

	_, ctx := test.New(t)
	pingAt, ok := mac.NextPingSlotAt(ctx, &ttnpb.EndDevice{
		Session: &ttnpb.Session{
			DevAddr: devAddr,
		},
		MACState: &ttnpb.MACState{
			PingSlotPeriodicity: pingSlotPeriodicity,
		},
	}, time.Now())
	if !ok {
		t.Fatal("Failed to compute ping slot")
	}
	now := pingAt.Add(-DefaultEU868RX1Delay.Duration()/2 - 2*NSScheduleWindow())
	clock := test.NewMockClock(now)
	defer SetMockClock(clock)()

	type DeviceDiffFunc func(ctx context.Context, expected, created, updated *ttnpb.EndDevice, down *ttnpb.DownlinkMessage, downAt time.Time) bool
	makeRemoveDownlinksDiff := func(n uint) DeviceDiffFunc {
		return func(ctx context.Context, expected, _, _ *ttnpb.EndDevice, _ *ttnpb.DownlinkMessage, _ time.Time) bool {
			_, a := test.MustNewTFromContext(ctx)
			if !a.So(len(expected.Session.QueuedApplicationDownlinks), should.BeGreaterThanOrEqualTo, n) {
				return false
			}
			expected.Session.QueuedApplicationDownlinks = expected.Session.QueuedApplicationDownlinks[n:]
			if len(expected.Session.QueuedApplicationDownlinks) == 0 {
				expected.Session.QueuedApplicationDownlinks = nil
			}
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
		expected.MACState.PendingApplicationDownlink = created.Session.QueuedApplicationDownlinks[0]
		return true
	})

	testErr := errors.New("test")
	testErrScheduleResponse := NsGsScheduleDownlinkResponse{
		Error: testErr,
	}
	oneSecondScheduleResponse := NsGsScheduleDownlinkResponse{
		Response: &ttnpb.ScheduleDownlinkResponse{
			Delay: time.Second,
		},
	}

	for _, tc := range []struct {
		Name                        string
		CreateDevice                SetDeviceRequest
		DeviceDiffs                 []DeviceDiffFunc
		ApplicationUplinkAssertions []func(context.Context, *ttnpb.EndDevice, *ttnpb.ApplicationUp) bool
		DownlinkAssertion           func(context.Context, TestEnvironment, *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool)
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								Confirmed:     true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								Confirmed:     true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								Confirmed:     true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second - time.Nanosecond),
							}),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_0_3,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_0_3_REV_A].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
				makeRemoveDownlinksDiff(2),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_0_3,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_0_3_REV_A].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastUp := LastUplink(dev.MACState.RecentUplinks...)
				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
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
							Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							Rx1DataRateIndex: lastUp.Settings.DataRateIndex,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     dev.MACState.CurrentParameters.Channels[lastUp.DeviceChannelIndex].DownlinkFrequency,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					testErrScheduleResponse,
					testErrScheduleResponse,
					testErrScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return nil, time.Time{}, test.AllTrue(
					ok,
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						append(lastUp.CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 3),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				setRxWindowsUnavailableDiff,
				removeMACQueueDiff,
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastUp := LastUplink(dev.MACState.RecentUplinks...)
				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
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
							Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							Rx1DataRateIndex: lastUp.Settings.DataRateIndex,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     dev.MACState.CurrentParameters.Channels[lastUp.DeviceChannelIndex].DownlinkFrequency,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					testErrScheduleResponse,
					testErrScheduleResponse,
					oneSecondScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, now.Add(time.Second), test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						append(lastUp.CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 2),
					assertReceiveScheduleDataSuccessAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, oneSecondScheduleResponse.Response,
						mac.EvtEnqueueDevStatusRequest,
					),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				setLastConfirmedDownlinkAtDiff,
				setLastDownlinkAtDiff,
				setRxWindowsUnavailableDiff,
				removeMACQueueDiff,
				makeRemoveDownlinksDiff(1),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								DataRateIndex: customCh.MinDataRateIndex,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastUp := LastUplink(dev.MACState.RecentUplinks...)
				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
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
							Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							Rx1DataRateIndex: lastUp.Settings.DataRateIndex,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     dev.MACState.CurrentParameters.Channels[lastUp.DeviceChannelIndex].DownlinkFrequency,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					testErrScheduleResponse,
					testErrScheduleResponse,
					oneSecondScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, now.Add(time.Second), test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						lastUp.CorrelationIDs,
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 2),
					assertReceiveScheduleDataSuccessAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, oneSecondScheduleResponse.Response,
						mac.EvtEnqueueDevStatusRequest,
					),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				makeSetLastNFCntDownDiff(0x25),
				setLastConfirmedDownlinkAtDiff,
				setLastDownlinkAtDiff,
				setRxWindowsUnavailableDiff,
				removeMACQueueDiff,
			},
		},

		// Adapted from https://github.com/TheThingsNetwork/lorawan-stack/issues/866#issue-461484955.
		{
			Name: "Class A/windows open/1.1/RX1,RX2 available/MAC answers/MAC requests/generic application downlink/data+MAC/RX2 does not fit/RX1/EU868",
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[ttnpb.DATA_RATE_6].Rate,
								DataRateIndex: ttnpb.DATA_RATE_6,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastUp := LastUplink(dev.MACState.RecentUplinks...)
				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
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
							Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							Rx1DataRateIndex: lastUp.Settings.DataRateIndex,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     dev.MACState.CurrentParameters.Channels[lastUp.DeviceChannelIndex].DownlinkFrequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					testErrScheduleResponse,
					testErrScheduleResponse,
					oneSecondScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, now.Add(time.Second), test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						append(lastUp.CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 2),
					assertReceiveScheduleDataSuccessAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, oneSecondScheduleResponse.Response,
						mac.EvtEnqueueDevStatusRequest,
					),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				setLastDownlinkAtDiff,
				setLastConfirmedDownlinkAtDiff,
				setRxWindowsUnavailableDiff,
				removeMACQueueDiff,
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
						CurrentParameters:   makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DesiredParameters:   makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
						DeviceClass:         ttnpb.CLASS_B,
						LoRaWANVersion:      ttnpb.MAC_V1_1,
						PingSlotPeriodicity: pingSlotPeriodicity,
						RecentUplinks: []*ttnpb.UplinkMessage{
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second),
								FCtrl: ttnpb.FCtrl{
									ClassB: true,
								},
							}),
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
							{
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
					func() []byte {
						b := []byte{
							/* MHDR */
							0b101_000_00,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0b1_0_0_1_0000,
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
							Class:            ttnpb.CLASS_B,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx2DataRateIndex: ttnpb.DATA_RATE_3,
							Rx2Frequency:     DefaultEU868RX2Frequency,
							AbsoluteTime:     TimePtr(pingAt),
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					oneSecondScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, now.Add(time.Second), test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
					),
					assertReceiveScheduleDataSuccessAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, oneSecondScheduleResponse.Response),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				setPendingApplicationDownlinkDiff,
				setLastNetworkInitiatedDownlinkAtDiff,
				setLastDownlinkAtDiff,
				setLastConfirmedDownlinkAtDiff,
				makeSetLastConfFCntDownDiff(0x42),
				makeRemoveDownlinksDiff(1),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
							}),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastUp := LastUplink(dev.MACState.RecentUplinks...)
				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
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
							Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							Rx1DataRateIndex: lastUp.Settings.DataRateIndex,
							Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
							Rx1Frequency:     dev.MACState.CurrentParameters.Channels[lastUp.DeviceChannelIndex].DownlinkFrequency,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					testErrScheduleResponse,
					testErrScheduleResponse,
					oneSecondScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, now.Add(time.Second), test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						append(lastUp.CorrelationIDs, dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs...),
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 2),
					assertReceiveScheduleDataSuccessAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, oneSecondScheduleResponse.Response,
						mac.EvtEnqueueDevStatusRequest,
					),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				makePendingMACCommandsDiff(
					ttnpb.CID_DEV_STATUS.MACCommand(),
				),
				makeSetLastNFCntDownDiff(0x24),
				setLastDownlinkAtDiff,
				setLastConfirmedDownlinkAtDiff,
				setRxWindowsUnavailableDiff,
				removeMACQueueDiff,
				makeRemoveDownlinksDiff(1),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second - time.Nanosecond),
							}),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
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
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					testErrScheduleResponse,
					testErrScheduleResponse,
					oneSecondScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, now.Add(time.Second), test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 2),
					assertReceiveScheduleDataSuccessAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, oneSecondScheduleResponse.Response),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				setLastDownlinkAtDiff,
				setLastNetworkInitiatedDownlinkAtDiff,
				setRxWindowsUnavailableDiff,
				makeRemoveDownlinksDiff(1),
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
					LastDevStatusReceivedAt: TimePtr(now),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second),
							}),
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
									AbsoluteTime: TimePtr(now.Add(InfrastructureDelay)),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
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
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx2DataRateIndex: ttnpb.DATA_RATE_0,
							Rx2Frequency:     DefaultEU868RX2Frequency,
							AbsoluteTime:     TimePtr(now.Add(InfrastructureDelay)),
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					testErrScheduleResponse,
					testErrScheduleResponse,
					oneSecondScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, now.Add(time.Second), test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 2),
					assertReceiveScheduleDataSuccessAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, oneSecondScheduleResponse.Response),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				appendRecentMACStateDownlinkDiff,
				setLastNetworkInitiatedDownlinkAtDiff,
				setLastDownlinkAtDiff,
				setRxWindowsUnavailableDiff,
				makeRemoveDownlinksDiff(1),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration()),
							}),
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
									AbsoluteTime: TimePtr(now.Add(DefaultEU868RX1Delay.Duration())),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
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
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							AbsoluteTime:     TimePtr(now.Add(DefaultEU868RX1Delay.Duration())),
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: testErr.WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineAborted(ulid.MustNew(0, test.Randy).String(), "aborted")),
								ttnpb.ErrorDetailsToProto(errors.DefineResourceExhausted(ulid.MustNew(0, test.Randy).String(), "resource exhausted")),
							},
						}),
					},
					NsGsScheduleDownlinkResponse{
						Error: testErr.WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineFailedPrecondition(ulid.MustNew(0, test.Randy).String(), "failed precondition")),
							},
						}),
					},
					NsGsScheduleDownlinkResponse{
						Error: testErr.WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
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
				return nil, time.Time{}, test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 3),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second),
							}),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					true,
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
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					NsGsScheduleDownlinkResponse{
						Error: testErr.WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineAborted(ulid.MustNew(0, test.Randy).String(), "aborted")),
								ttnpb.ErrorDetailsToProto(errors.DefineResourceExhausted(ulid.MustNew(0, test.Randy).String(), "resource exhausted")),
							},
						}),
					},
					NsGsScheduleDownlinkResponse{
						Error: testErr.WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
							PathErrors: []*ttnpb.ErrorDetails{
								ttnpb.ErrorDetailsToProto(errors.DefineCorruption(ulid.MustNew(0, test.Randy).String(), "corruption")), // retryable
							},
						}),
					},
					NsGsScheduleDownlinkResponse{
						Error: testErr.WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
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
				return nil, time.Time{}, test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						dev.Session.QueuedApplicationDownlinks[0].CorrelationIDs,
					),
					assertReceiveScheduleDataFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 3),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-time.Second),
							}),
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
									AbsoluteTime: TimePtr(now.Add(42 * time.Hour)),
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
							MakeDataUplink(DataUplinkConfig{
								DecodePayload: true,
								Matched:       true,
								MACVersion:    ttnpb.MAC_V1_1,
								DevAddr:       devAddr,
								DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
								Frequency:     customCh.UplinkFrequency,
								ChannelIndex:  customChIdx,
								RxMetadata:    RxMetadata[:],
								ReceivedAt:    now.Add(-time.Second),
							}),
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
									AbsoluteTime: TimePtr(now.Add(-2)),
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
									AbsoluteTime: TimePtr(now.Add(-1)),
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
							Payload: joinAcceptBytes,
							Request: ttnpb.JoinRequest{
								RawPayload: bytes.Repeat([]byte{0x42}, 23),
								DevAddr:    devAddr,
							},
						},
						RxWindowsAvailable: true,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						MakeJoinRequest(JoinRequestConfig{
							DecodePayload: true,
							JoinEUI:       types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
							DevEUI:        types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     DefaultEU868Channels[0].UplinkFrequency,
							RxMetadata:    RxMetadata[:],
							ReceivedAt:    now.Add(-time.Second),
						}),
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
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastUp := LastUplink(dev.RecentUplinks...)
				lastDown, ok := assertScheduleGateways(
					ctx,
					env,
					false,
					joinAcceptBytes,
					func(paths ...*ttnpb.DownlinkPath) *ttnpb.TxRequest {
						return &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							DownlinkPaths:    paths,
							Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							Rx1DataRateIndex: lastUp.Settings.DataRateIndex,
							Rx1Delay:         DefaultEU868JoinAcceptDelay,
							Rx1Frequency:     dev.PendingMACState.CurrentParameters.Channels[lastUp.DeviceChannelIndex].DownlinkFrequency,
							Rx2DataRateIndex: dev.PendingMACState.CurrentParameters.Rx2DataRateIndex,
							Rx2Frequency:     dev.PendingMACState.CurrentParameters.Rx2Frequency,
							FrequencyPlanID:  dev.FrequencyPlanID,
						}
					},
					testErrScheduleResponse,
					testErrScheduleResponse,
					oneSecondScheduleResponse,
				)
				if !a.So(ok, should.BeTrue) {
					t.Error("Scheduling assertion failed")
					return nil, time.Time{}, false
				}
				return lastDown, now.Add(time.Second), test.AllTrue(
					a.So(lastDown.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual,
						lastUp.CorrelationIDs,
					),
					assertReceiveScheduleJoinFailAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, testErr, 2),
					assertReceiveScheduleJoinSuccessAttemptEvents(ctx, env, lastDown, dev.EndDeviceIdentifiers, oneSecondScheduleResponse.Response),
				)
			},
			DeviceDiffs: []DeviceDiffFunc{
				func(ctx context.Context, expected, created, updated *ttnpb.EndDevice, down *ttnpb.DownlinkMessage, downAt time.Time) bool {
					expected.PendingMACState.RxWindowsAvailable = false
					expected.PendingMACState.PendingJoinRequest = &created.PendingMACState.QueuedJoinAccept.Request
					expected.PendingSession = &ttnpb.Session{
						DevAddr:     created.PendingMACState.QueuedJoinAccept.Request.DevAddr,
						SessionKeys: created.PendingMACState.QueuedJoinAccept.Keys,
					}
					expected.PendingMACState.QueuedJoinAccept = nil
					return true
				},
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name: tc.Name,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				errCh := make(chan error, 1)
				_, ctx, env, stop := StartTest(t, TestConfig{
					Context:       ctx,
					NetworkServer: DefaultConfig,
					TaskStarter: component.StartTaskFunc(func(conf *component.TaskConfig) {
						if conf.ID != DownlinkProcessTaskName {
							component.DefaultStartTask(conf)
							return
						}
						go func() {
							errCh <- conf.Func(conf.Context)
						}()
					}),
				})
				defer stop()

				var created *ttnpb.EndDevice
				if tc.CreateDevice.EndDevice != nil {
					created, ctx = MustCreateDevice(ctx, env.Devices, tc.CreateDevice.EndDevice, tc.CreateDevice.Paths...)
				}
				test.Must(nil, env.DownlinkTasks.Add(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
				}, time.Now(), true))

				var (
					down   *ttnpb.DownlinkMessage
					downAt time.Time
				)
				if tc.DownlinkAssertion != nil {
					var ok bool
					down, downAt, ok = tc.DownlinkAssertion(ctx, env, created)
					if !a.So(ok, should.BeTrue) {
						t.FailNow()
					}
				}

				select {
				case <-ctx.Done():
					t.Fatal("Timed out while waiting for processDownlinkTask to return")

				case err := <-errCh:
					if tc.ErrorAssertion == nil {
						a.So(err, should.BeNil)
					} else {
						a.So(tc.ErrorAssertion(t, err), should.BeTrue)
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
						msg := &ttnpb.Message{}
						test.Must(nil, lorawan.UnmarshalMessage(down.RawPayload, msg))
						switch msg.MType {
						case ttnpb.MType_CONFIRMED_DOWN, ttnpb.MType_UNCONFIRMED_DOWN:
							pld := msg.GetMACPayload()
							pld.FullFCnt = created.Session.LastNFCntDown&0xffff0000 | pld.FCnt
						case ttnpb.MType_JOIN_ACCEPT:
							msg.Payload = &ttnpb.Message_JoinAcceptPayload{
								JoinAcceptPayload: &ttnpb.JoinAcceptPayload{
									NetID:      created.PendingMACState.QueuedJoinAccept.Request.NetID,
									DevAddr:    created.PendingMACState.QueuedJoinAccept.Request.DevAddr,
									DLSettings: created.PendingMACState.QueuedJoinAccept.Request.DownlinkSettings,
									RxDelay:    created.PendingMACState.QueuedJoinAccept.Request.RxDelay,
									CFList:     created.PendingMACState.QueuedJoinAccept.Request.CFList,
								},
							}
						}
						down.RawPayload = nil
						down.Payload = msg
						expected.RecentDownlinks = AppendRecentDownlink(expected.RecentDownlinks, down, RecentDownlinkCount)
					}
					for _, diff := range tc.DeviceDiffs {
						if !a.So(diff(ctx, expected, created, updated, down, downAt), should.BeTrue) {
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
					return a.So(AssertProcessApplicationUps(ctx, link, func() []func(context.Context, *ttnpb.ApplicationUp) bool {
						var upAsserts []func(context.Context, *ttnpb.ApplicationUp) bool
						for _, assert := range tc.ApplicationUplinkAssertions {
							upAsserts = append(upAsserts, func(ctx context.Context, up *ttnpb.ApplicationUp) bool {
								_, a := test.MustNewTFromContext(ctx)
								return a.So(assert(ctx, created, up), should.BeTrue)
							})
						}
						return upAsserts
					}()...), should.BeTrue)
				}), should.BeTrue)
			},
		})
	}
}
