// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// MACCommandIdentifier_uplinkLength gives the payload length of a MAC command in the uplink direction.
// If the MAC command is not defined for this direction, it should not be in this list.
var MACCommandIdentifier_uplinkLength = map[MACCommandIdentifier]uint{
	CID_RESET:              1,
	CID_LINK_CHECK:         0,
	CID_LINK_ADR:           1,
	CID_DUTY_CYCLE:         0,
	CID_RX_PARAM_SETUP:     1,
	CID_DEV_STATUS:         2,
	CID_NEW_CHANNEL:        1,
	CID_RX_TIMING_SETUP:    0,
	CID_TX_PARAM_SETUP:     0,
	CID_DL_CHANNEL:         1,
	CID_REKEY:              1,
	CID_ADR_PARAM_SETUP:    0,
	CID_DEVICE_TIME:        0,
	CID_REJOIN_PARAM_SETUP: 1,
	CID_PING_SLOT_INFO:     1,
	CID_PING_SLOT_CHANNEL:  1,
	CID_BEACON_TIMING:      0,
	CID_BEACON_FREQ:        1,
	CID_DEVICE_MODE:        1,
}

// MACCommandIdentifier_downlinkLength gives the payload length of a MAC command in the downlink direction.
// If the MAC command is not defined for this direction, it should not be in this list.
var MACCommandIdentifier_downlinkLength = map[MACCommandIdentifier]uint{
	CID_RESET:              1,
	CID_LINK_CHECK:         2,
	CID_LINK_ADR:           4,
	CID_DUTY_CYCLE:         1,
	CID_RX_PARAM_SETUP:     4,
	CID_DEV_STATUS:         0,
	CID_NEW_CHANNEL:        5,
	CID_RX_TIMING_SETUP:    1,
	CID_TX_PARAM_SETUP:     1,
	CID_DL_CHANNEL:         4,
	CID_REKEY:              1,
	CID_ADR_PARAM_SETUP:    1,
	CID_DEVICE_TIME:        5,
	CID_FORCE_REJOIN:       2,
	CID_REJOIN_PARAM_SETUP: 1,
	CID_PING_SLOT_CHANNEL:  4,
	CID_BEACON_TIMING:      3,
	CID_BEACON_FREQ:        3,
	CID_DEVICE_MODE:        1,
}

// MACCommands is a slice of MAC Commands
type MACCommands []*MACCommand

