// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"math"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg MHDR) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if msg.MType > 7 {
		return nil, errors.Errorf("expected MType to be less or equal to 7, got %d", msg.MType)
	}
	if msg.Major > 4 {
		return nil, errors.Errorf("expected Major to be less or equal to 4, got %d", msg.Major)
	}
	return append(dst, byte(msg.MType)<<5|byte(msg.Major)), nil
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *MHDR) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("expected length of encoded MHDR to be equal to 1, got %d", len(b))
	}
	v := b[0]
	msg.MType = MType(v >> 5)
	msg.Major = Major(v & 3)
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg FCtrl) AppendLoRaWAN(dst []byte, isUplink bool, fOptsLen uint8) ([]byte, error) {
	if fOptsLen > 15 {
		return nil, errors.Errorf("expected fOptsLen be less or equal to 15, got %d", fOptsLen)
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
		return errors.Errorf("expected length of encoded FCtrl to be equal to 1, got %d", len(b))
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
	dst = append(dst, msg.DevAddr[:]...)
	fOptsLen := uint8(len(msg.FOpts))
	if fOptsLen > 15 {
		return nil, errors.Errorf("expected length of FOpts to be less or equal to 15, got %d", fOptsLen)
	}
	dst, err := msg.FCtrl.AppendLoRaWAN(dst, isUplink, fOptsLen)
	if err != nil {
		return nil, errors.NewWithCause("failed to encode FCtrl", err)
	}
	if msg.FCnt > math.MaxUint16 {
		return nil, errors.Errorf("expected FCnt to be less or equal to %d, got %d", math.MaxUint16, msg.FCnt)
	}
	dst = appendUint32(dst, msg.FCnt, 2)
	dst = append(dst, msg.FOpts...)
	return dst, nil
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *FHDR) UnmarshalLoRaWAN(b []byte, isUplink bool) error {
	n := len(b)
	if n < 7 || n > 23 {
		return errors.Errorf("expected length of encoded FHDR to be between 7 and 23, got %d", n)
	}
	copy(msg.DevAddr[:], b[0:4])
	if err := msg.FCtrl.UnmarshalLoRaWAN(b[4:5], isUplink); err != nil {
		return errors.NewWithCause("failed to decode FCtrl", err)
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
		return nil, errors.NewWithCause("failed to encode FHDR", err)
	}
	if msg.FPort > math.MaxUint8 {
		return nil, errors.Errorf("expected FPort to be less or equal to %d, got %d", math.MaxUint8, msg.FPort)
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
		return errors.Errorf("expected length of at least 7 to decode FHDR, got %d", n)
	}
	fOptsLen := b[4] & 0xf
	fhdrLen := fOptsLen + 7
	if n < fhdrLen {
		return errors.Errorf("expected length of at least %d bytes to decode FHDR(FOptsLen is %d), got %d.", fhdrLen, fOptsLen, n)
	}
	if err := msg.FHDR.UnmarshalLoRaWAN(b[0:fhdrLen], isUplink); err != nil {
		return errors.NewWithCause("failed to decode FHDR", err)
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
		return nil, errors.Errorf("expected Rx1DROffset to be less or equal to 7, got %d", msg.Rx1DROffset)
	}
	if msg.Rx2DR > 15 {
		return nil, errors.Errorf("expected Rx2DR to be less or equal to 15, got %d", msg.Rx2DR)
	}
	b := msg.Rx2DR
	b |= (msg.Rx1DROffset << 4)
	if msg.OptNeg {
		b |= (1 << 7)
	}
	return append(dst, byte(b)), nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg DLSettings) MarshalLoRaWAN() ([]byte, error) {
	return msg.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *DLSettings) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("expected length of encoded DLSettings to be equal to 1, got %d", len(b))
	}
	v := uint32(b[0])
	msg.OptNeg = (v >> 7) != 0
	msg.Rx1DROffset = (v >> 4) & 0x7
	msg.Rx2DR = v & 0xf
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg CFList) AppendLoRaWAN(dst []byte) ([]byte, error) {
	switch msg.Type {
	case 0:
		if len(msg.Freq) > 15 {
			return nil, errors.Errorf("expected length of frequencies to be less or equal to 15, got %d", len(msg.Freq))
		}
		for i, freq := range msg.Freq {
			if freq > maxUint24 {
				return nil, errors.Errorf("expected frequency nr. %d to be less or equal to %d, got %d", i, math.MaxUint8, freq)
			}
			dst = appendUint32(dst, freq, 3)
		}
	case 1:
		n := len(msg.ChMasks)
		if n > 96 {
			return nil, errors.Errorf("expected length of channel masks to be less or equal to 96, got %d", n)
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
		return errors.Errorf("expected length of encoded CFList to be equal to 16, got %d", n)
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
		return errors.Errorf("unknown CFListType %s", msg.Type)
	}
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg JoinAcceptPayload) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, msg.JoinNonce[:]...)
	dst = append(dst, msg.NetID[:]...)
	dst = append(dst, msg.DevAddr[:]...)
	dst, err := msg.DLSettings.AppendLoRaWAN(dst)
	if err != nil {
		return nil, errors.NewWithCause("failed to encode DLSettings", err)
	}
	if msg.RxDelay > math.MaxUint8 {
		return nil, errors.Errorf("expected RxDelay to be less or equal to %d, got %d", math.MaxUint8, msg.RxDelay)
	}
	dst = append(dst, byte(msg.RxDelay))
	if msg.GetCFList() != nil {
		dst, err = msg.CFList.AppendLoRaWAN(dst)
		if err != nil {
			return nil, errors.NewWithCause("failed to encode CFList", err)
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
		return errors.Errorf("expected length of encoded JoinAcceptPayload to be either 12 or 28, got %d", n)
	}
	copy(msg.JoinNonce[:], b[0:3])
	copy(msg.NetID[:], b[3:6])
	copy(msg.DevAddr[:], b[6:10])
	if err := msg.DLSettings.UnmarshalLoRaWAN(b[10:11]); err != nil {
		return errors.NewWithCause("failed to decode DLSettings", err)
	}
	msg.RxDelay = uint32(b[11])

	if n == 12 {
		return nil
	}
	msg.CFList = &CFList{}
	if err := msg.CFList.UnmarshalLoRaWAN(b[12:]); err != nil {
		return errors.NewWithCause("failed to decode CFList", err)
	}
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg JoinRequestPayload) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, msg.JoinEUI[:]...)
	dst = append(dst, msg.DevEUI[:]...)
	dst = append(dst, msg.DevNonce[:]...)
	return dst, nil
}

