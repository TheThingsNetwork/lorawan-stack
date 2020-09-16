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

func frequencyPlanMACCommands(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string, otaa bool) ([]MACCommander, []events.Builder) {
	switch fpID {
	case test.EUFrequencyPlanID:
		linkADRReq := &ttnpb.MACCommand_LinkADRReq{
			ChannelMask:   []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false},
			DataRateIndex: ttnpb.DATA_RATE_4,
			TxPowerIndex:  1,
			NbTrans:       1,
		}
		return []MACCommander{
				linkADRReq,
			}, []events.Builder{
				mac.EvtEnqueueLinkADRRequest.With(events.WithData(linkADRReq)),
			}
	case test.USFrequencyPlanID:
		linkADRReqs := []MACCommander{
			&ttnpb.MACCommand_LinkADRReq{
				ChannelMask:   []bool{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
				DataRateIndex: ttnpb.DATA_RATE_2,
				TxPowerIndex:  1,
				NbTrans:       1,
			},
		}
		if !otaa || phyVersion.Compare(ttnpb.PHY_V1_0_3_REV_A) < 0 {
			linkADRReqs = append([]MACCommander{
				&ttnpb.MACCommand_LinkADRReq{
					ChannelMask:        []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					ChannelMaskControl: 7,
					DataRateIndex:      ttnpb.DATA_RATE_2,
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
		return linkADRReqs, evBuilders
	case test.ASAUFrequencyPlanID:
		newChannelReq := &ttnpb.MACCommand_NewChannelReq{
			ChannelIndex:     7,
			Frequency:        924600000,
			MaxDataRateIndex: ttnpb.DATA_RATE_5,
		}
		linkADRReq := &ttnpb.MACCommand_LinkADRReq{
			ChannelMask:   []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false},
			DataRateIndex: ttnpb.DATA_RATE_4,
			TxPowerIndex:  1,
			NbTrans:       1,
		}
		return []MACCommander{
				newChannelReq,
				linkADRReq,
			}, []events.Builder{
				mac.EvtEnqueueNewChannelRequest.With(events.WithData(newChannelReq)),
				mac.EvtEnqueueLinkADRRequest.With(events.WithData(linkADRReq)),
			}
	default:
		panic(fmt.Errorf("unknown LinkADRReqs for %s frequency plan", fpID))
	}
}

type OTAAFlowTestConfig struct {
	CreateDevice *ttnpb.SetEndDeviceRequest
	Func         func(context.Context, TestEnvironment, *ttnpb.EndDevice, ttnpb.AsNs_LinkApplicationClient)

	UplinkMACCommanders   []MACCommander
	UplinkEventBuilders   []events.Builder
	DownlinkMACCommanders []MACCommander
	DownlinkEventBuilders []events.Builder
}

func makeOTAAFlowTest(conf OTAAFlowTestConfig) func(context.Context, TestEnvironment) {
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

		dev, ok := env.AssertSetDevice(ctx, true, conf.CreateDevice)
		if !a.So(ok, should.BeTrue) {
			t.Error("Failed to create device")
			return
		}
		t.Log("Device created")
		a.So(dev.CreatedAt, should.HappenAfter, start)
		a.So(dev.UpdatedAt, should.Equal, dev.CreatedAt)
		a.So([]time.Time{start, dev.CreatedAt, time.Now()}, should.BeChronological)
		a.So(dev, should.ResembleFields, &conf.CreateDevice.EndDevice, conf.CreateDevice.FieldMask.Paths)

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
				"GsNs-join-1",
				"GsNs-join-2",
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

		var upCmders []MACCommander
		var upEvBuilders []events.Builder
		var downCmders []MACCommander
		if dev.PendingMACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			rekeyInd := &ttnpb.MACCommand_RekeyInd{
				MinorVersion: ttnpb.MINOR_1,
			}
			upCmders = []MACCommander{
				rekeyInd,
			}

			rekeyConf := &ttnpb.MACCommand_RekeyConf{
				MinorVersion: ttnpb.MINOR_1,
			}
			upEvBuilders = []events.Builder{
				mac.EvtReceiveRekeyIndication.With(events.WithData(rekeyInd)),
				mac.EvtEnqueueRekeyConfirmation.With(events.WithData(rekeyConf)),
			}
			downCmders = []MACCommander{
				rekeyConf,
			}
		}

		fp := FrequencyPlan(dev.FrequencyPlanID)
		phy := LoRaWANBands[fp.BandID][dev.LoRaWANPHYVersion]

		deviceChannels, ok := ApplyCFList(dev.PendingMACState.PendingJoinRequest.CFList, phy, dev.PendingMACState.CurrentParameters.Channels...)
		if !a.So(ok, should.BeTrue) {
			t.Error("Failed to apply CFList")
			return
		}
		dev.PendingMACState.CurrentParameters.Channels = deviceChannels
		dev.EndDeviceIdentifiers.DevAddr = &dev.PendingSession.DevAddr
		dev, ok = env.AssertHandleDataUplink(ctx, DataUplinkAssertionConfig{
			Link:          link,
			Device:        dev,
			ChannelIndex:  2,
			DataRateIndex: ttnpb.DATA_RATE_1,
			RxMetadatas: [][]*ttnpb.RxMetadata{
				RxMetadata[:2],
				RxMetadata[2:],
			},
			CorrelationIDs: []string{"GsNs-data-0"},

			Confirmed:     true,
			Pending:       true,
			FRMPayload:    []byte("test"),
			FOpts:         MakeUplinkMACBuffer(phy, append(upCmders, conf.UplinkMACCommanders...)...),
			FCtrl:         ttnpb.FCtrl{ADR: true},
			FPort:         0x42,
			EventBuilders: append(upEvBuilders, conf.UplinkEventBuilders...),
		})
		if !a.So(ok, should.BeTrue) {
			t.Error("Data uplink assertion failed")
			return
		}

		fpCmders, fpEvBuilders := frequencyPlanMACCommands(dev.MACState.LoRaWANVersion, dev.LoRaWANPHYVersion, dev.FrequencyPlanID, true)
		downEvBuilders := append(conf.DownlinkEventBuilders, fpEvBuilders...)
		downCmders = append(downCmders, conf.DownlinkMACCommanders...)
		downCmders = append(downCmders, fpCmders...)

		fOpts := MakeDownlinkMACBuffer(phy, downCmders...)
		var frmPayload []byte
		if len(fOpts) > 15 {
			frmPayload = fOpts
			fOpts = nil
		}
		down := MakeDataDownlink(DataDownlinkConfig{
			DecodePayload: true,

			MACVersion: dev.MACState.LoRaWANVersion,
			DevAddr:    dev.Session.DevAddr,
			FCtrl: ttnpb.FCtrl{
				ADR: true,
				Ack: true,
			},
			FRMPayload:  frmPayload,
			FOpts:       fOpts,
			SessionKeys: &dev.Session.SessionKeys,
		})
		dev, ok = env.AssertScheduleDataDownlink(ctx, DataDownlinkAssertionConfig{
			SetRX1:      true,
			SetRX2:      true,
			Device:      dev,
			Class:       ttnpb.CLASS_A,
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
			events.ContextWithCorrelationID(ctx, LastDownlink(dev.RecentDownlinks...).CorrelationIDs...),
			events.WithIdentifiers(dev.EndDeviceIdentifiers)),
		) {
			t.Error("Data downlink event assertion failed")
			return
		}

		conf.Func(ctx, env, dev, link)
	}
}

func makeClassAOTAAFlowTest(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string) func(context.Context, TestEnvironment) {
	return makeOTAAFlowTest(OTAAFlowTestConfig{
		CreateDevice: &ttnpb.SetEndDeviceRequest{
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: *MakeOTAAIdentifiers(nil),
				FrequencyPlanID:      fpID,
				LoRaWANVersion:       macVersion,
				LoRaWANPHYVersion:    phyVersion,
				SupportsJoin:         true,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"supports_join",
				},
			},
		},
		DownlinkMACCommanders: []MACCommander{ttnpb.CID_DEV_STATUS},
		DownlinkEventBuilders: []events.Builder{mac.EvtEnqueueDevStatusRequest},
		Func: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice, link ttnpb.AsNs_LinkApplicationClient) {
		},
	})
}