// UnmarshalLoRaWAN unmarshals MAC commands from their LoRaWAN representation.
func (m *MACCommands) UnmarshalLoRaWAN(b []byte, isUplink bool) error {
	for len(b) > 0 {
		cid := MACCommandIdentifier(b[0])
		var rawPayload []byte
		cmd := &MACCommand{}
		var macPayload interface {
			UnmarshalLoRaWAN(b []byte) error
		}
		if cid >= 0x80 {
			cmd.Payload = &MACCommand_Proprietary_{Proprietary: &MACCommand_Proprietary{
				CID:        cid,
				RawPayload: b[1:],
			}}
			*m = append(*m, cmd)
			break
		} else if isUplink {
			payloadLen, ok := MACCommandIdentifier_uplinkLength[cid]
			if !ok {
				return errors.Errorf("Unknown Uplink MAC command with CID 0x%x", cid)
			}
			if len(b) < int(payloadLen)+1 {
				return errors.Errorf("Expected length of Uplink %s payload to be %d, got %d", cid, payloadLen-1, len(b)-1)
			}
			rawPayload = b[:payloadLen+1]
			b = b[payloadLen+1:]
			if payloadLen == 0 {
				cmd.Payload = &MACCommand_CID{CID: cid}
				*m = append(*m, cmd)
				continue
			}
			switch cid {
			case CID_RESET:
				pld := &MACCommand_ResetInd{}
				macPayload = pld
				cmd.Payload = &MACCommand_ResetInd_{ResetInd: pld}
			case CID_LINK_ADR:
				pld := &MACCommand_LinkADRAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_LinkAdrAns{LinkAdrAns: pld}
			case CID_RX_PARAM_SETUP:
				pld := &MACCommand_RxParamSetupAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_RxParamSetupAns_{RxParamSetupAns: pld}
			case CID_DEV_STATUS:
				pld := &MACCommand_DevStatusAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_DevStatusAns_{DevStatusAns: pld}
			case CID_NEW_CHANNEL:
				pld := &MACCommand_NewChannelAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_NewChannelAns_{NewChannelAns: pld}
			case CID_DL_CHANNEL:
				pld := &MACCommand_DLChannelAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_DlChannelAns{DlChannelAns: pld}
			case CID_REKEY:
				pld := &MACCommand_RekeyInd{}
				macPayload = pld
				cmd.Payload = &MACCommand_RekeyInd_{RekeyInd: pld}
			case CID_REJOIN_PARAM_SETUP:
				pld := &MACCommand_RejoinParamSetupAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_RejoinParamSetupAns_{RejoinParamSetupAns: pld}
			case CID_PING_SLOT_INFO:
				pld := &MACCommand_PingSlotInfoReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_PingSlotInfoReq_{PingSlotInfoReq: pld}
			case CID_PING_SLOT_CHANNEL:
				pld := &MACCommand_PingSlotChannelAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_PingSlotChannelAns_{PingSlotChannelAns: pld}
			case CID_BEACON_FREQ:
				pld := &MACCommand_BeaconFreqAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_BeaconFreqAns_{BeaconFreqAns: pld}
			case CID_DEVICE_MODE:
				pld := &MACCommand_DeviceModeInd{}
				macPayload = pld
				cmd.Payload = &MACCommand_DeviceModeInd_{DeviceModeInd: pld}
			}
		} else {
			payloadLen, ok := MACCommandIdentifier_downlinkLength[cid]
			if !ok {
				return errors.New("Unknown Downlink MAC command")
			}
			if len(b) < int(payloadLen)+1 {
				return errors.Errorf("Expected length of Downlink %s payload to be %d, got %d", cid, payloadLen-1, len(b)-1)
			}
			rawPayload = b[:payloadLen+1]
			b = b[payloadLen+1:]
			if payloadLen == 0 {
				cmd.Payload = &MACCommand_CID{CID: cid}
				*m = append(*m, cmd)
				continue
			}
			switch cid {
			case CID_RESET:
				pld := &MACCommand_ResetConf{}
				macPayload = pld
				cmd.Payload = &MACCommand_ResetConf_{ResetConf: pld}
			case CID_LINK_CHECK:
				pld := &MACCommand_LinkCheckAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_LinkCheckAns_{LinkCheckAns: pld}
			case CID_LINK_ADR:
				pld := &MACCommand_LinkADRReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_LinkAdrReq{LinkAdrReq: pld}
			case CID_DUTY_CYCLE:
				pld := &MACCommand_DutyCycleReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_DutyCycleReq_{DutyCycleReq: pld}
			case CID_RX_PARAM_SETUP:
				pld := &MACCommand_RxParamSetupReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_RxParamSetupReq_{RxParamSetupReq: pld}
			case CID_NEW_CHANNEL:
				pld := &MACCommand_NewChannelReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_NewChannelReq_{NewChannelReq: pld}
			case CID_RX_TIMING_SETUP:
				pld := &MACCommand_RxTimingSetupReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_RxTimingSetupReq_{RxTimingSetupReq: pld}
			case CID_TX_PARAM_SETUP:
				pld := &MACCommand_TxParamSetupReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_TxParamSetupReq_{TxParamSetupReq: pld}
			case CID_DL_CHANNEL:
				pld := &MACCommand_DLChannelReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_DlChannelReq{DlChannelReq: pld}
			case CID_REKEY:
				pld := &MACCommand_RekeyConf{}
				macPayload = pld
				cmd.Payload = &MACCommand_RekeyConf_{RekeyConf: pld}
			case CID_ADR_PARAM_SETUP:
				pld := &MACCommand_ADRParamSetupReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_AdrParamSetupReq{AdrParamSetupReq: pld}
			case CID_DEVICE_TIME:
				pld := &MACCommand_DeviceTimeAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_DeviceTimeAns_{DeviceTimeAns: pld}
			case CID_FORCE_REJOIN:
				pld := &MACCommand_ForceRejoinReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_ForceRejoinReq_{ForceRejoinReq: pld}
			case CID_REJOIN_PARAM_SETUP:
				pld := &MACCommand_RejoinParamSetupReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_RejoinParamSetupReq_{RejoinParamSetupReq: pld}
			case CID_PING_SLOT_CHANNEL:
				pld := &MACCommand_PingSlotChannelReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_PingSlotChannelReq_{PingSlotChannelReq: pld}
			case CID_BEACON_TIMING:
				pld := &MACCommand_BeaconTimingAns{}
				macPayload = pld
				cmd.Payload = &MACCommand_BeaconTimingAns_{BeaconTimingAns: pld}
			case CID_BEACON_FREQ:
				pld := &MACCommand_BeaconFreqReq{}
				macPayload = pld
				cmd.Payload = &MACCommand_BeaconFreqReq_{BeaconFreqReq: pld}
			case CID_DEVICE_MODE:
				pld := &MACCommand_DeviceModeConf{}
				macPayload = pld
				cmd.Payload = &MACCommand_DeviceModeConf_{DeviceModeConf: pld}
			}
		}
		if macPayload == nil {
			panic(errors.Errorf("No payload for 0x%X MAC command", byte(cid)))
		}
		err := macPayload.UnmarshalLoRaWAN(rawPayload)
		if err != nil {
			return errors.NewWithCause("Could not unmarshal MAC command payload", err)
		}
		*m = append(*m, cmd)
	}
	return nil
}

