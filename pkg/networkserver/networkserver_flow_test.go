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
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func makeOTAAFlowTest(createDevice *ttnpb.SetEndDeviceRequest, f func(context.Context, TestEnvironment, *ttnpb.EndDevice, ttnpb.AsNs_LinkApplicationClient)) func(context.Context, TestEnvironment) {
	return func(ctx context.Context, env TestEnvironment) {
		t, a := test.MustNewTFromContext(ctx)

		start := time.Now()

		linkCtx, closeLink := context.WithCancel(ctx)
		link, linkCIDs, ok := env.AssertLinkApplication(linkCtx, AppID)
		if !a.So(ok, should.BeTrue) || !a.So(link, should.NotBeNil) {
			t.Error("AS link assertion failed")
			closeLink()
			return
		}
		defer func() {
			closeLink()
			if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
				return a.So(ev.Data(), should.BeError) && test.AllTrue(
					a.So(errors.IsCanceled(ev.Data().(error)), should.BeTrue),
					a.So(ev, should.ResembleEvent, EvtEndApplicationLink.New(
						events.ContextWithCorrelationID(ctx, linkCIDs...),
						events.WithIdentifiers(AppID),
						events.WithData(ev.Data().(error)),
					)),
				)
			}), should.BeTrue) {
				t.Error("AS link end event assertion failed")
			}
		}()

		dev, ok := env.AssertSetDevice(ctx, true, createDevice)
		if !a.So(ok, should.BeTrue) {
			t.Error("Failed to create device")
			return
		}
		t.Log("Device created")
		a.So(dev.CreatedAt, should.HappenAfter, start)
		a.So(dev.UpdatedAt, should.Equal, dev.CreatedAt)
		a.So([]time.Time{start, dev.CreatedAt, time.Now()}, should.BeChronological)
		a.So(dev, should.ResembleFields, &createDevice.EndDevice, createDevice.FieldMask.Paths)

		dev, ok = env.AssertJoin(ctx, JoinAssertionConfig{
			Link:          link,
			Device:        dev,
			ChannelIndex:  1,
			DataRateIndex: ttnpb.DATA_RATE_2,
			RxMetadatas: [][]*ttnpb.RxMetadata{
				nil,
				RxMetadata[3:],
				RxMetadata[:3],
			},
			CorrelationIDs: []string{
				"GsNs-1",
				"GsNs-2",
			},

			ClusterResponse: &NsJsHandleJoinResponse{
				Response: &ttnpb.JoinResponse{
					RawPayload:     bytes.Repeat([]byte{0x42}, 33),
					SessionKeys:    *MakeSessionKeys(dev.LoRaWANVersion, true),
					Lifetime:       time.Hour,
					CorrelationIDs: []string{"NsJs-1", "NsJs-2"},
				},
			},
		})
		if !a.So(ok, should.BeTrue) {
			t.Error("Device failed to join")
			return
		}
		t.Logf("Device successfully joined. DevAddr: %s", dev.PendingSession.DevAddr)
		f(ctx, env, dev, link)
	}
}

