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
	"math"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg MHDR) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if msg.MType > 7 {
		return nil, errExpectedLowerOrEqual("MType", 7)(msg.MType)
	}
	if msg.Major > 4 {
		return nil, errExpectedLowerOrEqual("Major", 4)(msg.Major)
	}
	return append(dst, byte(msg.MType)<<5|byte(msg.Major)), nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg MHDR) MarshalLoRaWAN() ([]byte, error) {
	return msg.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *MHDR) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return errExpectedLengthEncodedEqual("MHDR", 1)(len(b))
	}
	v := b[0]
	msg.MType = MType(v >> 5)
	msg.Major = Major(v & 3)
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg FCtrl) AppendLoRaWAN(dst []byte, isUplink bool, fOptsLen uint8) ([]byte, error) {
	if fOptsLen > 15 {
		return nil, errExpectedLowerOrEqual("FOptsLen", 15)(fOptsLen)
	}
	b := fOptsLen
	if msg.ADR {
		b |= 1 << 7
	}
	if msg.Ack {
		b |= 1 << 5
	}
	if isUplink {
		if msg.ADRAckReq {
			b |= 1 << 6
		}
		if msg.ClassB {
			b |= 1 << 4
		}
	} else {
		if msg.FPending {
			b |= 1 << 4
		}
	}
	return append(dst, b), nil
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *FCtrl) UnmarshalLoRaWAN(b []byte, isUplink bool) error {
	if len(b) != 1 {
		return errExpectedLengthEncodedEqual("FCtrl", 1)(len(b))
	}
	v := b[0]
	msg.ADR = v&(1<<7) > 0
	msg.Ack = v&(1<<5) > 0
	if isUplink {
		msg.ADRAckReq = v&(1<<6) > 0
		msg.ClassB = v&(1<<4) > 0
	} else {
		msg.FPending = v&(1<<4) > 0
	}
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg FHDR) AppendLoRaWAN(dst []byte, isUplink bool) ([]byte, error) {
	dst = appendReverse(dst, msg.DevAddr[:]...)
	fOptsLen := uint8(len(msg.FOpts))
	if fOptsLen > 15 {
		return nil, errExpectedLowerOrEqual("FOptsLen", 15)(fOptsLen)
	}
	dst, err := msg.FCtrl.AppendLoRaWAN(dst, isUplink, fOptsLen)
	if err != nil {
		return nil, errFailedEncoding("FCtrl").WithCause(err)
	}
	if msg.FCnt > math.MaxUint16 {
		return nil, errExpectedLowerOrEqual("FCnt", math.MaxUint16)(msg.FCnt)
	}
	dst = appendUint32(dst, msg.FCnt, 2)
	dst = append(dst, msg.FOpts...)
	return dst, nil
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *FHDR) UnmarshalLoRaWAN(b []byte, isUplink bool) error {
	n := len(b)
	if n < 7 || n > 23 {
		return errExpectedLengthEncodedBound("FHDR", 7, 23)(n)
	}
	copyReverse(msg.DevAddr[:], b[0:4])
	if err := msg.FCtrl.UnmarshalLoRaWAN(b[4:5], isUplink); err != nil {
		return errFailedDecoding("FCtrl").WithCause(err)
	}
	msg.FCnt = parseUint32(b[5:7])
	msg.FOpts = make([]byte, 0, n-7)
	msg.FOpts = append(msg.FOpts, b[7:n]...)
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg MACPayload) AppendLoRaWAN(dst []byte, isUplink bool) ([]byte, error) {
	dst, err := msg.FHDR.AppendLoRaWAN(dst, isUplink)
	if err != nil {
		return nil, errFailedEncoding("FHDR").WithCause(err)
	}
	if msg.FPort > math.MaxUint8 {
		return nil, errExpectedLowerOrEqual("FPort", math.MaxUint8)(msg.FPort)
	}
	if len(msg.FRMPayload) > 0 || msg.FPort != 0 {
		dst = append(dst, byte(msg.FPort))
	}
	dst = append(dst, msg.FRMPayload...)
	return dst, nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg MACPayload) MarshalLoRaWAN(isUplink bool) ([]byte, error) {
	return msg.AppendLoRaWAN(make([]byte, 0, 1), isUplink)
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *MACPayload) UnmarshalLoRaWAN(b []byte, isUplink bool) error {
	n := uint8(len(b))
	if n < 7 {
		return errExpectedLengthEqual("FHDR", 7)(n)
	}
	fOptsLen := b[4] & 0xf
	fhdrLen := fOptsLen + 7
	if n < fhdrLen {
		return errExpectedLengthEqual("MACPayload", fhdrLen)(n)
	}
	if err := msg.FHDR.UnmarshalLoRaWAN(b[0:fhdrLen], isUplink); err != nil {
		return errFailedDecoding("FHDR").WithCause(err)
	}

	fPortIdx := fhdrLen
	if n >= fPortIdx {
		msg.FPort = uint32(b[fPortIdx])

		frmPayloadIdx := fPortIdx + 1
		if n >= frmPayloadIdx {
			msg.FRMPayload = make([]byte, 0, n-frmPayloadIdx)
			msg.FRMPayload = append(msg.FRMPayload, b[frmPayloadIdx:]...)
		}
	}
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg DLSettings) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if msg.Rx1DROffset > 7 {
		return nil, errExpectedLowerOrEqual("Rx1DROffset", 7)(msg.Rx1DROffset)
	}
	if msg.Rx2DR > 15 {
		return nil, errExpectedLowerOrEqual("Rx2DR", 15)(msg.Rx2DR)
	}
	b := byte(msg.Rx2DR)
	b |= byte(msg.Rx1DROffset << 4)
	if msg.OptNeg {
		b |= (1 << 7)
	}
	return append(dst, b), nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg DLSettings) MarshalLoRaWAN() ([]byte, error) {
	return msg.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *DLSettings) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return errExpectedLengthEncodedEqual("DLSettings", 1)(len(b))
	}
	v := uint32(b[0])
	msg.OptNeg = (v >> 7) != 0
	msg.Rx1DROffset = (v >> 4) & 0x7
	msg.Rx2DR = DataRateIndex(v & 0xf)
	return nil
}