// AppendLoRaWAN appends the LoRaWAN representation of the MAC commands to dst and returns the extended buffer.
func (m *MACCommands) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var err error
	for _, cmd := range *m {
		switch x := cmd.Payload.(type) {
		case *MACCommand_Proprietary_:
			dst, err = x.Proprietary.AppendLoRaWAN(dst)
		case *MACCommand_ResetInd_:
			dst, err = x.ResetInd.AppendLoRaWAN(dst)
		case *MACCommand_ResetConf_:
			dst, err = x.ResetConf.AppendLoRaWAN(dst)
		case *MACCommand_LinkCheckAns_:
			dst, err = x.LinkCheckAns.AppendLoRaWAN(dst)
		case *MACCommand_LinkAdrReq:
			dst, err = x.LinkAdrReq.AppendLoRaWAN(dst)
		case *MACCommand_LinkAdrAns:
			dst, err = x.LinkAdrAns.AppendLoRaWAN(dst)
		case *MACCommand_DutyCycleReq_:
			dst, err = x.DutyCycleReq.AppendLoRaWAN(dst)
		case *MACCommand_RxParamSetupReq_:
			dst, err = x.RxParamSetupReq.AppendLoRaWAN(dst)
		case *MACCommand_RxParamSetupAns_:
			dst, err = x.RxParamSetupAns.AppendLoRaWAN(dst)
		case *MACCommand_DevStatusAns_:
			dst, err = x.DevStatusAns.AppendLoRaWAN(dst)
		case *MACCommand_NewChannelReq_:
			dst, err = x.NewChannelReq.AppendLoRaWAN(dst)
		case *MACCommand_NewChannelAns_:
			dst, err = x.NewChannelAns.AppendLoRaWAN(dst)
		case *MACCommand_DlChannelReq:
			dst, err = x.DlChannelReq.AppendLoRaWAN(dst)
		case *MACCommand_DlChannelAns:
			dst, err = x.DlChannelAns.AppendLoRaWAN(dst)
		case *MACCommand_RxTimingSetupReq_:
			dst, err = x.RxTimingSetupReq.AppendLoRaWAN(dst)
		case *MACCommand_TxParamSetupReq_:
			dst, err = x.TxParamSetupReq.AppendLoRaWAN(dst)
		case *MACCommand_RekeyInd_:
			dst, err = x.RekeyInd.AppendLoRaWAN(dst)
		case *MACCommand_RekeyConf_:
			dst, err = x.RekeyConf.AppendLoRaWAN(dst)
		case *MACCommand_AdrParamSetupReq:
			dst, err = x.AdrParamSetupReq.AppendLoRaWAN(dst)
		case *MACCommand_DeviceTimeAns_:
			dst, err = x.DeviceTimeAns.AppendLoRaWAN(dst)
		case *MACCommand_ForceRejoinReq_:
			dst, err = x.ForceRejoinReq.AppendLoRaWAN(dst)
		case *MACCommand_RejoinParamSetupReq_:
			dst, err = x.RejoinParamSetupReq.AppendLoRaWAN(dst)
		case *MACCommand_RejoinParamSetupAns_:
			dst, err = x.RejoinParamSetupAns.AppendLoRaWAN(dst)
		case *MACCommand_PingSlotInfoReq_:
			dst, err = x.PingSlotInfoReq.AppendLoRaWAN(dst)
		case *MACCommand_PingSlotChannelReq_:
			dst, err = x.PingSlotChannelReq.AppendLoRaWAN(dst)
		case *MACCommand_PingSlotChannelAns_:
			dst, err = x.PingSlotChannelAns.AppendLoRaWAN(dst)
		case *MACCommand_BeaconTimingAns_:
			dst, err = x.BeaconTimingAns.AppendLoRaWAN(dst)
		case *MACCommand_BeaconFreqReq_:
			dst, err = x.BeaconFreqReq.AppendLoRaWAN(dst)
		case *MACCommand_BeaconFreqAns_:
			dst, err = x.BeaconFreqAns.AppendLoRaWAN(dst)
		case *MACCommand_DeviceModeInd_:
			dst, err = x.DeviceModeInd.AppendLoRaWAN(dst)
		case *MACCommand_DeviceModeConf_:
			dst, err = x.DeviceModeConf.AppendLoRaWAN(dst)
		case nil:
		default:
			return nil, errors.Errorf("MACCommand.Payload has unexpected type %T", x)
		}
		if err != nil {
			return nil, err
		}
	}
	return dst, nil
}

