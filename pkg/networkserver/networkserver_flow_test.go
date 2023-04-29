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
	"runtime"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/durationpb"
)

// frequencyPlanMACCommands generates the MAC command and event builders that are expected
// for a particular MAC and PHY version pair, in the provided frequency plan, if OTAA is enabled.
// The returned bool value represents the FPending flag, which may be set for pairs where
// the MAC commands do not fit a singular transmission.
func frequencyPlanMACCommands(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string, otaa bool) ([]MACCommander, []events.Builder, bool) {
	switch fpID {
	case test.EUFrequencyPlanID:
		linkADRReq := &ttnpb.MACCommand_LinkADRReq{
			ChannelMask:   []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false},
			DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			TxPowerIndex:  1,
			NbTrans:       1,
		}
		return []MACCommander{
				linkADRReq,
			}, []events.Builder{
				mac.EvtEnqueueLinkADRRequest.With(events.WithData(linkADRReq)),
			}, false
	case test.USFrequencyPlanID:
		linkADRReqs := []MACCommander{
			&ttnpb.MACCommand_LinkADRReq{
				ChannelMask:   []bool{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
				DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
				TxPowerIndex:  1,
				NbTrans:       1,
			},
		}
		beforeRP001_V1_0_3_REV_A := func(v ttnpb.PHYVersion) bool {
			switch v {
			case ttnpb.PHYVersion_TS001_V1_0,
				ttnpb.PHYVersion_TS001_V1_0_1,
				ttnpb.PHYVersion_RP001_V1_0_2,
				ttnpb.PHYVersion_RP001_V1_0_2_REV_B:
				return true
			default:
				return false
			}
		}(phyVersion)
		if !otaa || beforeRP001_V1_0_3_REV_A {
			linkADRReqs = append([]MACCommander{
				&ttnpb.MACCommand_LinkADRReq{
					ChannelMask:        []bool{false, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					ChannelMaskControl: 7,
					DataRateIndex:      ttnpb.DataRateIndex_DATA_RATE_3,
					TxPowerIndex:       1,
					NbTrans:            1,
				},
			}, linkADRReqs...)
		}
		var evBuilders []events.Builder
		for _, req := range linkADRReqs {
			req := req
			evBuilders = append(evBuilders, mac.EvtEnqueueLinkADRRequest.With(events.WithData(req)))
		}
		return linkADRReqs, evBuilders, false
	case test.ASAUFrequencyPlanID:
		maxEIRPIndex := ttnpb.DeviceEIRP(5)
		// The maximum EIRP in the RP001-1.0.2A version is 14 dB only, while the other versions use 16 dB.
		if phyVersion == ttnpb.PHYVersion_PHY_V1_0_2_REV_A {
			maxEIRPIndex = ttnpb.DeviceEIRP(4)
		}
		txParamSetupReq := &ttnpb.MACCommand_TxParamSetupReq{
			MaxEirpIndex:      maxEIRPIndex,
			DownlinkDwellTime: false,
			UplinkDwellTime:   false,
		}
		macCommanders := []MACCommander{txParamSetupReq}
		evBuilders := []events.Builder{mac.EvtEnqueueTxParamSetupRequest.With(events.WithData(txParamSetupReq))}

		newChannelReq := &ttnpb.MACCommand_NewChannelReq{
			ChannelIndex:     7,
			Frequency:        924600000,
			MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
		}
		// The boot time settings of RP001-1.0.2B devices enable downlink dwell time, which disables the
		// ability to include a CFList as part of the JoinAccept. This causes the non-default channels to be
		// included via NewChannelReq.
		if phyVersion == ttnpb.PHYVersion_PHY_V1_0_2_REV_B {
			newChannelReq = &ttnpb.MACCommand_NewChannelReq{
				ChannelIndex:     2,
				Frequency:        923600000,
				MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			}
		}

		macCommanders = append(macCommanders, newChannelReq)
		evBuilders = append(evBuilders, mac.EvtEnqueueNewChannelRequest.With(events.WithData(newChannelReq)))

		fPending := true
		// The boot time settings of RP001-1.0.2B devices enable downlink dwell time, which limits the number
		// of MAC commands that can be part of the downlink. As such, the LinkADRReq is not sent as part of
		// the initial downlink on this version.
		if phyVersion != ttnpb.PHYVersion_PHY_V1_0_2_REV_B {
			linkADRReq := &ttnpb.MACCommand_LinkADRReq{
				ChannelMask:   []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false},
				DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
				TxPowerIndex:  1,
				NbTrans:       1,
			}

			macCommanders = append(macCommanders, linkADRReq)
			evBuilders = append(evBuilders, mac.EvtEnqueueLinkADRRequest.With(events.WithData(linkADRReq)))
			fPending = false
		}
		return macCommanders, evBuilders, fPending
	default:
		panic(fmt.Errorf("unknown LinkADRReqs for %s frequency plan", fpID))
	}
}