var (
	errTooManyCFListFrequencies = unexpectedValue(
		errors.DefineInvalidArgument("cflist_frequencies", "too many CFList frequencies: expected 15 or less", valueKey),
	)
	errMaxCFListFrequency = unexpectedValue(
		errors.DefineInvalidArgument("max_cflist_frequency", "expected CFList frequency to be less or equal to MaxUint24", valueKey),
	)
)

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg CFList) AppendLoRaWAN(dst []byte) ([]byte, error) {
	switch msg.Type {
	case 0:
		if len(msg.Freq) > 15 {
			return nil, errTooManyCFListFrequencies(len(msg.Freq))
		}
		for _, freq := range msg.Freq {
			if freq > maxUint24 {
				return nil, errMaxCFListFrequency(freq)
			}
			dst = appendUint32(dst, freq, 3)
		}
	case 1:
		n := len(msg.ChMasks)
		if n > 96 {
			return nil, errExpectedLengthLowerOrEqual("CFListChMasks", 96)(n)
		}
		for i := uint(0); i < uint(n); i += 8 {
			var b byte
			for j := uint(0); j < 8; j++ {
				if msg.ChMasks[i+j] {
					b |= (1 << j)
				}
			}
			dst = append(dst, b)
		}
		// fill remaining space with 0's
		for i := 0; i < 15-(n+7)/8; i++ {
			dst = append(dst, 0)
		}
	}
	dst = append(dst, byte(msg.Type))
	return dst, nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg CFList) MarshalLoRaWAN() ([]byte, error) {
	return msg.AppendLoRaWAN(make([]byte, 0, 16))
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *CFList) UnmarshalLoRaWAN(b []byte) error {
	n := len(b)
	if n != 16 {
		return errExpectedLengthEncodedEqual("CFList", 16)(n)
	}
	msg.Type = CFListType(b[15])
	switch msg.Type {
	case 0:
		msg.Freq = make([]uint32, 0, 5)
		for i := 0; i < 15; i += 3 {
			msg.Freq = append(msg.Freq, parseUint32(b[i:i+3]))
		}
	case 1:
		msg.ChMasks = make([]bool, 0, 96)
		for _, m := range b[:12] {
			msg.ChMasks = append(msg.ChMasks,
				m&1 > 0,
				m&(1<<1) > 0,
				m&(1<<2) > 0,
				m&(1<<3) > 0,
				m&(1<<4) > 0,
				m&(1<<5) > 0,
				m&(1<<6) > 0,
				m&(1<<7) > 0,
			)
		}
	default:
		return errUnknown("CFListType")(msg.Type.String())
	}
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg JoinAcceptPayload) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = appendReverse(dst, msg.JoinNonce[:]...)
	dst = appendReverse(dst, msg.NetID[:]...)
	dst = appendReverse(dst, msg.DevAddr[:]...)
	dst, err := msg.DLSettings.AppendLoRaWAN(dst)
	if err != nil {
		return nil, errFailedEncoding("DLSettings").WithCause(err)
	}
	if msg.RxDelay > math.MaxUint8 {
		return nil, errExpectedLowerOrEqual("RxDelay", math.MaxUint8)(msg.RxDelay)
	}
	dst = append(dst, byte(msg.RxDelay))
	if msg.GetCFList() != nil {
		dst, err = msg.CFList.AppendLoRaWAN(dst)
		if err != nil {
			return nil, errFailedEncoding("CFList").WithCause(err)
		}
	}
	return dst, nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg JoinAcceptPayload) MarshalLoRaWAN() ([]byte, error) {
	if msg.GetCFList() != nil {
		return msg.AppendLoRaWAN(make([]byte, 0, 28))
	}
	return msg.AppendLoRaWAN(make([]byte, 0, 12))
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *JoinAcceptPayload) UnmarshalLoRaWAN(b []byte) error {
	n := len(b)
	if n != 12 && n != 28 {
		return errExpectedLengthEncodedTwoChoices("JoinAcceptPayload", 12, 28)(n)
	}
	copyReverse(msg.JoinNonce[:], b[0:3])
	copyReverse(msg.NetID[:], b[3:6])
	copyReverse(msg.DevAddr[:], b[6:10])
	if err := msg.DLSettings.UnmarshalLoRaWAN(b[10:11]); err != nil {
		return errFailedDecoding("DLSettings").WithCause(err)
	}
	msg.RxDelay = uint32(b[11])

	if n == 12 {
		return nil
	}
	msg.CFList = &CFList{}
	if err := msg.CFList.UnmarshalLoRaWAN(b[12:]); err != nil {
		return errFailedDecoding("CFList").WithCause(err)
	}
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg JoinRequestPayload) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = appendReverse(dst, msg.JoinEUI[:]...)
	dst = appendReverse(dst, msg.DevEUI[:]...)
	dst = appendReverse(dst, msg.DevNonce[:]...)
	return dst, nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg JoinRequestPayload) MarshalLoRaWAN() ([]byte, error) {
	return msg.AppendLoRaWAN(make([]byte, 0, 18))
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *JoinRequestPayload) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 18 {
		return errExpectedLengthEncodedEqual("JoinRequestPayload", 18)(len(b))
	}
	copyReverse(msg.JoinEUI[:], b[0:8])
	copyReverse(msg.DevEUI[:], b[8:16])
	copyReverse(msg.DevNonce[:], b[16:18])
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg RejoinRequestPayload) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(msg.RejoinType))
	switch msg.RejoinType {
	case 0, 2:
		dst = appendReverse(dst, msg.NetID[:]...)
		dst = appendReverse(dst, msg.DevEUI[:]...)
	case 1:
		dst = appendReverse(dst, msg.JoinEUI[:]...)
		dst = appendReverse(dst, msg.DevEUI[:]...)
	default:
		return nil, errUnknown("RejoinType")(msg.RejoinType)
	}
	if msg.RejoinCnt > math.MaxUint16 {
		return nil, errExpectedLowerOrEqual("RJcount1", math.MaxUint16)(msg.RejoinCnt)
	}
	dst = appendUint32(dst, msg.RejoinCnt, 2)
	return dst, nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg RejoinRequestPayload) MarshalLoRaWAN() ([]byte, error) {
	if msg.RejoinType == 1 {
		return msg.AppendLoRaWAN(make([]byte, 0, 19))
	}
	return msg.AppendLoRaWAN(make([]byte, 0, 14))
}

