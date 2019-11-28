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

package lorawan

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gpstime"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// fractStep defines (1/2)^8 second step used in DeviceTimeAns payload.
const fractStep = 3906250 * time.Nanosecond

// maxGPSTime defines the maximum time allowed in the DeviceTime MAC command.
const maxGPSTime int64 = 1<<32 - 1

// MACCommandDescriptor descibes a MAC command.
type MACCommandDescriptor struct {
	InitiatedByDevice bool
	ExpectAnswer      bool
	// UplinkLength is length of uplink payload.
	UplinkLength uint16
	// DownlinkLength is length of downlink payload.
	DownlinkLength uint16
	// AppendUplink appends uplink payload of cmd to b.
	AppendUplink func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error)
	// UnmarshalUplink unmarshals uplink payload b into cmd.
	UnmarshalUplink func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error
	// AppendDownlink appends uplink payload of cmd to b.
	AppendDownlink func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error)
	// UnmarshalDownlink unmarshals downlink payload b into cmd.
	UnmarshalDownlink func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error
}

// MACCommandSpec maps the ttnpb.CID of MACCommand to a *MACCommandDescriptor.
type MACCommandSpec map[ttnpb.MACCommandIdentifier]*MACCommandDescriptor

func newMACUnmarshaler(cid ttnpb.MACCommandIdentifier, name string, n uint8, f func(band.Band, []byte, *ttnpb.MACCommand) error) func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
	return func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
		if len(b) != int(n) {
			return errExpectedLengthEqual(name, int(n))(len(b))
		}
		cmd.CID = cid
		if f == nil {
			return nil
		}
		return f(phy, b, cmd)
	}
}