type OTAAFlowTestConfig struct {
	CreateDevice *ttnpb.SetEndDeviceRequest
	Func         func(context.Context, TestEnvironment, *ttnpb.EndDevice)

	UplinkMACCommanders       []MACCommander
	UplinkEventBuilders       []events.Builder
	DownlinkHeadMACCommanders []MACCommander
	DownlinkHeadEventBuilders []events.Builder
	DownlinkTailMACCommanders []MACCommander
	DownlinkTailEventBuilders []events.Builder
}

func makeOTAAFlowTest(conf OTAAFlowTestConfig) func(context.Context, TestEnvironment) {
	return func(ctx context.Context, env TestEnvironment) {
		t, a := test.MustNewTFromContext(ctx)

		start := time.Now()

		dev, err, ok := env.AssertSetDevice(ctx, true, conf.CreateDevice,
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
		)
		if !a.So(err, should.BeNil) || !a.So(ok, should.BeTrue) {
			t.Error("Failed to create device")
			return
		}
		t.Log("Device created")
		a.So(*ttnpb.StdTime(dev.CreatedAt), should.HappenAfter, start)
		a.So(dev.UpdatedAt, should.Resemble, dev.CreatedAt)
		a.So([]time.Time{start, *ttnpb.StdTime(dev.CreatedAt), time.Now()}, should.BeChronological)
		a.So(dev, should.ResembleFields, conf.CreateDevice.EndDevice, conf.CreateDevice.FieldMask.GetPaths())

		responseLen := lorawan.JoinAcceptWithCFListLength
		if dev.GetPendingMacState().GetPendingJoinRequest().GetCfList() == nil {
			responseLen = lorawan.JoinAcceptWithoutCFListLength
		}
		dev, ok = env.AssertJoin(ctx, JoinAssertionConfig{
			Device:        dev,
			ChannelIndex:  1,
			DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
			RxMetadatas: [][]*ttnpb.RxMetadata{
				nil,
				DefaultRxMetadata[3:],
				DefaultRxMetadata[:3],
			},
			CorrelationIDs: []string{
				"GsNs-join-1",
				"GsNs-join-2",
			},

			ClusterResponse: &NsJsHandleJoinResponse{
				Response: &ttnpb.JoinResponse{
					RawPayload: bytes.Repeat([]byte{0x42}, responseLen),
					SessionKeys: test.MakeSessionKeys(
						test.SessionKeysOptions.WithDefaultNwkKeys(dev.LorawanVersion),
					),
					Lifetime:       durationpb.New(time.Hour),
					CorrelationIds: []string{"NsJs-1", "NsJs-2"},
				},
			},
		})
		if !a.So(ok, should.BeTrue) {
			t.Error("Device failed to join")
			return
		}
		t.Logf("Device successfully joined. DevAddr: %s", types.MustDevAddr(dev.PendingSession.DevAddr).OrZero())

		var upCmders []MACCommander
		var upEvBuilders []events.Builder
		var downCmders []MACCommander
		if macspec.UseRekeyInd(dev.PendingMacState.LorawanVersion) {
			rekeyInd := &ttnpb.MACCommand_RekeyInd{
				MinorVersion: ttnpb.Minor_MINOR_1,
			}
			upCmders = []MACCommander{
				rekeyInd,
			}

			rekeyConf := &ttnpb.MACCommand_RekeyConf{
				MinorVersion: ttnpb.Minor_MINOR_1,
			}
			upEvBuilders = []events.Builder{
				mac.EvtReceiveRekeyIndication.With(events.WithData(rekeyInd)),
				mac.EvtEnqueueRekeyConfirmation.With(events.WithData(rekeyConf)),
			}
			downCmders = []MACCommander{
				rekeyConf,
			}
		}

		fp := test.FrequencyPlan(dev.FrequencyPlanId)
		phy := LoRaWANBands[fp.BandID][dev.LorawanPhyVersion]

		deviceChannels, ok := ApplyCFList(dev.PendingMacState.PendingJoinRequest.CfList, phy, dev.PendingMacState.CurrentParameters.Channels...)
		if !a.So(ok, should.BeTrue) {
			t.Error("Failed to apply CFList")
			return
		}
		dev.PendingMacState.CurrentParameters.Channels = deviceChannels
		dev.Ids.DevAddr = dev.PendingSession.DevAddr
		dev, ok = env.AssertHandleDataUplink(ctx, DataUplinkAssertionConfig{
			Device:        dev,
			ChannelIndex:  1,
			DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			RxMetadatas: [][]*ttnpb.RxMetadata{
				DefaultRxMetadata[:2],
				DefaultRxMetadata[2:],
				PacketBrokerRxMetadata[:],
			},
			CorrelationIDs: []string{"GsNs-data-0"},

			Confirmed:     true,
			Pending:       true,
			FRMPayload:    []byte("test"),
			FOpts:         MakeUplinkMACBuffer(phy, append(upCmders, conf.UplinkMACCommanders...)...),
			FCtrl:         &ttnpb.FCtrl{Adr: true},
			FPort:         0x42,
			EventBuilders: append(upEvBuilders, conf.UplinkEventBuilders...),
		})
		if !a.So(ok, should.BeTrue) {
			t.Error("Data uplink assertion failed")
			return
		}

		fpCmders, fpEvBuilders, fPending := frequencyPlanMACCommands(dev.MacState.LorawanVersion, dev.LorawanPhyVersion, dev.FrequencyPlanId, true)
		downEvBuilders := append(conf.DownlinkHeadEventBuilders, fpEvBuilders...)
		downEvBuilders = append(downEvBuilders, conf.DownlinkTailEventBuilders...)
		downCmders = append(downCmders, conf.DownlinkHeadMACCommanders...)
		downCmders = append(downCmders, fpCmders...)
		downCmders = append(downCmders, conf.DownlinkTailMACCommanders...)

		fOpts := MakeDownlinkMACBuffer(phy, downCmders...)
		var frmPayload []byte
		if len(fOpts) > 15 {
			frmPayload = fOpts
			fOpts = nil
		}
		down := MakeDataDownlink(&DataDownlinkConfig{
			DecodePayload: true,

			MACVersion: dev.MacState.LorawanVersion,
			DevAddr:    types.MustDevAddr(dev.Session.DevAddr).OrZero(),
			FCtrl: &ttnpb.FCtrl{
				Adr: true,
				Ack: true,

				FPending: fPending,
			},
			FRMPayload:  frmPayload,
			FOpts:       fOpts,
			SessionKeys: dev.Session.Keys,
		})
		dev, ok = env.AssertScheduleDataDownlink(ctx, DataDownlinkAssertionConfig{
			SetRX1:      !SkipRX1Window(ttnpb.DataRateIndex_DATA_RATE_1, dev.MacState, phy),
			SetRX2:      true,
			Device:      dev,
			Class:       ttnpb.Class_CLASS_A,
			Priority:    ttnpb.TxSchedulePriority_HIGHEST,
			Payload:     down.Payload,
			RawPayload:  down.RawPayload,
			PeerIndexes: []uint{0, 1},
			Responses: []NsGsScheduleDownlinkResponse{
				{
					Response: &ttnpb.ScheduleDownlinkResponse{},
				},
			},
		})
		if !a.So(ok, should.BeTrue) {
			t.Error("Data downlink assertion failed")
			return
		}
		if !a.So(env.Events, should.ReceiveEventsResembling, events.Builders(downEvBuilders).New(
			events.ContextWithCorrelationID(ctx, LastDownlink(dev.MacState.RecentDownlinks...).CorrelationIds...),
			events.WithIdentifiers(dev.Ids)),
		) {
			t.Error("Data downlink event assertion failed")
			return
		}

		if macspec.UseSharedFCntDown(dev.MacState.LorawanVersion) {
			if !a.So(env.AssertNsAsHandleUplink(ctx, dev.Ids.ApplicationIds, func(ctx context.Context, ups ...*ttnpb.ApplicationUp) bool {
				_, a := test.MustNewTFromContext(ctx)
				if !a.So(ups, should.HaveLength, 1) {
					return false
				}
				up := ups[0]
				return test.AllTrue(
					// TODO: Enable this assertion once https://github.com/TheThingsNetwork/lorawan-stack/issues/3416 is done.
					// a.So(up.CorrelationIDs, should.HaveSameElementsDeep, LastDownlink(dev.RecentDownlinks...).CorrelationIDs),
					a.So(up, should.Resemble, &ttnpb.ApplicationUp{
						EndDeviceIds:   dev.Ids,
						CorrelationIds: up.CorrelationIds,
						Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
							DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
								SessionKeyId: dev.Session.Keys.SessionKeyId,
							},
						},
					}),
				)
			}, nil), should.BeTrue) {
				t.Error("Failed to send queue invalidation to Application Server")
				return
			}
		}
		conf.Func(ctx, env, dev)
	}
}

