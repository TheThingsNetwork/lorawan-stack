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
	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
)

// Validate reports whether cid represents a valid MACCommandIdentifier.
func (cid MACCommandIdentifier) Validate() error {
	_, ok := MACCommandIdentifier_name[int32(cid)]
	if !ok || cid == CID_RFU_0 {
		return errors.Errorf("Unknown CID: %x", cid)
	}
	return nil
}

// InitiatedByDevice reports whether CID is initiated by device.
func (cid MACCommandIdentifier) InitiatedByDevice() bool {
	switch cid {
	case CID_LINK_ADR,
		CID_DUTY_CYCLE,
		CID_RX_PARAM_SETUP,
		CID_DEV_STATUS,
		CID_NEW_CHANNEL,
		CID_RX_TIMING_SETUP,
		CID_TX_PARAM_SETUP,
		CID_DL_CHANNEL,
		CID_ADR_PARAM_SETUP,
		CID_FORCE_REJOIN,
		CID_REJOIN_PARAM_SETUP,
		CID_PING_SLOT_INFO,
		CID_PING_SLOT_CHANNEL,
		CID_BEACON_TIMING,
		CID_BEACON_FREQ,
		CID_DEVICE_MODE:
		return false

	case CID_RESET,
		CID_LINK_CHECK,
		CID_REKEY,
		CID_DEVICE_TIME:
		return true

	default:
		panic(errors.Errorf("unknown CID: %s", cid))
	}
}

// MACCommandPayload interface is implemented by all MAC commands.
type MACCommandPayload interface {
	proto.Message
	MACCommand() *MACCommand
	AppendLoRaWAN(dst []byte) ([]byte, error)
	MarshalLoRaWAN() ([]byte, error)
	UnmarshalLoRaWAN(b []byte) error
}

// Validate reports whether cmd represents a valid MACCommandIdentifier.
func (cmd *MACCommand) Validate() error {
	if cmd.Payload == nil {
		return common.ErrMissingPayload.New(nil)
	}
	return cmd.CID().Validate()
}

// CID returns the CID of the embedded MAC command.
func (cmd *MACCommand) CID() MACCommandIdentifier {
	switch x := cmd.Payload.(type) {
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
	case *MACCommand_LinkADRReq_:
		return CID_LINK_ADR
	case *MACCommand_LinkADRAns_:
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
	case *MACCommand_ADRParamSetupReq_:
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
	default:
		panic(errors.Errorf("unmatched payload type: %T", x))
	}
}

// MACCommand returns a payload-less MAC command with this CID.
func (cid MACCommandIdentifier) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_CID{CID: cid}}
}

// MACCommand returns the Proprietary MAC command as a *MACCommand.
func (cmd *MACCommand_Proprietary) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_Proprietary_{Proprietary: cmd}}
}

// MACCommand returns the ResetInd MAC command as a *MACCommand.
func (cmd *MACCommand_ResetInd) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_ResetInd_{ResetInd: cmd}}
}

// MACCommand returns the ResetConf MAC command as a *MACCommand.
func (cmd *MACCommand_ResetConf) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_ResetConf_{ResetConf: cmd}}
}

// MACCommand returns the LinkCheckAns MAC command as a *MACCommand.
func (cmd *MACCommand_LinkCheckAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_LinkCheckAns_{LinkCheckAns: cmd}}
}

// MACCommand returns the LinkADRReq MAC command as a *MACCommand.
func (cmd *MACCommand_LinkADRReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_LinkADRReq_{LinkADRReq: cmd}}
}

// MACCommand returns the LinkADRAns MAC command as a *MACCommand.
func (cmd *MACCommand_LinkADRAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_LinkADRAns_{LinkADRAns: cmd}}
}

// MACCommand returns the DutyCycleReq MAC command as a *MACCommand.
func (cmd *MACCommand_DutyCycleReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DutyCycleReq_{DutyCycleReq: cmd}}
}

// MACCommand returns the RxParamSetupReq MAC command as a *MACCommand.
func (cmd *MACCommand_RxParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RxParamSetupReq_{RxParamSetupReq: cmd}}
}

// MACCommand returns the RxParamSetupAns MAC command as a *MACCommand.
func (cmd *MACCommand_RxParamSetupAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RxParamSetupAns_{RxParamSetupAns: cmd}}
}

// MACCommand returns the DevStatusAns MAC command as a *MACCommand.
func (cmd *MACCommand_DevStatusAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DevStatusAns_{DevStatusAns: cmd}}
}

