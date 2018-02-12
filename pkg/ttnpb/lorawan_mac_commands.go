// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import proto "github.com/gogo/protobuf/proto"


// MACCommandPayload interface is implemented by all MAC commands
type MACCommandPayload interface {
	proto.Message
	MACCommand() *MACCommand
	AppendLoRaWAN(dst []byte) ([]byte, error)
	MarshalLoRaWAN() ([]byte, error)
	UnmarshalLoRaWAN(b []byte) error
}

// CID returns the CID of the embedded MAC command
func (m *MACCommand) CID() MACCommandIdentifier {
	switch x := m.Payload.(type) {
	case *MACCommand_CID:
		return x.CID
	case *MACCommand_Proprietary_:
		return x.Proprietary.CID
	case *MACCommand_ResetInd_:
		return CID_RESET
	case *MACCommand_ResetConf_:
		return CID_RESET
	case *MACCommand_LinkCheckAns_:
		return CID_LINK_CHECK
	case *MACCommand_LinkAdrReq:
		return CID_LINK_ADR
	case *MACCommand_LinkAdrAns:
		return CID_LINK_ADR
	case *MACCommand_DutyCycleReq_:
		return CID_DUTY_CYCLE
	case *MACCommand_RxParamSetupReq_:
		return CID_RX_PARAM_SETUP
	case *MACCommand_RxParamSetupAns_:
		return CID_RX_PARAM_SETUP
	case *MACCommand_DevStatusAns_:
		return CID_DEV_STATUS
	case *MACCommand_NewChannelReq_:
		return CID_NEW_CHANNEL
	case *MACCommand_NewChannelAns_:
		return CID_NEW_CHANNEL
	case *MACCommand_DlChannelReq:
		return CID_DL_CHANNEL
	case *MACCommand_DlChannelAns:
		return CID_DL_CHANNEL
	case *MACCommand_RxTimingSetupReq_:
		return CID_RX_TIMING_SETUP
	case *MACCommand_TxParamSetupReq_:
		return CID_TX_PARAM_SETUP
	case *MACCommand_RekeyInd_:
		return CID_REKEY
	case *MACCommand_RekeyConf_:
		return CID_REKEY
	case *MACCommand_AdrParamSetupReq:
		return CID_ADR_PARAM_SETUP
	case *MACCommand_DeviceTimeAns_:
		return CID_DEVICE_TIME
	case *MACCommand_ForceRejoinReq_:
		return CID_FORCE_REJOIN
	case *MACCommand_RejoinParamSetupReq_:
		return CID_REJOIN_PARAM_SETUP
	case *MACCommand_RejoinParamSetupAns_:
		return CID_REJOIN_PARAM_SETUP
	case *MACCommand_PingSlotInfoReq_:
		return CID_PING_SLOT_INFO
	case *MACCommand_PingSlotChannelReq_:
		return CID_PING_SLOT_CHANNEL
	case *MACCommand_PingSlotChannelAns_:
		return CID_PING_SLOT_CHANNEL
	case *MACCommand_BeaconTimingAns_:
		return CID_BEACON_TIMING
	case *MACCommand_BeaconFreqReq_:
		return CID_BEACON_FREQ
	case *MACCommand_BeaconFreqAns_:
		return CID_BEACON_FREQ
	case *MACCommand_DeviceModeInd_:
		return CID_DEVICE_MODE
	case *MACCommand_DeviceModeConf_:
		return CID_DEVICE_MODE
	}
	return 0
}