func makeClassAOTAAFlowTest(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string) func(context.Context, TestEnvironment) {
	return makeOTAAFlowTest(OTAAFlowTestConfig{
		CreateDevice: &ttnpb.SetEndDeviceRequest{
			EndDevice: MakeOTAAEndDevice(
				EndDeviceOptions.WithFrequencyPlanId(fpID),
				EndDeviceOptions.WithLorawanVersion(macVersion),
				EndDeviceOptions.WithLorawanPhyVersion(phyVersion),
			),
			FieldMask: ttnpb.FieldMask(
				"frequency_plan_id",
				"lorawan_phy_version",
				"lorawan_version",
				"supports_join",
			),
		},
		DownlinkTailMACCommanders: []MACCommander{ttnpb.MACCommandIdentifier_CID_DEV_STATUS},
		DownlinkTailEventBuilders: []events.Builder{mac.EvtEnqueueDevStatusRequest},
		Func: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) {
		},
	})
}

func makeClassCOTAAFlowTest(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string) func(context.Context, TestEnvironment) {
	var upCmders []MACCommander
	var upEvBuilders []events.Builder
	var downCmders []MACCommander
	if macspec.UseDeviceModeInd(macVersion) {
		deviceModeInd := &ttnpb.MACCommand_DeviceModeInd{
			Class: ttnpb.Class_CLASS_C,
		}
		upCmders = []MACCommander{
			deviceModeInd,
		}

		deviceModeConf := &ttnpb.MACCommand_DeviceModeConf{
			Class: ttnpb.Class_CLASS_C,
		}
		upEvBuilders = []events.Builder{
			mac.EvtReceiveDeviceModeIndication.With(events.WithData(deviceModeInd)),
			mac.EvtClassCSwitch.With(events.WithData(ttnpb.Class_CLASS_A)),
			mac.EvtEnqueueDeviceModeConfirmation.With(events.WithData(deviceModeConf)),
		}
		downCmders = []MACCommander{
			deviceModeConf,
		}
	}
	return makeOTAAFlowTest(OTAAFlowTestConfig{
		CreateDevice: &ttnpb.SetEndDeviceRequest{
			EndDevice: MakeOTAAEndDevice(
				EndDeviceOptions.WithFrequencyPlanId(fpID),
				EndDeviceOptions.WithLorawanVersion(macVersion),
				EndDeviceOptions.WithLorawanPhyVersion(phyVersion),
				EndDeviceOptions.WithSupportsClassC(true),
			),
			FieldMask: ttnpb.FieldMask(
				"frequency_plan_id",
				"lorawan_phy_version",
				"lorawan_version",
				"supports_class_c",
				"supports_join",
			),
		},
		UplinkMACCommanders:       upCmders,
		UplinkEventBuilders:       upEvBuilders,
		DownlinkHeadMACCommanders: downCmders,
		DownlinkTailMACCommanders: []MACCommander{ttnpb.MACCommandIdentifier_CID_DEV_STATUS},
		DownlinkTailEventBuilders: []events.Builder{mac.EvtEnqueueDevStatusRequest},
		Func: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice) {
		},
	})
}