// MACCommand returns the NewChannelReq MAC command as a *MACCommand.
func (cmd *MACCommand_NewChannelReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_NewChannelReq_{NewChannelReq: cmd}}
}

// MACCommand returns the NewChannelAns MAC command as a *MACCommand.
func (cmd *MACCommand_NewChannelAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_NewChannelAns_{NewChannelAns: cmd}}
}

// MACCommand returns the DLChannelReq MAC command as a *MACCommand.
func (cmd *MACCommand_DLChannelReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DlChannelReq{DlChannelReq: cmd}}
}

// MACCommand returns the DLChannelAns MAC command as a *MACCommand.
func (cmd *MACCommand_DLChannelAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DlChannelAns{DlChannelAns: cmd}}
}

// MACCommand returns the RxTimingSetupReq MAC command as a *MACCommand.
func (cmd *MACCommand_RxTimingSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RxTimingSetupReq_{RxTimingSetupReq: cmd}}
}

// MACCommand returns the TxParamSetupReq MAC command as a *MACCommand.
func (cmd *MACCommand_TxParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_TxParamSetupReq_{TxParamSetupReq: cmd}}
}

// MACCommand returns the RekeyInd MAC command as a *MACCommand.
func (cmd *MACCommand_RekeyInd) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RekeyInd_{RekeyInd: cmd}}
}

// MACCommand returns the RekeyConf MAC command as a *MACCommand.
func (cmd *MACCommand_RekeyConf) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RekeyConf_{RekeyConf: cmd}}
}

// MACCommand returns the ADRParamSetupReq MAC command as a *MACCommand.
func (cmd *MACCommand_ADRParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_ADRParamSetupReq_{ADRParamSetupReq: cmd}}
}

// MACCommand returns the DeviceTimeAns MAC command as a *MACCommand.
func (cmd *MACCommand_DeviceTimeAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DeviceTimeAns_{DeviceTimeAns: cmd}}
}

// MACCommand returns the ForceRejoinReq MAC command as a *MACCommand.
func (cmd *MACCommand_ForceRejoinReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_ForceRejoinReq_{ForceRejoinReq: cmd}}
}

// MACCommand returns the RejoinParamSetupReq MAC command as a *MACCommand.
func (cmd *MACCommand_RejoinParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RejoinParamSetupReq_{RejoinParamSetupReq: cmd}}
}

// MACCommand returns the RejoinParamSetupAns MAC command as a *MACCommand.
func (cmd *MACCommand_RejoinParamSetupAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_RejoinParamSetupAns_{RejoinParamSetupAns: cmd}}
}

// MACCommand returns the PingSlotInfoReq MAC command as a *MACCommand.
func (cmd *MACCommand_PingSlotInfoReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_PingSlotInfoReq_{PingSlotInfoReq: cmd}}
}

// MACCommand returns the PingSlotChannelReq MAC command as a *MACCommand.
func (cmd *MACCommand_PingSlotChannelReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_PingSlotChannelReq_{PingSlotChannelReq: cmd}}
}

// MACCommand returns the PingSlotChannelAns MAC command as a *MACCommand.
func (cmd *MACCommand_PingSlotChannelAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_PingSlotChannelAns_{PingSlotChannelAns: cmd}}
}

// MACCommand returns the BeaconTimingAns MAC command as a *MACCommand.
func (cmd *MACCommand_BeaconTimingAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_BeaconTimingAns_{BeaconTimingAns: cmd}}
}

// MACCommand returns the BeaconFreqReq MAC command as a *MACCommand.
func (cmd *MACCommand_BeaconFreqReq) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_BeaconFreqReq_{BeaconFreqReq: cmd}}
}

// MACCommand returns the BeaconFreqAns MAC command as a *MACCommand.
func (cmd *MACCommand_BeaconFreqAns) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_BeaconFreqAns_{BeaconFreqAns: cmd}}
}

// MACCommand returns the DeviceModeInd MAC command as a *MACCommand.
func (cmd *MACCommand_DeviceModeInd) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DeviceModeInd_{DeviceModeInd: cmd}}
}

// MACCommand returns the DeviceModeConf MAC command as a *MACCommand.
func (cmd *MACCommand_DeviceModeConf) MACCommand() *MACCommand {
	return &MACCommand{Payload: &MACCommand_DeviceModeConf_{DeviceModeConf: cmd}}
}