// MarshalLoRaWAN marshals the LoRaWAN representation of the MAC commands
func (m *MACCommands) MarshalLoRaWAN() ([]byte, error) {
	if len(*m) == 0 {
		return []byte{}, nil // avoid allocation as most messages don't have MAC commands
	}
	return m.AppendLoRaWAN(make([]byte, 0, 15)) // most messages have MAC commands in the header, then 15 is enough
}

// AppendLoRaWAN appends the marshaled Proprietary CID and payload to the slice.
func (m *MACCommand_Proprietary) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(m.CID))
	dst = append(dst, m.RawPayload...)
	return dst, nil
}

// MarshalLoRaWAN marshals the Proprietary CID and payload.
func (m *MACCommand_Proprietary) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1+len(m.RawPayload)))
}

// UnmarshalLoRaWAN unmarshals the Proprietary CID and payload.
func (m *MACCommand_Proprietary) UnmarshalLoRaWAN(b []byte) error {
	n := uint8(len(b))
	if n < 1 {
		return errors.Errorf("expected length of encoded Proprietary command to be at least 1, got %d", n)
	}
	m.CID = MACCommandIdentifier(b[0])
	m.RawPayload = b[1:]
	return nil
}

func checkMACCommand(cid MACCommandIdentifier, name string, b []byte, expectedPayload uint8) error {
	n := uint8(len(b))
	if n != expectedPayload+1 {
		return errors.Errorf("expected length of encoded %s to be %d, got %d", name, expectedPayload+1, n)
	}
	if b[0] != byte(cid) {
		return errors.Errorf("expected CID of encoded %s payload to be 0x%X, got 0x%X", name, byte(cid), b[0])
	}
	return nil
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func setBit(b byte, i uint8) byte {
	b |= 1 << i
	return b
}

func getBit(b byte, i uint8) bool {
	return (b>>i)&1 == 1
}

// AppendLoRaWAN appends the marshaled ResetInd CID and payload to the slice.
func (m *MACCommand_ResetInd) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if m.MinorVersion > 15 {
		return nil, errors.Errorf("expected MinorVersion to be less or equal to 15, got %d", m.MinorVersion)
	}
	dst = append(dst, byte(CID_RESET))
	dst = append(dst, byte(m.MinorVersion))
	return dst, nil
}