func makeClassCOTAAFlowTest(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string, linkADRReqs ...*ttnpb.MACCommand_LinkADRReq) func(context.Context, TestEnvironment) {
	return makeOTAAFlowTest(&ttnpb.SetEndDeviceRequest{
		EndDevice: ttnpb.EndDevice{
			EndDeviceIdentifiers: *MakeOTAAIdentifiers(nil),
			FrequencyPlanID:      fpID,
			LoRaWANVersion:       macVersion,
			LoRaWANPHYVersion:    phyVersion,
			SupportsClassC:       true,
			SupportsJoin:         true,
		},
		FieldMask: pbtypes.FieldMask{
			Paths: []string{
				"frequency_plan_id",
				"lorawan_phy_version",
				"lorawan_version",
				"supports_class_c",
				"supports_join",
			},
		},
	}, func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice, link ttnpb.AsNs_LinkApplicationClient) {
		t, a := test.MustNewTFromContext(ctx)

		var upCmders []MACCommander
		var expectedUpEvBuilders []events.Builder
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
			expectedUpEvBuilders = append(expectedUpEvBuilders,
				mac.EvtReceiveRekeyIndication.With(events.WithData(rekeyInd)),
				mac.EvtEnqueueRekeyConfirmation.With(events.WithData(rekeyConf)),
				mac.EvtReceiveDeviceModeIndication.With(events.WithData(deviceModeInd)),
				mac.EvtClassCSwitch.With(events.WithData(ttnpb.CLASS_A)),
				mac.EvtEnqueueDeviceModeConfirmation.With(events.WithData(deviceModeConf)),
			)
			downCmders = append(downCmders,
				rekeyConf,
				deviceModeConf,
			)
		}

		fp := FrequencyPlan(fpID)
		phy := LoRaWANBands[fp.BandID][phyVersion]

		deviceChannels, ok := ApplyCFList(dev.PendingMACState.PendingJoinRequest.CFList, phy, dev.PendingMACState.CurrentParameters.Channels...)
		if !a.So(ok, should.BeTrue) {
			t.Fatal("Failed to apply CFList")
			return
		}
		upChIdx := uint8(2)
		upDRIdx := ttnpb.DATA_RATE_1
		upConf := DataUplinkConfig{
			Confirmed:      true,
			MACVersion:     macVersion,
			DevAddr:        dev.PendingMACState.PendingJoinRequest.DevAddr,
			FCtrl:          ttnpb.FCtrl{ADR: true},
			FPort:          0x42,
			FRMPayload:     []byte("test"),
			FOpts:          MakeUplinkMACBuffer(phy, upCmders...),
			DataRate:       phy.DataRates[upDRIdx].Rate,
			DataRateIndex:  upDRIdx,
			Frequency:      deviceChannels[upChIdx].UplinkFrequency,
			ChannelIndex:   upChIdx,
			RxMetadata:     RxMetadata[:2],
			CorrelationIDs: []string{"GsNs-data-0"},
		}
		start := time.Now()
		if !a.So(env.AssertHandleDataUplink(
			ctx,
			upConf,
			func(ctx context.Context, assertEvents func(...events.Event) bool, ups ...*ttnpb.UplinkMessage) bool {
				deduplicatedUpConf := upConf
				deduplicatedUpConf.DecodePayload = true
				deduplicatedUpConf.Matched = true
				for _, up := range ups[1:] {
					deduplicatedUpConf.RxMetadata = append(deduplicatedUpConf.RxMetadata, up.RxMetadata...)
				}
				deduplicatedUp := MakeDataUplink(deduplicatedUpConf)

				dev.EndDeviceIdentifiers.DevAddr = &dev.PendingMACState.PendingJoinRequest.DevAddr
				dev.MACState = dev.PendingMACState
				dev.MACState.CurrentParameters.Rx1Delay = dev.PendingMACState.PendingJoinRequest.RxDelay
				dev.MACState.CurrentParameters.Rx1DataRateOffset = dev.PendingMACState.PendingJoinRequest.DownlinkSettings.Rx1DROffset
				dev.MACState.CurrentParameters.Rx2DataRateIndex = dev.PendingMACState.PendingJoinRequest.DownlinkSettings.Rx2DR
				dev.MACState.PendingJoinRequest = nil
				dev.MACState.RecentUplinks = AppendRecentUplink(dev.MACState.RecentUplinks, deduplicatedUp, RecentUplinkCount)
				dev.Session = dev.PendingSession
				dev.PendingMACState = nil
				dev.PendingSession = nil
				dev.RecentUplinks = AppendRecentUplink(dev.RecentUplinks, deduplicatedUp, RecentUplinkCount)

				if !a.So(assertEvents(events.Builders(func() []events.Builder {
					evBuilders := []events.Builder{
						EvtReceiveDataUplink,
					}
					for range ups[1:] {
						evBuilders = append(evBuilders,
							EvtReceiveDataUplink,
							EvtDropDataUplink.With(events.WithData(ErrDuplicate)),
						)
					}
					return append(
						append(
							evBuilders,
							expectedUpEvBuilders...,
						),
						EvtProcessDataUplink,
					)
				}()).New(
					ctx,
					events.WithIdentifiers(dev.EndDeviceIdentifiers),
				)...), should.BeTrue) {
					t.Error("Uplink event assertion failed")
					return false
				}

				var appUp *ttnpb.ApplicationUp
				if !a.So(AssertProcessApplicationUp(ctx, link, func(ctx context.Context, up *ttnpb.ApplicationUp) bool {
					_, a := test.MustNewTFromContext(ctx)
					recvAt := up.GetUplinkMessage().GetReceivedAt()
					appUp = up
					return test.AllTrue(
						a.So(up.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, deduplicatedUp.CorrelationIDs),
						a.So(up.GetUplinkMessage().GetRxMetadata(), should.HaveSameElementsDeep, deduplicatedUp.RxMetadata),
						a.So([]time.Time{start, recvAt, time.Now()}, should.BeChronological),
						a.So(up, should.Resemble, &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
							CorrelationIDs:       up.CorrelationIDs,
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									Confirmed:    deduplicatedUp.Payload.MHDR.MType == ttnpb.MType_CONFIRMED_UP,
									FPort:        deduplicatedUp.Payload.GetMACPayload().FPort,
									FRMPayload:   deduplicatedUp.Payload.GetMACPayload().FRMPayload,
									ReceivedAt:   up.GetUplinkMessage().GetReceivedAt(),
									RxMetadata:   up.GetUplinkMessage().GetRxMetadata(),
									Settings:     deduplicatedUp.Settings,
									SessionKeyID: MakeSessionKeys(macVersion, false).SessionKeyID,
								},
							},
						}),
					)
				}), should.BeTrue) {
					t.Error("Failed to send data uplink to Application Server")
					return false
				}
				return a.So(env.Events, should.ReceiveEventFunc, test.MakeEventEqual(test.EventEqualConfig{
					Identifiers:    true,
					Data:           true,
					Origin:         true,
					Context:        true,
					Visibility:     true,
					Authentication: true,
					RemoteIP:       true,
					UserAgent:      true,
				}),
					EvtForwardDataUplink.New(
						link.Context(),
						events.WithIdentifiers(dev.EndDeviceIdentifiers),
						events.WithData(appUp),
					),
				)
			},
			RxMetadata[2:],
		), should.BeTrue) {
			return
		}

		downCmders = append(downCmders, ttnpb.CID_DEV_STATUS)
		expectedEvBuilders := []events.Builder{mac.EvtEnqueueDevStatusRequest}
		for _, cmd := range linkADRReqs {
			cmd := cmd
			downCmders = append(downCmders, cmd)
			expectedEvBuilders = append(expectedEvBuilders, mac.EvtEnqueueLinkADRRequest.With(events.WithData(cmd)))
		}

		paths := DownlinkPathsFromMetadata(RxMetadata[:]...)
		txReq := &ttnpb.TxRequest{
			Class:            ttnpb.CLASS_A,
			DownlinkPaths:    DownlinkProtoPaths(paths...),
			Rx1Delay:         dev.MACState.CurrentParameters.Rx1Delay,
			Rx1DataRateIndex: test.Must(phy.Rx1DataRate(upDRIdx, dev.MACState.CurrentParameters.Rx1DataRateOffset, dev.MACState.CurrentParameters.DownlinkDwellTime.GetValue())).(ttnpb.DataRateIndex),
			Rx1Frequency:     phy.DownlinkChannels[test.Must(phy.Rx1Channel(upChIdx)).(uint8)].Frequency,
			Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
			Rx2Frequency:     phy.DefaultRx2Parameters.Frequency,
			Priority:         ttnpb.TxSchedulePriority_HIGHEST,
			FrequencyPlanID:  fpID,
		}
		if !a.So(env.AssertScheduleDownlink(
			ctx,
			MakeDownlinkPathsWithPeerIndex(paths, UintRepeat(1, len(paths))...),
			func(ctx, reqCtx context.Context, down *ttnpb.DownlinkMessage) (NsGsScheduleDownlinkResponse, bool) {
				return NsGsScheduleDownlinkResponse{
						Response: &ttnpb.ScheduleDownlinkResponse{},
					}, test.AllTrue(
						a.So(events.CorrelationIDsFromContext(reqCtx), should.NotBeEmpty),
						a.So(down.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, LastUplink(dev.RecentUplinks...).CorrelationIDs),
						a.So(down, should.Resemble, MakeDataDownlink(macVersion, false, dev.Session.DevAddr, ttnpb.FCtrl{
							ADR: true,
							Ack: true,
						}, 0x00, 0x00, 0x00, nil, MakeDownlinkMACBuffer(phy, downCmders...), txReq, down.CorrelationIDs...)),
					)
			}), should.BeTrue) {
			t.Error("Failed to schedule downlink on Gateway Server")
			return
		}
		a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, 2+len(expectedEvBuilders), func(evs ...events.Event) bool {
			return a.So(evs, should.HaveSameElementsFunc, test.MakeEventEqual(test.EventEqualConfig{
				Identifiers:    true,
				Origin:         true,
				Context:        true,
				Visibility:     true,
				Authentication: true,
				RemoteIP:       true,
				UserAgent:      true,
			}), events.Builders(append(
				expectedEvBuilders,
				EvtScheduleDataDownlinkAttempt.With(events.WithData(txReq)),
				EvtScheduleDataDownlinkSuccess.With(events.WithData(&ttnpb.ScheduleDownlinkResponse{})),
			)).New(
				ctx,
				events.WithIdentifiers(dev.EndDeviceIdentifiers)),
			)
		}), should.BeTrue)
	})
}

