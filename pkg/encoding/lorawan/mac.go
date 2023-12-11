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
	"math"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/byteutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// fractStep defines (1/2)^8 second step used in DeviceTimeAns payload.
const fractStep = 3906250 * time.Nanosecond

// MACCommandDescriptor descibes a MAC command.
type MACCommandDescriptor struct {
	InitiatedByDevice bool
	// UplinkLength is length of uplink payload.
	UplinkLength uint16
	// DownlinkLength is length of downlink payload.
	DownlinkLength uint16
	// AppendUplink appends uplink payload of cmd to b.
	AppendUplink func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error)
	// UnmarshalUplink unmarshals uplink payload b into cmd.
	UnmarshalUplink func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error
	// AppendDownlink appends uplink payload of cmd to b.
	AppendDownlink func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error)
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
		cmd.Cid = cid
		if f == nil {
			return nil
		}
		return f(phy, b, cmd)
	}
}

// DefaultMACCommands contains all the default MAC commands.
var DefaultMACCommands = MACCommandSpec{
	ttnpb.MACCommandIdentifier_CID_RESET: &MACCommandDescriptor{
		InitiatedByDevice: true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetResetInd()
			if pld.MinorVersion > 15 {
				return nil, errExpectedLowerOrEqual("MinorVersion", 15)(pld.MinorVersion)
			}
			b = append(b, byte(pld.MinorVersion))
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_RESET, "ResetInd", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_ResetInd_{
				ResetInd: &ttnpb.MACCommand_ResetInd{
					MinorVersion: ttnpb.Minor(b[0] & 0xf),
				},
			}
			return nil
		}),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetResetConf()
			if pld.MinorVersion > 15 {
				return nil, errExpectedLowerOrEqual("MinorVersion", 15)(pld.MinorVersion)
			}
			b = append(b, byte(pld.MinorVersion))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_RESET, "ResetConf", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_ResetConf_{
				ResetConf: &ttnpb.MACCommand_ResetConf{
					MinorVersion: ttnpb.Minor(b[0] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_LINK_CHECK: &MACCommandDescriptor{
		InitiatedByDevice: true,

		AppendUplink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_LINK_CHECK, "LinkCheckReq", 0, nil),

		DownlinkLength: 2,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
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
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_LINK_CHECK, "LinkCheckAns", 2, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_LinkCheckAns_{
				LinkCheckAns: &ttnpb.MACCommand_LinkCheckAns{
					Margin:       uint32(b[0]),
					GatewayCount: uint32(b[1]),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_LINK_ADR: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetLinkAdrAns()
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
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_LINK_ADR, "LinkADRAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_LinkAdrAns{
				LinkAdrAns: &ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   b[0]&1 == 1,
					DataRateIndexAck: (b[0]>>1)&1 == 1,
					TxPowerIndexAck:  (b[0]>>2)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 4,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetLinkAdrReq()
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
			for i, v := range pld.ChannelMask {
				chMask[i/8] = chMask[i/8] ^ boolToByte(v)<<(i%8)
			}
			b = append(b, chMask...)
			b = append(b, byte((pld.ChannelMaskControl&0x7)<<4)^byte(pld.NbTrans&0xf))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_LINK_ADR, "LinkADRReq", 4, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			var chMask [16]bool
			for i := 0; i < 16; i++ {
				chMask[i] = (b[1+i/8]>>(i%8))&1 == 1
			}
			cmd.Payload = &ttnpb.MACCommand_LinkAdrReq{
				LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
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

	ttnpb.MACCommandIdentifier_CID_DUTY_CYCLE: &MACCommandDescriptor{
		InitiatedByDevice: false,

		AppendUplink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DUTY_CYCLE, "DutyCycleAns", 0, nil),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDutyCycleReq()
			if pld.MaxDutyCycle > 15 {
				return nil, errExpectedLowerOrEqual("MaxDutyCycle", 15)(pld.MaxDutyCycle)
			}
			b = append(b, byte(pld.MaxDutyCycle))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DUTY_CYCLE, "DutyCycleReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DutyCycleReq_{
				DutyCycleReq: &ttnpb.MACCommand_DutyCycleReq{
					MaxDutyCycle: ttnpb.AggregatedDutyCycle(b[0] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_RX_PARAM_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
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
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_RX_PARAM_SETUP, "RxParamSetupAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
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
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRxParamSetupReq()
			if pld.Rx1DataRateOffset > 7 {
				return nil, errExpectedLowerOrEqual("Rx1DROffset", 7)(pld.Rx1DataRateOffset)
			}
			if pld.Rx2DataRateIndex > 15 {
				return nil, errExpectedLowerOrEqual("Rx2DR", 15)(pld.Rx2DataRateIndex)
			}
			b = append(b, byte(pld.Rx2DataRateIndex)|byte(pld.Rx1DataRateOffset<<4))
			if pld.Rx2Frequency < 100000 || pld.Rx2Frequency > byteutil.MaxUint24*phy.FreqMultiplier {
				return nil, errExpectedBetween("Rx2Frequency", 100000, byteutil.MaxUint24*phy.FreqMultiplier)(pld.Rx2Frequency)
			}
			b = byteutil.AppendUint64(b, pld.Rx2Frequency/phy.FreqMultiplier, 3)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_RX_PARAM_SETUP, "RxParamSetupReq", 4, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RxParamSetupReq_{
				RxParamSetupReq: &ttnpb.MACCommand_RxParamSetupReq{
					Rx1DataRateOffset: ttnpb.DataRateOffset(uint32((b[0] >> 4) & 0x7)),
					Rx2DataRateIndex:  ttnpb.DataRateIndex(b[0] & 0xf),
					Rx2Frequency:      byteutil.ParseUint64(b[1:4]) * phy.FreqMultiplier,
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_DEV_STATUS: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 2,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
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
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DEV_STATUS, "DevStatusAns", 2, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
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

		AppendDownlink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DEV_STATUS, "DevStatusReq", 0, nil),
	},

	ttnpb.MACCommandIdentifier_CID_NEW_CHANNEL: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
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
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_NEW_CHANNEL, "NewChannelAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_NewChannelAns_{
				NewChannelAns: &ttnpb.MACCommand_NewChannelAns{
					FrequencyAck: b[0]&1 == 1,
					DataRateAck:  (b[0]>>1)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 5,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetNewChannelReq()
			if pld.ChannelIndex > math.MaxUint8 {
				return nil, errExpectedLowerOrEqual("ChannelIndex", math.MaxUint8)(pld.ChannelIndex)
			}
			b = append(b, byte(pld.ChannelIndex))

			if pld.Frequency > byteutil.MaxUint24*phy.FreqMultiplier {
				return nil, errExpectedLowerOrEqual("Frequency", byteutil.MaxUint24*phy.FreqMultiplier)(pld.Frequency)
			}
			b = byteutil.AppendUint64(b, pld.Frequency/phy.FreqMultiplier, 3)

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
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_NEW_CHANNEL, "NewChannelReq", 5, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_NewChannelReq_{
				NewChannelReq: &ttnpb.MACCommand_NewChannelReq{
					ChannelIndex:     uint32(b[0]),
					Frequency:        byteutil.ParseUint64(b[1:4]) * phy.FreqMultiplier,
					MinDataRateIndex: ttnpb.DataRateIndex(b[4] & 0xf),
					MaxDataRateIndex: ttnpb.DataRateIndex(b[4] >> 4),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,

		AppendUplink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP, "RxTimingSetupAns", 0, nil),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRxTimingSetupReq()
			if pld.Delay > 15 {
				return nil, errExpectedLowerOrEqual("Delay", 15)(pld.Delay)
			}
			b = append(b, byte(pld.Delay))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP, "RxTimingSetupReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RxTimingSetupReq_{
				RxTimingSetupReq: &ttnpb.MACCommand_RxTimingSetupReq{
					Delay: ttnpb.RxDelay(b[0] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,

		AppendUplink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP, "TxParamSetupAns", 0, nil),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetTxParamSetupReq()

			v := byte(pld.MaxEirpIndex)
			if pld.UplinkDwellTime {
				v |= (1 << 4)
			}
			if pld.DownlinkDwellTime {
				v |= (1 << 5)
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP, "TxParamSetupReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_TxParamSetupReq_{
				TxParamSetupReq: &ttnpb.MACCommand_TxParamSetupReq{
					MaxEirpIndex:      ttnpb.DeviceEIRP(b[0] & 0xf),
					UplinkDwellTime:   (b[0]>>4)&1 == 1,
					DownlinkDwellTime: (b[0]>>5)&1 == 1,
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_DL_CHANNEL: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDlChannelAns()
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
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DL_CHANNEL, "DLChannelAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DlChannelAns{
				DlChannelAns: &ttnpb.MACCommand_DLChannelAns{
					ChannelIndexAck: b[0]&1 == 1,
					FrequencyAck:    (b[0]>>1)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 4,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDlChannelReq()
			if pld.ChannelIndex > math.MaxUint8 {
				return nil, errExpectedLowerOrEqual("ChannelIndex", math.MaxUint8)(pld.ChannelIndex)
			}
			b = append(b, byte(pld.ChannelIndex))

			if pld.Frequency < 100000 || pld.Frequency > byteutil.MaxUint24*phy.FreqMultiplier {
				return nil, errExpectedBetween("Frequency", 100000, byteutil.MaxUint24*phy.FreqMultiplier)(pld.Frequency)
			}
			b = byteutil.AppendUint64(b, pld.Frequency/phy.FreqMultiplier, 3)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DL_CHANNEL, "DLChannelReq", 4, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DlChannelReq{
				DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
					ChannelIndex: uint32(b[0]),
					Frequency:    byteutil.ParseUint64(b[1:4]) * phy.FreqMultiplier,
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_REKEY: &MACCommandDescriptor{
		InitiatedByDevice: true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRekeyInd()
			if pld.MinorVersion > 15 {
				return nil, errExpectedLowerOrEqual("MinorVersion", 15)(pld.MinorVersion)
			}
			b = append(b, byte(pld.MinorVersion))
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_REKEY, "RekeyInd", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RekeyInd_{
				RekeyInd: &ttnpb.MACCommand_RekeyInd{
					MinorVersion: ttnpb.Minor(b[0] & 0xf),
				},
			}
			return nil
		}),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRekeyConf()
			if pld.MinorVersion > 15 {
				return nil, errExpectedLowerOrEqual("MinorVersion", 15)(pld.MinorVersion)
			}
			b = append(b, byte(pld.MinorVersion))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_REKEY, "RekeyConf", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RekeyConf_{
				RekeyConf: &ttnpb.MACCommand_RekeyConf{
					MinorVersion: ttnpb.Minor(b[0] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_ADR_PARAM_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,

		AppendUplink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_ADR_PARAM_SETUP, "ADRParamSetupAns", 0, nil),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetAdrParamSetupReq()
			if 1 > pld.AdrAckDelayExponent || pld.AdrAckDelayExponent > 32768 {
				return nil, errExpectedBetween("ADRAckDelay", 1, 32768)(pld.AdrAckDelayExponent)
			}
			v := byte(pld.AdrAckDelayExponent)

			if 1 > pld.AdrAckLimitExponent || pld.AdrAckLimitExponent > 32768 {
				return nil, errExpectedBetween("ADRAckLimit", 1, 32768)(pld.AdrAckLimitExponent)
			}
			v |= byte(pld.AdrAckLimitExponent) << 4

			b = append(b, v)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_ADR_PARAM_SETUP, "ADRParamSetupReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_AdrParamSetupReq{
				AdrParamSetupReq: &ttnpb.MACCommand_ADRParamSetupReq{
					AdrAckDelayExponent: ttnpb.ADRAckDelayExponent(b[0] & 0xf),
					AdrAckLimitExponent: ttnpb.ADRAckLimitExponent(b[0] >> 4),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_DEVICE_TIME: &MACCommandDescriptor{
		InitiatedByDevice: true,

		AppendUplink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DEVICE_TIME, "DeviceTimeReq", 0, nil),

		DownlinkLength: 5,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDeviceTimeAns()

			t := gpstime.ToGPS(*ttnpb.StdTime(pld.Time))
			sec := t / time.Second
			if sec > math.MaxUint32 {
				return nil, errExpectedLowerOrEqual("Time", uint32(math.MaxUint32))(sec)
			}
			b = byteutil.AppendUint32(b, uint32(sec), 4)
			b = append(b, byte((t-sec*time.Second)/fractStep))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DEVICE_TIME, "DeviceTimeAns", 5, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DeviceTimeAns_{
				DeviceTimeAns: &ttnpb.MACCommand_DeviceTimeAns{
					Time: timestamppb.New(gpstime.Parse(time.Duration(byteutil.ParseUint32(b[0:4]))*time.Second + time.Duration(b[4])*fractStep)),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_FORCE_REJOIN: &MACCommandDescriptor{
		InitiatedByDevice: false,

		DownlinkLength: 2,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
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
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_FORCE_REJOIN, "ForceRejoinReq", 2, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_ForceRejoinReq_{
				ForceRejoinReq: &ttnpb.MACCommand_ForceRejoinReq{
					PeriodExponent: ttnpb.RejoinPeriodExponent(uint32(b[0] >> 3)),
					MaxRetries:     uint32(b[0] & 0x7),
					RejoinType:     ttnpb.RejoinRequestType(b[1] >> 4),
					DataRateIndex:  ttnpb.DataRateIndex(b[1] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_REJOIN_PARAM_SETUP: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRejoinParamSetupAns()

			var v byte
			if pld.MaxTimeExponentAck {
				v |= 1
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_REJOIN_PARAM_SETUP, "RejoinParamSetupAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RejoinParamSetupAns_{
				RejoinParamSetupAns: &ttnpb.MACCommand_RejoinParamSetupAns{
					MaxTimeExponentAck: b[0]&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
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
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_REJOIN_PARAM_SETUP, "RejoinParamSetupReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_RejoinParamSetupReq_{
				RejoinParamSetupReq: &ttnpb.MACCommand_RejoinParamSetupReq{
					MaxTimeExponent:  ttnpb.RejoinTimeExponent(uint32(b[0] >> 4)),
					MaxCountExponent: ttnpb.RejoinCountExponent(uint32(b[0] & 0xf)),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_PING_SLOT_INFO: &MACCommandDescriptor{
		InitiatedByDevice: true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetPingSlotInfoReq()

			if pld.Period > 7 {
				return nil, errExpectedLowerOrEqual("Period", 15)(pld.Period)
			}
			b = append(b, byte(pld.Period))
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_PING_SLOT_INFO, "PingSlotInfoReq", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_PingSlotInfoReq_{
				PingSlotInfoReq: &ttnpb.MACCommand_PingSlotInfoReq{
					Period: ttnpb.PingSlotPeriod(b[0] & 0x7),
				},
			}
			return nil
		}),

		AppendDownlink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_PING_SLOT_INFO, "PingSlotInfoAns", 0, nil),
	},

	ttnpb.MACCommandIdentifier_CID_PING_SLOT_CHANNEL: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
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
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_PING_SLOT_CHANNEL, "PingSlotChannelAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_PingSlotChannelAns_{
				PingSlotChannelAns: &ttnpb.MACCommand_PingSlotChannelAns{
					FrequencyAck:     b[0]&1 == 1,
					DataRateIndexAck: (b[0]>>1)&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 4,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetPingSlotChannelReq()

			if pld.Frequency < 100000 || pld.Frequency > byteutil.MaxUint24*phy.FreqMultiplier {
				return nil, errExpectedBetween("Frequency", 100000, byteutil.MaxUint24*phy.FreqMultiplier)(pld.Frequency)
			}
			b = byteutil.AppendUint64(b, pld.Frequency/phy.FreqMultiplier, 3)

			if pld.DataRateIndex > 15 {
				return nil, errExpectedLowerOrEqual("DataRateIndex", 15)(pld.DataRateIndex)
			}
			b = append(b, byte(pld.DataRateIndex))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_PING_SLOT_CHANNEL, "PingSlotChannelReq", 4, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_PingSlotChannelReq_{
				PingSlotChannelReq: &ttnpb.MACCommand_PingSlotChannelReq{
					Frequency:     byteutil.ParseUint64(b[0:3]) * phy.FreqMultiplier,
					DataRateIndex: ttnpb.DataRateIndex(b[3] & 0xf),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_BEACON_TIMING: &MACCommandDescriptor{
		InitiatedByDevice: true,

		AppendUplink: func(phy band.Band, b []byte, _ *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_BEACON_TIMING, "BeaconTimingReq", 0, nil),

		DownlinkLength: 3,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetBeaconTimingAns()

			if pld.Delay > math.MaxUint16 {
				return nil, errExpectedLowerOrEqual("Delay", math.MaxUint16)(pld.Delay)
			}
			b = byteutil.AppendUint32(b, pld.Delay, 2)

			if pld.ChannelIndex > math.MaxUint8 {
				return nil, errExpectedLowerOrEqual("ChannelIndex", math.MaxUint8)(pld.ChannelIndex)
			}
			b = append(b, byte(pld.ChannelIndex))

			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_BEACON_TIMING, "BeaconTimingAns", 3, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_BeaconTimingAns_{
				BeaconTimingAns: &ttnpb.MACCommand_BeaconTimingAns{
					Delay:        byteutil.ParseUint32(b[0:2]),
					ChannelIndex: uint32(b[2]),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_BEACON_FREQ: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetBeaconFreqAns()
			var v byte
			if pld.FrequencyAck {
				v |= 1
			}
			b = append(b, v)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_BEACON_FREQ, "BeaconFreqAns", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_BeaconFreqAns_{
				BeaconFreqAns: &ttnpb.MACCommand_BeaconFreqAns{
					FrequencyAck: b[0]&1 == 1,
				},
			}
			return nil
		}),

		DownlinkLength: 3,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetBeaconFreqReq()
			if pld.Frequency < 100000 || pld.Frequency > byteutil.MaxUint24*phy.FreqMultiplier {
				return nil, errExpectedBetween("Frequency", 100000, byteutil.MaxUint24*phy.FreqMultiplier)(pld.Frequency)
			}
			b = byteutil.AppendUint64(b, pld.Frequency/phy.FreqMultiplier, 3)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_BEACON_FREQ, "BeaconFreqReq", 3, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_BeaconFreqReq_{
				BeaconFreqReq: &ttnpb.MACCommand_BeaconFreqReq{
					Frequency: byteutil.ParseUint64(b[0:3]) * phy.FreqMultiplier,
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_DEVICE_MODE: &MACCommandDescriptor{
		InitiatedByDevice: true,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDeviceModeInd()
			b = append(b, byte(pld.Class))
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DEVICE_MODE, "DeviceModeInd", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DeviceModeInd_{
				DeviceModeInd: &ttnpb.MACCommand_DeviceModeInd{
					Class: ttnpb.Class(b[0]),
				},
			}
			return nil
		}),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetDeviceModeConf()
			b = append(b, byte(pld.Class))
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(ttnpb.MACCommandIdentifier_CID_DEVICE_MODE, "DeviceModeConf", 1, func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
			cmd.Payload = &ttnpb.MACCommand_DeviceModeConf_{
				DeviceModeConf: &ttnpb.MACCommand_DeviceModeConf{
					Class: ttnpb.Class(b[0]),
				},
			}
			return nil
		}),
	},

	ttnpb.MACCommandIdentifier_CID_RELAY_CONF: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRelayConfAns()
			var status byte
			if pld.SecondChannelFrequencyAck {
				status |= 1
			}
			if pld.SecondChannelAckOffsetAck {
				status |= (1 << 1)
			}
			if pld.SecondChannelDataRateIndexAck {
				status |= (1 << 2)
			}
			if pld.SecondChannelIndexAck {
				status |= (1 << 3)
			}
			if pld.DefaultChannelIndexAck {
				status |= (1 << 4)
			}
			if pld.CadPeriodicityAck {
				status |= (1 << 5)
			}
			b = append(b, status)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_CONF,
			"RelayConfAns",
			1,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				cmd.Payload = &ttnpb.MACCommand_RelayConfAns_{
					RelayConfAns: &ttnpb.MACCommand_RelayConfAns{
						SecondChannelFrequencyAck:     b[0]&1 == 1,
						SecondChannelAckOffsetAck:     (b[0]>>1)&1 == 1,
						SecondChannelDataRateIndexAck: (b[0]>>2)&1 == 1,
						SecondChannelIndexAck:         (b[0]>>3)&1 == 1,
						DefaultChannelIndexAck:        (b[0]>>4)&1 == 1,
						CadPeriodicityAck:             (b[0]>>5)&1 == 1,
					},
				}
				return nil
			},
		),

		DownlinkLength: 5,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRelayConfReq()
			conf := pld.GetConfiguration()
			if conf == nil {
				b = append(b, 0x00, 0x00, 0x00, 0x00, 0x00)
				return b, nil
			}
			var chSettings uint16
			chSettings |= 1 << 13 // StartStop
			if conf.CadPeriodicity > 5 {
				return nil, errExpectedLowerOrEqual("CADPeriodicity", 5)(conf.CadPeriodicity)
			}
			chSettings |= uint16(conf.CadPeriodicity&0x7) << 10
			if conf.DefaultChannelIndex > 1 {
				return nil, errExpectedLowerOrEqual("DefaultChannelIndex", 1)(conf.DefaultChannelIndex)
			}
			chSettings |= uint16(conf.DefaultChannelIndex&0x1) << 9
			var secondChFrequency uint64
			if secondCh := conf.SecondChannel; secondCh != nil {
				chSettings |= 1 << 7 // SecondChannelIndex
				if secondCh.DataRateIndex > 15 {
					return nil, errExpectedLowerOrEqual("SecondChannelDataRateIndex", 15)(secondCh.DataRateIndex)
				}
				chSettings |= uint16(secondCh.DataRateIndex&0xf) << 3
				if secondCh.AckOffset > 5 {
					return nil, errExpectedLowerOrEqual("SecondChannelAckOffset", 5)(secondCh.AckOffset)
				}
				chSettings |= uint16(secondCh.AckOffset & 0x7)
				if secondCh.Frequency < 100000 || secondCh.Frequency > byteutil.MaxUint24*phy.FreqMultiplier {
					return nil, errExpectedBetween(
						"SecondChannelFrequency", 100000, byteutil.MaxUint24*phy.FreqMultiplier,
					)(secondCh.Frequency)
				}
				secondChFrequency = secondCh.Frequency
			}
			b = byteutil.AppendUint16(b, chSettings, 2)
			b = byteutil.AppendUint64(b, secondChFrequency/phy.FreqMultiplier, 3)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_CONF,
			"RelayConfReq",
			5,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				payload := &ttnpb.MACCommand_RelayConfReq_{
					RelayConfReq: &ttnpb.MACCommand_RelayConfReq{},
				}
				cmd.Payload = payload
				chSettings := byteutil.ParseUint16(b[0:2])
				if chSettings&(1<<13) == 0 {
					return nil
				}
				conf := &ttnpb.MACCommand_RelayConfReq_Configuration{
					SecondChannel:       nil,
					DefaultChannelIndex: uint32(chSettings>>9) & 0x1,
					CadPeriodicity:      ttnpb.RelayCADPeriodicity(chSettings>>10) & 0x7,
				}
				if chSettings&(1<<7) != 0 {
					conf.SecondChannel = &ttnpb.RelaySecondChannel{
						Frequency:     byteutil.ParseUint64(b[2:5]) * phy.FreqMultiplier,
						AckOffset:     ttnpb.RelaySecondChAckOffset(chSettings) & 0x7,
						DataRateIndex: ttnpb.DataRateIndex(chSettings>>3) & 0xf,
					}
				}
				payload.RelayConfReq.Configuration = conf
				return nil
			},
		),
	},
	ttnpb.MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 1,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRelayEndDeviceConfAns()
			var status byte
			if pld.SecondChannelFrequencyAck {
				status |= 1
			}
			if pld.SecondChannelDataRateIndexAck {
				status |= (1 << 1)
			}
			if pld.SecondChannelIndexAck {
				status |= (1 << 2)
			}
			if pld.BackoffAck {
				status |= (1 << 3)
			}
			b = append(b, status)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF,
			"RelayEndDeviceConfAns",
			1,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				cmd.Payload = &ttnpb.MACCommand_RelayEndDeviceConfAns_{
					RelayEndDeviceConfAns: &ttnpb.MACCommand_RelayEndDeviceConfAns{
						SecondChannelFrequencyAck:     b[0]&1 == 1,
						SecondChannelDataRateIndexAck: (b[0]>>1)&1 == 1,
						SecondChannelIndexAck:         (b[0]>>2)&1 == 1,
						BackoffAck:                    (b[0]>>3)&1 == 1,
					},
				}
				return nil
			},
		),

		DownlinkLength: 6,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRelayEndDeviceConfReq()
			conf := pld.GetConfiguration()
			if conf == nil {
				b = append(b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
				return b, nil
			}
			var mode byte
			switch conf.Mode.(type) {
			case *ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Always:
				mode |= 1 << 2
			case *ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Dynamic:
				mode |= 2 << 2
				smartEnableLevel := conf.GetDynamic().GetSmartEnableLevel()
				if smartEnableLevel > 3 {
					return nil, errExpectedLowerOrEqual("SmartEnableLevel", 3)(smartEnableLevel)
				}
				mode |= byte(conf.GetDynamic().GetSmartEnableLevel() & 0x3)
			case *ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_EndDeviceControlled:
				mode |= 3 << 2
			default:
				return nil, errExpectedLengthLowerOrEqual("mode", 3)(conf.Mode)
			}
			b = append(b, mode)
			var chSettings uint16
			if conf.Backoff > 63 {
				return nil, errExpectedLowerOrEqual("Backoff", 63)(conf.Backoff)
			}
			chSettings |= uint16(conf.Backoff&0x3f) << 9
			var secondChFrequency uint64
			if secondCh := conf.SecondChannel; secondCh != nil {
				chSettings |= 1 << 7 // SecondChannelIndex
				if secondCh.DataRateIndex > 15 {
					return nil, errExpectedLowerOrEqual("SecondChannelDataRateIndex", 15)(secondCh.DataRateIndex)
				}
				chSettings |= uint16(secondCh.DataRateIndex&0xf) << 3
				if secondCh.Frequency < 100000 || secondCh.Frequency > byteutil.MaxUint24*phy.FreqMultiplier {
					return nil, errExpectedBetween(
						"SecondChannelFrequency", 100000, byteutil.MaxUint24*phy.FreqMultiplier,
					)(secondCh.Frequency)
				}
				if secondCh.AckOffset > 5 {
					return nil, errExpectedLowerOrEqual("SecondChannelAckOffset", 5)(secondCh.AckOffset)
				}
				chSettings |= uint16(secondCh.AckOffset & 0x7)
				secondChFrequency = secondCh.Frequency
			}
			b = byteutil.AppendUint16(b, chSettings, 2)
			b = byteutil.AppendUint64(b, secondChFrequency/phy.FreqMultiplier, 3)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF,
			"RelayEndDeviceConfReq",
			6,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				payload := &ttnpb.MACCommand_RelayEndDeviceConfReq_{
					RelayEndDeviceConfReq: &ttnpb.MACCommand_RelayEndDeviceConfReq{},
				}
				cmd.Payload = payload
				var conf *ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration
				switch mode := b[0] >> 2; mode {
				case 0:
					return nil
				case 1:
					conf = &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
						Mode: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Always{},
					}
				case 2:
					conf = &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
						Mode: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Dynamic{
							Dynamic: &ttnpb.RelayEndDeviceDynamicMode{
								SmartEnableLevel: ttnpb.RelaySmartEnableLevel(b[0] & 0x3),
							},
						},
					}
				case 3:
					conf = &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
						Mode: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_EndDeviceControlled{},
					}
				default:
					return errExpectedLengthLowerOrEqual("mode", 3)(mode)
				}
				payload.RelayEndDeviceConfReq.Configuration = conf
				chSettings := byteutil.ParseUint16(b[1:3])
				conf.Backoff = uint32(chSettings>>9) & 0x3f
				if chSettings&(1<<7) != 0 {
					conf.SecondChannel = &ttnpb.RelaySecondChannel{
						Frequency:     byteutil.ParseUint64(b[3:6]) * phy.FreqMultiplier,
						AckOffset:     ttnpb.RelaySecondChAckOffset(chSettings) & 0x7,
						DataRateIndex: ttnpb.DataRateIndex(chSettings>>3) & 0xf,
					}
				}
				return nil
			}),
	},
	ttnpb.MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST,
			"RelayUpdateUplinkListAns",
			0,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				cmd.Payload = &ttnpb.MACCommand_RelayUpdateUplinkListAns_{
					RelayUpdateUplinkListAns: &ttnpb.MACCommand_RelayUpdateUplinkListAns{},
				}
				return nil
			},
		),

		DownlinkLength: 26,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRelayUpdateUplinkListReq()
			if pld.RuleIndex > 15 {
				return nil, errExpectedLowerOrEqual("RuleIndex", 15)(pld.RuleIndex)
			}
			b = append(b, byte(pld.RuleIndex))
			var uplinkLimit byte = 0x3f
			if limits := pld.ForwardLimits; limits != nil {
				if limits.BucketSize > 3 {
					return nil, errExpectedLowerOrEqual("BucketSize", 3)(limits.BucketSize)
				}
				uplinkLimit = byte(limits.BucketSize&0x3) << 6
				if limits.ReloadRate > 62 {
					return nil, errExpectedLowerOrEqual("ReloadRate", 62)(limits.ReloadRate)
				}
				uplinkLimit |= byte(limits.ReloadRate)
			}
			b = append(b, uplinkLimit)
			if n := len(pld.DevAddr); n != 4 {
				return nil, errExpectedLengthEncodedEqual("DevAddr", 4)(n)
			}
			devAddr := make([]byte, 4)
			copyReverse(devAddr, pld.DevAddr)
			b = append(b, devAddr...)
			b = byteutil.AppendUint32(b, pld.WFCnt, 4)
			if n := len(pld.RootWorSKey); n != 16 {
				return nil, errExpectedLengthEncodedEqual("RootWorSKey", 16)(n)
			}
			b = append(b, pld.RootWorSKey...)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST,
			"RelayUpdateUplinkListReq",
			26,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				req := &ttnpb.MACCommand_RelayUpdateUplinkListReq{
					DevAddr: make([]byte, 4),
				}
				cmd.Payload = &ttnpb.MACCommand_RelayUpdateUplinkListReq_{
					RelayUpdateUplinkListReq: req,
				}
				req.RuleIndex = uint32(b[0])
				if req.RuleIndex > 7 {
					return errExpectedLowerOrEqual("RuleIndex", 7)(req.RuleIndex)
				}
				uplinkLimit := b[1]
				if (uplinkLimit & 0x3f) != 0x3f {
					req.ForwardLimits = &ttnpb.RelayUplinkForwardLimits{
						BucketSize: ttnpb.RelayLimitBucketSize(uplinkLimit >> 6),
						ReloadRate: uint32(uplinkLimit & 0x3f),
					}
				}
				copyReverse(req.DevAddr[:], b[2:6])
				req.WFCnt = byteutil.ParseUint32(b[6:10])
				req.RootWorSKey = b[10:26]
				return nil
			},
		),
	},
	ttnpb.MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 5,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRelayCtrlUplinkListAns()
			var status byte
			if pld.RuleIndexAck {
				status |= 1
			}
			b = append(b, status)
			b = byteutil.AppendUint32(b, pld.WFCnt, 4)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST,
			"RelayCtrlUplinkListAns",
			5,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				cmd.Payload = &ttnpb.MACCommand_RelayCtrlUplinkListAns_{
					RelayCtrlUplinkListAns: &ttnpb.MACCommand_RelayCtrlUplinkListAns{
						RuleIndexAck: b[0]&1 == 1,
						WFCnt:        byteutil.ParseUint32(b[1:5]),
					},
				}
				return nil
			},
		),

		DownlinkLength: 1,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			req := cmd.GetRelayCtrlUplinkListReq()
			if req.RuleIndex > 15 {
				return nil, errExpectedLowerOrEqual("RuleIndex", 7)(req.RuleIndex)
			}
			var action byte
			action |= byte(req.RuleIndex) & 0xf
			if req.Action > 1 {
				return nil, errExpectedLowerOrEqual("Action", 1)(req.Action)
			}
			action |= byte(req.Action&0x1) << 4
			b = append(b, action)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST,
			"RelayCtrlUplinkListReq",
			1,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				cmd.Payload = &ttnpb.MACCommand_RelayCtrlUplinkListReq_{
					RelayCtrlUplinkListReq: &ttnpb.MACCommand_RelayCtrlUplinkListReq{
						RuleIndex: uint32(b[0] & 0xf),
						Action:    ttnpb.RelayCtrlUplinkListAction(b[0] >> 4),
					},
				}
				return nil
			},
		),
	},
	ttnpb.MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT: &MACCommandDescriptor{
		InitiatedByDevice: false,

		UplinkLength: 0,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT,
			"RelayConfigureFwdLimitAns",
			0,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				cmd.Payload = &ttnpb.MACCommand_RelayConfigureFwdLimitAns_{
					RelayConfigureFwdLimitAns: &ttnpb.MACCommand_RelayConfigureFwdLimitAns{},
				}
				return nil
			},
		),

		DownlinkLength: 5,
		AppendDownlink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRelayConfigureFwdLimitReq()
			if pld.ResetLimitCounter > 3 {
				return nil, errExpectedLowerOrEqual("ResetLimitCounter", 3)(pld.ResetLimitCounter)
			}
			var overallReloadRate byte = 0x7f
			var overallBucketSize byte
			if overall := pld.OverallLimits; overall != nil {
				if overall.ReloadRate > 126 {
					return nil, errExpectedLowerOrEqual("ReloadRate", 126)(overall.ReloadRate)
				}
				overallReloadRate = byte(overall.ReloadRate & 0x7f)
				if overall.BucketSize > 3 {
					return nil, errExpectedLowerOrEqual("BucketSize", 3)(overall.BucketSize)
				}
				overallBucketSize = byte(overall.BucketSize & 0x3)
			}
			var globalUplinkReloadRate byte = 0x7f
			var globalUplinkBucketSize byte
			if global := pld.GlobalUplinkLimits; global != nil {
				if global.ReloadRate > 126 {
					return nil, errExpectedLowerOrEqual("ReloadRate", 126)(global.ReloadRate)
				}
				globalUplinkReloadRate = byte(global.ReloadRate & 0x7f)
				if global.BucketSize > 3 {
					return nil, errExpectedLowerOrEqual("BucketSize", 3)(global.BucketSize)
				}
				globalUplinkBucketSize = byte(global.BucketSize & 0x3)
			}
			var notifyReloadRate byte = 0x7f
			var notifyBucketSize byte
			if notify := pld.NotifyLimits; notify != nil {
				if notify.ReloadRate > 126 {
					return nil, errExpectedLowerOrEqual("ReloadRate", 126)(notify.ReloadRate)
				}
				notifyReloadRate = byte(notify.ReloadRate & 0x7f)
				if notify.BucketSize > 3 {
					return nil, errExpectedLowerOrEqual("BucketSize", 3)(notify.BucketSize)
				}
				notifyBucketSize = byte(notify.BucketSize & 0x3)
			}
			var joinRequestLimits byte = 0x7f
			var joinRequestBucketSize byte
			if joinReq := pld.JoinRequestLimits; joinReq != nil {
				if joinReq.ReloadRate > 126 {
					return nil, errExpectedLowerOrEqual("ReloadRate", 126)(joinReq.ReloadRate)
				}
				joinRequestLimits = byte(joinReq.ReloadRate & 0x7f)
				if joinReq.BucketSize > 3 {
					return nil, errExpectedLowerOrEqual("BucketSize", 3)(joinReq.BucketSize)
				}
				joinRequestBucketSize = byte(joinReq.BucketSize & 0x3)
			}
			var reloadRate uint32
			reloadRate |= uint32(overallReloadRate)
			reloadRate |= uint32(globalUplinkReloadRate) << 7
			reloadRate |= uint32(notifyReloadRate) << 14
			reloadRate |= uint32(joinRequestLimits) << 21
			reloadRate |= uint32(pld.ResetLimitCounter&0x3) << 28
			b = byteutil.AppendUint32(b, reloadRate, 4)
			var bucketSize byte
			bucketSize |= overallBucketSize
			bucketSize |= globalUplinkBucketSize << 2
			bucketSize |= notifyBucketSize << 4
			bucketSize |= joinRequestBucketSize << 6
			b = append(b, bucketSize)
			return b, nil
		},
		UnmarshalDownlink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT,
			"RelayConfigureFwdLimitReq",
			5,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				req := &ttnpb.MACCommand_RelayConfigureFwdLimitReq{}
				cmd.Payload = &ttnpb.MACCommand_RelayConfigureFwdLimitReq_{
					RelayConfigureFwdLimitReq: req,
				}
				bucketSize, reloadRate := b[4], byteutil.ParseUint32(b[0:4])
				req.ResetLimitCounter = ttnpb.RelayResetLimitCounter((reloadRate >> 28) & 0x3)
				if reloadRate := reloadRate & 0x7f; reloadRate != 0x7f {
					req.OverallLimits = &ttnpb.RelayForwardLimits{
						BucketSize: ttnpb.RelayLimitBucketSize(bucketSize & 0x3),
						ReloadRate: reloadRate,
					}
				}
				if reloadRate := (reloadRate >> 7) & 0x7f; reloadRate != 0x7f {
					req.GlobalUplinkLimits = &ttnpb.RelayForwardLimits{
						BucketSize: ttnpb.RelayLimitBucketSize((bucketSize >> 2) & 0x3),
						ReloadRate: reloadRate,
					}
				}
				if reloadRate := (reloadRate >> 14) & 0x7f; reloadRate != 0x7f {
					req.NotifyLimits = &ttnpb.RelayForwardLimits{
						BucketSize: ttnpb.RelayLimitBucketSize((bucketSize >> 4) & 0x3),
						ReloadRate: reloadRate,
					}
				}
				if reloadRate := (reloadRate >> 21) & 0x7f; reloadRate != 0x7f {
					req.JoinRequestLimits = &ttnpb.RelayForwardLimits{
						BucketSize: ttnpb.RelayLimitBucketSize((bucketSize >> 6) & 0x3),
						ReloadRate: reloadRate,
					}
				}
				return nil
			},
		),
	},
	ttnpb.MACCommandIdentifier_CID_RELAY_NOTIFY_NEW_END_DEVICE: &MACCommandDescriptor{
		InitiatedByDevice: true,

		UplinkLength: 6,
		AppendUplink: func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
			pld := cmd.GetRelayNotifyNewEndDeviceReq()
			var powerLevel uint16
			if pld.Snr < -20 || pld.Snr > 11 {
				return nil, errExpectedBetween("SNR", -20, 11)(pld.Snr)
			}
			powerLevel |= uint16(pld.Snr+20) & 0x1f
			if pld.Rssi < -142 || pld.Rssi > -15 {
				return nil, errExpectedBetween("RSSI", -142, -15)(pld.Rssi)
			}
			powerLevel |= uint16(-pld.Rssi-15) & 0x7f << 5
			b = byteutil.AppendUint16(b, powerLevel, 2)
			if n := len(pld.DevAddr); n != 4 {
				return nil, errExpectedLengthEncodedEqual("DevAddr", 4)(n)
			}
			devAddr := make([]byte, 4)
			copyReverse(devAddr, pld.DevAddr)
			b = append(b, devAddr...)
			return b, nil
		},
		UnmarshalUplink: newMACUnmarshaler(
			ttnpb.MACCommandIdentifier_CID_RELAY_NOTIFY_NEW_END_DEVICE,
			"RelayNotifyNewEndDeviceReq",
			6,
			func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) error {
				req := &ttnpb.MACCommand_RelayNotifyNewEndDeviceReq{}
				cmd.Payload = &ttnpb.MACCommand_RelayNotifyNewEndDeviceReq_{
					RelayNotifyNewEndDeviceReq: req,
				}
				powerLevel := byteutil.ParseUint16(b[0:2])
				req.Snr = int32(powerLevel&0x1f) - 20
				req.Rssi = -int32(powerLevel>>5&0x7f) - 15
				req.DevAddr = make([]byte, 4)
				copyReverse(req.DevAddr, b[2:6])
				return nil
			},
		),
	},
}