var errEncryptedJoinAcceptPayloadLength = unexpectedValue(
	errors.DefineInvalidArgument("encrypted_joinacceptpayload_length", "encrypted JoinAcceptPayload should have length of 16 or 32", valueKey),
)

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
// If message type is a join-accept, only the Encrypted field is populated in
// the payload. You should decrypt that value and supply it to UnmarshalLoRaWAN of the payload struct to populate it. MIC should be set manually(i.e. msg.MIC = decrypted[len(decrypted)-4:]) after decryption.
func (msg *RejoinRequestPayload) UnmarshalLoRaWAN(b []byte) error {
	msg.RejoinType = RejoinType(b[0])
	switch msg.RejoinType {
	case 0, 2:
		if len(b) != 14 {
			return errExpectedLengthEqual("RejoinRequestPayload", 14)(len(b))
		}
		copyReverse(msg.NetID[:], b[1:4])
		copyReverse(msg.DevEUI[:], b[4:12])
		msg.RejoinCnt = parseUint32(b[12:14])
	case 1:
		if len(b) != 19 {
			return errExpectedLengthEqual("RejoinRequestPayload", 19)(len(b))
		}
		copyReverse(msg.JoinEUI[:], b[1:9])
		copyReverse(msg.DevEUI[:], b[9:17])
		msg.RejoinCnt = parseUint32(b[17:19])
	default:
		return errUnknown("RejoinType")(msg.RejoinType.String())
	}
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg Message) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst, err := msg.MHDR.AppendLoRaWAN(dst)
	if err != nil {
		return nil, errFailedEncoding("MHDR").WithCause(err)
	}
	switch msg.MType {
	case MType_CONFIRMED_DOWN, MType_UNCONFIRMED_DOWN:
		pld := msg.GetMACPayload()
		if pld == nil {
			return nil, errMissing("MACPayload")
		}
		dst, err = pld.AppendLoRaWAN(dst, false)
		if err != nil {
			return nil, errFailedEncoding("MACPayload").WithCause(err)
		}
	case MType_CONFIRMED_UP, MType_UNCONFIRMED_UP:
		pld := msg.GetMACPayload()
		if pld == nil {
			return nil, errMissing("MACPayload")
		}
		dst, err = pld.AppendLoRaWAN(dst, true)
		if err != nil {
			return nil, errFailedEncoding("uplink MACPayload").WithCause(err)
		}
	case MType_JOIN_REQUEST:
		pld := msg.GetJoinRequestPayload()
		if pld == nil {
			return nil, errMissing("JoinRequestPayload")
		}
		dst, err = pld.AppendLoRaWAN(dst)
		if err != nil {
			return nil, errFailedEncoding("JoinRequestPayload").WithCause(err)
		}
	case MType_REJOIN_REQUEST:
		pld := msg.GetRejoinRequestPayload()
		if pld == nil {
			return nil, errMissing("RejoinRequestPayload")
		}
		dst, err = pld.AppendLoRaWAN(dst)
		if err != nil {
			return nil, errFailedEncoding("RejoinRequestPayload").WithCause(err)
		}
	case MType_JOIN_ACCEPT:
		pld := msg.GetJoinAcceptPayload()
		if pld == nil {
			return nil, errMissing("JoinAcceptPayload")
		}
		n := len(pld.Encrypted)
		if n != 16 && n != 32 {
			return nil, errEncryptedJoinAcceptPayloadLength(n)
		}
		dst = append(dst, pld.Encrypted...)
	default:
		return nil, errUnknown("MType")(msg.MType.String())
	}
	dst = append(dst, msg.MIC...)
	return dst, nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg Message) MarshalLoRaWAN() ([]byte, error) {
	switch msg.MType {
	case MType_CONFIRMED_DOWN, MType_UNCONFIRMED_DOWN, MType_CONFIRMED_UP, MType_UNCONFIRMED_UP:
		// MHDR(1) + Payload(up to 250) + MIC(4)
		return msg.AppendLoRaWAN(make([]byte, 0, 255))
	case MType_JOIN_REQUEST:
		// MHDR(1) + Payload(18) + MIC(4)
		return msg.AppendLoRaWAN(make([]byte, 0, 23))
	case MType_REJOIN_REQUEST:
		// MHDR(1) + Payload(14|19) + MIC(4)
		return msg.AppendLoRaWAN(make([]byte, 0, 24))
	case MType_JOIN_ACCEPT:
		// MHDR(1) + Encrypted payload(16|32)
		return msg.AppendLoRaWAN(make([]byte, 0, 33))
	default:
		return nil, errUnknown("MType")(msg.MType.String())
	}
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
// If message type is a join-accept, only the Encrypted field is populated in
// the payload. You should decrypt that value and supply it to UnmarshalLoRaWAN of the payload struct to populate it. MIC should be set manually(i.e. msg.MIC = decrypted[len(decrypted)-4:]) after decryption.
func (msg *Message) UnmarshalLoRaWAN(b []byte) error {
	n := len(b)
	if n == 0 {
		return errMissing("PHYPayload")
	}
	if err := msg.MHDR.UnmarshalLoRaWAN(b[0:1]); err != nil {
		return errFailedDecoding("MHDR").WithCause(err)
	}
	switch msg.MHDR.MType {
	case MType_CONFIRMED_DOWN, MType_UNCONFIRMED_DOWN:
		pld := &MACPayload{}
		if err := pld.UnmarshalLoRaWAN(b[1:n-4], false); err != nil {
			return errFailedDecoding("MACPayload").WithCause(err)
		}
		msg.Payload = &Message_MACPayload{pld}
		msg.MIC = b[n-4:]
	case MType_CONFIRMED_UP, MType_UNCONFIRMED_UP:
		pld := &MACPayload{}
		if err := pld.UnmarshalLoRaWAN(b[1:n-4], true); err != nil {
			return errFailedDecoding("MACPayload").WithCause(err)
		}
		msg.Payload = &Message_MACPayload{pld}
		msg.MIC = b[n-4:]
	case MType_JOIN_REQUEST:
		if n != 23 {
			return errExpectedLengthEqual("JoinRequestPHYPayload", 23)(n)
		}
		pld := &JoinRequestPayload{}
		if err := pld.UnmarshalLoRaWAN(b[1:19]); err != nil {
			return errFailedDecoding("JoinRequestPayload").WithCause(err)
		}
		msg.Payload = &Message_JoinRequestPayload{pld}
		msg.MIC = b[19:]
	case MType_REJOIN_REQUEST:
		var micIdx int
		if b[1] == 1 {
			// MHDR(1) + Payload(19)
			micIdx = 20
		} else {
			// MHDR(1) + Payload(14)
			micIdx = 15
		}
		if n != micIdx+4 {
			return errExpectedLengthTwoChoices("RejoinRequestPHYPayload", 19, 24)(n)
		}
		pld := &RejoinRequestPayload{}
		if err := pld.UnmarshalLoRaWAN(b[1:micIdx]); err != nil {
			return errFailedDecoding("RejoinRequestPayload").WithCause(err)
		}
		msg.Payload = &Message_RejoinRequestPayload{pld}
		msg.MIC = b[micIdx:]
	case MType_JOIN_ACCEPT:
		if n != 17 && n != 33 {
			return errExpectedLengthTwoChoices("JoinAcceptPHYPayload", 17, 33)(n)
		}
		msg.Payload = &Message_JoinAcceptPayload{&JoinAcceptPayload{Encrypted: b[1:]}}
	default:
		return errUnknown("MType")(msg.MType.String())
	}
	return nil
}
