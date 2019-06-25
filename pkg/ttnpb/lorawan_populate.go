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

package ttnpb

import (
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// NewPopulatedFrequency returns a uint64 in range [100000, 1677721600].
func NewPopulatedFrequency(r randyEndDevice, _ bool) uint64 {
	return 100000 + uint64(r.Int63()%(1677721600-100000))
}

func NewPopulatedDataRateIndex(r randyEndDevice, _ bool) DataRateIndex {
	return DataRateIndex(r.Intn(16))
}

func NewPopulatedChannelIndex(r randyEndDevice, _ bool) uint32 {
	return r.Uint32() % 256
}

func NewPopulatedMACCommand_ResetInd(r randyLorawan, _ bool) *MACCommand_ResetInd {
	out := &MACCommand_ResetInd{}
	out.MinorVersion = MINOR_1
	return out
}

func NewPopulatedMACCommand_ResetConf(r randyLorawan, _ bool) *MACCommand_ResetConf {
	out := &MACCommand_ResetConf{}
	out.MinorVersion = MINOR_1
	return out
}

func NewPopulatedMACCommand_LinkCheckAns(r randyLorawan, _ bool) *MACCommand_LinkCheckAns {
	out := &MACCommand_LinkCheckAns{}
	out.Margin = r.Uint32() % 255
	out.GatewayCount = r.Uint32() % 256
	return out
}

func NewPopulatedMACCommand_LinkADRReq(r randyLorawan, easy bool) *MACCommand_LinkADRReq {
	out := &MACCommand_LinkADRReq{}
	out.DataRateIndex = NewPopulatedDataRateIndex(r, easy)
	out.TxPowerIndex = r.Uint32() % 16
	out.ChannelMask = make([]bool, r.Intn(17))
	for i := 0; i < len(out.ChannelMask); i++ {
		out.ChannelMask[i] = r.Intn(2) == 0
	}
	out.ChannelMaskControl = r.Uint32() % 8
	out.NbTrans = r.Uint32() % 16
	return out
}

func NewPopulatedMACCommand_RxParamSetupReq(r randyLorawan, easy bool) *MACCommand_RxParamSetupReq {
	out := &MACCommand_RxParamSetupReq{}
	out.Rx2DataRateIndex = NewPopulatedDataRateIndex(r, easy)
	out.Rx1DataRateOffset = r.Uint32() % 8
	out.Rx2Frequency = NewPopulatedFrequency(r, easy)
	return out
}

func NewPopulatedMACCommand_DevStatusAns(r randyLorawan, _ bool) *MACCommand_DevStatusAns {
	out := &MACCommand_DevStatusAns{}
	out.Battery = r.Uint32() % 256
	out.Margin = r.Int31() % 32
	if r.Intn(2) == 0 {
		out.Margin *= -1
	}
	return out
}

func NewPopulatedMACCommand_NewChannelReq(r randyLorawan, easy bool) *MACCommand_NewChannelReq {
	out := &MACCommand_NewChannelReq{}
	out.ChannelIndex = NewPopulatedChannelIndex(r, easy)
	out.Frequency = NewPopulatedFrequency(r, easy)
	out.MinDataRateIndex = NewPopulatedDataRateIndex(r, easy)
	out.MaxDataRateIndex = NewPopulatedDataRateIndex(r, easy)
	return out
}

func NewPopulatedMACCommand_DLChannelReq(r randyLorawan, easy bool) *MACCommand_DLChannelReq {
	out := &MACCommand_DLChannelReq{}
	out.ChannelIndex = NewPopulatedChannelIndex(r, easy)
	out.Frequency = NewPopulatedFrequency(r, easy)
	return out
}

func NewPopulatedMACCommand_ForceRejoinReq(r randyLorawan, easy bool) *MACCommand_ForceRejoinReq {
	out := &MACCommand_ForceRejoinReq{}
	out.RejoinType = RejoinType(r.Intn(3))
	out.DataRateIndex = NewPopulatedDataRateIndex(r, easy)
	out.MaxRetries = r.Uint32() % 8
	out.PeriodExponent = RejoinPeriodExponent(r.Intn(8))
	return out
}

func NewPopulatedMACCommand_PingSlotChannelReq(r randyLorawan, easy bool) *MACCommand_PingSlotChannelReq {
	out := &MACCommand_PingSlotChannelReq{}
	out.Frequency = NewPopulatedFrequency(r, easy)
	out.DataRateIndex = NewPopulatedDataRateIndex(r, easy)
	return out
}

func NewPopulatedMACCommand_BeaconTimingAns(r randyLorawan, easy bool) *MACCommand_BeaconTimingAns {
	out := &MACCommand_BeaconTimingAns{}
	out.Delay = r.Uint32() % 65536
	out.ChannelIndex = r.Uint32() % 256
	return out
}

func NewPopulatedMACCommand_BeaconFreqReq(r randyLorawan, easy bool) *MACCommand_BeaconFreqReq {
	out := &MACCommand_BeaconFreqReq{}
	out.Frequency = NewPopulatedFrequency(r, easy)
	return out
}

func NewPopulatedMACCommand(r randyLorawan, easy bool) *MACCommand {
	out := &MACCommand{}
	out.CID = MACCommandIdentifier([]int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 32}[r.Intn(20)])
	switch out.CID {
	case CID_RESET:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_ResetInd_{ResetInd: NewPopulatedMACCommand_ResetInd(r, easy)}
		} else {
			out.Payload = &MACCommand_ResetConf_{ResetConf: NewPopulatedMACCommand_ResetConf(r, easy)}
		}
	case CID_LINK_CHECK:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_LinkCheckAns_{LinkCheckAns: NewPopulatedMACCommand_LinkCheckAns(r, easy)}
		}
	case CID_LINK_ADR:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_LinkADRReq_{LinkADRReq: NewPopulatedMACCommand_LinkADRReq(r, easy)}
		} else {
			out.Payload = &MACCommand_LinkADRAns_{LinkADRAns: NewPopulatedMACCommand_LinkADRAns(r, easy)}
		}
	case CID_DUTY_CYCLE:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_DutyCycleReq_{DutyCycleReq: NewPopulatedMACCommand_DutyCycleReq(r, easy)}
		}
	case CID_RX_PARAM_SETUP:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_RxParamSetupReq_{RxParamSetupReq: NewPopulatedMACCommand_RxParamSetupReq(r, easy)}
		} else {
			out.Payload = &MACCommand_RxParamSetupAns_{RxParamSetupAns: NewPopulatedMACCommand_RxParamSetupAns(r, easy)}
		}
	case CID_DEV_STATUS:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_DevStatusAns_{DevStatusAns: NewPopulatedMACCommand_DevStatusAns(r, easy)}
		}
	case CID_NEW_CHANNEL:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_NewChannelReq_{NewChannelReq: NewPopulatedMACCommand_NewChannelReq(r, easy)}
		} else {
			out.Payload = &MACCommand_NewChannelAns_{NewChannelAns: NewPopulatedMACCommand_NewChannelAns(r, easy)}
		}
	case CID_DL_CHANNEL:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_DLChannelReq_{DLChannelReq: NewPopulatedMACCommand_DLChannelReq(r, easy)}
		} else {
			out.Payload = &MACCommand_DLChannelAns_{DLChannelAns: NewPopulatedMACCommand_DLChannelAns(r, easy)}
		}
	case CID_RX_TIMING_SETUP:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_RxTimingSetupReq_{RxTimingSetupReq: NewPopulatedMACCommand_RxTimingSetupReq(r, easy)}
		}
	case CID_TX_PARAM_SETUP:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_TxParamSetupReq_{TxParamSetupReq: NewPopulatedMACCommand_TxParamSetupReq(r, easy)}
		}
	case CID_REKEY:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_RekeyInd_{RekeyInd: NewPopulatedMACCommand_RekeyInd(r, easy)}
		} else {
			out.Payload = &MACCommand_RekeyConf_{RekeyConf: NewPopulatedMACCommand_RekeyConf(r, easy)}
		}
	case CID_ADR_PARAM_SETUP:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_ADRParamSetupReq_{ADRParamSetupReq: NewPopulatedMACCommand_ADRParamSetupReq(r, easy)}
		}
	case CID_DEVICE_TIME:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_DeviceTimeAns_{DeviceTimeAns: NewPopulatedMACCommand_DeviceTimeAns(r, easy)}
		}
	case CID_FORCE_REJOIN:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_ForceRejoinReq_{ForceRejoinReq: NewPopulatedMACCommand_ForceRejoinReq(r, easy)}
		}
	case CID_REJOIN_PARAM_SETUP:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_RejoinParamSetupReq_{RejoinParamSetupReq: NewPopulatedMACCommand_RejoinParamSetupReq(r, easy)}
		} else {
			out.Payload = &MACCommand_RejoinParamSetupAns_{RejoinParamSetupAns: NewPopulatedMACCommand_RejoinParamSetupAns(r, easy)}
		}
	case CID_PING_SLOT_INFO:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_PingSlotInfoReq_{PingSlotInfoReq: NewPopulatedMACCommand_PingSlotInfoReq(r, easy)}
		}
	case CID_PING_SLOT_CHANNEL:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_PingSlotChannelReq_{PingSlotChannelReq: NewPopulatedMACCommand_PingSlotChannelReq(r, easy)}
		} else {
			out.Payload = &MACCommand_PingSlotChannelAns_{PingSlotChannelAns: NewPopulatedMACCommand_PingSlotChannelAns(r, easy)}
		}
	case CID_BEACON_TIMING:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_BeaconTimingAns_{BeaconTimingAns: NewPopulatedMACCommand_BeaconTimingAns(r, easy)}
		}
	case CID_BEACON_FREQ:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_BeaconFreqReq_{BeaconFreqReq: NewPopulatedMACCommand_BeaconFreqReq(r, easy)}
		} else {
			out.Payload = &MACCommand_BeaconFreqAns_{BeaconFreqAns: NewPopulatedMACCommand_BeaconFreqAns(r, easy)}
		}
	case CID_DEVICE_MODE:
		if r.Intn(2) == 1 {
			out.Payload = &MACCommand_DeviceModeInd_{DeviceModeInd: NewPopulatedMACCommand_DeviceModeInd(r, easy)}
		} else {
			out.Payload = &MACCommand_DeviceModeConf_{DeviceModeConf: NewPopulatedMACCommand_DeviceModeConf(r, easy)}
		}
	default:
		if r.Intn(2) == 1 {
			b := make([]byte, r.Intn(100))
			for i := 0; i < len(b); i++ {
				b[i] = byte(r.Int63())
			}
			out.Payload = &MACCommand_RawPayload{RawPayload: b}
		}
	}
	return out
}