var (
	errDecodingMACCommand = errors.DefineInvalidArgument("decoding_mac_command", "decode MAC command with CID `{cid}`")
	errNoUnmarshaler      = errors.DefineNotFound("no_unmarshaler", "no unmarshaler available for MAC command with CID `{cid}`")
)

func (spec MACCommandSpec) read(phy band.Band, r io.Reader, isUplink bool, cmd *ttnpb.MACCommand) error {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return err
	}

	cid := ttnpb.MACCommandIdentifier(b[0])

	desc, ok := spec[cid]
	if !ok || desc == nil {
		b, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		payload := &ttnpb.MACCommand_RawPayload{
			RawPayload: b,
		}
		return cmd.SetFields(&ttnpb.MACCommand{Payload: payload, Cid: cid}, ttnpb.MACCommandFieldPathsTopLevel...)
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
	if unmarshaler == nil {
		return errNoUnmarshaler.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cid)))
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
		return errDecodingMACCommand.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cid))).WithCause(err)
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
	errEncodingMACCommand = errors.DefineInvalidArgument("encoding_mac_command", "encode MAC command with CID `{cid}`")
	errUnknownMACCommand  = errors.DefineInvalidArgument("unknown_mac_command", "unknown MAC command CID `{cid}`")
	errMACCommandUplink   = errors.DefineInvalidArgument("mac_command_uplink", "invalid uplink MAC command CID `{cid}`")
	errMACCommandDownlink = errors.DefineInvalidArgument("mac_command_downlink", "invalid downlink MAC command CID `{cid}`")
)