// MarshalLoRaWAN marshals the ResetInd CID and payload.
func (m *MACCommand_ResetInd) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 2))
}

// UnmarshalLoRaWAN unmarshals the ResetInd CID and payload.
func (m *MACCommand_ResetInd) UnmarshalLoRaWAN(b []byte) error {
	if err := checkMACCommand(CID_RESET, "ResetInd", b, 1); err != nil {
		return err
	}
	m.MinorVersion = uint32(b[1] & 0xf)
	return nil
}

// AppendLoRaWAN appends the marshaled ResetConf CID and payload to the slice.
func (m *MACCommand_ResetConf) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if m.MinorVersion > 15 {
		return nil, errors.Errorf("expected MinorVersion to be less or equal to 15, got %d", m.MinorVersion)
	}
	dst = append(dst, byte(CID_RESET))
	dst = append(dst, byte(m.MinorVersion))
	return dst, nil
}

// MarshalLoRaWAN marshals the ResetConf CID and payload.
func (m *MACCommand_ResetConf) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 2))
}

// UnmarshalLoRaWAN unmarshals the ResetConf CID and payload.
func (m *MACCommand_ResetConf) UnmarshalLoRaWAN(b []byte) error {
	if err := checkMACCommand(CID_RESET, "ResetConf", b, 1); err != nil {
		return err
	}
	m.MinorVersion = uint32(b[1] & 0xf)
	return nil
}

// AppendLoRaWAN appends the marshaled LinkCheckAns CID and payload to the slice.
func (m *MACCommand_LinkCheckAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if m.Margin > 254 {
		return nil, errors.Errorf("expected Margin to be less or equal to 254, got %d", m.Margin)
	}
	if m.GatewayCount > 255 {
		return nil, errors.Errorf("expected GatewayCount to be less or equal to 255, got %d", m.GatewayCount)
	}
	dst = append(dst, byte(CID_LINK_CHECK))
	dst = append(dst, byte(m.Margin), byte(m.GatewayCount))
	return dst, nil
}

// MarshalLoRaWAN marshals the LinkCheckAns CID and payload.
func (m *MACCommand_LinkCheckAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 3))
}

// UnmarshalLoRaWAN unmarshals the LinkCheckAns CID and payload.
func (m *MACCommand_LinkCheckAns) UnmarshalLoRaWAN(b []byte) error {
	if err := checkMACCommand(CID_LINK_CHECK, "LinkCheckAns", b, 2); err != nil {
		return err
	}
	m.Margin = uint32(b[1])
	m.GatewayCount = uint32(b[2])
	return nil
}

// AppendLoRaWAN appends the marshaled LinkADRReq CID and payload to the slice.
func (m *MACCommand_LinkADRReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_LINK_ADR))
	if m.DataRateIndex > 15 {
		return nil, errors.Errorf("expected DataRateIndex to be less or equal to 15, got %d", m.DataRateIndex)
	}
	if m.TxPowerIndex > 15 {
		return nil, errors.Errorf("expected TxPowerIndex to be less or equal to 15, got %d", m.TxPowerIndex)
	}
	if len(m.ChannelMask) > 16 {
		return nil, errors.Errorf("expected ChannelMask to be shorter or equal to 16, got %d", len(m.ChannelMask))
	}
	if m.ChannelMaskControl > 7 {
		return nil, errors.Errorf("expected ChannelMaskControl to be less or equal to 7, got %d", m.ChannelMaskControl)
	}
	if m.NbTrans > 15 {
		return nil, errors.Errorf("expected NbTrans to be less or equal to 15, got %d", m.NbTrans)
	}
	dst = append(dst, byte((m.DataRateIndex&0xf)<<4)^byte(m.TxPowerIndex&0xf))
	chMask := make([]byte, 2)
	for i := uint8(0); i < 16 && i < uint8(len(m.ChannelMask)); i++ {
		chMask[i/8] = chMask[i/8] ^ boolToByte(m.ChannelMask[i])<<(i%8)
	}
	dst = append(dst, chMask...)
	dst = append(dst, byte((m.ChannelMaskControl&0x7)<<4)^byte(m.NbTrans&0xf))
	return dst, nil
}