func NewPopulatedFHDR(r randyLorawan, _ bool) *FHDR {
	out := &FHDR{}
	out.DevAddr = *types.NewPopulatedDevAddr(r)
	out.FCtrl = *NewPopulatedFCtrl(r, false)
	out.FCnt = uint32(uint16(r.Uint32()))
	out.FOpts = make([]byte, r.Intn(15))
	for i := 0; i < len(out.FOpts); i++ {
		out.FOpts[i] = byte(128 + r.Intn(128))
	}
	return out
}

func NewPopulatedMACPayload(r randyLorawan, easy bool) *MACPayload {
	out := &MACPayload{}
	out.FHDR = *NewPopulatedFHDR(r, easy)
	out.FPort = uint32(r.Intn(225))
	out.FRMPayload = make([]byte, 10)
	for i := 0; i < len(out.FRMPayload); i++ {
		if r.Intn(2) == 0 {
			out.FRMPayload[i] = byte(1 + r.Intn(15))
		} else {
			out.FRMPayload[i] = byte(128 + r.Intn(128))
		}
	}
	return out
}

func NewPopulatedTxRequest(r randyLorawan, easy bool) *TxRequest {
	out := &TxRequest{}
	out.Class = []Class{CLASS_A, CLASS_B, CLASS_C}[r.Intn(3)]
	if out.Class == CLASS_A || r.Intn(2) == 0 {
		uplinkToken := make([]byte, 10)
		for i := 0; i < len(uplinkToken); i++ {
			if r.Intn(2) == 0 {
				uplinkToken[i] = byte(1 + r.Intn(15))
			} else {
				uplinkToken[i] = byte(128 + r.Intn(128))
			}
		}
		out.DownlinkPaths = []*DownlinkPath{
			{
				Path: &DownlinkPath_UplinkToken{
					UplinkToken: uplinkToken,
				},
			},
		}
	} else {
		out.DownlinkPaths = []*DownlinkPath{
			{
				Path: &DownlinkPath_Fixed{
					Fixed: &GatewayAntennaIdentifiers{
						GatewayIdentifiers: *NewPopulatedGatewayIdentifiers(r, false),
						AntennaIndex:       uint32(r.Intn(2)),
					},
				},
			},
		}
		if out.Class == CLASS_C && r.Intn(2) == 0 {
			out.AbsoluteTime = pbtypes.NewPopulatedStdTime(r, false)
		}
	}
	out.Rx1Delay = []RxDelay{RX_DELAY_1, RX_DELAY_2, RX_DELAY_5}[r.Intn(3)]
	out.Rx1DataRateIndex = []DataRateIndex{DATA_RATE_0, DATA_RATE_1, DATA_RATE_2}[r.Intn(3)]
	out.Rx1Frequency = []uint64{868100000, 868300000, 868500000}[r.Intn(3)]
	out.Rx2DataRateIndex = []DataRateIndex{DATA_RATE_0, DATA_RATE_1, DATA_RATE_2}[r.Intn(3)]
	out.Rx2Frequency = []uint64{868100000, 868300000, 868500000}[r.Intn(3)]
	out.Priority = []TxSchedulePriority{TxSchedulePriority_LOW, TxSchedulePriority_NORMAL, TxSchedulePriority_HIGH}[r.Intn(3)]
	return out
}

