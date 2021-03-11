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

func TestProcessDownlinkTask(t *testing.T) {
	// TODO: Refactor. (https://github.com/TheThingsNetwork/lorawan-stack/issues/2475)

	const appIDString = "process-downlink-test-app-id"
	appID := ttnpb.ApplicationIdentifiers{ApplicationID: appIDString}
	const devID = "process-downlink-test-dev-id"

	joinAcceptBytes := append([]byte{0b001_000_00}, bytes.Repeat([]byte{0x42}, 32)...)

	sessionKeys := &ttnpb.SessionKeys{
		FNwkSIntKey: &ttnpb.KeyEnvelope{
			Key: &test.DefaultFNwkSIntKey,
		},
		NwkSEncKey: &ttnpb.KeyEnvelope{
			Key: &test.DefaultNwkSEncKey,
		},
		SNwkSIntKey: &ttnpb.KeyEnvelope{
			Key: &test.DefaultSNwkSIntKey,
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
			downlinkPaths = DownlinkPathsFromMetadata(DefaultRxMetadata[:]...)
		} else {
			for i, ids := range DefaultGatewayAntennaIdentifiers {
				ids := ids
				downlinkPaths = append(downlinkPaths, DownlinkPath{
					GatewayIdentifiers: &ids.GatewayIdentifiers,
					DownlinkPath: &ttnpb.DownlinkPath{
						Path: &ttnpb.DownlinkPath_Fixed{
							Fixed: &DefaultGatewayAntennaIdentifiers[i],
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

	attemptEventEqual := test.MakeEventEqual(test.EventEqualConfig{
		Identifiers:    true,
		CorrelationIDs: true,
		Origin:         true,
		Context:        true,
		Visibility:     true,
		Authentication: true,
		RemoteIP:       true,
		UserAgent:      true,
	})
	makeAssertReceiveScheduleFailAttemptEvents := func(attempt, fail events.Builder) func(context.Context, TestEnvironment, *ttnpb.DownlinkMessage, ttnpb.EndDeviceIdentifiers, error, uint) bool {
		return func(ctx context.Context, env TestEnvironment, down *ttnpb.DownlinkMessage, ids ttnpb.EndDeviceIdentifiers, err error, n uint) bool {
			_, a := test.MustNewTFromContext(ctx)
			ctx = events.ContextWithCorrelationID(ctx, down.CorrelationIDs...)
			evIDOpt := events.WithIdentifiers(ids)
			for i := uint(0); i < n; i++ {
				if !test.AllTrue(
					a.So(env.Events, should.ReceiveEventFunc, attemptEventEqual,
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
				a.So(env.Events, should.ReceiveEventFunc, attemptEventEqual,
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
			DevAddr: test.DefaultDevAddr,
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
		Name                       string
		CreateDevice               *ttnpb.EndDevice
		DeviceDiffs                []DeviceDiffFunc
		ApplicationUplinkAssertion func(context.Context, *ttnpb.EndDevice, ...*ttnpb.ApplicationUp) ([]events.Event, bool)
		DownlinkAssertion          func(context.Context, TestEnvironment, *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool)
		ErrorAssertion             func(*testing.T, error) bool
	}{
		{
			Name: "no device",
		},

		{
			Name: "no MAC state",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
				},
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
			},
		},

		{
			Name: "Class A/windows closed",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_0_3,
						}),
					},
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
					LastNFCntDown: 0x24,
					SessionKeys:   *sessionKeys,
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/no uplink",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
				},
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters:  makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
					DesiredParameters:  makeEU868macParameters(ttnpb.PHY_V1_1_REV_B),
					DeviceClass:        ttnpb.CLASS_A,
					LoRaWANVersion:     ttnpb.MAC_V1_1,
					RxWindowsAvailable: true,
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_0_3,
						}),
					},
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
					LastNFCntDown: 0x24,
					SessionKeys:   *sessionKeys,
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/no session",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/RX1,RX2 expired",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second - time.Nanosecond),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_0_3,
						}),
					},
				},
				MACSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity:  DurationPtr(0),
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
					LastNFCntDown: 0x24,
					SessionKeys:   *sessionKeys,
				},
			},
		},

		{
			Name: "Class A/windows open/1.1/RX1,RX2 available/no MAC/no application downlink",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_0_3,
						}),
					},
					RxWindowsAvailable: true,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
					StatusTimePeriodicity:  DurationPtr(0),
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
					LastNFCntDown: 0x24,
					SessionKeys:   *sessionKeys,
				},
			},
		},

		{
			Name: "Class A/windows open/1.0.3/RX1,RX2 available/no MAC/generic application downlink/FCnt too low",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_0_3_REV_A].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_0_3,
						}),
					},
					RxWindowsAvailable: true,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
					StatusTimePeriodicity:  DurationPtr(0),
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
			DeviceDiffs: []DeviceDiffFunc{
				makeRemoveDownlinksDiff(2),
			},
			ApplicationUplinkAssertion: func(ctx context.Context, dev *ttnpb.EndDevice, ups ...*ttnpb.ApplicationUp) ([]events.Event, bool) {
				_, a := test.MustNewTFromContext(ctx)
				return nil, a.So(ups, should.HaveLength, 1) &&
					a.So(ups[0], should.Resemble, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
						CorrelationIDs:       LastUplink(dev.MACState.RecentUplinks...).CorrelationIDs,
						Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
							DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
								LastFCntDown: dev.Session.LastNFCntDown,
								Downlinks:    dev.Session.QueuedApplicationDownlinks,
								SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
							},
						},
					})
			},
		},

		{
			Name: "Class A/windows open/1.0.3/RX1,RX2 available/no MAC/generic application downlink/application downlink exceeds length limit",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_0_3_REV_A].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_0_3,
						}),
					},
					RxWindowsAvailable: true,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
					StatusTimePeriodicity:  DurationPtr(0),
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
			DeviceDiffs: []DeviceDiffFunc{
				makeRemoveDownlinksDiff(1),
			},
			ApplicationUplinkAssertion: func(ctx context.Context, dev *ttnpb.EndDevice, ups ...*ttnpb.ApplicationUp) ([]events.Event, bool) {
				_, a := test.MustNewTFromContext(ctx)
				return nil, a.So(ups, should.HaveLength, 1) &&
					a.So(ups[0], should.Resemble, &ttnpb.ApplicationUp{
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

		{
			Name: "Class A/windows open/1.1/RX1,RX2 available/MAC answers/MAC requests/generic application downlink/data+MAC/RX1,RX2/EU868/scheduling fail",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0110,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							test.DefaultNwkSEncKey,
							test.DefaultDevAddr,
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0110,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							test.DefaultNwkSEncKey,
							test.DefaultDevAddr,
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
							/*** FCtrl ***/
							0b1_0_0_1_0110,
							/*** FCnt ***/
							0x25, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							test.DefaultNwkSEncKey,
							test.DefaultDevAddr,
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[ttnpb.DATA_RATE_6].Rate,
							DataRateIndex: ttnpb.DATA_RATE_6,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0110,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							test.DefaultNwkSEncKey,
							test.DefaultDevAddr,
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second),
							FCtrl: ttnpb.FCtrl{
								ClassB: true,
							},
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() / 10),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
							/*** FCtrl ***/
							0b1_0_0_0_0110,
							/*** FCnt ***/
							0x42, 0x00,
						}

						/** FOpts **/
						b = append(b, test.Must(crypto.EncryptDownlink(
							test.DefaultNwkSEncKey,
							test.DefaultDevAddr,
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second - time.Nanosecond),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration()),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			ApplicationUplinkAssertion: func(ctx context.Context, dev *ttnpb.EndDevice, ups ...*ttnpb.ApplicationUp) ([]events.Event, bool) {
				_, a := test.MustNewTFromContext(ctx)
				return nil, a.So(ups, should.HaveLength, 1) &&
					a.So(ups[0], should.Resemble, &ttnpb.ApplicationUp{
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

		{
			Name: "Class C/windows closed/1.1/no MAC answers/MAC requests/classBC application downlink/forced gateways/MAC/RXC/EU868/retryable error",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-DefaultEU868RX1Delay.Duration() - time.Second),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
								Gateways: DefaultGatewayAntennaIdentifiers[:],
							},
						},
					},
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
							test.DefaultDevAddr[3], test.DefaultDevAddr[2], test.DefaultDevAddr[1], test.DefaultDevAddr[0],
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
							test.DefaultSNwkSIntKey,
							test.DefaultDevAddr,
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
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-time.Second),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
		},

		{
			Name: "Class C/windows open/1.1/RX1,RX2 available/no MAC/expired application downlinks",
			CreateDevice: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &test.DefaultDevAddr,
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
							DevAddr:       test.DefaultDevAddr,
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							Frequency:     customCh.UplinkFrequency,
							ChannelIndex:  customChIdx,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-time.Second),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
		},

		{
			Name: "join-accept/windows open/RX1,RX2 available/active session/EU868",
			CreateDevice: &ttnpb.EndDevice{
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
						DevAddr: test.DefaultDevAddr,
					},
					RxWindowsAvailable: true,
					RecentUplinks: []*ttnpb.UplinkMessage{
						MakeJoinRequest(JoinRequestConfig{
							DecodePayload: true,
							JoinEUI:       types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
							DevEUI:        types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
							DataRate:      LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B].DataRates[customCh.MinDataRateIndex].Rate,
							DataRateIndex: customCh.MinDataRateIndex,
							Frequency:     DefaultEU868Channels[0].UplinkFrequency,
							RxMetadata:    DefaultRxMetadata[:],
							ReceivedAt:    now.Add(-time.Second),
						}),
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						MakeDataDownlink(DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MAC_V1_1,
						}),
					},
				},
				Session: &ttnpb.Session{
					DevAddr:       test.DefaultDevAddr,
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
				MACState:     MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B),
				SupportsJoin: true,
			},
			DownlinkAssertion: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) (*ttnpb.DownlinkMessage, time.Time, bool) {
				a := assertions.New(test.MustTFromContext(ctx))

				lastUp := LastUplink(dev.PendingMACState.RecentUplinks...)
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
						DevAddr:     created.PendingMACState.QueuedJoinAccept.DevAddr,
						SessionKeys: created.PendingMACState.QueuedJoinAccept.Keys,
					}
					expected.PendingMACState.QueuedJoinAccept = nil
					expected.PendingMACState.RecentDownlinks = AppendRecentDownlink(expected.PendingMACState.RecentDownlinks, down, RecentDownlinkCount)
					return true
				},
			},
			ApplicationUplinkAssertion: func(ctx context.Context, dev *ttnpb.EndDevice, ups ...*ttnpb.ApplicationUp) ([]events.Event, bool) {
				a := assertions.New(test.MustTFromContext(ctx))
				ids := ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: dev.ApplicationIdentifiers,
					DeviceID:               dev.DeviceID,
					DevEUI:                 dev.DevEUI,
					JoinEUI:                dev.JoinEUI,
					DevAddr:                &dev.PendingMACState.QueuedJoinAccept.DevAddr,
				}
				cids := LastUplink(dev.PendingMACState.RecentUplinks...).CorrelationIDs
				recvAt := LastUplink(dev.PendingMACState.RecentUplinks...).ReceivedAt

				ok := a.So(ups, should.Resemble, []*ttnpb.ApplicationUp{
					{
						EndDeviceIdentifiers: ids,
						CorrelationIDs:       cids,
						Up: &ttnpb.ApplicationUp_JoinAccept{
							JoinAccept: &ttnpb.ApplicationJoinAccept{
								AppSKey:              dev.PendingMACState.QueuedJoinAccept.Keys.AppSKey,
								InvalidatedDownlinks: dev.Session.QueuedApplicationDownlinks,
								SessionKeyID:         dev.PendingMACState.QueuedJoinAccept.Keys.SessionKeyID,
								ReceivedAt:           recvAt,
							},
						},
					},
				})
				if !ok {
					return nil, false
				}
				return []events.Event{
					EvtForwardJoinAccept.With(
						events.WithIdentifiers(ids),
						events.WithData(&ttnpb.ApplicationUp{
							EndDeviceIdentifiers: ids,
							CorrelationIDs:       cids,
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									InvalidatedDownlinks: dev.Session.QueuedApplicationDownlinks,
									SessionKeyID:         dev.PendingMACState.QueuedJoinAccept.Keys.SessionKeyID,
									ReceivedAt:           recvAt,
								},
							},
						}),
					).New(ctx),
				}, true
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name: tc.Name,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				errCh := make(chan error, 1)
				_, ctx, env, stop := StartTest(ctx, TestConfig{
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
				if tc.CreateDevice != nil {
					created, ctx = MustCreateDevice(ctx, env.Devices, tc.CreateDevice)
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
				if tc.CreateDevice != nil {
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
					expected.UpdatedAt = now
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
									NetID:      created.PendingMACState.QueuedJoinAccept.NetID,
									DevAddr:    created.PendingMACState.QueuedJoinAccept.DevAddr,
									DLSettings: created.PendingMACState.QueuedJoinAccept.Request.DownlinkSettings,
									RxDelay:    created.PendingMACState.QueuedJoinAccept.Request.RxDelay,
									CFList:     created.PendingMACState.QueuedJoinAccept.Request.CFList,
								},
							}
						}
						down.RawPayload = nil
						down.Payload = msg
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
				if tc.ApplicationUplinkAssertion != nil {
					var evs []events.Event
					a.So(env.AssertNsAsHandleUplink(ctx, appID, func(ctx context.Context, ups ...*ttnpb.ApplicationUp) bool {
						var ok bool
						evs, ok = tc.ApplicationUplinkAssertion(ctx, created, ups...)
						return ok
					}, nil), should.BeTrue)
					if len(evs) > 0 {
						a.So(env.Events, should.ReceiveEventsResembling, evs)
					}
				}
			},
		})
	}
}