// MarshalLoRaWAN implements the encoding.LoRaWANMarshaler interface.
func (msg JoinRequestPayload) MarshalLoRaWAN() ([]byte, error) {
	return msg.AppendLoRaWAN(make([]byte, 0, 18))
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
func (msg *JoinRequestPayload) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 18 {
		return errors.Errorf("expected length of encoded Join-Request payload to be 18, got %d", len(b))
	}
	copy(msg.JoinEUI[:], b[0:8])
	copy(msg.DevEUI[:], b[8:16])
	copy(msg.DevNonce[:], b[16:18])
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg RejoinRequestPayload) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(msg.RejoinType))
	switch msg.RejoinType {
	case 0, 2:
		dst = append(dst, msg.NetID[:]...)
		dst = append(dst, msg.DevEUI[:]...)
	case 1:
		dst = append(dst, msg.JoinEUI[:]...)
		dst = append(dst, msg.DevEUI[:]...)
	default:
		return nil, errors.Errorf("unknown RejoinType %s", msg.RejoinType)
	}
	if msg.RejoinCnt > math.MaxUint16 {
		return nil, errors.Errorf("expected RJcount1 to be less or equal to %d, got %d", math.MaxUint16, msg.RejoinCnt)
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

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
// If message type is a Join-Accept, only the Encrypted field is populated in
// the payload. You should decrypt that value and supply it to UnmarshalLoRaWAN of the payload struct to populate it. MIC should be set manually(i.e. msg.MIC = decrypted[len(decrypted)-4:]) after decryption.
func (msg *RejoinRequestPayload) UnmarshalLoRaWAN(b []byte) error {
	msg.RejoinType = RejoinType(b[0])
	switch msg.RejoinType {
	case 0, 2:
		if len(b) != 14 {
			return errors.Errorf("expected payload length of 14 bytes, got %d", len(b))
		}
		copy(msg.NetID[:], b[1:4])
		copy(msg.DevEUI[:], b[4:12])
		msg.RejoinCnt = parseUint32(b[12:14])
	case 1:
		if len(b) != 19 {
			return errors.Errorf("expected payload length of 19 bytes, got %d", len(b))
		}
		copy(msg.JoinEUI[:], b[1:9])
		copy(msg.DevEUI[:], b[9:17])
		msg.RejoinCnt = parseUint32(b[17:19])
	default:
		return errors.Errorf("unknown RejoinType value: %s", msg.RejoinType)
	}
	return nil
}

// AppendLoRaWAN implements the encoding.LoRaWANAppender interface.
func (msg Message) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst, err := msg.MHDR.AppendLoRaWAN(dst)
	if err != nil {
		return nil, errors.NewWithCause("failed to encode MHDR", err)
	}
	switch msg.MType {
	case MType_CONFIRMED_DOWN, MType_UNCONFIRMED_DOWN:
		pld := msg.GetMACPayload()
		if pld == nil {
			return nil, errors.New("MACPayload is empty")
		}
		dst, err = pld.AppendLoRaWAN(dst, false)
		if err != nil {
			return nil, errors.NewWithCause("failed to encode downlink MACPayload", err)
		}
	case MType_CONFIRMED_UP, MType_UNCONFIRMED_UP:
		pld := msg.GetMACPayload()
		if pld == nil {
			return nil, errors.New("MACPayload is empty")
		}
		dst, err = pld.AppendLoRaWAN(dst, true)
		if err != nil {
			return nil, errors.NewWithCause("failed to encode uplink MACPayload", err)
		}
	case MType_JOIN_REQUEST:
		pld := msg.GetJoinRequestPayload()
		if pld == nil {
			return nil, errors.New("Join-Request payload is empty")
		}
		dst, err = pld.AppendLoRaWAN(dst)
		if err != nil {
			return nil, errors.NewWithCause("failed to encode Join-Request payload", err)
		}
	case MType_REJOIN_REQUEST:
		pld := msg.GetRejoinRequestPayload()
		if pld == nil {
			return nil, errors.New("RejoinRequestPayload is empty")
		}
		dst, err = pld.AppendLoRaWAN(dst)
		if err != nil {
			return nil, errors.NewWithCause("failed to encode Rejoin-Request payload", err)
		}
	case MType_JOIN_ACCEPT:
		pld := msg.GetJoinAcceptPayload()
		if pld == nil {
			return nil, errors.New("Join-Accept payload is empty")
		}
		n := len(pld.Encrypted)
		if n != 16 && n != 32 {
			return nil, errors.Errorf("expected length of encrypted Join-Accept payload to be equal to 16 or 32, got %d", n)
		}
		dst = append(dst, pld.Encrypted...)
	default:
		return nil, errors.Errorf("unknown MType %s", msg.MType)
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
		return nil, errors.Errorf("unknown MType %s", msg.MType)
	}
}