func TestFlow(t *testing.T) {
	ForEachFrequencyPlanLoRaWANVersionPair(t, func(makeName func(...string) string, fpID string, _ *frequencyplans.FrequencyPlan, phy *band.Band, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion) {
		for flowName, handleFlowTest := range map[string]func(context.Context, TestEnvironment){
			MakeTestCaseName("Class C", "OTAA"): makeClassCOTAAFlowTest(macVersion, phyVersion, fpID, func() []*ttnpb.MACCommand_LinkADRReq {
				switch phy.ID {
				case band.EU_863_870:
					return []*ttnpb.MACCommand_LinkADRReq{
						{
							ChannelMask:   []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false},
							DataRateIndex: ttnpb.DATA_RATE_4,
							TxPowerIndex:  1,
							NbTrans:       1,
						},
					}
				case band.US_902_928:
					return []*ttnpb.MACCommand_LinkADRReq{
						{
							ChannelMask:   []bool{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
							DataRateIndex: ttnpb.DATA_RATE_2,
							TxPowerIndex:  1,
							NbTrans:       1,
						},
					}
				default:
					t.Skipf("Unknown LinkADRReqs for %s band", phy.ID)
					panic("unreachable")
				}
			}()...),
		} {
			handleFlowTest := handleFlowTest
			test.RunSubtest(t, test.SubtestConfig{
				Name:     makeName(flowName),
				Parallel: true,
				Timeout:  (1 << 14) * test.Delay,
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					nsConf := DefaultConfig
					nsConf.NetID = test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)
					nsConf.DeduplicationWindow = (1 << 4) * test.Delay
					nsConf.CooldownWindow = (1 << 9) * test.Delay

					_, ctx, env, stop := StartTest(t, TestConfig{
						Context:       ctx,
						NetworkServer: nsConf,
					})
					defer stop()

					handleFlowTest(ctx, env)
				},
			})
		}
	})
}