func NewPopulatedTxSettings(r randyLorawan, easy bool) *TxSettings {
	out := &TxSettings{
		Downlink: &TxSettings_Downlink{
			TxPower:            float32(r.Int31()),
			InvertPolarization: r.Intn(2) == 0,
		},
	}
	switch r.Intn(2) {
	case 0:
		out.DataRate.Modulation = &DataRate_FSK{
			FSK: &FSKDataRate{
				BitRate: 50000,
			},
		}
	case 1:
		out.DataRate.Modulation = &DataRate_LoRa{
			LoRa: &LoRaDataRate{
				Bandwidth:       []uint32{125000, 250000, 500000}[r.Intn(3)],
				SpreadingFactor: uint32(r.Intn(6) + 7),
			},
		}
		out.CodingRate = fmt.Sprintf("4/%d", r.Intn(4)+5)
	}
	out.Frequency = uint64(r.Uint32())
	out.CodingRate = fmt.Sprintf("4/%d", r.Intn(4)+5)
	out.DataRateIndex = NewPopulatedDataRateIndex(r, false) % 6
	return out
}

func NewPopulatedMessage_MACPayload(r randyLorawan) *Message_MACPayload {
	return &Message_MACPayload{NewPopulatedMACPayload(r, false)}
}

func NewPopulatedJoinRequestPayload(r randyLorawan, easy bool) *JoinRequestPayload {
	out := &JoinRequestPayload{}
	out.JoinEUI = *types.NewPopulatedEUI64(r)
	out.DevEUI = *types.NewPopulatedEUI64(r)
	out.DevNonce = *types.NewPopulatedDevNonce(r)
	return out
}