// UnmarshalLoRaWAN implements the encoding.LoRaWANUnmarshaler interface.
// If message type is a Join-Accept, only the Encrypted field is populated in
// the payload. You should decrypt that value and supply it to UnmarshalLoRaWAN of the payload struct to populate it. MIC should be set manually(i.e. msg.MIC = decrypted[len(decrypted)-4:]) after decryption.
func (msg *Message) UnmarshalLoRaWAN(b []byte) error {
	n := len(b)
	if n < 1 {
		return errors.Errorf("expected length of PHYPayload to be at least 1, got %d", n)
	}
	if err := msg.MHDR.UnmarshalLoRaWAN(b[0:1]); err != nil {
		return errors.NewWithCause("failed to decode MHDR", err)
	}
	switch msg.MHDR.MType {
	case MType_CONFIRMED_DOWN, MType_UNCONFIRMED_DOWN:
		pld := &MACPayload{}
		if err := pld.UnmarshalLoRaWAN(b[1:n-4], false); err != nil {
			return errors.NewWithCause("failed to decode downlink MACPayload", err)
		}
		msg.Payload = &Message_MACPayload{pld}
		msg.MIC = b[n-4:]
	case MType_CONFIRMED_UP, MType_UNCONFIRMED_UP:
		pld := &MACPayload{}
		if err := pld.UnmarshalLoRaWAN(b[1:n-4], true); err != nil {
			return errors.NewWithCause("failed to decode uplink MACPayload", err)
		}
		msg.Payload = &Message_MACPayload{pld}
		msg.MIC = b[n-4:]
	case MType_JOIN_REQUEST:
		if n != 23 {
			// MHDR(1) + Payload(18) + MIC(4)
			return errors.Errorf("expected length of Join-Request PHYPayload to be 23, got %d", n)
		}
		pld := &JoinRequestPayload{}
		if err := pld.UnmarshalLoRaWAN(b[1:19]); err != nil {
			return errors.NewWithCause("failed to decode Join-Request MACPayload", err)
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
			return errors.Errorf("expected length of Rejoin-Request PHYPayload to be %d, got %d", micIdx+4, n)
		}
		pld := &RejoinRequestPayload{}
		if err := pld.UnmarshalLoRaWAN(b[1:micIdx]); err != nil {
			return errors.NewWithCause("failed to decode Rejoin-Request MACPayload", err)
		}
		msg.Payload = &Message_RejoinRequestPayload{pld}
		msg.MIC = b[micIdx:]
	case MType_JOIN_ACCEPT:
		if n != 17 && n != 33 {
			// MHDR(1) + Payload(16|32)
			return errors.Errorf("expected length of Join-Accept PHYPayload to be equal to 17 or 33, got %d", n)
		}
		msg.Payload = &Message_JoinAcceptPayload{&JoinAcceptPayload{Encrypted: b[1:]}}
	default:
		return errors.Errorf("unknown MType %s", msg.MType)
	}
	return nil
}
