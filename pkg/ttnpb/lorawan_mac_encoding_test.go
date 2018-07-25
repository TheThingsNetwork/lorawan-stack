// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb_test

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gpstime"
	. "go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestLoRaWANEncodingMAC(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Payload interface {
			MACCommand() *MACCommand
		}
		Bytes    []byte
		IsUplink bool
	}{
		{
			"ResetConf",
			&MACCommand_ResetConf{MinorVersion: 1},
			[]byte{0x01, 1},
			false,
		},
		{
			"ResetInd",
			&MACCommand_ResetInd{MinorVersion: 1},
			[]byte{0x01, 1},
			true,
		},
		{
			"ResetConf",
			&MACCommand_ResetConf{MinorVersion: 1},
			[]byte{0x01, 1},
			false,
		},
		{
			"LinkCheckReq",
			CID_LINK_CHECK,
			[]byte{0x02},
			true,
		},
		{
			"LinkCheckAns",
			&MACCommand_LinkCheckAns{Margin: 20, GatewayCount: 3},
			[]byte{0x02, 20, 3},
			false,
		},
		{
			"LinkADRReq",
			&MACCommand_LinkADRReq{
				DataRateIndex: 5,
				TxPowerIndex:  2,
				ChannelMask: []bool{
					false, false, true, false, false, false, false, false,
					false, true, false, false, false, false, false, false,
				},
				ChannelMaskControl: 1,
				NbTrans:            1,
			},
			[]byte{0x03, 0x52, 0x04, 0x02, 0x11},
			false,
		},
		{
			"LinkADRAns",
			&MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			[]byte{0x03, 0x07},
			true,
		},
		{
			"DutyCycleReq",
			&MACCommand_DutyCycleReq{
				MaxDutyCycle: 0x0d,
			},
			[]byte{0x04, 0x0d},
			false,
		},
		{
			"DutyCycleAns",
			CID_DUTY_CYCLE,
			[]byte{0x04},
			true,
		},
		{
			"RxParamSetupReq",
			&MACCommand_RxParamSetupReq{
				Rx1DataRateOffset: 5,
				Rx2DataRateIndex:  12,
				Rx2Frequency:      1677702600,
			},
			[]byte{0x05, 0x5c, 0x42, 0xff, 0xff},
			false,
		},
		{
			"RxParamSetupAns",
			&MACCommand_RxParamSetupAns{
				Rx2FrequencyAck:      true,
				Rx2DataRateIndexAck:  false,
				Rx1DataRateOffsetAck: true,
			},
			[]byte{0x05, 0x05},
			true,
		},
		{
			"DevStatusReq",
			CID_DEV_STATUS,
			[]byte{0x06},
			false,
		},
		{
			"DevStatusAns",
			&MACCommand_DevStatusAns{
				Battery: 0x42,
				Margin:  -16,
			},
			[]byte{0x06, 0x42, 0x2f},
			true,
		},
		{
			"NewChannelReq",
			&MACCommand_NewChannelReq{
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
			&MACCommand_NewChannelReq{
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
			&MACCommand_NewChannelAns{
				FrequencyAck: false,
				DataRateAck:  true,
			},
			[]byte{0x07, 0x2},
			true,
		},
		{
			"RxTimingSetupReq",
			&MACCommand_RxTimingSetupReq{
				Delay: 0xf,
			},
			[]byte{0x08, 0xf},
			false,
		},
		{
			"RxTimingSetupAns",
			CID_RX_TIMING_SETUP,
			[]byte{0x08},
			true,
		},
		{
			"TxParamSetupReq",
			&MACCommand_TxParamSetupReq{
				MaxEIRPIndex:      DEVICE_EIRP_36,
				UplinkDwellTime:   false,
				DownlinkDwellTime: true,
			},
			[]byte{0x09, 0x2f},
			false,
		},
		{
			"TxParamSetupAns",
			CID_TX_PARAM_SETUP,
			[]byte{0x09},
			true,
		},
		{
			"DLChannelReq",
			&MACCommand_DLChannelReq{
				ChannelIndex: 0x4,
				Frequency:    0x42ffff * 100,
			},
			[]byte{0x0A, 0x4, 0xff, 0xff, 0x42},
			false,
		},
		{
			"DLChannelAns",
			&MACCommand_DLChannelAns{
				ChannelIndexAck: false,
				FrequencyAck:    true,
			},
			[]byte{0x0A, 0x2},
			true,
		},
		{
			"RekeyInd",
			&MACCommand_RekeyInd{MinorVersion: 1},
			[]byte{0x0B, 1},
			true,
		},
		{
			"RekeyConf",
			&MACCommand_RekeyConf{MinorVersion: 1},
			[]byte{0x0B, 1},
			false,
		},
		{
			"ADRParamSetupReq",
			&MACCommand_ADRParamSetupReq{
				ADRAckDelayExponent: ADR_ACK_DELAY_4,
				ADRAckLimitExponent: ADR_ACK_LIMIT_16,
			},
			[]byte{0x0C, 0x42},
			false,
		},
		{
			"ADRParamSetupAns",
			CID_ADR_PARAM_SETUP,
			[]byte{0x0C},
			true,
		},
		{
			"DeviceTimeReq",
			CID_DEVICE_TIME,
			[]byte{0x0D},
			true,
		},
		{
			"DeviceTimeAns",
			&MACCommand_DeviceTimeAns{
				Time: gpstime.Parse(0x42ffffff).Add(0x42 * time.Duration(math.Pow(0.5, 8)*float64(time.Second))).UTC(),
			},
			[]byte{0x0D, 0xff, 0xff, 0xff, 0x42, 0x42},
			false,
		},
		{
			"ForceRejoinReq",
			&MACCommand_ForceRejoinReq{
				MaxRetries:     0x7,
				PeriodExponent: 0x7,
				DataRateIndex:  0xd,
				RejoinType:     0x7,
			},
			[]byte{0x0E, 0x3f, 0x7d},
			false,
		},
		{
			"RejoinParamSetupReq",
			&MACCommand_RejoinParamSetupReq{
				MaxTimeExponent:  0x4,
				MaxCountExponent: 0x2,
			},
			[]byte{0x0F, 0x42},
			false,
		},
		{
			"RejoinParamSetupAns",
			&MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: true,
			},
			[]byte{0x0F, 0x1},
			true,
		},
		{
			"PingSlotInfoReq",
			&MACCommand_PingSlotInfoReq{
				Period: PingSlotPeriod(0x7),
			},
			[]byte{0x10, 0x7},
			true,
		},
		{
			"PingSlotInfoAns",
			CID_PING_SLOT_INFO,
			[]byte{0x10},
			false,
		},
		{
			"PingSlotChannelReq",
			&MACCommand_PingSlotChannelReq{
				Frequency:     0x42ffff,
				DataRateIndex: 0xf,
			},
			[]byte{0x11, 0xff, 0xff, 0x42, 0xf},
			false,
		},
		{
			"PingSlotChannelAns",
			&MACCommand_PingSlotChannelAns{
				DataRateIndexAck: true,
				FrequencyAck:     false,
			},
			[]byte{0x11, 0x2},
			true,
		},
		{
			"BeaconTimingReq",
			CID_BEACON_TIMING,
			[]byte{0x12},
			true,
		},
		{
			"BeaconTimingAns",
			&MACCommand_BeaconTimingAns{
				Delay:        0x42ff,
				ChannelIndex: 0x42,
			},
			[]byte{0x12, 0xff, 0x42, 0x42},
			false,
		},
		{
			"BeaconFreqReq",
			&MACCommand_BeaconFreqReq{
				Frequency: 0x42ffff,
			},
			[]byte{0x13, 0xff, 0xff, 0x42},
			false,
		},
		{
			"BeaconFreqAns",
			&MACCommand_BeaconFreqAns{
				FrequencyAck: true,
			},
			[]byte{0x13, 0x01},
			true,
		},
		{
			"DeviceModeInd",
			&MACCommand_DeviceModeInd{
				Class: CLASS_A,
			},
			[]byte{0x20, 0x00},
			true,
		},
		{
			"DeviceModeConf",
			&MACCommand_DeviceModeConf{
				Class: CLASS_C,
			},
			[]byte{0x20, 0x02},
			false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			cmd := tc.Payload.MACCommand()
			if !a.So(cmd, should.NotBeNil) {
				t.FailNow()
			}

			b, err := cmd.MarshalLoRaWAN()
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, tc.Bytes)
			}

			b, err = cmd.AppendLoRaWAN([]byte{})
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, tc.Bytes)
			}

			cmd = &MACCommand{}
			err = cmd.UnmarshalLoRaWAN(tc.Bytes, tc.IsUplink)
			if a.So(err, should.BeNil) {
				if pld := cmd.GetDeviceTimeAns(); pld != nil {
					pld.Time = pld.Time.UTC()
				}
				a.So(cmd, should.Resemble, tc.Payload.MACCommand())
			}

			cmd = &MACCommand{}
			err = ReadMACCommand(bytes.NewReader(tc.Bytes), tc.IsUplink, cmd)
			if pld := cmd.GetDeviceTimeAns(); pld != nil {
				pld.Time = pld.Time.UTC()
			}
			if a.So(err, should.BeNil) {
				a.So(cmd, should.Resemble, tc.Payload.MACCommand())
			}
		})
	}
}