func NewPopulatedMessage_JoinRequestPayload(r randyLorawan) *Message_JoinRequestPayload {
	return &Message_JoinRequestPayload{NewPopulatedJoinRequestPayload(r, false)}
}

func NewPopulatedDLSettings(r randyLorawan, easy bool) *DLSettings {
	out := &DLSettings{}
	out.Rx1DROffset = uint32(r.Intn(8))
	out.Rx2DR = NewPopulatedDataRateIndex(r, easy)
	return out
}

func NewPopulatedCFList(r randyLorawan, easy bool) *CFList {
	out := &CFList{}
	out.Type = CFListType(r.Intn(2))
	switch out.Type {
	case 0:
		out.Freq = make([]uint32, 5)
		for i := 0; i < len(out.Freq); i++ {
			out.Freq[i] = uint32(r.Intn(0xfff))
		}
	case 1:
		out.ChMasks = make([]bool, 96)
		for i := 0; i < len(out.ChMasks); i++ {
			out.ChMasks[i] = (r.Intn(2) == 0)
		}
	default:
		panic("unreachable")
	}
	return out
}

func NewPopulatedJoinAcceptPayload(r randyLorawan, easy bool) *JoinAcceptPayload {
	out := &JoinAcceptPayload{}
	out.JoinNonce = *types.NewPopulatedJoinNonce(r)
	out.NetID = *types.NewPopulatedNetID(r)
	out.DevAddr = *types.NewPopulatedDevAddr(r)
	out.DLSettings = *NewPopulatedDLSettings(r, easy)
	out.RxDelay = RxDelay(r.Intn(16))
	if r.Intn(10) != 0 {
		out.CFList = NewPopulatedCFList(r, false)
	}
	return out
}
func NewPopulatedMessage_JoinAcceptPayload(r randyLorawan) *Message_JoinAcceptPayload {
	return &Message_JoinAcceptPayload{NewPopulatedJoinAcceptPayload(r, false)}
}