// GetActualPayload returns the actual payload of the embedded MAC command
func (m *MACCommand) GetActualPayload() MACCommandPayload {
	switch x := m.Payload.(type) {
	case *MACCommand_CID:
		return nil
	case *MACCommand_Proprietary_:
		return x.Proprietary
	case *MACCommand_ResetInd_:
		return x.ResetInd
	case *MACCommand_ResetConf_:
		return x.ResetConf
	case *MACCommand_LinkCheckAns_:
		return x.LinkCheckAns
	case *MACCommand_LinkAdrReq:
		return x.LinkAdrReq
	case *MACCommand_LinkAdrAns:
		return x.LinkAdrAns
	case *MACCommand_DutyCycleReq_:
		return x.DutyCycleReq
	case *MACCommand_RxParamSetupReq_:
		return x.RxParamSetupReq
	case *MACCommand_RxParamSetupAns_:
		return x.RxParamSetupAns
	case *MACCommand_DevStatusAns_:
		return x.DevStatusAns
	case *MACCommand_NewChannelReq_:
		return x.NewChannelReq
	case *MACCommand_NewChannelAns_:
		return x.NewChannelAns
	case *MACCommand_DlChannelReq:
		return x.DlChannelReq
	case *MACCommand_DlChannelAns:
		return x.DlChannelAns
	case *MACCommand_RxTimingSetupReq_:
		return x.RxTimingSetupReq
	case *MACCommand_TxParamSetupReq_:
		return x.TxParamSetupReq
	case *MACCommand_RekeyInd_:
		return x.RekeyInd
	case *MACCommand_RekeyConf_:
		return x.RekeyConf
	case *MACCommand_AdrParamSetupReq:
		return x.AdrParamSetupReq
	case *MACCommand_DeviceTimeAns_:
		return x.DeviceTimeAns
	case *MACCommand_ForceRejoinReq_:
		return x.ForceRejoinReq
	case *MACCommand_RejoinParamSetupReq_:
		return x.RejoinParamSetupReq
	case *MACCommand_RejoinParamSetupAns_:
		return x.RejoinParamSetupAns
	case *MACCommand_PingSlotInfoReq_:
		return x.PingSlotInfoReq
	case *MACCommand_PingSlotChannelReq_:
		return x.PingSlotChannelReq
	case *MACCommand_PingSlotChannelAns_:
		return x.PingSlotChannelAns
	case *MACCommand_BeaconTimingAns_:
		return x.BeaconTimingAns
	case *MACCommand_BeaconFreqReq_:
		return x.BeaconFreqReq
	case *MACCommand_BeaconFreqAns_:
		return x.BeaconFreqAns
	case *MACCommand_DeviceModeInd_:
		return x.DeviceModeInd
	case *MACCommand_DeviceModeConf_:
		return x.DeviceModeConf
	}
	return nil
}

// MACCommand returns a payload-less MAC command with this CID
func (m MACCommandIdentifier) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_CID{CID: m}}
}

// MACCommand returns the Proprietary MAC command as a *MACCommand
func (m *MACCommand_Proprietary) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_Proprietary_{Proprietary: m}}
}

// MACCommand returns the ResetInd MAC command as a *MACCommand
func (m *MACCommand_ResetInd) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_ResetInd_{ResetInd: m}}
}

// MACCommand returns the ResetConf MAC command as a *MACCommand
func (m *MACCommand_ResetConf) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_ResetConf_{ResetConf: m}}
}

// MACCommand returns the LinkCheckAns MAC command as a *MACCommand
func (m *MACCommand_LinkCheckAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_LinkCheckAns_{LinkCheckAns: m}}
}

// MACCommand returns the LinkADRReq MAC command as a *MACCommand
func (m *MACCommand_LinkADRReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_LinkAdrReq{LinkAdrReq: m}}
}

// MACCommand returns the LinkADRAns MAC command as a *MACCommand
func (m *MACCommand_LinkADRAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_LinkAdrAns{LinkAdrAns: m}}
}

// MACCommand returns the DutyCycleReq MAC command as a *MACCommand
func (m *MACCommand_DutyCycleReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DutyCycleReq_{DutyCycleReq: m}}
}

// MACCommand returns the RxParamSetupReq MAC command as a *MACCommand
func (m *MACCommand_RxParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RxParamSetupReq_{RxParamSetupReq: m}}
}

// MACCommand returns the RxParamSetupAns MAC command as a *MACCommand
func (m *MACCommand_RxParamSetupAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RxParamSetupAns_{RxParamSetupAns: m}}
}

// MACCommand returns the DevStatusAns MAC command as a *MACCommand
func (m *MACCommand_DevStatusAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DevStatusAns_{DevStatusAns: m}}
}