// MarshalLoRaWAN marshals the LinkADRReq CID and payload.
func (m *MACCommand_LinkADRReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 5))
}

// UnmarshalLoRaWAN unmarshals the LinkADRReq CID and payload.
func (m *MACCommand_LinkADRReq) UnmarshalLoRaWAN(b []byte) error {
	if err := checkMACCommand(CID_LINK_ADR, "LinkADRReq", b, 4); err != nil {
		return err
	}
	m.DataRateIndex = uint32(b[1] >> 4)
	m.TxPowerIndex = uint32(b[1] & 0xf)
	var chMask [16]bool
	for i := uint8(0); i < 16; i++ {
		if (b[2+i/8]>>(i%8))&1 == 1 {
			chMask[i] = true
		}
	}
	m.ChannelMask = chMask[:]
	m.ChannelMaskControl = uint32((b[4] >> 4) & 0x7)
	m.NbTrans = uint32(b[4] & 0xf)
	return nil
}

// AppendLoRaWAN appends the marshaled LinkADRAns CID and payload to the slice.
func (m *MACCommand_LinkADRAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_LINK_ADR))
	var status byte
	if m.ChannelMaskAck {
		status = setBit(status, 0)
	}
	if m.DataRateIndexAck {
		status = setBit(status, 1)
	}
	if m.TxPowerIndexAck {
		status = setBit(status, 2)
	}
	dst = append(dst, status)
	return dst, nil
}

// MarshalLoRaWAN marshals the LinkADRAns CID and payload.
func (m *MACCommand_LinkADRAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 2))
}

// UnmarshalLoRaWAN unmarshals the LinkADRAns CID and payload.
func (m *MACCommand_LinkADRAns) UnmarshalLoRaWAN(b []byte) error {
	if err := checkMACCommand(CID_LINK_ADR, "LinkADRAns", b, 1); err != nil {
		return err
	}
	m.ChannelMaskAck = getBit(b[1], 0)
	m.DataRateIndexAck = getBit(b[1], 1)
	m.TxPowerIndexAck = getBit(b[1], 2)
	return nil
}

// AppendLoRaWAN appends the marshaled DutyCycleReq CID and payload to the slice.
func (m *MACCommand_DutyCycleReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_DUTY_CYCLE))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the DutyCycleReq CID and payload.
func (m *MACCommand_DutyCycleReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the DutyCycleReq CID and payload.
func (m *MACCommand_DutyCycleReq) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled RxParamSetupReq CID and payload to the slice.
func (m *MACCommand_RxParamSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_RX_PARAM_SETUP))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the RxParamSetupReq CID and payload.
func (m *MACCommand_RxParamSetupReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the RxParamSetupReq CID and payload.
func (m *MACCommand_RxParamSetupReq) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled RxParamSetupAns CID and payload to the slice.
func (m *MACCommand_RxParamSetupAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_RX_PARAM_SETUP))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the RxParamSetupAns CID and payload.
func (m *MACCommand_RxParamSetupAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the RxParamSetupAns CID and payload.
func (m *MACCommand_RxParamSetupAns) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled DevStatusAns CID and payload to the slice.
func (m *MACCommand_DevStatusAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_DEV_STATUS))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the DevStatusAns CID and payload.
func (m *MACCommand_DevStatusAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the DevStatusAns CID and payload.
func (m *MACCommand_DevStatusAns) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled NewChannelReq CID and payload to the slice.
func (m *MACCommand_NewChannelReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_NEW_CHANNEL))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the NewChannelReq CID and payload.
func (m *MACCommand_NewChannelReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the NewChannelReq CID and payload.
func (m *MACCommand_NewChannelReq) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled NewChannelAns CID and payload to the slice.
func (m *MACCommand_NewChannelAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_NEW_CHANNEL))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the NewChannelAns CID and payload.
func (m *MACCommand_NewChannelAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the NewChannelAns CID and payload.
func (m *MACCommand_NewChannelAns) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled DLChannelReq CID and payload to the slice.
func (m *MACCommand_DLChannelReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_DL_CHANNEL))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the DLChannelReq CID and payload.
func (m *MACCommand_DLChannelReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the DLChannelReq CID and payload.
func (m *MACCommand_DLChannelReq) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled DLChannelAns CID and payload to the slice.
func (m *MACCommand_DLChannelAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_DL_CHANNEL))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the DLChannelAns CID and payload.
func (m *MACCommand_DLChannelAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the DLChannelAns CID and payload.
func (m *MACCommand_DLChannelAns) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled RxTimingSetupReq CID and payload to the slice.
func (m *MACCommand_RxTimingSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_RX_TIMING_SETUP))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the RxTimingSetupReq CID and payload.
func (m *MACCommand_RxTimingSetupReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the RxTimingSetupReq CID and payload.
func (m *MACCommand_RxTimingSetupReq) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled TxParamSetupReq CID and payload to the slice.
func (m *MACCommand_TxParamSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_TX_PARAM_SETUP))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the TxParamSetupReq CID and payload.
func (m *MACCommand_TxParamSetupReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1)) // TODO
}

