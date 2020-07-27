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
	"reflect"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

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
	xp.UniqueID = ""
	yp.UniqueID = ""
	xp.Data = nil
	yp.Data = nil
	xp.Time = time.Time{}
	yp.Time = time.Time{}
	xp.Authentication = nil
	yp.Authentication = nil

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
		xp.UniqueID = ""
		yp.UniqueID = ""
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

func makeClassCOTAAFlowTest(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fpID string, linkADRReqs ...*ttnpb.MACCommand_LinkADRReq) func(context.Context, TestEnvironment) {
	return func(ctx context.Context, env TestEnvironment) {
		t := test.MustTFromContext(ctx)
		a := assertions.New(t)

		start := time.Now()

		linkCtx, closeLink := context.WithCancel(ctx)
		link, linkEndEvent, ok := env.AssertLinkApplication(linkCtx, AppID)
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

		joinReq, ok := env.AssertJoin(ctx, link, linkCtx, ids, fpID, macVersion, phyVersion, 1, ttnpb.DATA_RATE_2, makeAssertFlowTestEventEqual(t))
		if !a.So(ok, should.BeTrue) {
			t.Error("Device failed to join")
			return
		}
		t.Logf("Device successfully joined. DevAddr: %s", joinReq.DevAddr)

		ids = *MakeOTAAIdentifiers(&joinReq.DevAddr)
		fp := FrequencyPlan(fpID)
		phy := LoRaWANBands[fp.BandID][phyVersion]

		upChs := DeviceDesiredChannels(phy, fp, env.Config.DefaultMACSettings.Parse())
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
				EvtReceiveRekeyIndication.NewWithIdentifiersAndData(ctx, ids, rekeyInd),
				EvtEnqueueRekeyConfirmation.NewWithIdentifiersAndData(ctx, ids, rekeyConf),
				EvtReceiveDeviceModeIndication.NewWithIdentifiersAndData(ctx, ids, deviceModeInd),
				EvtClassCSwitch.NewWithIdentifiersAndData(ctx, ids, ttnpb.CLASS_A),
				EvtEnqueueDeviceModeConfirmation.NewWithIdentifiersAndData(ctx, ids, deviceModeConf),
			)
			downCmders = append(downCmders,
				rekeyConf,
				deviceModeConf,
			)
		}
		makeUplink := func(matched bool, rxMetadata ...*ttnpb.RxMetadata) *ttnpb.UplinkMessage {
			msg := MakeDataUplink(DataUplinkConfig{
				MACVersion:    macVersion,
				DecodePayload: matched,
				Confirmed:     true,
				DevAddr:       joinReq.DevAddr,
				FCtrl:         ttnpb.FCtrl{ADR: true},
				FPort:         0x42,
				FRMPayload:    []byte("test"),
				FOpts:         MakeUplinkMACBuffer(phy, upCmders...),
				DataRate:      phy.DataRates[upDRIdx].Rate,
				DataRateIndex: upDRIdx,
				Frequency:     upChs[upChIdx].UplinkFrequency,
				ChannelIndex:  upChIdx,
				RxMetadata:    rxMetadata,
			})
			if matched {
				return WithMatchedUplinkSettings(msg, upChIdx, upDRIdx)
			}
			return msg
		}
		expectedUp := makeUplink(true, RxMetadata[:]...)
		start = time.Now()
		if !a.So(env.AssertSendDataUplink(ctx, link, linkCtx, ids, makeUplink,
			makeAssertFlowTestEventEqual(t),
			append(expectedUpEvs,
				EvtProcessDataUplink.NewWithIdentifiersAndData(ctx, ids, expectedUp),
			)...), should.BeTrue) {
			t.Error("Failed to process data uplink")
			return
		}

		var expectedEvs []events.Event
		if !a.So(AssertProcessApplicationUp(ctx, link, func(ctx context.Context, up *ttnpb.ApplicationUp) bool {
			expectedEvs = append(expectedEvs, EvtForwardDataUplink.NewWithIdentifiersAndData(linkCtx, up.EndDeviceIdentifiers, up))

			t := test.MustTFromContext(ctx)
			t.Helper()
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
		expectedEvs = append(expectedEvs, EvtEnqueueDevStatusRequest.NewWithIdentifiersAndData(ctx, ids, nil))
		for _, cmd := range linkADRReqs {
			cmd := cmd
			downCmders = append(downCmders, cmd)
			expectedEvs = append(expectedEvs, EvtEnqueueLinkADRRequest.NewWithIdentifiersAndData(ctx, ids, cmd))
		}

		paths := DownlinkPathsFromMetadata(RxMetadata[:]...)
		txReq := &ttnpb.TxRequest{
			Class:            ttnpb.CLASS_A,
			DownlinkPaths:    DownlinkProtoPaths(paths...),
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
				EvtScheduleDataDownlinkAttempt.NewWithIdentifiersAndData(ctx, ids, txReq),
				EvtScheduleDataDownlinkSuccess.NewWithIdentifiersAndData(ctx, ids, &ttnpb.ScheduleDownlinkResponse{}),
			))
		}), should.BeTrue)
	}
}

func TestFlow(t *testing.T) {
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
	for flowName, handleFlowTest := range map[string]func(context.Context, TestEnvironment){
		"Class C/OTAA/MAC:1.0.3/PHY:1.0.3-a/FP:EU868": makeClassCOTAAFlowTest(ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A, test.EUFrequencyPlanID, eu868LinkADRReqs...),
		"Class C/OTAA/MAC:1.0.4/PHY:1.0.3-a/FP:US915": makeClassCOTAAFlowTest(ttnpb.MAC_V1_0_4, ttnpb.PHY_V1_0_3_REV_A, test.USFrequencyPlanID, us915LinkADRReqs...),
		"Class C/OTAA/MAC:1.1/PHY:1.1-b/FP:EU868":     makeClassCOTAAFlowTest(ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B, test.EUFrequencyPlanID, eu868LinkADRReqs...),
	} {
		t.Run(flowName, func(t *testing.T) {
			t.Parallel()

			nsConf := DefaultConfig
			nsConf.NetID = test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)
			nsConf.DeduplicationWindow = (1 << 4) * test.Delay
			nsConf.CooldownWindow = (1 << 9) * test.Delay

			_, ctx, env, stop := StartTest(t, TestConfig{
				NetworkServer: nsConf,
				Timeout:       (1 << 13) * test.Delay,
			})
			defer stop()

			handleFlowTest(ctx, env)
		})
	}
}