// MACCommand returns the NewChannelReq MAC command as a *MACCommand
func (m *MACCommand_NewChannelReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_NewChannelReq_{NewChannelReq: m}}
}

// MACCommand returns the NewChannelAns MAC command as a *MACCommand
func (m *MACCommand_NewChannelAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_NewChannelAns_{NewChannelAns: m}}
}

// MACCommand returns the DLChannelReq MAC command as a *MACCommand
func (m *MACCommand_DLChannelReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DlChannelReq{DlChannelReq: m}}
}

// MACCommand returns the DLChannelAns MAC command as a *MACCommand
func (m *MACCommand_DLChannelAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DlChannelAns{DlChannelAns: m}}
}

// MACCommand returns the RxTimingSetupReq MAC command as a *MACCommand
func (m *MACCommand_RxTimingSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RxTimingSetupReq_{RxTimingSetupReq: m}}
}

// MACCommand returns the TxParamSetupReq MAC command as a *MACCommand
func (m *MACCommand_TxParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_TxParamSetupReq_{TxParamSetupReq: m}}
}

// MACCommand returns the RekeyInd MAC command as a *MACCommand
func (m *MACCommand_RekeyInd) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RekeyInd_{RekeyInd: m}}
}

// MACCommand returns the RekeyConf MAC command as a *MACCommand
func (m *MACCommand_RekeyConf) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RekeyConf_{RekeyConf: m}}
}

// MACCommand returns the ADRParamSetupReq MAC command as a *MACCommand
func (m *MACCommand_ADRParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_AdrParamSetupReq{AdrParamSetupReq: m}}
}

// MACCommand returns the DeviceTimeAns MAC command as a *MACCommand
func (m *MACCommand_DeviceTimeAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DeviceTimeAns_{DeviceTimeAns: m}}
}

// MACCommand returns the ForceRejoinReq MAC command as a *MACCommand
func (m *MACCommand_ForceRejoinReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_ForceRejoinReq_{ForceRejoinReq: m}}
}

// MACCommand returns the RejoinParamSetupReq MAC command as a *MACCommand
func (m *MACCommand_RejoinParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RejoinParamSetupReq_{RejoinParamSetupReq: m}}
}

// MACCommand returns the RejoinParamSetupAns MAC command as a *MACCommand
func (m *MACCommand_RejoinParamSetupAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RejoinParamSetupAns_{RejoinParamSetupAns: m}}
}

// MACCommand returns the PingSlotInfoReq MAC command as a *MACCommand
func (m *MACCommand_PingSlotInfoReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_PingSlotInfoReq_{PingSlotInfoReq: m}}
}

// MACCommand returns the PingSlotChannelReq MAC command as a *MACCommand
func (m *MACCommand_PingSlotChannelReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_PingSlotChannelReq_{PingSlotChannelReq: m}}
}

// MACCommand returns the PingSlotChannelAns MAC command as a *MACCommand
func (m *MACCommand_PingSlotChannelAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_PingSlotChannelAns_{PingSlotChannelAns: m}}
}

// MACCommand returns the BeaconTimingAns MAC command as a *MACCommand
func (m *MACCommand_BeaconTimingAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_BeaconTimingAns_{BeaconTimingAns: m}}
}

// MACCommand returns the BeaconFreqReq MAC command as a *MACCommand
func (m *MACCommand_BeaconFreqReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_BeaconFreqReq_{BeaconFreqReq: m}}
}

// MACCommand returns the BeaconFreqAns MAC command as a *MACCommand
func (m *MACCommand_BeaconFreqAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_BeaconFreqAns_{BeaconFreqAns: m}}
}

// MACCommand returns the DeviceModeInd MAC command as a *MACCommand
func (m *MACCommand_DeviceModeInd) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DeviceModeInd_{DeviceModeInd: m}}
}

// MACCommand returns the DeviceModeConf MAC command as a *MACCommand
func (m *MACCommand_DeviceModeConf) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DeviceModeConf_{DeviceModeConf: m}}
}
