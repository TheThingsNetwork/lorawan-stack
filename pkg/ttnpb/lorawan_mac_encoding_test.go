// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/ttnpb"
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
			[]byte{0x03, 0x7},
			true,
		},
		// {
		// 	&MACCommand_DutyCycleReq{},
		// 	&MACCommand_DutyCycleReq{},
		// 	[]byte{0x04},
		// },
		// {
		// 	&MACCommand_RxParamSetupReq{},
		// 	&MACCommand_RxParamSetupReq{},
		// 	[]byte{0x05},
		// },
		// {
		// 	&MACCommand_RxParamSetupAns{},
		// 	&MACCommand_RxParamSetupAns{},
		// 	[]byte{0x05},
		// },
		// {
		// 	&MACCommand_DevStatusAns{},
		// 	&MACCommand_DevStatusAns{},
		// 	[]byte{0x06},
		// },
		// {
		// 	&MACCommand_NewChannelReq{},
		// 	&MACCommand_NewChannelReq{},
		// 	[]byte{0x07},
		// },
		// {
		// 	&MACCommand_NewChannelAns{},
		// 	&MACCommand_NewChannelAns{},
		// 	[]byte{0x07},
		// },
		// {
		// 	&MACCommand_RxTimingSetupReq{},
		// 	&MACCommand_RxTimingSetupReq{},
		// 	[]byte{0x08},
		// },
		// {
		// 	&MACCommand_TxParamSetupReq{},
		// 	&MACCommand_TxParamSetupReq{},
		// 	[]byte{0x09},
		// },
		// {
		// 	&MACCommand_DLChannelReq{},
		// 	&MACCommand_DLChannelReq{},
		// 	[]byte{0x0A},
		// },
		// {
		// 	&MACCommand_DLChannelAns{},
		// 	&MACCommand_DLChannelAns{},
		// 	[]byte{0x0A},
		// },
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
		// {
		// 	&MACCommand_ADRParamSetupReq{},
		// 	&MACCommand_ADRParamSetupReq{},
		// 	[]byte{0x0C},
		// },
		// {
		// 	&MACCommand_DeviceTimeAns{},
		// 	&MACCommand_DeviceTimeAns{},
		// 	[]byte{0x0D},
		// },
		// {
		// 	&MACCommand_ForceRejoinReq{},
		// 	&MACCommand_ForceRejoinReq{},
		// 	[]byte{0x0E},
		// },
		// {
		// 	&MACCommand_RejoinParamSetupReq{},
		// 	&MACCommand_RejoinParamSetupReq{},
		// 	[]byte{0x0F},
		// },
		// {
		// 	&MACCommand_RejoinParamSetupAns{},
		// 	&MACCommand_RejoinParamSetupAns{},
		// 	[]byte{0x0F},
		// },
		// {
		// 	&MACCommand_PingSlotInfoReq{},
		// 	&MACCommand_PingSlotInfoReq{},
		// 	[]byte{0x10},
		// },
		// {
		// 	&MACCommand_PingSlotChannelReq{},
		// 	&MACCommand_PingSlotChannelReq{},
		// 	[]byte{0x11},
		// },
		// {
		// 	&MACCommand_PingSlotChannelAns{},
		// 	&MACCommand_PingSlotChannelAns{},
		// 	[]byte{0x11},
		// },
		// {
		// 	&MACCommand_BeaconTimingAns{},
		// 	&MACCommand_BeaconTimingAns{},
		// 	[]byte{0x12},
		// },
		// {
		// 	&MACCommand_BeaconFreqReq{},
		// 	&MACCommand_BeaconFreqReq{},
		// 	[]byte{0x13},
		// },
		// {
		// 	&MACCommand_BeaconFreqAns{},
		// 	&MACCommand_BeaconFreqAns{},
		// 	[]byte{0x13},
		// },
		// {
		// 	&MACCommand_DeviceModeInd{},
		// 	&MACCommand_DeviceModeInd{},
		// 	[]byte{0x20},
		// },
		// {
		// 	&MACCommand_DeviceModeConf{},
		// 	&MACCommand_DeviceModeConf{},
		// 	[]byte{0x20},
		// },
	} {
		t.Run(strings.TrimPrefix(fmt.Sprintf("%T", tc.Command), "*ttnpb.MACCommand_"), func(t *testing.T) {
			a := assertions.New(t)

			b, err := tc.Command.MarshalLoRaWAN()
			a.So(err, should.BeNil)
			a.So(b, should.NotBeNil)
			a.So(b, should.Resemble, tc.Bytes)

			cmd := tc.Empty
			a.So(cmd.UnmarshalLoRaWAN(b), should.BeNil)
			a.So(cmd, should.Resemble, tc.Command)

			ret, err := tc.Command.AppendLoRaWAN(make([]byte, 0))
			a.So(err, should.BeNil)
			a.So(ret, should.Resemble, tc.Bytes)

			cmds := MACCommands{tc.Command.MACCommand()}
			cmdsb, err := cmds.MarshalLoRaWAN()
			a.So(err, should.BeNil)
			a.So(cmdsb, should.Resemble, tc.Bytes)

			var cmds2 MACCommands
			err = cmds2.UnmarshalLoRaWAN(cmdsb, tc.IsUplink)
			a.So(err, should.BeNil)
			a.So(cmds2, should.Resemble, cmds)
		})
	}
}
