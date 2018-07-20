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

package ttnpb

import (
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

func NewPopulatedFHDR(r randyLorawan, easy bool) *FHDR {
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

func NewPopulatedTxSettings(r randyLorawan, easy bool) *TxSettings {
	out := &TxSettings{}
	switch r.Intn(2) {
	case 0:
		out.Modulation = Modulation_FSK
	case 1:
		out.Modulation = Modulation_LORA
	}
	out.Bandwidth = r.Uint32()
	out.SpreadingFactor = uint32(r.Intn(6) + 7)
	out.BitRate = r.Uint32()
	out.CodingRate = fmt.Sprintf("4/%d", r.Intn(4)+5)
	out.Frequency = uint64(r.Uint32())
	out.TxPower = r.Int31()
	if r.Intn(2) == 0 {
		out.TxPower *= -1
	}
	out.PolarizationInversion = r.Intn(2) == 0
	out.ChannelIndex = r.Uint32() % 255
	out.DataRateIndex = r.Uint32() % 8
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
	out.Rx2DR = uint32(r.Intn(16))
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
	out.RxDelay = uint32(r.Intn(256))
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
	b, err := mhdr.AppendLoRaWAN(b)
	if err != nil {
		return nil, err
	}
	b, err = fhdr.AppendLoRaWAN(b, isUplink)
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

	b, err := macMICPayload(out.MHDR, pld.MACPayload.FHDR, uint8(pld.MACPayload.FPort), pld.MACPayload.FRMPayload, true)
	if err != nil {
		panic(errors.NewWithCause(err, "failed to compute payload for MIC computation"))
	}
	var confFCnt uint32
	if pld.MACPayload.Ack {
		confFCnt = pld.MACPayload.FCnt % (1 << 16)
	}
	mic, err := crypto.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey, confFCnt, txDrIdx, txChIdx, pld.MACPayload.DevAddr, pld.MACPayload.FCnt, b)
	if err != nil {
		panic(errors.NewWithCause(err, "failed to compute MIC"))
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
		panic(errors.NewWithCause(err, "failed to compute payload for MIC computation"))
	}
	mic, err := crypto.ComputeDownlinkMIC(sNwkSIntKey, pld.MACPayload.DevAddr, pld.MACPayload.FCnt, b)
	if err != nil {
		panic(errors.NewWithCause(err, "failed to compute MIC"))
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
		return NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(r.Intn(256)), uint8(r.Intn(256)), true)
	case MType_UNCONFIRMED_DOWN:
		return NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), false)
	case MType_CONFIRMED_DOWN:
		return NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), true)
	case MType_JOIN_REQUEST:
		return NewPopulatedMessageJoinRequest(r)
	case MType_JOIN_ACCEPT:
		return NewPopulatedMessageJoinAccept(r, true)
	case MType_REJOIN_REQUEST:
		return NewPopulatedMessageRejoinRequest(r, RejoinType(r.Intn(3)))
	}
	panic("unreachable")
}