func NewPopulatedRejoinRequestPayloadType(r randyLorawan, typ RejoinType) *RejoinRequestPayload {
	out := &RejoinRequestPayload{}
	out.RejoinType = typ
	switch typ {
	case 0, 2:
		out.JoinEUI = types.EUI64{}
		out.NetID = *types.NewPopulatedNetID(r)
		out.DevEUI = *types.NewPopulatedEUI64(r)
		out.RejoinCnt = uint32(uint16(r.Uint32()))
	case 1:
		out.NetID = types.NetID{}
		out.JoinEUI = *types.NewPopulatedEUI64(r)
		out.DevEUI = *types.NewPopulatedEUI64(r)
		out.RejoinCnt = uint32(uint16(r.Uint32()))
	}
	return out
}

func NewPopulatedRejoinRequestPayload(r randyLorawan, easy bool) *RejoinRequestPayload {
	return NewPopulatedRejoinRequestPayloadType(r, RejoinType(r.Intn(3)))
}
func NewPopulatedMessage_RejoinRequestPayload(r randyLorawan) *Message_RejoinRequestPayload {
	return &Message_RejoinRequestPayload{NewPopulatedRejoinRequestPayload(r, false)}
}
func NewPopulatedMessage_RejoinRequestPayloadType(r randyLorawan, typ RejoinType) *Message_RejoinRequestPayload {
	return &Message_RejoinRequestPayload{NewPopulatedRejoinRequestPayloadType(r, typ)}
}

func macMICPayload(mhdr MHDR, fhdr FHDR, fPort byte, frmPayload []byte, isUplink bool) ([]byte, error) {
	b := make([]byte, 0, 4)
	b, err := PopulatorConfig.LoRaWAN.AppendMHDR(b, mhdr)
	if err != nil {
		return nil, err
	}
	if isUplink {
		b, err = PopulatorConfig.LoRaWAN.AppendFHDR(b, fhdr, false)
	} else {
		b, err = PopulatorConfig.LoRaWAN.AppendFHDR(b, fhdr, false)
	}
	if err != nil {
		return nil, err
	}
	b = append(b, fPort)
	b = append(b, frmPayload...)
	return b, nil
}

func NewPopulatedMessageUplink(r randyLorawan, sNwkSIntKey, fNwkSIntKey types.AES128Key, txDrIdx, txChIdx uint8, confirmed bool) *Message {
	out := &Message{}
	out.MHDR = *NewPopulatedMHDR(r, false)
	if confirmed {
		out.MHDR.MType = MType_CONFIRMED_UP
	} else {
		out.MHDR.MType = MType_UNCONFIRMED_UP
	}
	pld := NewPopulatedMessage_MACPayload(r)
	pld.MACPayload.FHDR.FCtrl = FCtrl{
		ADR:       r.Intn(2) == 0,
		ADRAckReq: r.Intn(2) == 0,
		ClassB:    r.Intn(2) == 0,
		Ack:       r.Intn(2) == 0,
	}

	b, err := macMICPayload(out.MHDR, pld.MACPayload.FHDR, uint8(pld.MACPayload.FPort), pld.MACPayload.FRMPayload, false)
	if err != nil {
		panic(fmt.Sprintf("failed to compute payload for MIC computation: %s", err))
	}
	var confFCnt uint32
	if pld.MACPayload.Ack {
		confFCnt = pld.MACPayload.FCnt % (1 << 16)
	}
	mic, err := PopulatorConfig.LoRaWAN.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey, confFCnt, txDrIdx, txChIdx, pld.MACPayload.DevAddr, pld.MACPayload.FCnt, b)
	if err != nil {
		panic(fmt.Sprintf("failed to compute MIC: %s", err))
	}
	out.MIC = mic[:]
	out.Payload = pld
	return out
}