// UnmarshalLoRaWAN unmarshals the TxParamSetupReq CID and payload.
func (m *MACCommand_TxParamSetupReq) UnmarshalLoRaWAN(b []byte) error {
	return nil // TODO
}

// AppendLoRaWAN appends the marshaled RekeyInd CID and payload to the slice.
func (m *MACCommand_RekeyInd) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if m.MinorVersion > 15 {
		return nil, errors.Errorf("expected MinorVersion to be less or equal to 15, got %d", m.MinorVersion)
	}
	dst = append(dst, byte(CID_REKEY))
	dst = append(dst, byte(m.MinorVersion))
	return dst, nil
}

// MarshalLoRaWAN marshals the RekeyInd CID and payload.
func (m *MACCommand_RekeyInd) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 2))
}

// UnmarshalLoRaWAN unmarshals the RekeyInd CID and payload.
func (m *MACCommand_RekeyInd) UnmarshalLoRaWAN(b []byte) error {
	if err := checkMACCommand(CID_REKEY, "RekeyInd", b, 1); err != nil {
		return err
	}
	m.MinorVersion = uint32(b[1] & 0xf)
	return nil
}

// AppendLoRaWAN appends the marshaled RekeyConf CID and payload to the slice.
func (m *MACCommand_RekeyConf) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if m.MinorVersion > 15 {
		return nil, errors.Errorf("expected MinorVersion to be less or equal to 15, got %d", m.MinorVersion)
	}
	dst = append(dst, byte(CID_REKEY))
	dst = append(dst, byte(m.MinorVersion))
	return dst, nil
}

// MarshalLoRaWAN marshals the RekeyConf CID and payload.
func (m *MACCommand_RekeyConf) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 2))
}

// UnmarshalLoRaWAN unmarshals the RekeyConf CID and payload.
func (m *MACCommand_RekeyConf) UnmarshalLoRaWAN(b []byte) error {
	if err := checkMACCommand(CID_REKEY, "RekeyConf", b, 1); err != nil {
		return err
	}
	m.MinorVersion = uint32(b[1] & 0xf)
	return nil
}