// DefaultMACCommands contains all the default MAC commands.
var DefaultMACCommands = MACCommandSpec{
	ttnpb.CID_RESET: &MACCommandDescriptor{
		InitiatedByDevice: true,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetResetInd()
			if pld.MinorVersion > 15 {
				return nil, errExpectedLowerOrEqual("MinorVersion", 15)(pld.MinorVersion)
			}
			b = append(b, byte(pld.MinorVersion))
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_RESET, "ResetInd", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_ResetInd_{
				ResetInd: &ttnpb.MACCommand_ResetInd{
					MinorVersion: ttnpb.Minor(b[0] & 0xf),
				},
			}
			return nil
		}),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetResetConf()
			if pld.MinorVersion > 15 {
				return nil, errExpectedLowerOrEqual("MinorVersion", 15)(pld.MinorVersion)
			}
			b = append(b, byte(pld.MinorVersion))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_RESET, "ResetConf", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_ResetConf_{
				ResetConf: &ttnpb.MACCommand_ResetConf{
					MinorVersion: ttnpb.Minor(b[0] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_LINK_CHECK: &MACCommandDescriptor{
		InitiatedByDevice: true,
		ExpectAnswer:      true,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_LINK_CHECK, "LinkCheckReq", 0, nil),

		DownlinkLength: 2,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetLinkCheckAns()
			if pld.Margin > 254 {
				return nil, errExpectedLowerOrEqual("Margin", 254)(pld.Margin)
			}
			if pld.GatewayCount > 255 {
				return nil, errExpectedLowerOrEqual("GatewayCount", 255)(pld.GatewayCount)
			}
			b = append(b, byte(pld.Margin), byte(pld.GatewayCount))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_LINK_CHECK, "LinkCheckAns", 2, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_LinkCheckAns_{
				LinkCheckAns: &ttnpb.MACCommand_LinkCheckAns{
					Margin:       uint32(b[0]),
					GatewayCount: uint32(b[1]),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_LINK_ADR: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetLinkADRAns()
			var status byte
			if pld.ChannelMaskAck {
				status |= 1
			}
			if pld.DataRateIndexAck {
				status |= (1 << 1)
			}
			if pld.TxPowerIndexAck {
				status |= (1 << 2)
			}
			b = append(b, status)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_LINK_ADR, "LinkADRAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_LinkADRAns_{
				LinkADRAns: &ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   b[0]&1 == 1,
					DataRateIndexAck: (b[0]>>1)&1 == 1,
					TxPowerIndexAck:  (b[0]>>2)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 4,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetLinkADRReq()
			if pld.DataRateIndex > 15 {
				return nil, errExpectedLowerOrEqual("DataRateIndex", 15)(pld.DataRateIndex)
			}
			if pld.TxPowerIndex > 15 {
				return nil, errExpectedLowerOrEqual("TxPowerIndex", 15)(pld.TxPowerIndex)
			}
			if len(pld.ChannelMask) > 16 {
				return nil, errExpectedLowerOrEqual("length of ChannelMask", "16 bits")(len(pld.ChannelMask))
			}
			if pld.ChannelMaskControl > 7 {
				return nil, errExpectedLowerOrEqual("ChannelMaskControl", 7)(pld.ChannelMaskControl)
			}
			if pld.NbTrans > 15 {
				return nil, errExpectedLowerOrEqual("NbTrans", 15)(pld.NbTrans)
			}
			b = append(b, byte((pld.DataRateIndex&0xf)<<4)^byte(pld.TxPowerIndex&0xf))
			chMask := make([]byte, 2)
			for i := uint8(0); i < 16 && i < uint8(len(pld.ChannelMask)); i++ {
				chMask[i/8] = chMask[i/8] ^ boolToByte(pld.ChannelMask[i])<<(i%8)
			}
			b = append(b, chMask...)
			b = append(b, byte((pld.ChannelMaskControl&0x7)<<4)^byte(pld.NbTrans&0xf))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_LINK_ADR, "LinkADRReq", 4, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			var chMask [16]bool
			for i := uint8(0); i < 16; i++ {
				if (b[1+i/8]>>(i%8))&1 == 1 {
					chMask[i] = true
				}
			}
			cmd.Payload = &ttnpb.MACCommand_LinkADRReq_{
				LinkADRReq: &ttnpb.MACCommand_LinkADRReq{
					DataRateIndex:      ttnpb.DataRateIndex(b[0] >> 4),
					TxPowerIndex:       uint32(b[0] & 0xf),
					ChannelMask:        chMask[:],
					ChannelMaskControl: uint32((b[3] >> 4) & 0x7),
					NbTrans:            uint32(b[3] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_DUTY_CYCLE: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_DUTY_CYCLE, "DutyCycleAns", 0, nil),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDutyCycleReq()
			if pld.MaxDutyCycle > 15 {
				return nil, errExpectedLowerOrEqual("MaxDutyCycle", 15)(pld.MaxDutyCycle)
			}
			b = append(b, byte(pld.MaxDutyCycle))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_DUTY_CYCLE, "DutyCycleReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DutyCycleReq_{
				DutyCycleReq: &ttnpb.MACCommand_DutyCycleReq{
					MaxDutyCycle: ttnpb.AggregatedDutyCycle(b[0] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_RX_PARAM_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRxParamSetupAns()
			var v byte
			if pld.Rx2FrequencyAck {
				v |= 1
			}
			if pld.Rx2DataRateIndexAck {
				v |= (1 << 1)
			}
			if pld.Rx1DataRateOffsetAck {
				v |= (1 << 2)
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_RX_PARAM_SETUP, "RxParamSetupAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RxParamSetupAns_{
				RxParamSetupAns: &ttnpb.MACCommand_RxParamSetupAns{
					Rx2FrequencyAck:      b[0]&1 == 1,
					Rx2DataRateIndexAck:  (b[0]>>1)&1 == 1,
					Rx1DataRateOffsetAck: (b[0]>>2)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 4,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRxParamSetupReq()
			if pld.Rx1DataRateOffset > 7 {
				return nil, errExpectedLowerOrEqual("Rx1DROffset", 7)(pld.Rx1DataRateOffset)
			}
			if pld.Rx2DataRateIndex > 15 {
				return nil, errExpectedLowerOrEqual("Rx2DR", 15)(pld.Rx2DataRateIndex)
			}
			b = append(b, byte(pld.Rx2DataRateIndex)|byte(pld.Rx1DataRateOffset<<4))
			if pld.Rx2Frequency < 100000 || pld.Rx2Frequency > maxUint24*phy.FreqMultiplier {
				return nil, errExpectedBetween("Rx2Frequency", 100000, maxUint24*phy.FreqMultiplier)(pld.Rx2Frequency)
			}
			b = appendUint64(b, pld.Rx2Frequency/phy.FreqMultiplier, 3)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_RX_PARAM_SETUP, "RxParamSetupReq", 4, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RxParamSetupReq_{
				RxParamSetupReq: &ttnpb.MACCommand_RxParamSetupReq{
					Rx1DataRateOffset: uint32((b[0] >> 4) & 0x7),
					Rx2DataRateIndex:  ttnpb.DataRateIndex(b[0] & 0xf),
					Rx2Frequency:      parseUint64(b[1:4]) * phy.FreqMultiplier,
				},
			}
			return nil
		}),
	},

	ttnpb.CID_DEV_STATUS: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 2,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDevStatusAns()
			if pld.Battery > math.MaxUint8 {
				return nil, errExpectedLowerOrEqual("Battery", math.MaxUint8)(math.MaxUint8)
			}
			if pld.Margin < -32 || pld.Margin > 31 {
				return nil, errExpectedBetween("Margin", -32, 31)(pld.Margin)
			}
			b = append(b, byte(pld.Battery))
			if pld.Margin < 0 {
				b = append(b, byte(-(pld.Margin+1)|(1<<5)))
			} else {
				b = append(b, byte(pld.Margin))
			}
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_DEV_STATUS, "DevStatusAns", 2, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			margin := int32(b[1] & 0x1f)
			if (b[1]>>5)&1 == 1 {
				margin = -margin - 1
			}
			cmd.Payload = &ttnpb.MACCommand_DevStatusAns_{
				DevStatusAns: &ttnpb.MACCommand_DevStatusAns{
					Battery: uint32(b[0]),
					Margin:  margin,
				},
			}
			return nil
		}),

		DownlinkLength: 0,
		AppendDownlink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_DEV_STATUS, "DevStatusReq", 0, nil),
	},

	ttnpb.CID_NEW_CHANNEL: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetNewChannelAns()
			var v byte
			if pld.FrequencyAck {
				v |= 1
			}
			if pld.DataRateAck {
				v |= (1 << 1)
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_NEW_CHANNEL, "NewChannelAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_NewChannelAns_{
				NewChannelAns: &ttnpb.MACCommand_NewChannelAns{
					FrequencyAck: b[0]&1 == 1,
					DataRateAck:  (b[0]>>1)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 5,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetNewChannelReq()
			if pld.ChannelIndex > math.MaxUint8 {
				return nil, errExpectedLowerOrEqual("ChannelIndex", math.MaxUint8)(pld.ChannelIndex)
			}
			b = append(b, byte(pld.ChannelIndex))

			if pld.Frequency > maxUint24*phy.FreqMultiplier {
				return nil, errExpectedLowerOrEqual("Frequency", maxUint24*phy.FreqMultiplier)(pld.Frequency)
			}
			b = appendUint64(b, pld.Frequency/phy.FreqMultiplier, 3)

			if pld.MinDataRateIndex > 15 {
				return nil, errExpectedLowerOrEqual("MinDataRateIndex", 15)(pld.MinDataRateIndex)
			}
			v := byte(pld.MinDataRateIndex)

			if pld.MaxDataRateIndex > 15 {
				return nil, errExpectedLowerOrEqual("MaxDataRateIndex", 15)(pld.MaxDataRateIndex)
			}
			v |= byte(pld.MaxDataRateIndex) << 4
			b = append(b, v)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_NEW_CHANNEL, "NewChannelReq", 5, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_NewChannelReq_{
				NewChannelReq: &ttnpb.MACCommand_NewChannelReq{
					ChannelIndex:     uint32(b[0]),
					Frequency:        parseUint64(b[1:4]) * phy.FreqMultiplier,
					MinDataRateIndex: ttnpb.DataRateIndex(b[4] & 0xf),
					MaxDataRateIndex: ttnpb.DataRateIndex(b[4] >> 4),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_RX_TIMING_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_RX_TIMING_SETUP, "RxTimingSetupAns", 0, nil),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRxTimingSetupReq()
			if pld.Delay > 15 {
				return nil, errExpectedLowerOrEqual("Delay", 15)(pld.Delay)
			}
			b = append(b, byte(pld.Delay))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_RX_TIMING_SETUP, "RxTimingSetupReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RxTimingSetupReq_{
				RxTimingSetupReq: &ttnpb.MACCommand_RxTimingSetupReq{
					Delay: ttnpb.RxDelay(b[0] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_TX_PARAM_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_TX_PARAM_SETUP, "TxParamSetupAns", 0, nil),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetTxParamSetupReq()

			v := byte(pld.MaxEIRPIndex)
			if pld.UplinkDwellTime {
				v |= (1 << 4)
			}
			if pld.DownlinkDwellTime {
				v |= (1 << 5)
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_TX_PARAM_SETUP, "TxParamSetupReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_TxParamSetupReq_{
				TxParamSetupReq: &ttnpb.MACCommand_TxParamSetupReq{
					MaxEIRPIndex:      ttnpb.DeviceEIRP(b[0] & 0xf),
					UplinkDwellTime:   (b[0]>>4)&1 == 1,
					DownlinkDwellTime: (b[0]>>5)&1 == 1,
				},
			}
			return nil
		}),
	},

	ttnpb.CID_DL_CHANNEL: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDLChannelAns()
			var v byte
			if pld.ChannelIndexAck {
				v |= 1
			}
			if pld.FrequencyAck {
				v |= (1 << 1)
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_DL_CHANNEL, "DLChannelAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DLChannelAns_{
				DLChannelAns: &ttnpb.MACCommand_DLChannelAns{
					ChannelIndexAck: b[0]&1 == 1,
					FrequencyAck:    (b[0]>>1)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 4,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDLChannelReq()
			if pld.ChannelIndex > math.MaxUint8 {
				return nil, errExpectedLowerOrEqual("ChannelIndex", math.MaxUint8)(pld.ChannelIndex)
			}
			b = append(b, byte(pld.ChannelIndex))

			if pld.Frequency < 100000 || pld.Frequency > maxUint24*phy.FreqMultiplier {
				return nil, errExpectedBetween("Frequency", 100000, maxUint24*phy.FreqMultiplier)(pld.Frequency)
			}
			b = appendUint64(b, pld.Frequency/phy.FreqMultiplier, 3)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_DL_CHANNEL, "DLChannelReq", 4, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DLChannelReq_{
				DLChannelReq: &ttnpb.MACCommand_DLChannelReq{
					ChannelIndex: uint32(b[0]),
					Frequency:    parseUint64(b[1:4]) * phy.FreqMultiplier,
				},
			}
			return nil
		}),
	},

	ttnpb.CID_REKEY: &MACCommandDescriptor{
		InitiatedByDevice: true,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRekeyInd()
			if pld.MinorVersion > 15 {
				return nil, errExpectedLowerOrEqual("MinorVersion", 15)(pld.MinorVersion)
			}
			b = append(b, byte(pld.MinorVersion))
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_REKEY, "RekeyInd", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RekeyInd_{
				RekeyInd: &ttnpb.MACCommand_RekeyInd{
					MinorVersion: ttnpb.Minor(b[0] & 0xf),
				},
			}
			return nil
		}),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRekeyConf()
			if pld.MinorVersion > 15 {
				return nil, errExpectedLowerOrEqual("MinorVersion", 15)(pld.MinorVersion)
			}
			b = append(b, byte(pld.MinorVersion))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_REKEY, "RekeyConf", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RekeyConf_{
				RekeyConf: &ttnpb.MACCommand_RekeyConf{
					MinorVersion: ttnpb.Minor(b[0] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_ADR_PARAM_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_ADR_PARAM_SETUP, "ADRParamSetupAns", 0, nil),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetADRParamSetupReq()
			if 1 > pld.ADRAckDelayExponent || pld.ADRAckDelayExponent > 32768 {
				return nil, errExpectedBetween("ADRAckDelay", 1, 32768)(pld.ADRAckDelayExponent)
			}
			v := byte(pld.ADRAckDelayExponent)

			if 1 > pld.ADRAckLimitExponent || pld.ADRAckLimitExponent > 32768 {
				return nil, errExpectedBetween("ADRAckLimit", 1, 32768)(pld.ADRAckLimitExponent)
			}
			v |= byte(pld.ADRAckLimitExponent) << 4

			b = append(b, v)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_ADR_PARAM_SETUP, "ADRParamSetupReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_ADRParamSetupReq_{
				ADRParamSetupReq: &ttnpb.MACCommand_ADRParamSetupReq{
					ADRAckDelayExponent: ttnpb.ADRAckDelayExponent(b[0] & 0xf),
					ADRAckLimitExponent: ttnpb.ADRAckLimitExponent(b[0] >> 4),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_DEVICE_TIME: &MACCommandDescriptor{
		InitiatedByDevice: true,
		ExpectAnswer:      true,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_DEVICE_TIME, "DeviceTimeReq", 0, nil),

		DownlinkLength: 5,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDeviceTimeAns()

			sec := gpstime.ToGPS(pld.Time)
			if sec > maxGPSTime {
				return nil, errExpectedLowerOrEqual("Time", maxGPSTime)(sec)
			}
			b = appendUint32(b, uint32(sec), 4)
			b = append(b, byte(time.Duration(pld.Time.Nanosecond())/fractStep))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_DEVICE_TIME, "DeviceTimeAns", 5, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DeviceTimeAns_{
				DeviceTimeAns: &ttnpb.MACCommand_DeviceTimeAns{
					Time: gpstime.Parse(int64(parseUint32(b[0:4]))).Add(time.Duration(b[4]) * fractStep),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_FORCE_REJOIN: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      false,

		DownlinkLength: 2,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetForceRejoinReq()

			if pld.PeriodExponent > 7 {
				return nil, errExpectedLowerOrEqual("PeriodExponent", 7)(pld.PeriodExponent)
			}
			// First byte
			v := byte(pld.PeriodExponent) << 3

			if pld.MaxRetries > 7 {
				return nil, errExpectedLowerOrEqual("MaxRetries", 7)(pld.MaxRetries)
			}
			v |= byte(pld.MaxRetries)
			b = append(b, v)

			if pld.RejoinType > 2 {
				return nil, errExpectedLowerOrEqual("RejoinType", 2)(pld.RejoinType)
			}
			// Second byte
			v = byte(pld.RejoinType) << 4

			if pld.DataRateIndex > 15 {
				return nil, errExpectedLowerOrEqual("DataRateIndex", 15)(pld.DataRateIndex)
			}
			v |= byte(pld.DataRateIndex)
			b = append(b, v)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_FORCE_REJOIN, "ForceRejoinReq", 2, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_ForceRejoinReq_{
				ForceRejoinReq: &ttnpb.MACCommand_ForceRejoinReq{
					PeriodExponent: ttnpb.RejoinPeriodExponent(uint32(b[0] >> 3)),
					MaxRetries:     uint32(b[0] & 0x7),
					RejoinType:     ttnpb.RejoinType(b[1] >> 4),
					DataRateIndex:  ttnpb.DataRateIndex(b[1] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_REJOIN_PARAM_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRejoinParamSetupAns()

			var v byte
			if pld.MaxTimeExponentAck {
				v |= 1
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_REJOIN_PARAM_SETUP, "RejoinParamSetupAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RejoinParamSetupAns_{
				RejoinParamSetupAns: &ttnpb.MACCommand_RejoinParamSetupAns{
					MaxTimeExponentAck: b[0]&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRejoinParamSetupReq()

			if pld.MaxTimeExponent > 15 {
				return nil, errExpectedLowerOrEqual("MaxTimeExponent", 15)(pld.MaxTimeExponent)
			}
			v := byte(pld.MaxTimeExponent) << 4

			if pld.MaxCountExponent > 15 {
				return nil, errExpectedLowerOrEqual("MaxCountExponent", 15)(pld.MaxCountExponent)
			}
			v |= byte(pld.MaxCountExponent)
			b = append(b, v)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_REJOIN_PARAM_SETUP, "RejoinParamSetupReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RejoinParamSetupReq_{
				RejoinParamSetupReq: &ttnpb.MACCommand_RejoinParamSetupReq{
					MaxTimeExponent:  ttnpb.RejoinTimeExponent(uint32(b[0] >> 4)),
					MaxCountExponent: ttnpb.RejoinCountExponent(uint32(b[0] & 0xf)),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_PING_SLOT_INFO: &MACCommandDescriptor{
		InitiatedByDevice: true,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetPingSlotInfoReq()

			if pld.Period > 7 {
				return nil, errExpectedLowerOrEqual("Period", 15)(pld.Period)
			}
			b = append(b, byte(pld.Period))
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_PING_SLOT_INFO, "PingSlotInfoReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_PingSlotInfoReq_{
				PingSlotInfoReq: &ttnpb.MACCommand_PingSlotInfoReq{
					Period: ttnpb.PingSlotPeriod(b[0] & 0x7),
				},
			}
			return nil
		}),

		DownlinkLength: 0,
		AppendDownlink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_PING_SLOT_INFO, "PingSlotInfoAns", 0, nil),
	},

	ttnpb.CID_PING_SLOT_CHANNEL: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetPingSlotChannelAns()

			var v byte
			if pld.FrequencyAck {
				v |= 1
			}
			if pld.DataRateIndexAck {
				v |= (1 << 1)
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_PING_SLOT_CHANNEL, "PingSlotChannelAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_PingSlotChannelAns_{
				PingSlotChannelAns: &ttnpb.MACCommand_PingSlotChannelAns{
					FrequencyAck:     b[0]&1 == 1,
					DataRateIndexAck: (b[0]>>1)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 4,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetPingSlotChannelReq()

			if pld.Frequency < 100000 || pld.Frequency > maxUint24*phy.FreqMultiplier {
				return nil, errExpectedBetween("Frequency", 100000, maxUint24*phy.FreqMultiplier)(pld.Frequency)
			}
			b = appendUint64(b, pld.Frequency/phy.FreqMultiplier, 3)

			if pld.DataRateIndex > 15 {
				return nil, errExpectedLowerOrEqual("DataRateIndex", 15)(pld.DataRateIndex)
			}
			b = append(b, byte(pld.DataRateIndex))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_PING_SLOT_CHANNEL, "PingSlotChannelReq", 4, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_PingSlotChannelReq_{
				PingSlotChannelReq: &ttnpb.MACCommand_PingSlotChannelReq{
					Frequency:     parseUint64(b[0:3]) * phy.FreqMultiplier,
					DataRateIndex: ttnpb.DataRateIndex(b[3] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_BEACON_TIMING: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, _ ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_BEACON_TIMING, "BeaconTimingReq", 0, nil),

		DownlinkLength: 3,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetBeaconTimingAns()

			if pld.Delay > math.MaxUint16 {
				return nil, errExpectedLowerOrEqual("Delay", math.MaxUint16)(pld.Delay)
			}
			b = appendUint32(b, pld.Delay, 2)

			if pld.ChannelIndex > math.MaxUint8 {
				return nil, errExpectedLowerOrEqual("ChannelIndex", math.MaxUint8)(pld.ChannelIndex)
			}
			b = append(b, byte(pld.ChannelIndex))

			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_BEACON_TIMING, "BeaconTimingAns", 3, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_BeaconTimingAns_{
				BeaconTimingAns: &ttnpb.MACCommand_BeaconTimingAns{
					Delay:        parseUint32(b[0:2]),
					ChannelIndex: uint32(b[2]),
				},
			}
			return nil
		}),
	},

	ttnpb.CID_BEACON_FREQ: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetBeaconFreqAns()
			var v byte
			if pld.FrequencyAck {
				v |= 1
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_BEACON_FREQ, "BeaconFreqAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_BeaconFreqAns_{
				BeaconFreqAns: &ttnpb.MACCommand_BeaconFreqAns{
					FrequencyAck: b[0]&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 3,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetBeaconFreqReq()
			if pld.Frequency < 100000 || pld.Frequency > maxUint24*phy.FreqMultiplier {
				return nil, errExpectedBetween("Frequency", 100000, maxUint24*phy.FreqMultiplier)(pld.Frequency)
			}
			b = appendUint64(b, pld.Frequency/phy.FreqMultiplier, 3)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_BEACON_FREQ, "BeaconFreqReq", 3, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_BeaconFreqReq_{
				BeaconFreqReq: &ttnpb.MACCommand_BeaconFreqReq{
					Frequency: parseUint64(b[0:3]) * phy.FreqMultiplier,
				},
			}
			return nil
		}),
	},

	ttnpb.CID_DEVICE_MODE: &MACCommandDescriptor{
		InitiatedByDevice: false,
		ExpectAnswer:      true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDeviceModeInd()
			b = append(b, byte(pld.Class))
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.CID_DEVICE_MODE, "DeviceModeInd", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DeviceModeInd_{
				DeviceModeInd: &ttnpb.MACCommand_DeviceModeInd{
					Class: ttnpb.Class(b[0]),
				},
			}
			return nil
		}),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDeviceModeConf()
			b = append(b, byte(pld.Class))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.CID_DEVICE_MODE, "DeviceModeConf", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DeviceModeConf_{
				DeviceModeConf: &ttnpb.MACCommand_DeviceModeConf{
					Class: ttnpb.Class(b[0]),
				},
			}
			return nil
		}),
	},
}

var errDecodingMACCommand = errors.DefineInvalidArgument("decoding_mac_command", "could not decode MAC command with CID `{cid}`")

func (spec MACCommandSpec) read(phy band.Band, r io.Reader, isUplink bool, cmd *ttnpb.MACCommand) error {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return err
	}

	ret := ttnpb.MACCommand{
		CID: ttnpb.MACCommandIdentifier(b[0]),
	}

	desc, ok := spec[ret.CID]
	if !ok || desc == nil {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		ret.Payload = &ttnpb.MACCommand_RawPayload{
			RawPayload: b,
		}
		*cmd = ret
		return nil
	}

	var n uint16
	var unmarshaler func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error
	if isUplink {
		n = desc.UplinkLength
		unmarshaler = desc.UnmarshalUplink
	} else {
		n = desc.DownlinkLength
		unmarshaler = desc.UnmarshalDownlink
	}

	if n == 0 {
		b = nil
	} else {
		b = make([]byte, n)
		_, err = r.Read(b)
		if err != nil {
			return err
		}
	}
	if err := unmarshaler(phy, b, cmd); err != nil {
		return errDecodingMACCommand.WithAttributes("cid", fmt.Sprintf("0x%X", int32(ret.CID))).WithCause(err)
	}
	return nil
}

// ReadUplink reads an uplink MACCommand from r into cmd and returns any errors encountered.
func (spec MACCommandSpec) ReadUplink(phy band.Band, r io.Reader, cmd *ttnpb.MACCommand) error {
	return spec.read(phy, r, true, cmd)
}

// ReadDownlink reads a downlink MACCommand from r into cmd and returns any errors encountered.
func (spec MACCommandSpec) ReadDownlink(phy band.Band, r io.Reader, cmd *ttnpb.MACCommand) error {
	return spec.read(phy, r, false, cmd)
}

var (
	errEncodingMACCommand = errors.DefineInvalidArgument("encoding_mac_command", "could not encode MAC command with CID `{cid}`")
	errUnknownMACCommand  = errors.DefineInvalidArgument("unknown_mac_command", "unknown MAC command CID `{cid}`")
	errMACCommandUplink   = errors.DefineInvalidArgument("mac_command_uplink", "invalid uplink MAC command CID `{cid}`")
	errMACCommandDownlink = errors.DefineInvalidArgument("mac_command_downlink", "invalid downlink MAC command CID `{cid}`")
)

func (spec MACCommandSpec) append(phy band.Band, b []byte, isUplink bool, cmd ttnpb.MACCommand) ([]byte, error) {
	desc, ok := spec[cmd.CID]
	if !ok || desc == nil {
		return nil, errUnknownMACCommand.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cmd.CID)))
	}
	b = append(b, byte(cmd.CID))

	var appender func(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error)
	if isUplink {
		appender = desc.AppendUplink
		if appender == nil {
			return nil, errMACCommandUplink.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cmd.CID)))
		}
	} else {
		appender = desc.AppendDownlink
		if appender == nil {
			return nil, errMACCommandDownlink.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cmd.CID)))
		}
	}

	b, err := appender(phy, b, cmd)
	if err != nil {
		return nil, errEncodingMACCommand.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cmd.CID))).WithCause(err)
	}
	return b, nil
}

// AppendUplink encodes uplink MAC command cmd, appends it to b and returns any errors encountered.
func (spec MACCommandSpec) AppendUplink(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
	return spec.append(phy, b, true, cmd)
}

// AppendDownlink encodes downlink MAC command cmd, appends it to b and returns any errors encountered.
func (spec MACCommandSpec) AppendDownlink(phy band.Band, b []byte, cmd ttnpb.MACCommand) ([]byte, error) {
	return spec.append(phy, b, false, cmd)
}