func NewPopulatedMessageDownlink(r randyLorawan, sNwkSIntKey types.AES128Key, confirmed bool) *Message {
	out := &Message{}
	out.MHDR = *NewPopulatedMHDR(r, false)
	if confirmed {
		out.MHDR.MType = MType_CONFIRMED_DOWN
	} else {
		out.MHDR.MType = MType_UNCONFIRMED_DOWN
	}
	pld := NewPopulatedMessage_MACPayload(r)
	pld.MACPayload.FHDR.FCtrl = FCtrl{
		ADR:      r.Intn(2) == 0,
		FPending: r.Intn(2) == 0,
		Ack:      r.Intn(2) == 0,
	}
	b, err := macMICPayload(out.MHDR, pld.MACPayload.FHDR, uint8(pld.MACPayload.FPort), pld.MACPayload.FRMPayload, false)
	if err != nil {
		panic(fmt.Sprintf("failed to compute payload for MIC computation: %s", err))
	}
	mic, err := PopulatorConfig.LoRaWAN.ComputeDownlinkMIC(sNwkSIntKey, pld.MACPayload.DevAddr, 0, pld.MACPayload.FCnt, b)
	if err != nil {
		panic(fmt.Sprintf("failed to compute MIC: %s", err))
	}
	out.MIC = mic[:]
	out.Payload = pld
	return out
}

func NewPopulatedMessageJoinRequest(r randyLorawan) *Message {
	out := &Message{}
	out.MHDR = *NewPopulatedMHDR(r, false)
	out.MHDR.MType = MType_JOIN_REQUEST
	out.MIC = make([]byte, 4)
	for i := 0; i < 4; i++ {
		out.MIC[i] = byte(r.Intn(256))
	}
	pld := NewPopulatedMessage_JoinRequestPayload(r)
	pld.JoinRequestPayload = NewPopulatedJoinRequestPayload(r, false)
	out.Payload = pld
	return out
}

func NewPopulatedMessageJoinAccept(r randyLorawan, decrypted bool) *Message {
	out := &Message{}
	out.MHDR = *NewPopulatedMHDR(r, false)
	out.MHDR.MType = MType_JOIN_ACCEPT
	var pld *JoinAcceptPayload
	if decrypted {
		pld = NewPopulatedJoinAcceptPayload(r, false)
		out.MIC = make([]byte, 4)
		for i := 0; i < 4; i++ {
			out.MIC[i] = byte(r.Intn(256))
		}
		pld.Rx1DROffset %= 8
	} else {
		pld = &JoinAcceptPayload{
			Encrypted: []byte{42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42},
		}
	}
	out.Payload = &Message_JoinAcceptPayload{pld}
	return out
}

func NewPopulatedMessageRejoinRequest(r randyLorawan, typ RejoinType) *Message {
	out := &Message{}
	out.MHDR = *NewPopulatedMHDR(r, false)
	out.MHDR.MType = MType_REJOIN_REQUEST
	out.MIC = make([]byte, 4)
	for i := 0; i < 4; i++ {
		out.MIC[i] = byte(r.Intn(256))
	}
	out.Payload = NewPopulatedMessage_RejoinRequestPayloadType(r, typ)
	return out
}

// NewPopulatedMessage is used for compatibility with gogoproto, and in cases, where the
// contents of the message are not important. It's advised to use one of:
// - NewPopulatedMessageUplink
// - NewPopulatedMessageDownlink
// - NewPopulatedMessageJoinRequest
// - NewPopulatedMessageJoinAccept
// - NewPopulatedMessageRejoinRequest
func NewPopulatedMessage(r randyLorawan, easy bool) *Message {
	switch MType(r.Intn(7)) {
	case MType_UNCONFIRMED_UP:
		return NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(r.Intn(256)), uint8(r.Intn(256)), false)
	case MType_CONFIRMED_UP:
		return NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(r.Intn(256)), uint8(r.Intn(256)), false)
	case MType_UNCONFIRMED_DOWN:
		return NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), false)
	case MType_CONFIRMED_DOWN:
		return NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), false)
	case MType_JOIN_REQUEST:
		return NewPopulatedMessageJoinRequest(r)
	case MType_JOIN_ACCEPT:
		return NewPopulatedMessageJoinAccept(r, false)
	case MType_REJOIN_REQUEST:
		return NewPopulatedMessageRejoinRequest(r, RejoinType(r.Intn(3)))
	}
	panic("unreachable")
}
