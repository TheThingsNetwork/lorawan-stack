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

package networkserver

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var _ ttnpb.NsGsServer = &MockNsGsServer{}

type MockNsGsServer struct {
	ScheduleDownlinkFunc func(context.Context, *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error)
}

func (m *MockNsGsServer) ScheduleDownlink(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	if m.ScheduleDownlinkFunc == nil {
		return nil, nil
	}
	return m.ScheduleDownlinkFunc(ctx, msg)
}

func makeScheduleDownlinkSequence(fs ...func(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error)) func(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	var i uint64
	return func(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
		defer atomic.AddUint64(&i, 1)
		return fs[i](ctx, msg)
	}
}

func TestProcessDownlinkTask(t *testing.T) {
	gateways := [...]ttnpb.GatewayIdentifiers{
		{
			GatewayID: "test-gtw-0",
		},
		{
			GatewayID: "test-gtw-1",
		},
		{
			GatewayID: "test-gtw-2",
		},
		{
			GatewayID: "test-gtw-3",
		},
		{
			GatewayID: "test-gtw-4",
		},
		{
			GatewayID: "test-gtw-5",
		},
	}

	phy := test.Must(test.Must(band.GetByID(band.EU_863_870)).(band.Band).Version(ttnpb.PHY_V1_1_REV_B)).(band.Band)

	channels := [16]*ttnpb.MACParameters_Channel{
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
	}

	type nsKey struct{}
	type deviceKey struct{}
	type popCallKey struct{}
	type setByIDCallKey struct{}
	type getPeerCallKey struct{}
	type scheduleDownlinkCallKey struct{}

	now := time.Now()

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		PopFunc          func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error
		SetByIDFunc      func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		GetPeerFunc      func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer
		ContextAssertion func(ctx context.Context) bool
		ErrorAssertion   func(t *testing.T, err error) bool
	}{
		{
			Name: "1.1/data downlink/Class A/application downlink",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DeviceClass:        ttnpb.CLASS_A,
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						{
							DeviceChannelIndex: 3,
							RxMetadata: []*ttnpb.RxMetadata{
								{
									GatewayIdentifiers:     gateways[1],
									SNR:                    -9,
									UplinkToken:            []byte("testToken1"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[3],
									SNR:                    -5.3,
									UplinkToken:            []byte("testToken3"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[5],
									SNR:                    12,
									UplinkToken:            []byte("testToken5"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER,
								},
								{
									GatewayIdentifiers:     gateways[0],
									SNR:                    5.2,
									UplinkToken:            []byte("testToken0"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[2],
									SNR:                    6.3,
									UplinkToken:            []byte("testToken2"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[4],
									SNR:                    -7,
									UplinkToken:            []byte("testToken4"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
							},
							Settings: ttnpb.TxSettings{
								DataRateIndex: ttnpb.DATA_RATE_0,
							},
							CorrelationIDs: []string{"testCorrelationUpID1", "testCorrelationUpID2"},
							ReceivedAt:     now.Add(-time.Second),
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationAppDownID1", "testCorrelationAppDownID2"},
							Priority:       ttnpb.TxSchedulePriority_HIGHEST,
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, now)
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"pending_mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret, should.NotBeNil) || !a.So(ret.RecentDownlinks, should.HaveLength, 1) {
					t.FailNow()
				}
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 5)
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID2")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID2")

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}
				fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
				rx1DRIdx := test.Must(phy.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)
				rx1Freq := channels[int(test.Must(phy.Rx1Channel(3)).(uint8))].DownlinkFrequency
				genDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
					phy.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
					phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
				)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to generate Rx2 payload: %s", err)
				}

				expected := CopyEndDevice(pb)
				expected.MACState.PendingApplicationDownlink = nil
				expected.MACState.PendingRequests = nil
				expected.MACState.QueuedJoinAccept = nil
				expected.MACState.QueuedResponses = nil
				expected.MACState.RxWindowsAvailable = false
				expected.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{}
				expected.RecentDownlinks = append(expected.RecentDownlinks, &ttnpb.DownlinkMessage{
					RawPayload:     genDown.Payload,
					CorrelationIDs: ret.RecentDownlinks[0].CorrelationIDs,
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							Rx1DataRateIndex: rx1DRIdx,
							Rx1Delay:         ttnpb.RX_DELAY_3,
							Rx1Frequency:     rx1Freq,
							Rx2DataRateIndex: ttnpb.DATA_RATE_1,
							Rx2Frequency:     42,
							Priority:         ttnpb.TxSchedulePriority_NORMAL,
							DownlinkPaths: []*ttnpb.DownlinkPath{
								{
									Path: &ttnpb.DownlinkPath_UplinkToken{
										UplinkToken: []byte("testToken4"),
									},
								},
							},
						},
					},
				})
				a.So(ret, should.Resemble, expected)

				return ret, nil
			},

			GetPeerFunc: func() func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
				// We need these to ensure equality under '==' for the peers returned for gateway[1] and gateway[2]
				gs124 := &MockNsGsServer{}
				peer124 := test.Must(test.NewGRPCServerPeer(test.Context(), gs124, ttnpb.RegisterNsGsServer)).(cluster.Peer)

				return func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, getPeerCallKey{}, 1)

					pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
					rx1DRIdx := test.Must(phy.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)
					rx1Freq := channels[int(test.Must(phy.Rx1Channel(3)).(uint8))].DownlinkFrequency
					drIdx := ttnpb.DATA_RATE_1
					if rx1DRIdx < drIdx {
						drIdx = rx1DRIdx
					}
					genDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
						phy.DataRates[drIdx].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
					)
					if !a.So(err, should.BeNil) {
						t.Fatalf("Failed to generate Rx2 payload: %s", err)
					}

					switch uid := unique.ID(ctx, ids); uid {
					case unique.ID(ctx, gateways[0]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 0)
						return nil

					case unique.ID(ctx, gateways[1]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 1)
						gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
							t.Fatal("ScheduleDownlink must not be called")
							panic("Unreachable")
						}
						return peer124

					case unique.ID(ctx, gateways[2]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 2)
						gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
							defer func() {
								gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
									defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

									a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
									a.So(msg.CorrelationIDs, should.HaveLength, 5)
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
									a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
										RawPayload:     genDown.Payload,
										CorrelationIDs: msg.CorrelationIDs,
										Settings: &ttnpb.DownlinkMessage_Request{
											Request: &ttnpb.TxRequest{
												Class:            ttnpb.CLASS_A,
												Rx1DataRateIndex: rx1DRIdx,
												Rx1Delay:         ttnpb.RX_DELAY_3,
												Rx1Frequency:     rx1Freq,
												Rx2DataRateIndex: ttnpb.DATA_RATE_1,
												Rx2Frequency:     42,
												Priority:         ttnpb.TxSchedulePriority_NORMAL,
												DownlinkPaths: []*ttnpb.DownlinkPath{
													{
														Path: &ttnpb.DownlinkPath_UplinkToken{
															UplinkToken: []byte("testToken4"),
														},
													},
												},
											},
										},
									})
									return &ttnpb.ScheduleDownlinkResponse{}, nil
								}
							}()

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)
							a.So(msg.CorrelationIDs, should.HaveLength, 5)
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
							a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
								RawPayload:     genDown.Payload,
								CorrelationIDs: msg.CorrelationIDs,
								Settings: &ttnpb.DownlinkMessage_Request{
									Request: &ttnpb.TxRequest{
										Class:            ttnpb.CLASS_A,
										Rx1DataRateIndex: rx1DRIdx,
										Rx1Delay:         ttnpb.RX_DELAY_3,
										Rx1Frequency:     rx1Freq,
										Rx2DataRateIndex: ttnpb.DATA_RATE_1,
										Rx2Frequency:     42,
										Priority:         ttnpb.TxSchedulePriority_NORMAL,
										DownlinkPaths: []*ttnpb.DownlinkPath{
											{
												Path: &ttnpb.DownlinkPath_UplinkToken{
													UplinkToken: []byte("testToken1"),
												},
											},
											{
												Path: &ttnpb.DownlinkPath_UplinkToken{
													UplinkToken: []byte("testToken2"),
												},
											},
										},
									},
								},
							})
							return nil, errors.New("ScheduleDownlink error")
						}
						return peer124

					case unique.ID(ctx, gateways[3]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 3)
						return test.Must(test.NewGRPCServerPeer(ctx, &MockNsGsServer{
							ScheduleDownlinkFunc: func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     genDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Delay:         ttnpb.RX_DELAY_3,
											Rx1Frequency:     rx1Freq,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											Priority:         ttnpb.TxSchedulePriority_NORMAL,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken3"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},
						}, ttnpb.RegisterNsGsServer)).(*test.MockPeer)

					case unique.ID(ctx, gateways[4]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 4)
						return peer124

					default:
						t.Fatalf("Unknown gateway `%s` requested", uid)
						panic("Unreachable")
					}
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 5) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 3)
			},
		},

		{
			Name: "1.0.2/data downlink/Class A/application downlink/FCnt too low",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
						LastNFCntDown: 42,
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DeviceClass:        ttnpb.CLASS_A,
						LoRaWANVersion:     ttnpb.MAC_V1_0_2,
						RxWindowsAvailable: true,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						{
							DeviceChannelIndex: 3,
							RxMetadata: []*ttnpb.RxMetadata{
								{
									GatewayIdentifiers:     gateways[1],
									SNR:                    -9,
									UplinkToken:            []byte("testToken1"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[3],
									SNR:                    -5.3,
									UplinkToken:            []byte("testToken3"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[5],
									SNR:                    12,
									UplinkToken:            []byte("testToken5"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER,
								},
								{
									GatewayIdentifiers:     gateways[0],
									SNR:                    5.2,
									UplinkToken:            []byte("testToken0"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[2],
									SNR:                    6.3,
									UplinkToken:            []byte("testToken2"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[4],
									SNR:                    -7,
									UplinkToken:            []byte("testToken4"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
							},
							Settings: ttnpb.TxSettings{
								DataRateIndex: ttnpb.DATA_RATE_0,
							},
							CorrelationIDs: []string{"testCorrelationUpID1", "testCorrelationUpID2"},
							ReceivedAt:     now.Add(-time.Second),
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							FPort:          1,
							FCnt:           1,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationAppDownID1", "testCorrelationAppDownID2"},
							Priority:       ttnpb.TxSchedulePriority_LOW,
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, now)
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"pending_mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.HaveSameErrorDefinitionAs, errNoDownlink)

				return ret, nil
			},

			GetPeerFunc: func() func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
				// We need these to ensure equality under '==' for the peers returned for gateway[1] and gateway[2]
				gs124 := &MockNsGsServer{}
				peer124 := test.Must(test.NewGRPCServerPeer(test.Context(), gs124, ttnpb.RegisterNsGsServer)).(cluster.Peer)

				return func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, getPeerCallKey{}, 1)

					pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
					rx1DRIdx := test.Must(phy.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)
					rx1Freq := channels[int(test.Must(phy.Rx1Channel(3)).(uint8))].DownlinkFrequency
					drIdx := ttnpb.DATA_RATE_1
					if rx1DRIdx < drIdx {
						drIdx = rx1DRIdx
					}
					genDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
						phy.DataRates[drIdx].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
					)
					if !a.So(err, should.BeNil) {
						t.Fatalf("Failed to generate Rx2 payload: %s", err)
					}

					switch uid := unique.ID(ctx, ids); uid {
					case unique.ID(ctx, gateways[0]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 0)
						return nil

					case unique.ID(ctx, gateways[1]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 1)
						gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
							t.Fatal("ScheduleDownlink must not be called")
							panic("Unreachable")
						}
						return peer124

					case unique.ID(ctx, gateways[2]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 2)
						gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
							defer func() {
								gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
									defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

									a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
									a.So(msg.CorrelationIDs, should.HaveLength, 5)
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
									a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
										RawPayload:     genDown.Payload,
										CorrelationIDs: msg.CorrelationIDs,
										Settings: &ttnpb.DownlinkMessage_Request{
											Request: &ttnpb.TxRequest{
												Class:            ttnpb.CLASS_A,
												Rx1DataRateIndex: rx1DRIdx,
												Rx1Delay:         ttnpb.RX_DELAY_3,
												Rx1Frequency:     rx1Freq,
												Rx2DataRateIndex: ttnpb.DATA_RATE_1,
												Rx2Frequency:     42,
												Priority:         ttnpb.TxSchedulePriority_LOW,
												DownlinkPaths: []*ttnpb.DownlinkPath{
													{
														Path: &ttnpb.DownlinkPath_UplinkToken{
															UplinkToken: []byte("testToken4"),
														},
													},
												},
											},
										},
									})
									return &ttnpb.ScheduleDownlinkResponse{}, nil
								}
							}()

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)
							a.So(msg.CorrelationIDs, should.HaveLength, 5)
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
							a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
								RawPayload:     genDown.Payload,
								CorrelationIDs: msg.CorrelationIDs,
								Settings: &ttnpb.DownlinkMessage_Request{
									Request: &ttnpb.TxRequest{
										Class:            ttnpb.CLASS_A,
										Rx1DataRateIndex: rx1DRIdx,
										Rx1Delay:         ttnpb.RX_DELAY_3,
										Rx1Frequency:     rx1Freq,
										Rx2DataRateIndex: ttnpb.DATA_RATE_1,
										Rx2Frequency:     42,
										Priority:         ttnpb.TxSchedulePriority_LOW,
										DownlinkPaths: []*ttnpb.DownlinkPath{
											{
												Path: &ttnpb.DownlinkPath_UplinkToken{
													UplinkToken: []byte("testToken1"),
												},
											},
											{
												Path: &ttnpb.DownlinkPath_UplinkToken{
													UplinkToken: []byte("testToken2"),
												},
											},
										},
									},
								},
							})
							return nil, errors.New("ScheduleDownlink error")
						}
						return peer124

					case unique.ID(ctx, gateways[3]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 3)
						return test.Must(test.NewGRPCServerPeer(ctx, &MockNsGsServer{
							ScheduleDownlinkFunc: func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     genDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Delay:         ttnpb.RX_DELAY_3,
											Rx1Frequency:     rx1Freq,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											Priority:         ttnpb.TxSchedulePriority_LOW,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken3"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},
						}, ttnpb.RegisterNsGsServer)).(*test.MockPeer)

					case unique.ID(ctx, gateways[4]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 4)
						return peer124

					default:
						t.Fatalf("Unknown gateway `%s` requested", uid)
						panic("Unreachable")
					}
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 0) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)
			},
		},

		{
			Name: "1.1/data downlink/Class C/Rx1/application downlink/uplink-token downlink path",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DeviceClass:        ttnpb.CLASS_C,
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						{
							DeviceChannelIndex: 3,
							RxMetadata: []*ttnpb.RxMetadata{
								{
									GatewayIdentifiers:     gateways[1],
									SNR:                    -9,
									UplinkToken:            []byte("testToken1"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[3],
									SNR:                    -5.3,
									UplinkToken:            []byte("testToken3"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[5],
									SNR:                    12,
									UplinkToken:            []byte("testToken5"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER,
								},
								{
									GatewayIdentifiers:     gateways[0],
									SNR:                    5.2,
									UplinkToken:            []byte("testToken0"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[2],
									SNR:                    6.3,
									UplinkToken:            []byte("testToken2"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[4],
									SNR:                    -7,
									UplinkToken:            []byte("testToken4"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
							},
							Settings: ttnpb.TxSettings{
								DataRateIndex: ttnpb.DATA_RATE_0,
							},
							CorrelationIDs: []string{"testCorrelationUpID1", "testCorrelationUpID2"},
							ReceivedAt:     now.Add(-time.Second),
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationAppDownID1", "testCorrelationAppDownID2"},
							Priority:       ttnpb.TxSchedulePriority_NORMAL,
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, now)
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"pending_mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret, should.NotBeNil) || !a.So(ret.RecentDownlinks, should.HaveLength, 1) {
					t.FailNow()
				}
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 5)
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID2")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID2")

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}
				fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
				rx1DRIdx := test.Must(phy.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)
				rx1Freq := channels[int(test.Must(phy.Rx1Channel(3)).(uint8))].DownlinkFrequency
				genDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
					phy.DataRates[rx1DRIdx].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
					phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
				)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to generate Rx1 payload: %s", err)
				}

				expected := CopyEndDevice(pb)
				expected.MACState.PendingApplicationDownlink = nil
				expected.MACState.PendingRequests = nil
				expected.MACState.QueuedJoinAccept = nil
				expected.MACState.QueuedResponses = nil
				expected.MACState.RxWindowsAvailable = false
				expected.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{}
				expected.RecentDownlinks = append(expected.RecentDownlinks, &ttnpb.DownlinkMessage{
					RawPayload:     genDown.Payload,
					CorrelationIDs: ret.RecentDownlinks[0].CorrelationIDs,
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							Rx1Delay:         ttnpb.RX_DELAY_3,
							Rx1DataRateIndex: rx1DRIdx,
							Rx1Frequency:     rx1Freq,
							Priority:         ttnpb.TxSchedulePriority_NORMAL,
							DownlinkPaths: []*ttnpb.DownlinkPath{
								{
									Path: &ttnpb.DownlinkPath_UplinkToken{
										UplinkToken: []byte("testToken4"),
									},
								},
							},
						},
					},
				})
				a.So(ret, should.Resemble, expected)

				return ret, nil
			},

			GetPeerFunc: func() func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
				// We need these to ensure equality under '==' for the peers returned for gateway[1] and gateway[2]
				gs124 := &MockNsGsServer{}
				peer124 := test.Must(test.NewGRPCServerPeer(test.Context(), gs124, ttnpb.RegisterNsGsServer)).(cluster.Peer)
				once := &sync.Once{}

				return func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, getPeerCallKey{}, 1)

					a.So(role, should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER)

					pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}
					fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
					rx1DRIdx := test.Must(phy.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)
					rx1Freq := channels[int(test.Must(phy.Rx1Channel(3)).(uint8))].DownlinkFrequency
					genDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
						phy.DataRates[rx1DRIdx].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
					)
					if !a.So(err, should.BeNil) {
						t.Fatalf("Failed to generate Rx1 payload: %s", err)
					}

					once.Do(func() {
						gs124.ScheduleDownlinkFunc = makeScheduleDownlinkSequence(
							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     genDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1Delay:         ttnpb.RX_DELAY_3,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Frequency:     rx1Freq,
											Priority:         ttnpb.TxSchedulePriority_NORMAL,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken1"),
													},
												},
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken2"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},

							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     genDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1Delay:         ttnpb.RX_DELAY_3,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Frequency:     rx1Freq,
											Priority:         ttnpb.TxSchedulePriority_NORMAL,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken4"),
													},
												},
											},
										},
									},
								})
								return &ttnpb.ScheduleDownlinkResponse{}, nil
							},
						)
					})

					switch uid := unique.ID(ctx, ids); uid {
					case unique.ID(ctx, gateways[0]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 0)
						return nil

					case unique.ID(ctx, gateways[1]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 1)
						return peer124

					case unique.ID(ctx, gateways[2]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 2)
						return peer124

					case unique.ID(ctx, gateways[3]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 3)
						return test.Must(test.NewGRPCServerPeer(ctx, &MockNsGsServer{
							ScheduleDownlinkFunc: func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     genDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1Delay:         ttnpb.RX_DELAY_3,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Frequency:     rx1Freq,
											Priority:         ttnpb.TxSchedulePriority_NORMAL,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken3"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},
						}, ttnpb.RegisterNsGsServer)).(*test.MockPeer)

					case unique.ID(ctx, gateways[4]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 4)
						return peer124

					default:
						t.Fatalf("Unknown gateway `%s` requested", uid)
						panic("Unreachable")
					}
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 5) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 3)
			},
		},

		{
			Name: "1.1/data downlink/Class C/Rx1,Rx2/application downlink/uplink-token downlink path",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DeviceClass:        ttnpb.CLASS_C,
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						{
							DeviceChannelIndex: 3,
							RxMetadata: []*ttnpb.RxMetadata{
								{
									GatewayIdentifiers:     gateways[1],
									SNR:                    -9,
									UplinkToken:            []byte("testToken1"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[3],
									SNR:                    -5.3,
									UplinkToken:            []byte("testToken3"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[5],
									SNR:                    12,
									UplinkToken:            []byte("testToken5"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER,
								},
								{
									GatewayIdentifiers:     gateways[0],
									SNR:                    5.2,
									UplinkToken:            []byte("testToken0"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[2],
									SNR:                    6.3,
									UplinkToken:            []byte("testToken2"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[4],
									SNR:                    -7,
									UplinkToken:            []byte("testToken4"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
							},
							Settings: ttnpb.TxSettings{
								DataRateIndex: ttnpb.DATA_RATE_0,
							},
							CorrelationIDs: []string{"testCorrelationUpID1", "testCorrelationUpID2"},
							ReceivedAt:     now.Add(-time.Second),
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationAppDownID1", "testCorrelationAppDownID2"},
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, now)
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"pending_mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret, should.NotBeNil) || !a.So(ret.RecentDownlinks, should.HaveLength, 1) {
					t.FailNow()
				}
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 5)
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID2")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID2")

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}
				fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
				rx2GenDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
					phy.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
					phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
				)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to generate Rx2 payload: %s", err)
				}

				expected := CopyEndDevice(pb)
				expected.MACState.PendingApplicationDownlink = nil
				expected.MACState.PendingRequests = nil
				expected.MACState.QueuedJoinAccept = nil
				expected.MACState.QueuedResponses = nil
				expected.MACState.RxWindowsAvailable = false
				expected.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{}
				expected.RecentDownlinks = append(expected.RecentDownlinks, &ttnpb.DownlinkMessage{
					RawPayload:     rx2GenDown.Payload,
					CorrelationIDs: ret.RecentDownlinks[0].CorrelationIDs,
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_C,
							Rx2Frequency:     42,
							Rx2DataRateIndex: ttnpb.DATA_RATE_1,
							DownlinkPaths: []*ttnpb.DownlinkPath{
								{
									Path: &ttnpb.DownlinkPath_UplinkToken{
										UplinkToken: []byte("testToken4"),
									},
								},
							},
						},
					},
				})
				a.So(ret, should.Resemble, expected)

				return ret, nil
			},

			GetPeerFunc: func() func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
				// We need these to ensure equality under '==' for the peers returned for gateway[1] and gateway[2]
				gs124 := &MockNsGsServer{}
				peer124 := test.Must(test.NewGRPCServerPeer(test.Context(), gs124, ttnpb.RegisterNsGsServer)).(cluster.Peer)
				gs3 := &MockNsGsServer{}
				peer3 := test.Must(test.NewGRPCServerPeer(test.Context(), gs3, ttnpb.RegisterNsGsServer)).(cluster.Peer)
				once := &sync.Once{}
				var rx2 bool

				return func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, getPeerCallKey{}, 1)

					a.So(role, should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER)

					pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}
					fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
					rx1DRIdx := test.Must(phy.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)
					rx1Freq := channels[int(test.Must(phy.Rx1Channel(3)).(uint8))].DownlinkFrequency
					rx1GenDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
						phy.DataRates[rx1DRIdx].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
					)
					if !a.So(err, should.BeNil) {
						t.Fatalf("Failed to generate Rx1 payload: %s", err)
					}
					rx2GenDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
						phy.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
					)
					if !a.So(err, should.BeNil) {
						t.Fatalf("Failed to generate Rx2 payload: %s", err)
					}

					once.Do(func() {
						gs3.ScheduleDownlinkFunc = makeScheduleDownlinkSequence(
							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx1GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1Delay:         ttnpb.RX_DELAY_3,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Frequency:     rx1Freq,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken3"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},

							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 4)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx2GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_C,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken3"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},
						)

						gs124.ScheduleDownlinkFunc = makeScheduleDownlinkSequence(
							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx1GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1Delay:         ttnpb.RX_DELAY_3,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Frequency:     rx1Freq,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken1"),
													},
												},
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken2"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},

							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx1GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1Delay:         ttnpb.RX_DELAY_3,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Frequency:     rx1Freq,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken4"),
													},
												},
											},
										},
									},
								})
								rx2 = true
								return nil, errors.New("ScheduleDownlink error")
							},

							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 3)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx2GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_C,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken1"),
													},
												},
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken2"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},

							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 5)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx2GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_C,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken4"),
													},
												},
											},
										},
									},
								})
								return &ttnpb.ScheduleDownlinkResponse{}, nil
							},
						)
					})

					switch uid := unique.ID(ctx, ids); uid {
					case unique.ID(ctx, gateways[0]):
						if !rx2 {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 0)
						} else {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 5)
						}
						return nil

					case unique.ID(ctx, gateways[1]):
						if !rx2 {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 1)
						} else {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 6)
						}
						return peer124

					case unique.ID(ctx, gateways[2]):
						if !rx2 {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 2)
						} else {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 7)
						}
						return peer124

					case unique.ID(ctx, gateways[3]):
						if !rx2 {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 3)
						} else {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 8)
						}
						return peer3

					case unique.ID(ctx, gateways[4]):
						if !rx2 {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 4)
						} else {
							a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 9)
						}
						return peer124

					default:
						t.Fatalf("Unknown gateway `%s` requested", uid)
						panic("Unreachable")
					}
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 10) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 6)
			},
		},

		{
			Name: "1.1/multicast data downlink/Class C/Rx2/application downlink/absolute time",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:         ttnpb.RX_DELAY_3,
							Rx2DataRateIndex: ttnpb.DATA_RATE_1,
							Rx2Frequency:     42,
							Channels:         channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:         ttnpb.RX_DELAY_3,
							Rx2DataRateIndex: ttnpb.DATA_RATE_1,
							Rx2Frequency:     42,
							Channels:         channels[:],
						},
						DeviceClass:        ttnpb.CLASS_C,
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: false,
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationAppDownID1", "testCorrelationAppDownID2"},
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								Gateways: []*ttnpb.GatewayAntennaIdentifiers{
									{
										GatewayIdentifiers: gateways[0],
										AntennaIndex:       2,
									},
								},
								AbsoluteTime: timePtr(now.Add(20 * time.Second)),
							},
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, now)
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"pending_mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret, should.NotBeNil) || !a.So(ret.RecentDownlinks, should.HaveLength, 1) {
					t.FailNow()
				}
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 3)
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID2")

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}
				fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
				rx2GenDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
					phy.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
					phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
				)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to generate Rx2 payload: %s", err)
				}

				expected := CopyEndDevice(pb)
				expected.MACState.PendingApplicationDownlink = nil
				expected.MACState.PendingRequests = nil
				expected.MACState.QueuedJoinAccept = nil
				expected.MACState.QueuedResponses = nil
				expected.MACState.RxWindowsAvailable = false
				expected.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{}
				expected.RecentDownlinks = append(expected.RecentDownlinks, &ttnpb.DownlinkMessage{
					RawPayload:     rx2GenDown.Payload,
					CorrelationIDs: ret.RecentDownlinks[0].CorrelationIDs,
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_C,
							Rx2Frequency:     42,
							Rx2DataRateIndex: ttnpb.DATA_RATE_1,
							DownlinkPaths: []*ttnpb.DownlinkPath{
								{
									Path: &ttnpb.DownlinkPath_Fixed{
										Fixed: &ttnpb.GatewayAntennaIdentifiers{
											GatewayIdentifiers: gateways[0],
											AntennaIndex:       2,
										},
									},
								},
							},
							AbsoluteTime: timePtr(now.Add(20 * time.Second)),
						},
					},
				})
				a.So(ret, should.Resemble, expected)

				return ret, nil
			},

			GetPeerFunc: func() func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
				// We need these to ensure equality under '==' for the peers returned for gateway[1] and gateway[2]
				gs124 := &MockNsGsServer{}
				peer124 := test.Must(test.NewGRPCServerPeer(test.Context(), gs124, ttnpb.RegisterNsGsServer)).(cluster.Peer)
				once := &sync.Once{}

				return func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, getPeerCallKey{}, 1)

					a.So(role, should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER)

					pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}
					fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
					rx2GenDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
						phy.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
					)
					if !a.So(err, should.BeNil) {
						t.Fatalf("Failed to generate Rx2 payload: %s", err)
					}

					once.Do(func() {
						gs124.ScheduleDownlinkFunc = makeScheduleDownlinkSequence(
							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)
								a.So(msg.CorrelationIDs, should.HaveLength, 3)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx2GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_C,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIdentifiers: gateways[0],
															AntennaIndex:       2,
														},
													},
												},
											},
											AbsoluteTime: timePtr(now.UTC().Add(20 * time.Second)),
										},
									},
								})
								return &ttnpb.ScheduleDownlinkResponse{}, nil
							},
						)
					})

					switch uid := unique.ID(ctx, ids); uid {
					case unique.ID(ctx, gateways[0]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 0)
						return peer124

					default:
						t.Fatalf("Unknown gateway `%s` requested", uid)
						panic("Unreachable")
					}
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)
			},
		},

		{
			Name: "1.1/data downlink/Class C/Rx2/application downlink/forced downlink path",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          ttnpb.RX_DELAY_3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DeviceClass:        ttnpb.CLASS_C,
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						{
							DeviceChannelIndex: 3,
							RxMetadata: []*ttnpb.RxMetadata{
								{
									GatewayIdentifiers:     gateways[1],
									SNR:                    -9,
									UplinkToken:            []byte("testToken1"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[3],
									SNR:                    -5.3,
									UplinkToken:            []byte("testToken3"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[5],
									SNR:                    12,
									UplinkToken:            []byte("testToken5"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER,
								},
								{
									GatewayIdentifiers:     gateways[0],
									SNR:                    5.2,
									UplinkToken:            []byte("testToken0"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[2],
									SNR:                    6.3,
									UplinkToken:            []byte("testToken2"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[4],
									SNR:                    -7,
									UplinkToken:            []byte("testToken4"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
							},
							Settings: ttnpb.TxSettings{
								DataRateIndex: ttnpb.DATA_RATE_0,
							},
							CorrelationIDs: []string{"testCorrelationUpID1", "testCorrelationUpID2"},
							ReceivedAt:     now.Add(-time.Second),
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationAppDownID1", "testCorrelationAppDownID2"},
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								Gateways: []*ttnpb.GatewayAntennaIdentifiers{
									{
										GatewayIdentifiers: gateways[0],
										AntennaIndex:       2,
									},
									{
										GatewayIdentifiers: gateways[1],
										AntennaIndex:       3,
									},
									{
										GatewayIdentifiers: gateways[2],
										AntennaIndex:       1,
									},
									{
										GatewayIdentifiers: gateways[3],
										AntennaIndex:       1,
									},
									{
										GatewayIdentifiers: gateways[4],
										AntennaIndex:       2,
									},
								},
							},
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, now)
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"pending_mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret, should.NotBeNil) || !a.So(ret.RecentDownlinks, should.HaveLength, 1) {
					t.FailNow()
				}
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 5)
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID2")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationAppDownID2")

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}
				fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
				rx2GenDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
					phy.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
					phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
				)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to generate Rx2 payload: %s", err)
				}

				expected := CopyEndDevice(pb)
				expected.MACState.PendingApplicationDownlink = nil
				expected.MACState.PendingRequests = nil
				expected.MACState.QueuedJoinAccept = nil
				expected.MACState.QueuedResponses = nil
				expected.MACState.RxWindowsAvailable = false
				expected.QueuedApplicationDownlinks = []*ttnpb.ApplicationDownlink{}
				expected.RecentDownlinks = append(expected.RecentDownlinks, &ttnpb.DownlinkMessage{
					RawPayload:     rx2GenDown.Payload,
					CorrelationIDs: ret.RecentDownlinks[0].CorrelationIDs,
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_C,
							Rx2Frequency:     42,
							Rx2DataRateIndex: ttnpb.DATA_RATE_1,
							DownlinkPaths: []*ttnpb.DownlinkPath{
								{
									Path: &ttnpb.DownlinkPath_Fixed{
										Fixed: &ttnpb.GatewayAntennaIdentifiers{
											GatewayIdentifiers: gateways[4],
											AntennaIndex:       2,
										},
									},
								},
							},
						},
					},
				})
				a.So(ret, should.Resemble, expected)

				return ret, nil
			},

			GetPeerFunc: func() func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
				// We need these to ensure equality under '==' for the peers returned for gateway[1] and gateway[2]
				gs124 := &MockNsGsServer{}
				peer124 := test.Must(test.NewGRPCServerPeer(test.Context(), gs124, ttnpb.RegisterNsGsServer)).(cluster.Peer)
				gs3 := &MockNsGsServer{}
				peer3 := test.Must(test.NewGRPCServerPeer(test.Context(), gs3, ttnpb.RegisterNsGsServer)).(cluster.Peer)
				once := &sync.Once{}

				return func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, getPeerCallKey{}, 1)

					a.So(role, should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER)

					pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}
					fp := test.Must(ns.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
					rx2GenDown, err := ns.generateDownlink(ctx, CopyEndDevice(pb), phy,
						phy.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						phy.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks()),
					)
					if !a.So(err, should.BeNil) {
						t.Fatalf("Failed to generate Rx2 payload: %s", err)
					}

					once.Do(func() {
						gs3.ScheduleDownlinkFunc = makeScheduleDownlinkSequence(
							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx2GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_C,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIdentifiers: gateways[3],
															AntennaIndex:       1,
														},
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},

							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 4)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx2GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_C,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIdentifiers: gateways[3],
															AntennaIndex:       1,
														},
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},
						)

						gs124.ScheduleDownlinkFunc = makeScheduleDownlinkSequence(
							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx2GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_C,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIdentifiers: gateways[1],
															AntennaIndex:       3,
														},
													},
												},
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIdentifiers: gateways[2],
															AntennaIndex:       1,
														},
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},

							func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
								a.So(msg.CorrelationIDs, should.HaveLength, 5)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationAppDownID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     rx2GenDown.Payload,
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_C,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_Fixed{
														Fixed: &ttnpb.GatewayAntennaIdentifiers{
															GatewayIdentifiers: gateways[4],
															AntennaIndex:       2,
														},
													},
												},
											},
										},
									},
								})
								return &ttnpb.ScheduleDownlinkResponse{}, nil
							},
						)
					})

					switch uid := unique.ID(ctx, ids); uid {
					case unique.ID(ctx, gateways[0]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 0)
						return nil

					case unique.ID(ctx, gateways[1]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 1)
						return peer124

					case unique.ID(ctx, gateways[2]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 2)
						return peer124

					case unique.ID(ctx, gateways[3]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 3)
						return peer3

					case unique.ID(ctx, gateways[4]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 4)
						return peer124

					default:
						t.Fatalf("Unknown gateway `%s` requested", uid)
						panic("Unreachable")
					}
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 5) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 3)
			},
		},

		{
			Name: "1.1/join accept",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					PendingMACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DeviceClass:        ttnpb.CLASS_A,
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
						QueuedJoinAccept: &ttnpb.MACState_JoinAccept{
							Payload: []byte("testJoinAccept"),
						},
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						ttnpb.NewPopulatedUplinkMessage(test.Randy, false),
						{
							DeviceChannelIndex: 3,
							RxMetadata: []*ttnpb.RxMetadata{
								{
									GatewayIdentifiers:     gateways[1],
									SNR:                    -9,
									UplinkToken:            []byte("testToken1"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[3],
									SNR:                    -5.3,
									UplinkToken:            []byte("testToken3"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[5],
									SNR:                    12,
									UplinkToken:            []byte("testToken5"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER,
								},
								{
									GatewayIdentifiers:     gateways[0],
									SNR:                    5.2,
									UplinkToken:            []byte("testToken0"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
								},
								{
									GatewayIdentifiers:     gateways[2],
									SNR:                    6.3,
									UplinkToken:            []byte("testToken2"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
								{
									GatewayIdentifiers:     gateways[4],
									SNR:                    -7,
									UplinkToken:            []byte("testToken4"),
									DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
								},
							},
							Settings: ttnpb.TxSettings{
								DataRateIndex: ttnpb.DATA_RATE_0,
							},
							CorrelationIDs: []string{"testCorrelationUpID1", "testCorrelationUpID2"},
							ReceivedAt:     now.Add(-time.Second),
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_JOIN_REQUEST,
								},
								Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{}},
							},
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationAppDownID1", "testCorrelationAppDownID2"},
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, now)
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"last_dev_status_received_at",
					"lorawan_phy_version",
					"mac_settings",
					"mac_state",
					"pending_mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"pending_mac_state.pending_join_request",
					"pending_mac_state.queued_join_accept",
					"pending_mac_state.rx_windows_available",
					"pending_session.dev_addr",
					"pending_session.session_keys",
					"recent_downlinks",
				})
				if !a.So(ret, should.NotBeNil) || !a.So(ret.RecentDownlinks, should.HaveLength, 1) {
					t.FailNow()
				}
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 3)
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID1")
				a.So(ret.RecentDownlinks[0].CorrelationIDs, should.Contain, "testCorrelationUpID2")

				rx1DRIdx := test.Must(phy.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)
				rx1Freq := channels[int(test.Must(phy.Rx1Channel(3)).(uint8))].DownlinkFrequency

				expected := CopyEndDevice(pb)
				expected.PendingMACState.PendingJoinRequest = &expected.PendingMACState.QueuedJoinAccept.Request
				expected.PendingSession = &ttnpb.Session{
					DevAddr:     expected.PendingMACState.QueuedJoinAccept.Request.DevAddr,
					SessionKeys: expected.PendingMACState.QueuedJoinAccept.Keys,
				}
				expected.PendingMACState.PendingApplicationDownlink = nil
				expected.PendingMACState.PendingRequests = nil
				expected.PendingMACState.QueuedJoinAccept = nil
				expected.PendingMACState.QueuedResponses = nil
				expected.PendingMACState.RxWindowsAvailable = false
				expected.RecentDownlinks = append(expected.RecentDownlinks, &ttnpb.DownlinkMessage{
					RawPayload:     []byte("testJoinAccept"),
					CorrelationIDs: ret.RecentDownlinks[0].CorrelationIDs,
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							Rx1DataRateIndex: rx1DRIdx,
							Rx1Delay:         ttnpb.RxDelay(phy.JoinAcceptDelay1 / time.Second),
							Rx1Frequency:     rx1Freq,
							Rx2DataRateIndex: ttnpb.DATA_RATE_1,
							Rx2Frequency:     42,
							Priority:         ttnpb.TxSchedulePriority_HIGH,
							DownlinkPaths: []*ttnpb.DownlinkPath{
								{
									Path: &ttnpb.DownlinkPath_UplinkToken{
										UplinkToken: []byte("testToken4"),
									},
								},
							},
						},
					},
				})
				a.So(ret, should.Resemble, expected)

				return ret, nil
			},

			GetPeerFunc: func() func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
				// We need these to ensure equality under '==' for the peers returned for gateway[1] and gateway[2]
				gs124 := &MockNsGsServer{}
				peer124 := test.Must(test.NewGRPCServerPeer(test.Context(), gs124, ttnpb.RegisterNsGsServer)).(cluster.Peer)

				return func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, getPeerCallKey{}, 1)

					rx1DRIdx := test.Must(phy.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)
					rx1Freq := channels[int(test.Must(phy.Rx1Channel(3)).(uint8))].DownlinkFrequency

					switch uid := unique.ID(ctx, ids); uid {
					case unique.ID(ctx, gateways[0]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 0)
						return nil

					case unique.ID(ctx, gateways[1]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 1)
						gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
							t.Fatal("ScheduleDownlink must not be called")
							panic("Unreachable")
						}
						return peer124

					case unique.ID(ctx, gateways[2]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 2)
						gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
							defer func() {
								gs124.ScheduleDownlinkFunc = func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
									defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

									a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
									a.So(msg.CorrelationIDs, should.HaveLength, 3)
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
									a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
									a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
										RawPayload:     []byte("testJoinAccept"),
										CorrelationIDs: msg.CorrelationIDs,
										Settings: &ttnpb.DownlinkMessage_Request{
											Request: &ttnpb.TxRequest{
												Class:            ttnpb.CLASS_A,
												Rx1DataRateIndex: rx1DRIdx,
												Rx1Delay:         ttnpb.RxDelay(phy.JoinAcceptDelay1 / time.Second),
												Rx1Frequency:     rx1Freq,
												Rx2DataRateIndex: ttnpb.DATA_RATE_1,
												Rx2Frequency:     42,
												Priority:         ttnpb.TxSchedulePriority_HIGH,
												DownlinkPaths: []*ttnpb.DownlinkPath{
													{
														Path: &ttnpb.DownlinkPath_UplinkToken{
															UplinkToken: []byte("testToken4"),
														},
													},
												},
											},
										},
									})
									return &ttnpb.ScheduleDownlinkResponse{}, nil
								}
							}()

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)
							a.So(msg.CorrelationIDs, should.HaveLength, 3)
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
							a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
							a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
								RawPayload:     []byte("testJoinAccept"),
								CorrelationIDs: msg.CorrelationIDs,
								Settings: &ttnpb.DownlinkMessage_Request{
									Request: &ttnpb.TxRequest{
										Class:            ttnpb.CLASS_A,
										Rx1DataRateIndex: rx1DRIdx,
										Rx1Delay:         ttnpb.RxDelay(phy.JoinAcceptDelay1 / time.Second),
										Rx1Frequency:     rx1Freq,
										Rx2DataRateIndex: ttnpb.DATA_RATE_1,
										Rx2Frequency:     42,
										Priority:         ttnpb.TxSchedulePriority_HIGH,
										DownlinkPaths: []*ttnpb.DownlinkPath{
											{
												Path: &ttnpb.DownlinkPath_UplinkToken{
													UplinkToken: []byte("testToken1"),
												},
											},
											{
												Path: &ttnpb.DownlinkPath_UplinkToken{
													UplinkToken: []byte("testToken2"),
												},
											},
										},
									},
								},
							})
							return nil, errors.New("ScheduleDownlink error")
						}
						return peer124

					case unique.ID(ctx, gateways[3]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 3)
						return test.Must(test.NewGRPCServerPeer(ctx, &MockNsGsServer{
							ScheduleDownlinkFunc: func(_ context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)
								a.So(msg.CorrelationIDs, should.HaveLength, 3)
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID1")
								a.So(msg.CorrelationIDs, should.Contain, "testCorrelationUpID2")
								a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
									RawPayload:     []byte("testJoinAccept"),
									CorrelationIDs: msg.CorrelationIDs,
									Settings: &ttnpb.DownlinkMessage_Request{
										Request: &ttnpb.TxRequest{
											Class:            ttnpb.CLASS_A,
											Rx1DataRateIndex: rx1DRIdx,
											Rx1Delay:         ttnpb.RxDelay(phy.JoinAcceptDelay1 / time.Second),
											Rx1Frequency:     rx1Freq,
											Rx2DataRateIndex: ttnpb.DATA_RATE_1,
											Priority:         ttnpb.TxSchedulePriority_HIGH,
											Rx2Frequency:     42,
											DownlinkPaths: []*ttnpb.DownlinkPath{
												{
													Path: &ttnpb.DownlinkPath_UplinkToken{
														UplinkToken: []byte("testToken3"),
													},
												},
											},
										},
									},
								})
								return nil, errors.New("ScheduleDownlink error")
							},
						}, ttnpb.RegisterNsGsServer)).(*test.MockPeer)

					case unique.ID(ctx, gateways[4]):
						a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 4)
						return peer124

					default:
						t.Fatalf("Unknown gateway `%s` requested", uid)
						panic("Unreachable")
					}
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, getPeerCallKey{}), should.Equal, 5) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 3)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t),
					&component.Config{},
					component.WithClusterNew(func(context.Context, *config.ServiceBase, ...rpcserver.Registerer) (cluster.Cluster, error) {
						return &test.MockCluster{
							GetPeerFunc: tc.GetPeerFunc,
						}, nil
					}),
				),
				&Config{
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks: &MockDownlinkTaskQueue{
						PopFunc: tc.PopFunc,
					},
					DownlinkPriorities: DownlinkPriorityConfig{
						JoinAccept:             "high",
						MACCommands:            "above_normal",
						MaxApplicationDownlink: "normal",
					},
					Devices: &MockDeviceRegistry{
						SetByIDFunc: tc.SetByIDFunc,
					},
					DefaultMACSettings: MACSettingConfig{
						StatusTimePeriodicity:  DurationPtr(0),
						StatusCountPeriodicity: func(v uint32) *uint32 { return &v }(0),
					},
				},
			)).(*NetworkServer)
			ns.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)
			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return context.WithValue(ctx, nsKey{}, ns)
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, popCallKey{})
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, setByIDCallKey{})
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, scheduleDownlinkCallKey{})
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, getPeerCallKey{})
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, now.Add(Timeout))
				_ = cancel
				return ctx
			})
			test.Must(nil, ns.Start())
			defer ns.Close()

			ctx := test.ContextWithT(ns.FillContext(ns.Context()), t)

			err := ns.processDownlinkTask(ctx)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
			}
			a.So(tc.ContextAssertion(ctx), should.BeTrue)
		})
	}
}

