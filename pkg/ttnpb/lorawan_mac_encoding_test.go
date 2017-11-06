// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb_test

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/timeutil"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestLoRaWANEncodingMAC(t *testing.T) {
	for _, tc := range []struct {
		Command interface {
			AppendLoRaWAN(dst []byte) ([]byte, error)
			MarshalLoRaWAN() ([]byte, error)
			MACCommand() *MACCommand
		}
		Empty interface {
			UnmarshalLoRaWAN(b []byte) error
		}
		Bytes    []byte
		IsUplink bool
	}{
		{
			&MACCommand_Proprietary{CID: 0x80, RawPayload: []byte{1, 2, 3, 4}},
			&MACCommand_Proprietary{},
			[]byte{0x80, 1, 2, 3, 4},
			true,
		},
		{
			&MACCommand_ResetInd{MinorVersion: 1},
			&MACCommand_ResetInd{},
			[]byte{0x01, 1},
			true,
		},
		{
			&MACCommand_ResetConf{MinorVersion: 1},
			&MACCommand_ResetConf{},
			[]byte{0x01, 1},
			false,
		},
		{
			&MACCommand_LinkCheckAns{Margin: 20, GatewayCount: 3},
			&MACCommand_LinkCheckAns{},
			[]byte{0x02, 20, 3},
			false,
		},
		{
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
			&MACCommand_LinkADRReq{},
			[]byte{0x03, 0x52, 0x04, 0x02, 0x11},
			false,
		},
		{
			&MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			&MACCommand_LinkADRAns{},
			[]byte{0x03, 0x07},
			true,
		},
		{
			&MACCommand_DutyCycleReq{
				MaxDutyCycle: 0x0d,
			},
			&MACCommand_DutyCycleReq{},
			[]byte{0x04, 0x0d},
			false,
		},
		{
			&MACCommand_RxParamSetupReq{
				Rx1DataRateOffset: 0x5,
				Rx2DataRateIndex:  0xd,
				Rx2Frequency:      0x42ffff * 100,
			},
			&MACCommand_RxParamSetupReq{},
			[]byte{0x05, 0x5d, 0xff, 0xff, 0x42},
			false,
		},
		{
			&MACCommand_RxParamSetupAns{
				Rx2FrequencyAck:      true,
				Rx2DataRateIndexAck:  false,
				Rx1DataRateOffsetAck: true,
			},
			&MACCommand_RxParamSetupAns{},
			[]byte{0x05, 0x05},
			true,
		},
		{
			&MACCommand_DevStatusAns{
				Battery: 0x42,
				Margin:  -16,
			},
			&MACCommand_DevStatusAns{},
			[]byte{0x06, 0x42, 0x2f},
			true,
		},
		{
			&MACCommand_NewChannelReq{
				ChannelIndex:     0xf,
				Frequency:        0x42ffff * 100,
				MaxDataRateIndex: 0x4,
				MinDataRateIndex: 0x2,
			},
			&MACCommand_NewChannelReq{},
			[]byte{0x07, 0xf, 0xff, 0xff, 0x42, 0x42},
			false,
		},
		{
			&MACCommand_NewChannelAns{
				FrequencyAck: false,
				DataRateAck:  true,
			},
			&MACCommand_NewChannelAns{},
			[]byte{0x07, 0x2},
			true,
		},
		{
			&MACCommand_RxTimingSetupReq{
				Delay: 0xf,
			},
			&MACCommand_RxTimingSetupReq{},
			[]byte{0x08, 0xf},
			false,
		},
		{
			&MACCommand_TxParamSetupReq{
				MaxEIRPIndex:      0xf,
				UplinkDwellTime:   false,
				DownlinkDwellTime: true,
			},
			&MACCommand_TxParamSetupReq{},
			[]byte{0x09, 0x2f},
			false,
		},
		{
			&MACCommand_DLChannelReq{
				ChannelIndex: 0x4,
				Frequency:    0x42ffff * 100,
			},
			&MACCommand_DLChannelReq{},
			[]byte{0x0A, 0x4, 0xff, 0xff, 0x42},
			false,
		},
		{
			&MACCommand_DLChannelAns{
				ChannelIndexAck: false,
				FrequencyAck:    true,
			},
			&MACCommand_DLChannelAns{},
			[]byte{0x0A, 0x2},
			true,
		},
		{
			&MACCommand_RekeyInd{MinorVersion: 1},
			&MACCommand_RekeyInd{},
			[]byte{0x0B, 1},
			true,
		},
		{
			&MACCommand_RekeyConf{MinorVersion: 1},
			&MACCommand_RekeyConf{},
			[]byte{0x0B, 1},
			false,
		},
		{
			&MACCommand_ADRParamSetupReq{
				ADRAckDelayExponent: 0x2,
				ADRAckLimitExponent: 0x4,
			},
			&MACCommand_ADRParamSetupReq{},
			[]byte{0x0C, 0x42},
			false,
		},
		{
			&MACCommand_DeviceTimeAns{
				Time: timeutil.GPS(0x42ffffff).Add(0x42 * time.Duration(math.Pow(0.5, 8)*float64(time.Second))).UTC(),
			},
			&MACCommand_DeviceTimeAns{},
			[]byte{0x0D, 0xff, 0xff, 0xff, 0x42, 0x42},
			false,
		},
		{
			&MACCommand_ForceRejoinReq{
				MaxRetries:     0x7,
				PeriodExponent: 0x7,
				DataRateIndex:  0xd,
				RejoinType:     0x7,
			},
			&MACCommand_ForceRejoinReq{},
			[]byte{0x0E, 0x3f, 0x7d},
			false,
		},
		{
			&MACCommand_RejoinParamSetupReq{
				MaxTimeExponent:  0x4,
				MaxCountExponent: 0x2,
			},
			&MACCommand_RejoinParamSetupReq{},
			[]byte{0x0F, 0x42},
			false,
		},
		{
			&MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: true,
			},
			&MACCommand_RejoinParamSetupAns{},
			[]byte{0x0F, 0x1},
			true,
		},
		{
			&MACCommand_PingSlotInfoReq{
				Period: PingSlotPeriod(0x7),
			},
			&MACCommand_PingSlotInfoReq{},
			[]byte{0x10, 0x7},
			true,
		},
		{
			&MACCommand_PingSlotChannelReq{
				Frequency:     0x42ffff,
				DataRateIndex: 0xf,
			},
			&MACCommand_PingSlotChannelReq{},
			[]byte{0x11, 0xff, 0xff, 0x42, 0xf},
			false,
		},
		{
			&MACCommand_PingSlotChannelAns{
				DataRateIndexAck: true,
				FrequencyAck:     false,
			},
			&MACCommand_PingSlotChannelAns{},
			[]byte{0x11, 0x2},
			true,
		},
		{
			&MACCommand_BeaconTimingAns{
				Delay:        0x42ff,
				ChannelIndex: 0x42,
			},
			&MACCommand_BeaconTimingAns{},
			[]byte{0x12, 0xff, 0x42, 0x42},
			false,
		},
		{
			&MACCommand_BeaconFreqReq{
				Frequency: 0x42ffff,
			},
			&MACCommand_BeaconFreqReq{},
			[]byte{0x13, 0xff, 0xff, 0x42},
			false,
		},
		{
			&MACCommand_BeaconFreqAns{
				FrequencyAck: true,
			},
			&MACCommand_BeaconFreqAns{},
			[]byte{0x13, 0x01},
			true,
		},
		{
			&MACCommand_DeviceModeInd{
				Class: CLASS_A,
			},
			&MACCommand_DeviceModeInd{},
			[]byte{0x20, 0x00},
			true,
		},
		{
			&MACCommand_DeviceModeConf{
				Class: CLASS_C,
			},
			&MACCommand_DeviceModeConf{},
			[]byte{0x20, 0x02},
			false,
		},
	} {
		t.Run(strings.TrimPrefix(fmt.Sprintf("%T", tc.Command), "*ttnpb.MACCommand_"), func(t *testing.T) {
			a := assertions.New(t)

			b, err := tc.Command.MarshalLoRaWAN()
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, tc.Bytes)
			}

			cmd := tc.Empty
			if a.So(cmd.UnmarshalLoRaWAN(b), should.BeNil) {
				if cmd, ok := cmd.(*MACCommand_DeviceTimeAns); ok {
					cmd.Time = cmd.Time.UTC()
				}
				a.So(cmd, should.Resemble, tc.Command)
			}

			ret, err := tc.Command.AppendLoRaWAN(make([]byte, 0))
			if a.So(err, should.BeNil) {
				a.So(ret, should.Resemble, tc.Bytes)
			}

			cmds := MACCommands{tc.Command.MACCommand()}
			cmdsb, err := cmds.MarshalLoRaWAN()
			if a.So(err, should.BeNil) {
				a.So(cmdsb, should.Resemble, tc.Bytes)
			}

			var cmds2 MACCommands
			err = cmds2.UnmarshalLoRaWAN(cmdsb, tc.IsUplink)
			for _, cmd := range cmds2 {
				if pld := cmd.GetDeviceTimeAns(); pld != nil {
					pld.Time = pld.Time.UTC()
				}
			}
			if a.So(err, should.BeNil) {
				a.So(cmds2, should.Resemble, cmds)
			}
		})
	}
}