func makeClassCOTAAFlowTest(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string) func(context.Context, TestEnvironment) {
	var upCmders []MACCommander
	var upEvBuilders []events.Builder
	var downCmders []MACCommander
	if macVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
		deviceModeInd := &ttnpb.MACCommand_DeviceModeInd{
			Class: ttnpb.CLASS_C,
		}
		upCmders = []MACCommander{
			deviceModeInd,
		}

		deviceModeConf := &ttnpb.MACCommand_DeviceModeConf{
			Class: ttnpb.CLASS_C,
		}
		upEvBuilders = []events.Builder{
			mac.EvtReceiveDeviceModeIndication.With(events.WithData(deviceModeInd)),
			mac.EvtClassCSwitch.With(events.WithData(ttnpb.CLASS_A)),
			mac.EvtEnqueueDeviceModeConfirmation.With(events.WithData(deviceModeConf)),
		}
		downCmders = []MACCommander{
			deviceModeConf,
		}
	}
	return makeOTAAFlowTest(OTAAFlowTestConfig{
		CreateDevice: &ttnpb.SetEndDeviceRequest{
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
		},
		UplinkMACCommanders:   upCmders,
		UplinkEventBuilders:   upEvBuilders,
		DownlinkMACCommanders: append(downCmders, ttnpb.CID_DEV_STATUS),
		DownlinkEventBuilders: []events.Builder{mac.EvtEnqueueDevStatusRequest},
		Func: func(ctx context.Context, env TestEnvironment, dev *ttnpb.EndDevice, link ttnpb.AsNs_LinkApplicationClient) {
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
							d = ttnpb.RX_DELAY_4
						case cpus >= 12:
							d = ttnpb.RX_DELAY_15
						default:
							d = ttnpb.RxDelay(cpus + 3)
						}
						return &d
					}()
					nsConf.NetID = test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)
					nsConf.DeduplicationWindow = (1 << 8) * test.Delay
					nsConf.CooldownWindow = (1 << 11) * test.Delay

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