func TestFlow(t *testing.T) {
	ForEachFrequencyPlanLoRaWANVersionPair(t, func(makeName func(...string) string, fpID string, _ *frequencyplans.FrequencyPlan, phy *band.Band, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion) {
		for flowName, handleFlowTest := range map[string]func(context.Context, TestEnvironment){
			MakeTestCaseName("Class A", "OTAA"): makeClassAOTAAFlowTest(macVersion, phyVersion, fpID),
			MakeTestCaseName("Class C", "OTAA"): makeClassCOTAAFlowTest(macVersion, phyVersion, fpID),
		} {
			handleFlowTest := handleFlowTest
			test.RunSubtest(t, test.SubtestConfig{
				Name:     makeName(flowName),
				Parallel: true,
				Timeout:  (1 << 17) * test.Delay,
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					nsConf := DefaultConfig
					nsConf.DefaultMACSettings.DesiredRx1Delay = func() *ttnpb.RxDelay {
						var d ttnpb.RxDelay
						switch cpus := runtime.NumCPU(); {
						case cpus <= 1:
							d = ttnpb.RxDelay_RX_DELAY_4
						case cpus >= 12:
							d = ttnpb.RxDelay_RX_DELAY_15
						default:
							d = ttnpb.RxDelay(cpus + 3)
						}
						return &d
					}()
					nsConf.NetID = test.Must(types.NewNetID(2, []byte{3})).(types.NetID)
					nsConf.ClusterID = "test-cluster"
					nsConf.DeduplicationWindow = (1 << 8) * test.Delay
					nsConf.CooldownWindow = (1 << 11) * test.Delay

					_, ctx, env, stop := StartTest(ctx, TestConfig{
						NetworkServer: nsConf,
						Component: component.Config{
							ServiceBase: config.ServiceBase{
								FrequencyPlans: config.FrequencyPlansConfig{
									ConfigSource: "static",
									Static:       test.StaticFrequencyPlans,
								},
							},
						},
					})
					defer stop()
					handleFlowTest(ctx, env)
				},
			})
		}
	})
}