func (spec MACCommandSpec) append(phy band.Band, b []byte, isUplink bool, cmd *ttnpb.MACCommand) ([]byte, error) {
	desc, ok := spec[cmd.Cid]
	if !ok || desc == nil {
		return nil, errUnknownMACCommand.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cmd.Cid)))
	}
	b = append(b, byte(cmd.Cid))

	var appender func(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error)
	if isUplink {
		appender = desc.AppendUplink
		if appender == nil {
			return nil, errMACCommandUplink.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cmd.Cid)))
		}
	} else {
		appender = desc.AppendDownlink
		if appender == nil {
			return nil, errMACCommandDownlink.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cmd.Cid)))
		}
	}

	b, err := appender(phy, b, cmd)
	if err != nil {
		return nil, errEncodingMACCommand.WithAttributes("cid", fmt.Sprintf("0x%X", int32(cmd.Cid))).WithCause(err)
	}
	return b, nil
}

// AppendUplink encodes uplink MAC command cmd, appends it to b and returns any errors encountered.
func (spec MACCommandSpec) AppendUplink(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
	return spec.append(phy, b, true, cmd)
}

// AppendDownlink encodes downlink MAC command cmd, appends it to b and returns any errors encountered.
func (spec MACCommandSpec) AppendDownlink(phy band.Band, b []byte, cmd *ttnpb.MACCommand) ([]byte, error) {
	return spec.append(phy, b, false, cmd)
}