// AppendLoRaWAN appends the marshaled ADRParamSetupReq CID and payload to the slice.
func (m *MACCommand_ADRParamSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_ADR_PARAM_SETUP))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the ADRParamSetupReq CID and payload.
func (m *MACCommand_ADRParamSetupReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the ADRParamSetupReq CID and payload.
func (m *MACCommand_ADRParamSetupReq) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled DeviceTimeAns CID and payload to the slice.
func (m *MACCommand_DeviceTimeAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_DEVICE_TIME))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the DeviceTimeAns CID and payload.
func (m *MACCommand_DeviceTimeAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the DeviceTimeAns CID and payload.
func (m *MACCommand_DeviceTimeAns) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled ForceRejoinReq CID and payload to the slice.
func (m *MACCommand_ForceRejoinReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_FORCE_REJOIN))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the ForceRejoinReq CID and payload.
func (m *MACCommand_ForceRejoinReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the ForceRejoinReq CID and payload.
func (m *MACCommand_ForceRejoinReq) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled RejoinParamSetupReq CID and payload to the slice.
func (m *MACCommand_RejoinParamSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_REJOIN_PARAM_SETUP))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the RejoinParamSetupReq CID and payload.
func (m *MACCommand_RejoinParamSetupReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the RejoinParamSetupReq CID and payload.
func (m *MACCommand_RejoinParamSetupReq) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled RejoinParamSetupAns CID and payload to the slice.
func (m *MACCommand_RejoinParamSetupAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_REJOIN_PARAM_SETUP))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the RejoinParamSetupAns CID and payload.
func (m *MACCommand_RejoinParamSetupAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the RejoinParamSetupAns CID and payload.
func (m *MACCommand_RejoinParamSetupAns) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled PingSlotInfoReq CID and payload to the slice.
func (m *MACCommand_PingSlotInfoReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_PING_SLOT_INFO))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the PingSlotInfoReq CID and payload.
func (m *MACCommand_PingSlotInfoReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the PingSlotInfoReq CID and payload.
func (m *MACCommand_PingSlotInfoReq) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled PingSlotChannelReq CID and payload to the slice.
func (m *MACCommand_PingSlotChannelReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_PING_SLOT_CHANNEL))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the PingSlotChannelReq CID and payload.
func (m *MACCommand_PingSlotChannelReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the PingSlotChannelReq CID and payload.
func (m *MACCommand_PingSlotChannelReq) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled PingSlotChannelAns CID and payload to the slice.
func (m *MACCommand_PingSlotChannelAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_PING_SLOT_CHANNEL))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the PingSlotChannelAns CID and payload.
func (m *MACCommand_PingSlotChannelAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the PingSlotChannelAns CID and payload.
func (m *MACCommand_PingSlotChannelAns) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled BeaconTimingAns CID and payload to the slice.
func (m *MACCommand_BeaconTimingAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_BEACON_TIMING))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the BeaconTimingAns CID and payload.
func (m *MACCommand_BeaconTimingAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the BeaconTimingAns CID and payload.
func (m *MACCommand_BeaconTimingAns) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled BeaconFreqReq CID and payload to the slice.
func (m *MACCommand_BeaconFreqReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_BEACON_FREQ))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the BeaconFreqReq CID and payload.
func (m *MACCommand_BeaconFreqReq) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the BeaconFreqReq CID and payload.
func (m *MACCommand_BeaconFreqReq) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled BeaconFreqAns CID and payload to the slice.
func (m *MACCommand_BeaconFreqAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_BEACON_FREQ))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the BeaconFreqAns CID and payload.
func (m *MACCommand_BeaconFreqAns) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the BeaconFreqAns CID and payload.
func (m *MACCommand_BeaconFreqAns) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled DeviceModeInd CID and payload to the slice.
func (m *MACCommand_DeviceModeInd) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_DEVICE_MODE))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the DeviceModeInd CID and payload.
func (m *MACCommand_DeviceModeInd) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the DeviceModeInd CID and payload.
func (m *MACCommand_DeviceModeInd) UnmarshalLoRaWAN(b []byte) error {
	return nil
}

// AppendLoRaWAN appends the marshaled DeviceModeConf CID and payload to the slice.
func (m *MACCommand_DeviceModeConf) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(CID_DEVICE_MODE))
	dst = append(dst, byte(0)) // TODO
	return dst, nil
}

// MarshalLoRaWAN marshals the DeviceModeConf CID and payload.
func (m *MACCommand_DeviceModeConf) MarshalLoRaWAN() ([]byte, error) {
	return m.AppendLoRaWAN(make([]byte, 0, 1))
}

// UnmarshalLoRaWAN unmarshals the DeviceModeConf CID and payload.
func (m *MACCommand_DeviceModeConf) UnmarshalLoRaWAN(b []byte) error {
	return nil
}