func TestGenerateDownlink(t *testing.T) {
	phy := test.Must(test.Must(band.GetByID(band.EU_863_870)).(band.Band).Version(ttnpb.PHY_V1_1_REV_B)).(band.Band)

	encodeMessage := func(msg *ttnpb.Message, ver ttnpb.MACVersion, confFCnt uint32) []byte {
		msg = deepcopy.Copy(msg).(*ttnpb.Message)
		mac := msg.GetMACPayload()

		if len(mac.FRMPayload) > 0 && mac.FPort == 0 {
			var key types.AES128Key
			switch ver {
			case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
				key = FNwkSIntKey
			case ttnpb.MAC_V1_1:
				key = NwkSEncKey
			default:
				panic(fmt.Errorf("unknown version %s", ver))
			}

			var err error
			mac.FRMPayload, err = crypto.EncryptDownlink(key, mac.DevAddr, mac.FCnt, mac.FRMPayload)
			if err != nil {
				t.Fatal("Failed to encrypt downlink FRMPayload")
			}
		}

		b, err := lorawan.MarshalMessage(*msg)
		if err != nil {
			t.Fatal("Failed to marshal downlink")
		}

		var key types.AES128Key
		switch ver {
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			key = FNwkSIntKey
		case ttnpb.MAC_V1_1:
			key = SNwkSIntKey
		default:
			panic(fmt.Errorf("unknown version %s", ver))
		}

		mic, err := crypto.ComputeDownlinkMIC(key, mac.DevAddr, confFCnt, mac.FCnt, b)
		if err != nil {
			t.Fatal("Failed to compute MIC")
		}
		return append(b, mic[:]...)
	}

	encodeMAC := func(phy band.Band, cmds ...*ttnpb.MACCommand) (b []byte) {
		for _, cmd := range cmds {
			b = test.Must(lorawan.DefaultMACCommands.AppendDownlink(phy, b, *cmd)).([]byte)
		}
		return
	}

	for _, tc := range []struct {
		Name                         string
		Device                       *ttnpb.EndDevice
		Context                      context.Context
		Bytes                        []byte
		ApplicationDownlinkAssertion func(t *testing.T, down *ttnpb.ApplicationDownlink) bool
		DeviceAssertion              func(*testing.T, *ttnpb.EndDevice) bool
		Error                        error
	}{
		{
			Name:    "1.1/no app downlink/no MAC/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session:           ttnpb.NewPopulatedSession(test.Randy, false),
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Error: errNoDownlink,
		},
		{
			Name:    "1.1/no app downlink/status after 1 downlink/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:      ttnpb.MAC_V1_1,
					LastDevStatusFCntUp: 2,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 4,
				},
				LoRaWANPHYVersion:       ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:         band.EU_863_870,
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Error: errNoDownlink,
		},
		{
			Name:    "1.1/no app downlink/status after an hour/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity: DurationPtr(24 * time.Hour),
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				LoRaWANPHYVersion:       ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:         band.EU_863_870,
				LastDevStatusReceivedAt: TimePtr(time.Now()),
				Session:                 ttnpb.NewPopulatedSession(test.Randy, false),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Error: errNoDownlink,
		},
		{
			Name:    "1.1/no app downlink/no MAC/ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:     ttnpb.MAC_V1_1,
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       DevAddr,
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_CONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{
							MACPayload: &ttnpb.MACPayload{
								FHDR: ttnpb.FHDR{
									FCnt: 24,
								},
							},
						},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
							},
							FCnt: 42,
						},
					},
				},
			}, ttnpb.MAC_V1_1, 24),
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       DevAddr,
						LastNFCntDown: 42,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCnt: 24,
									},
								},
							},
						},
					}},
				})
			},
		},
		{
			Name:    "1.1/unconfirmed app downlink/no MAC/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:     ttnpb.MAC_V1_1,
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					{
						Confirmed:  false,
						FCnt:       42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
							},
							FCnt: 42,
						},
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			}, ttnpb.MAC_V1_1, 0),
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  false,
					FCnt:       42,
					FPort:      1,
					FRMPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
				})
			},
		},
		{
			Name:    "1.1/unconfirmed app downlink/no MAC/ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:     ttnpb.MAC_V1_1,
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					{
						Confirmed:  false,
						FCnt:       42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_CONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{
							MACPayload: &ttnpb.MACPayload{
								FHDR: ttnpb.FHDR{
									FCnt: 24,
								},
							},
						},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
							},
							FCnt: 42,
						},
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			}, ttnpb.MAC_V1_1, 24),
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  false,
					FCnt:       42,
					FPort:      1,
					FRMPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:     ttnpb.MAC_V1_1,
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCnt: 24,
									},
								},
							},
						},
					}},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
				})
			},
		},
		{
			Name:    "1.1/confirmed app downlink/no MAC/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					{
						Confirmed:  true,
						FCnt:       42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
							},
							FCnt: 42,
						},
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			}, ttnpb.MAC_V1_1, 0),
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  true,
					FCnt:       42,
					FPort:      1,
					FRMPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MACState, should.NotBeNil) || !a.So(dev.MACState.LastConfirmedDownlinkAt, should.NotBeNil) {
					t.FailNow()
				}
				now := time.Now()
				a.So([]time.Time{now.Add(-time.Minute), *dev.MACState.LastConfirmedDownlinkAt, now}, should.BeChronological)
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:          ttnpb.MAC_V1_1,
						LastConfirmedDownlinkAt: dev.MACState.LastConfirmedDownlinkAt,
						PendingApplicationDownlink: &ttnpb.ApplicationDownlink{
							Confirmed:  true,
							FCnt:       42,
							FPort:      1,
							FRMPayload: []byte("test"),
						},
					},
					Session: &ttnpb.Session{
						DevAddr:          DevAddr,
						LastConfFCntDown: 42,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
				})
			},
		},
		{
			Name:    "1.1/confirmed app downlink/no MAC/ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:     ttnpb.MAC_V1_1,
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					{
						Confirmed:  true,
						FCnt:       42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_CONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{
							MACPayload: &ttnpb.MACPayload{
								FHDR: ttnpb.FHDR{
									FCnt: 24,
								},
							},
						},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
							},
							FCnt: 42,
						},
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			}, ttnpb.MAC_V1_1, 24),
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  true,
					FCnt:       42,
					FPort:      1,
					FRMPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MACState, should.NotBeNil) || !a.So(dev.MACState.LastConfirmedDownlinkAt, should.NotBeNil) {
					t.FailNow()
				}
				now := time.Now()
				a.So([]time.Time{now.Add(-time.Minute), *dev.MACState.LastConfirmedDownlinkAt, now}, should.BeChronological)
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:          ttnpb.MAC_V1_1,
						RxWindowsAvailable:      true,
						LastConfirmedDownlinkAt: dev.MACState.LastConfirmedDownlinkAt,
						PendingApplicationDownlink: &ttnpb.ApplicationDownlink{
							Confirmed:  true,
							FCnt:       42,
							FPort:      1,
							FRMPayload: []byte("test"),
						},
					},
					Session: &ttnpb.Session{
						DevAddr:          DevAddr,
						LastConfFCntDown: 42,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion:          ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:            band.EU_863_870,
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCnt: 24,
									},
								},
							},
						},
					}},
				})
			},
		},
		{
			Name:    "1.1/no app downlink/status(count)/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:      ttnpb.MAC_V1_1,
					LastDevStatusFCntUp: 4,
				},
				Session: &ttnpb.Session{
					DevAddr:       DevAddr,
					LastFCntUp:    99,
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
							},
							FCnt: 42,
						},
						FPort: 0,
						FRMPayload: encodeMAC(
							phy,
							ttnpb.CID_DEV_STATUS.MACCommand(),
						),
					},
				},
			}, ttnpb.MAC_V1_1, 0),
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MACState, should.NotBeNil) || !a.So(dev.MACState.LastConfirmedDownlinkAt, should.NotBeNil) {
					t.FailNow()
				}
				now := time.Now()
				a.So([]time.Time{now.Add(-time.Minute), *dev.MACState.LastConfirmedDownlinkAt, now}, should.BeChronological)
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					MACSettings: &ttnpb.MACSettings{
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:          ttnpb.MAC_V1_1,
						LastConfirmedDownlinkAt: dev.MACState.LastConfirmedDownlinkAt,
						LastDevStatusFCntUp:     4,
						PendingRequests: []*ttnpb.MACCommand{
							ttnpb.CID_DEV_STATUS.MACCommand(),
						},
					},
					Session: &ttnpb.Session{
						DevAddr:       DevAddr,
						LastFCntUp:    99,
						LastNFCntDown: 42,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
				})
			},
		},
		{
			Name:    "1.1/no app downlink/status(time/zero time)/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity: DurationPtr(time.Nanosecond),
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					DevAddr:       DevAddr,
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
							},
							FCnt: 42,
						},
						FPort: 0,
						FRMPayload: encodeMAC(
							phy,
							ttnpb.CID_DEV_STATUS.MACCommand(),
						),
					},
				},
			}, ttnpb.MAC_V1_1, 0),
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MACState, should.NotBeNil) || !a.So(dev.MACState.LastConfirmedDownlinkAt, should.NotBeNil) {
					t.FailNow()
				}
				now := time.Now()
				a.So([]time.Time{now.Add(-time.Minute), *dev.MACState.LastConfirmedDownlinkAt, now}, should.BeChronological)
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					MACSettings: &ttnpb.MACSettings{
						StatusTimePeriodicity: DurationPtr(time.Nanosecond),
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:          ttnpb.MAC_V1_1,
						LastConfirmedDownlinkAt: dev.MACState.LastConfirmedDownlinkAt,
						PendingRequests: []*ttnpb.MACCommand{
							ttnpb.CID_DEV_STATUS.MACCommand(),
						},
					},
					Session: &ttnpb.Session{
						DevAddr:       DevAddr,
						LastNFCntDown: 42,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
				})
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t),
					&component.Config{},
				),
				&Config{
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks:       &MockDownlinkTaskQueue{},
					Devices:             &MockDeviceRegistry{},
					DefaultMACSettings: MACSettingConfig{
						StatusTimePeriodicity:  DurationPtr(0),
						StatusCountPeriodicity: func(v uint32) *uint32 { return &v }(0),
					},
				},
			)).(*NetworkServer)
			ns.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)
			test.Must(nil, ns.Start())
			defer ns.Close()

			dev := CopyEndDevice(tc.Device)
			_, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			genDown, err := ns.generateDownlink(tc.Context, dev, phy, math.MaxUint16, math.MaxUint16)
			if tc.Error != nil {
				a.So(err, should.EqualErrorOrDefinition, tc.Error)
				a.So(genDown, should.BeNil)
				return
			}

			if !a.So(err, should.BeNil) || !a.So(genDown, should.NotBeNil) {
				t.FailNow()
			}

			a.So(genDown.Payload, should.Resemble, tc.Bytes)
			if tc.ApplicationDownlinkAssertion != nil {
				a.So(tc.ApplicationDownlinkAssertion(t, genDown.ApplicationDownlink), should.BeTrue)
			} else {
				a.So(genDown.ApplicationDownlink, should.BeNil)
			}

			if tc.DeviceAssertion != nil {
				a.So(tc.DeviceAssertion(t, dev), should.BeTrue)
			} else {
				a.So(dev, should.Resemble, tc.Device)
			}
		})
	}
}
