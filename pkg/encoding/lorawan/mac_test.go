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

package lorawan_test

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	. "go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLoRaWANEncodingMAC(t *testing.T) {
	phy := test.Must(band.Get(band.EU_863_870, ttnpb.PHYVersion_RP001_V1_1_REV_B))

	for _, tc := range []struct {
		Name    string
		Payload interface {
			MACCommand() *ttnpb.MACCommand
		}
		Bytes    []byte
		IsUplink bool
	}{
		{
			"ResetConf",
			&ttnpb.MACCommand_ResetConf{MinorVersion: 1},
			[]byte{0x01, 1},
			false,
		},
		{
			"ResetInd",
			&ttnpb.MACCommand_ResetInd{MinorVersion: 1},
			[]byte{0x01, 1},
			true,
		},
		{
			"ResetConf v1.2",
			&ttnpb.MACCommand_ResetConf{MinorVersion: 2, Cipher: 3},
			[]byte{0x01, 0x32},
			false,
		},
		{
			"ResetInd v1.2",
			&ttnpb.MACCommand_ResetInd{MinorVersion: 2, Cipher: 3},
			[]byte{0x01, 0x32},
			true,
		},
		{
			"LinkCheckReq",
			ttnpb.MACCommandIdentifier_CID_LINK_CHECK,
			[]byte{0x02},
			true,
		},
		{
			"LinkCheckAns",
			&ttnpb.MACCommand_LinkCheckAns{Margin: 20, GatewayCount: 3},
			[]byte{0x02, 20, 3},
			false,
		},
		{
			"LinkADRReq",
			&ttnpb.MACCommand_LinkADRReq{
				DataRateIndex: 0b0101,
				TxPowerIndex:  0b0010,
				ChannelMask: []bool{
					false, false, true, false, false, false, false, false,
					false, true, false, false, false, false, false, false,
				},
				ChannelMaskControl: 1,
				NbTrans:            1,
			},
			[]byte{0x03, 0b0101_0010, 0b00000100, 0b00000010, 0b0_001_0001},
			false,
		},
		{
			"LinkADRAns",
			&ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			[]byte{0x03, 0x07},
			true,
		},
		{
			"DutyCycleReq",
			&ttnpb.MACCommand_DutyCycleReq{
				MaxDutyCycle: 0x0d,
			},
			[]byte{0x04, 0x0d},
			false,
		},
		{
			"DutyCycleAns",
			ttnpb.MACCommandIdentifier_CID_DUTY_CYCLE,
			[]byte{0x04},
			true,
		},
		{
			"RxParamSetupReq",
			&ttnpb.MACCommand_RxParamSetupReq{
				Rx1DataRateOffset: 5,
				Rx2DataRateIndex:  12,
				Rx2Frequency:      1677702600,
			},
			[]byte{0x05, 0x5c, 0x42, 0xff, 0xff},
			false,
		},
		{
			"RxParamSetupAns",
			&ttnpb.MACCommand_RxParamSetupAns{
				Rx2FrequencyAck:      true,
				Rx2DataRateIndexAck:  false,
				Rx1DataRateOffsetAck: true,
			},
			[]byte{0x05, 0x05},
			true,
		},
		{
			"DevStatusReq",
			ttnpb.MACCommandIdentifier_CID_DEV_STATUS,
			[]byte{0x06},
			false,
		},
		{
			"DevStatusAns",
			&ttnpb.MACCommand_DevStatusAns{
				Battery: 0x42,
				Margin:  -16,
			},
			[]byte{0x06, 0x42, 0x2f},
			true,
		},
		{
			"NewChannelReq",
			&ttnpb.MACCommand_NewChannelReq{
				ChannelIndex:     0xf,
				Frequency:        0x42ffff * 100,
				MaxDataRateIndex: 0x4,
				MinDataRateIndex: 0x2,
			},
			[]byte{0x07, 0xf, 0xff, 0xff, 0x42, 0x42},
			false,
		},
		{
			"NewChannelReq/Freq 0",
			&ttnpb.MACCommand_NewChannelReq{
				ChannelIndex:     0xf,
				Frequency:        0x0,
				MaxDataRateIndex: 0x4,
				MinDataRateIndex: 0x2,
			},
			[]byte{0x07, 0xf, 0x0, 0x0, 0x0, 0x42},
			false,
		},
		{
			"NewChannelAns",
			&ttnpb.MACCommand_NewChannelAns{
				FrequencyAck: false,
				DataRateAck:  true,
			},
			[]byte{0x07, 0x2},
			true,
		},
		{
			"RxTimingSetupReq",
			&ttnpb.MACCommand_RxTimingSetupReq{
				Delay: 0xf,
			},
			[]byte{0x08, 0xf},
			false,
		},
		{
			"RxTimingSetupAns",
			ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP,
			[]byte{0x08},
			true,
		},
		{
			"TxParamSetupReq",
			&ttnpb.MACCommand_TxParamSetupReq{
				MaxEirpIndex:      ttnpb.DeviceEIRP_DEVICE_EIRP_36,
				UplinkDwellTime:   false,
				DownlinkDwellTime: true,
			},
			[]byte{0x09, 0x2f},
			false,
		},
		{
			"TxParamSetupAns",
			ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP,
			[]byte{0x09},
			true,
		},
		{
			"DLChannelReq",
			&ttnpb.MACCommand_DLChannelReq{
				ChannelIndex: 0x4,
				Frequency:    0x42ffff * 100,
			},
			[]byte{0x0A, 0x4, 0xff, 0xff, 0x42},
			false,
		},
		{
			"DLChannelAns",
			&ttnpb.MACCommand_DLChannelAns{
				ChannelIndexAck: false,
				FrequencyAck:    true,
			},
			[]byte{0x0A, 0x2},
			true,
		},
		{
			"RekeyInd",
			&ttnpb.MACCommand_RekeyInd{MinorVersion: 1},
			[]byte{0x0B, 1},
			true,
		},
		{
			"RekeyConf",
			&ttnpb.MACCommand_RekeyConf{MinorVersion: 1},
			[]byte{0x0B, 1},
			false,
		},
		{
			"RekeyInd v1.2",
			&ttnpb.MACCommand_RekeyInd{MinorVersion: 2, Cipher: 3},
			[]byte{0x0B, 0x32},
			true,
		},
		{
			"RekeyConf v1.2",
			&ttnpb.MACCommand_RekeyConf{MinorVersion: 2, Cipher: 3},
			[]byte{0x0B, 0x32},
			false,
		},
		{
			"ADRParamSetupReq",
			&ttnpb.MACCommand_ADRParamSetupReq{
				AdrAckDelayExponent: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4,
				AdrAckLimitExponent: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16,
			},
			[]byte{0x0C, 0x42},
			false,
		},
		{
			"ADRParamSetupAns",
			ttnpb.MACCommandIdentifier_CID_ADR_PARAM_SETUP,
			[]byte{0x0C},
			true,
		},
		{
			"DeviceTimeReq",
			ttnpb.MACCommandIdentifier_CID_DEVICE_TIME,
			[]byte{0x0D},
			true,
		},
		{
			"DeviceTimeAns",
			&ttnpb.MACCommand_DeviceTimeAns{
				Time: timestamppb.New(gpstime.Parse(0x42ffffff*time.Second + 0x42*time.Duration(math.Pow(0.5, 8)*float64(time.Second)))),
			},
			[]byte{0x0D, 0xff, 0xff, 0xff, 0x42, 0x42},
			false,
		},
		{
			"ForceRejoinReq",
			&ttnpb.MACCommand_ForceRejoinReq{
				MaxRetries:     0x7,
				PeriodExponent: 0x7,
				DataRateIndex:  0xd,
				RejoinType:     0x2,
			},
			[]byte{0x0E, 0x3f, 0x2d},
			false,
		},
		{
			"RejoinParamSetupReq",
			&ttnpb.MACCommand_RejoinParamSetupReq{
				MaxTimeExponent:  0x4,
				MaxCountExponent: 0x2,
			},
			[]byte{0x0F, 0x42},
			false,
		},
		{
			"RejoinParamSetupAns",
			&ttnpb.MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: true,
			},
			[]byte{0x0F, 0x1},
			true,
		},
		{
			"PingSlotInfoReq",
			&ttnpb.MACCommand_PingSlotInfoReq{
				Period: ttnpb.PingSlotPeriod(0x7),
			},
			[]byte{0x10, 0x7},
			true,
		},
		{
			"PingSlotInfoAns",
			ttnpb.MACCommandIdentifier_CID_PING_SLOT_INFO,
			[]byte{0x10},
			false,
		},
		{
			"PingSlotChannelReq",
			&ttnpb.MACCommand_PingSlotChannelReq{
				Frequency:     0x1a2bff9c, // 0x42ffff * 100
				DataRateIndex: 0xf,
			},
			[]byte{0x11, 0xff, 0xff, 0x42, 0xf},
			false,
		},
		{
			"PingSlotChannelAns",
			&ttnpb.MACCommand_PingSlotChannelAns{
				DataRateIndexAck: true,
				FrequencyAck:     false,
			},
			[]byte{0x11, 0x2},
			true,
		},
		{
			"BeaconTimingReq",
			ttnpb.MACCommandIdentifier_CID_BEACON_TIMING,
			[]byte{0x12},
			true,
		},
		{
			"BeaconTimingAns",
			&ttnpb.MACCommand_BeaconTimingAns{
				Delay:        0x42ff,
				ChannelIndex: 0x42,
			},
			[]byte{0x12, 0xff, 0x42, 0x42},
			false,
		},
		{
			"BeaconFreqReq",
			&ttnpb.MACCommand_BeaconFreqReq{
				Frequency: 0x42ffff * 100,
			},
			[]byte{0x13, 0xff, 0xff, 0x42},
			false,
		},
		{
			"BeaconFreqAns",
			&ttnpb.MACCommand_BeaconFreqAns{
				FrequencyAck: true,
			},
			[]byte{0x13, 0x01},
			true,
		},
		{
			"DeviceModeInd",
			&ttnpb.MACCommand_DeviceModeInd{
				Class: ttnpb.Class_CLASS_A,
			},
			[]byte{0x20, 0x00},
			true,
		},
		{
			"DeviceModeConf",
			&ttnpb.MACCommand_DeviceModeConf{
				Class: ttnpb.Class_CLASS_C,
			},
			[]byte{0x20, 0x02},
			false,
		},

		{
			"RelayConfReqDisabled",
			&ttnpb.MACCommand_RelayConfReq{},
			[]byte{0x40, 0x00, 0x00, 0x00, 0x00, 0x00},
			false,
		},
		{
			"RelayConfReqNoSecondCh",
			&ttnpb.MACCommand_RelayConfReq{
				Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
					SecondChannel:       nil,
					DefaultChannelIndex: 0x01,
					CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
				},
			},
			[]byte{0x40, 0x00, 0x2e, 0x00, 0x00, 0x00},
			false,
		},
		{
			"RelayConfReqSecondCh",
			&ttnpb.MACCommand_RelayConfReq{
				Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
					SecondChannel: &ttnpb.RelaySecondChannel{
						AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_200,
						Frequency:     868100000,
						DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					},
					DefaultChannelIndex: 0x01,
					CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
				},
			},
			[]byte{0x40, 0x89, 0x2e, 0x28, 0x76, 0x84},
			false,
		},
		{
			"RelayConfAns",
			&ttnpb.MACCommand_RelayConfAns{
				SecondChannelFrequencyAck:     true,
				SecondChannelAckOffsetAck:     false,
				SecondChannelDataRateIndexAck: true,
				SecondChannelIndexAck:         false,
				DefaultChannelIndexAck:        true,
				CadPeriodicityAck:             true,
			},
			[]byte{0x40, 0x35},
			true,
		},
		{
			"EndDeviceConfReqDisabled",
			&ttnpb.MACCommand_RelayEndDeviceConfReq{},
			[]byte{0x41, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			false,
		},
		{
			"EndDeviceConfReqNoSecondCh",
			&ttnpb.MACCommand_RelayEndDeviceConfReq{
				Configuration: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
					Mode: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Dynamic{
						Dynamic: &ttnpb.RelayEndDeviceDynamicMode{
							SmartEnableLevel: ttnpb.RelaySmartEnableLevel_RELAY_SMART_ENABLE_LEVEL_64,
						},
					},
					Backoff: 48,
				},
			},
			[]byte{0x41, 0x0b, 0x00, 0x60, 0x00, 0x00, 0x00},
			false,
		},
		{
			"EndDeviceConfReqSecondCh",
			&ttnpb.MACCommand_RelayEndDeviceConfReq{
				Configuration: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
					Mode:    &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_EndDeviceControlled{},
					Backoff: 48,
					SecondChannel: &ttnpb.RelaySecondChannel{
						AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_200,
						Frequency:     868100000,
						DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					},
				},
			},
			[]byte{0x41, 0x0c, 0x89, 0x60, 0x28, 0x76, 0x84},
			false,
		},
		{
			"EndDeviceConfAns",
			&ttnpb.MACCommand_RelayEndDeviceConfAns{
				SecondChannelFrequencyAck:     true,
				SecondChannelDataRateIndexAck: true,
				SecondChannelIndexAck:         false,
				BackoffAck:                    true,
			},
			[]byte{0x41, 0x0b},
			true,
		},
		{
			"UpdateUplinkListReqNoForwardLimits",
			&ttnpb.MACCommand_RelayUpdateUplinkListReq{
				RuleIndex: 1,
				DevAddr:   []byte{0x41, 0x42, 0x43, 0x44},
				WFCnt:     0x00be00ef,
				RootWorSKey: []byte{
					0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
				},
			},
			[]byte{
				0x43,                   // CID
				0x01,                   // RuleIndex
				0x3f,                   // No limits
				0x44, 0x43, 0x42, 0x41, // DevAddr
				0xef, 0x00, 0xbe, 0x00, // WFCnt
				0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // RootWorSKey
			},
			false,
		},
		{
			"UpdateUplinkListReqForwardLimits",
			&ttnpb.MACCommand_RelayUpdateUplinkListReq{
				RuleIndex: 1,
				ForwardLimits: &ttnpb.RelayUplinkForwardLimits{
					BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
					ReloadRate: 48,
				},
				DevAddr: []byte{0x41, 0x42, 0x43, 0x44},
				WFCnt:   0x00be00ef,
				RootWorSKey: []byte{
					0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
				},
			},
			[]byte{
				0x43,                   // CID
				0x01,                   // RuleIndex
				0x70,                   // BucketSize = 2, ReloadRate = 48
				0x44, 0x43, 0x42, 0x41, // DevAddr
				0xef, 0x00, 0xbe, 0x00, // WFCnt
				0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // RootWorSKey
			},
			false,
		},
		{
			"UpdateUplinkListAns",
			&ttnpb.MACCommand_RelayUpdateUplinkListAns{},
			[]byte{0x43},
			true,
		},
		{
			"RelayCtrlUplinkListReqRemoveTrustedEndDevice",
			&ttnpb.MACCommand_RelayCtrlUplinkListReq{
				RuleIndex: 2,
				Action:    ttnpb.RelayCtrlUplinkListAction_RELAY_CTRL_UPLINK_LIST_ACTION_REMOVE_TRUSTED_END_DEVICE,
			},
			[]byte{0x44, 0x12},
			false,
		},
		{
			"RelayCtrlUplinkListReqReadWFCnt",
			&ttnpb.MACCommand_RelayCtrlUplinkListReq{
				RuleIndex: 3,
				Action:    ttnpb.RelayCtrlUplinkListAction_RELAY_CTRL_UPLINK_LIST_ACTION_READ_W_F_CNT,
			},
			[]byte{0x44, 0x03},
			false,
		},
		{
			"RelayCtrlUplinkListAnsAccept",
			&ttnpb.MACCommand_RelayCtrlUplinkListAns{
				RuleIndexAck: true,
				WFCnt:        0x11223344,
			},
			[]byte{0x44, 0x01, 0x44, 0x33, 0x22, 0x11},
			true,
		},
		{
			"RelayCtrlUplinkListAnsReject",
			&ttnpb.MACCommand_RelayCtrlUplinkListAns{
				RuleIndexAck: false,
				WFCnt:        0,
			},
			[]byte{0x44, 0x00, 0x00, 0x00, 0x00, 0x00},
			true,
		},
		{
			"ConfigureFwdLimitReqNoLimits",
			&ttnpb.MACCommand_RelayConfigureFwdLimitReq{},
			[]byte{0x45, 0xff, 0xff, 0xff, 0x0f, 0x00},
			false,
		},
		{
			"ConfigureFwdLimitReqLimits",
			&ttnpb.MACCommand_RelayConfigureFwdLimitReq{
				ResetLimitCounter: ttnpb.RelayResetLimitCounter_RELAY_RESET_LIMIT_COUNTER_NO_RESET,
				JoinRequestLimits: &ttnpb.RelayForwardLimits{
					BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_1,
					ReloadRate: 12,
				},
				NotifyLimits: &ttnpb.RelayForwardLimits{
					BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
					ReloadRate: 23,
				},
				GlobalUplinkLimits: &ttnpb.RelayForwardLimits{
					BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
					ReloadRate: 34,
				},
				OverallLimits: &ttnpb.RelayForwardLimits{
					BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_12,
					ReloadRate: 45,
				},
			},
			[]byte{0x45, 0x2d, 0xd1, 0x85, 0x31, 0x1b},
			false,
		},
		{
			"ConfigureFwdLimitAns",
			&ttnpb.MACCommand_RelayConfigureFwdLimitAns{},
			[]byte{0x45},
			true,
		},
		{
			"NotifyNewEndDeviceReq",
			&ttnpb.MACCommand_RelayNotifyNewEndDeviceReq{
				DevAddr: []byte{0x41, 0x42, 0x43, 0x44},
				Snr:     6,
				Rssi:    -64,
			},
			[]byte{0x46, 0x3a, 0x06, 0x44, 0x43, 0x42, 0x41},
			true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			cmd := tc.Payload.MACCommand()
			if !a.So(cmd, should.NotBeNil) {
				t.FailNow()
			}

			desc := DefaultMACCommands[ttnpb.MACCommandIdentifier(tc.Bytes[0])]
			if !a.So(desc, should.NotBeNil) {
				t.FailNow()
			}

			appender := DefaultMACCommands.AppendUplink
			reader := DefaultMACCommands.ReadUplink
			if !tc.IsUplink {
				appender = DefaultMACCommands.AppendDownlink
				reader = DefaultMACCommands.ReadDownlink
			}

			b, err := appender(phy, []byte{}, cmd)
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, tc.Bytes)
			}

			cmd = &ttnpb.MACCommand{}
			err = reader(phy, bytes.NewReader(tc.Bytes), cmd)
			if a.So(err, should.BeNil) {
				a.So(cmd, should.Resemble, tc.Payload.MACCommand())
			}
		})
	}
}
