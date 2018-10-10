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

	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
)

// MACCommandDescriptor descibes a MAC command.
type MACCommandDescriptor struct {
	InitiatedByDevice bool
	UplinkLength      uint16
	DownlinkLength    uint16
	NewUplink         func() lorawan.AppendUnmarshaler
	NewDownlink       func() lorawan.AppendUnmarshaler
}

// MACCommandSpec maps the CID of MACCommand to a *MACCommandDescriptor.
type MACCommandSpec [0xff + 1]*MACCommandDescriptor

var DefaultMACCommands MACCommandSpec

func init() {
	DefaultMACCommands[CID_RESET] = &MACCommandDescriptor{
		InitiatedByDevice: true,
		UplinkLength:      1,
		DownlinkLength:    1,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_ResetInd_{
				ResetInd: &MACCommand_ResetInd{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_ResetConf_{
				ResetConf: &MACCommand_ResetConf{},
			}
		},
	}
	DefaultMACCommands[CID_LINK_CHECK] = &MACCommandDescriptor{
		InitiatedByDevice: true,
		UplinkLength:      0,
		DownlinkLength:    2,
		NewUplink:         nil,
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_LinkCheckAns_{
				LinkCheckAns: &MACCommand_LinkCheckAns{},
			}
		},
	}
	DefaultMACCommands[CID_LINK_ADR] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    4,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_LinkADRAns_{
				LinkADRAns: &MACCommand_LinkADRAns{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_LinkADRReq_{
				LinkADRReq: &MACCommand_LinkADRReq{},
			}
		},
	}
	DefaultMACCommands[CID_DUTY_CYCLE] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      0,
		DownlinkLength:    1,
		NewUplink:         nil,
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_DutyCycleReq_{
				DutyCycleReq: &MACCommand_DutyCycleReq{},
			}
		},
	}
	DefaultMACCommands[CID_RX_PARAM_SETUP] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    4,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_RxParamSetupAns_{
				RxParamSetupAns: &MACCommand_RxParamSetupAns{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_RxParamSetupReq_{
				RxParamSetupReq: &MACCommand_RxParamSetupReq{},
			}
		},
	}
	DefaultMACCommands[CID_DEV_STATUS] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      2,
		DownlinkLength:    0,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_DevStatusAns_{
				DevStatusAns: &MACCommand_DevStatusAns{},
			}
		},
		NewDownlink: nil,
	}
	DefaultMACCommands[CID_NEW_CHANNEL] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    5,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_NewChannelAns_{
				NewChannelAns: &MACCommand_NewChannelAns{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_NewChannelReq_{
				NewChannelReq: &MACCommand_NewChannelReq{},
			}
		},
	}
	DefaultMACCommands[CID_RX_TIMING_SETUP] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      0,
		DownlinkLength:    1,
		NewUplink:         nil,
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_RxTimingSetupReq_{
				RxTimingSetupReq: &MACCommand_RxTimingSetupReq{},
			}
		},
	}
	DefaultMACCommands[CID_TX_PARAM_SETUP] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      0,
		DownlinkLength:    1,
		NewUplink:         nil,
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_TxParamSetupReq_{
				TxParamSetupReq: &MACCommand_TxParamSetupReq{},
			}
		},
	}
	DefaultMACCommands[CID_DL_CHANNEL] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    4,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_DlChannelAns{
				DlChannelAns: &MACCommand_DLChannelAns{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_DlChannelReq{
				DlChannelReq: &MACCommand_DLChannelReq{},
			}
		},
	}
	DefaultMACCommands[CID_REKEY] = &MACCommandDescriptor{
		InitiatedByDevice: true,
		UplinkLength:      1,
		DownlinkLength:    1,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_RekeyInd_{
				RekeyInd: &MACCommand_RekeyInd{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_RekeyConf_{
				RekeyConf: &MACCommand_RekeyConf{},
			}
		},
	}
	DefaultMACCommands[CID_ADR_PARAM_SETUP] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      0,
		DownlinkLength:    1,
		NewUplink:         nil,
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_ADRParamSetupReq_{
				ADRParamSetupReq: &MACCommand_ADRParamSetupReq{},
			}
		},
	}
	DefaultMACCommands[CID_DEVICE_TIME] = &MACCommandDescriptor{
		InitiatedByDevice: true,
		UplinkLength:      0,
		DownlinkLength:    5,
		NewUplink:         nil,
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_DeviceTimeAns_{
				DeviceTimeAns: &MACCommand_DeviceTimeAns{},
			}
		},
	}
	DefaultMACCommands[CID_FORCE_REJOIN] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      0,
		DownlinkLength:    2,
		NewUplink:         nil,
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_ForceRejoinReq_{
				ForceRejoinReq: &MACCommand_ForceRejoinReq{},
			}
		},
	}
	DefaultMACCommands[CID_REJOIN_PARAM_SETUP] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    1,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_RejoinParamSetupAns_{
				RejoinParamSetupAns: &MACCommand_RejoinParamSetupAns{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_RejoinParamSetupReq_{
				RejoinParamSetupReq: &MACCommand_RejoinParamSetupReq{},
			}
		},
	}
	DefaultMACCommands[CID_PING_SLOT_INFO] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    0,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_PingSlotInfoReq_{
				PingSlotInfoReq: &MACCommand_PingSlotInfoReq{},
			}
		},
		NewDownlink: nil,
	}
	DefaultMACCommands[CID_PING_SLOT_CHANNEL] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    4,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_PingSlotChannelAns_{
				PingSlotChannelAns: &MACCommand_PingSlotChannelAns{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_PingSlotChannelReq_{
				PingSlotChannelReq: &MACCommand_PingSlotChannelReq{},
			}
		},
	}
	DefaultMACCommands[CID_BEACON_TIMING] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      0,
		DownlinkLength:    3,
		NewUplink:         nil,
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_BeaconTimingAns_{
				BeaconTimingAns: &MACCommand_BeaconTimingAns{},
			}
		},
	}
	DefaultMACCommands[CID_BEACON_FREQ] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    3,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_BeaconFreqAns_{
				BeaconFreqAns: &MACCommand_BeaconFreqAns{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_BeaconFreqReq_{
				BeaconFreqReq: &MACCommand_BeaconFreqReq{},
			}
		},
	}
	DefaultMACCommands[CID_DEVICE_MODE] = &MACCommandDescriptor{
		InitiatedByDevice: false,
		UplinkLength:      1,
		DownlinkLength:    1,
		NewUplink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_DeviceModeInd_{
				DeviceModeInd: &MACCommand_DeviceModeInd{},
			}
		},
		NewDownlink: func() lorawan.AppendUnmarshaler {
			return &MACCommand_DeviceModeConf_{
				DeviceModeConf: &MACCommand_DeviceModeConf{},
			}
		},
	}
}

// Validate reports whether cid represents a valid MACCommandIdentifier.
func (cid MACCommandIdentifier) Validate() error {
	if cid < 0x00 || cid > 0xff {
		return errExpectedBetween("CID", "0x00", "0xFF")(fmt.Sprintf("0x%X", int32(cid)))
	}
	return nil
}

// InitiatedByDevice reports whether CID is initiated by device.
// InitiatedByDevice returns false, false if cid is not known.
// The second return value is true if cid is known.
func (cid MACCommandIdentifier) InitiatedByDevice() (bool, bool) {
	if cid < 0 || int(cid) >= len(DefaultMACCommands) {
		return false, false
	}
	desc := DefaultMACCommands[cid]
	if desc == nil {
		return false, false
	}
	return desc.InitiatedByDevice, true
}

// MACCommand returns a payload-less MAC command with this CID as a *MACCommand.
func (cid MACCommandIdentifier) MACCommand() *MACCommand {
	return &MACCommand{
		CID: cid,
	}
}

// MACCommand returns a MAC command with specified CID as a *MACCommand.
func (pld *MACCommand_RawPayload) MACCommand(cid MACCommandIdentifier) *MACCommand {
	return &MACCommand{
		CID:     cid,
		Payload: pld,
	}
}

// MACCommand_Payload returns the ResetInd MAC command as a isMACCommand_Payload.
func (pld *MACCommand_ResetInd) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_ResetInd_{
		ResetInd: pld,
	}
}

// MACCommand returns the ResetInd MAC command as a *MACCommand.
func (pld *MACCommand_ResetInd) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_RESET,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the ResetConf MAC command as a isMACCommand_Payload.
func (pld *MACCommand_ResetConf) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_ResetConf_{
		ResetConf: pld,
	}
}

// MACCommand returns the ResetConf MAC command as a *MACCommand.
func (pld *MACCommand_ResetConf) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_RESET,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the LinkCheckAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_LinkCheckAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_LinkCheckAns_{
		LinkCheckAns: pld,
	}
}

// MACCommand returns the LinkCheckAns MAC command as a *MACCommand.
func (pld *MACCommand_LinkCheckAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_LINK_CHECK,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the LinkADRReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_LinkADRReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_LinkADRReq_{
		LinkADRReq: pld,
	}
}

// MACCommand returns the LinkADRReq MAC command as a *MACCommand.
func (pld *MACCommand_LinkADRReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_LINK_ADR,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the LinkADRAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_LinkADRAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_LinkADRAns_{
		LinkADRAns: pld,
	}
}

// MACCommand returns the LinkADRAns MAC command as a *MACCommand.
func (pld *MACCommand_LinkADRAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_LINK_ADR,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the DutyCycleReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_DutyCycleReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_DutyCycleReq_{
		DutyCycleReq: pld,
	}
}

// MACCommand returns the DutyCycleReq MAC command as a *MACCommand.
func (pld *MACCommand_DutyCycleReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_DUTY_CYCLE,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the RxParamSetupReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_RxParamSetupReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_RxParamSetupReq_{
		RxParamSetupReq: pld,
	}
}

// MACCommand returns the RxParamSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_RxParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_RX_PARAM_SETUP,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the RxParamSetupAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_RxParamSetupAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_RxParamSetupAns_{
		RxParamSetupAns: pld,
	}
}

// MACCommand returns the RxParamSetupAns MAC command as a *MACCommand.
func (pld *MACCommand_RxParamSetupAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_RX_PARAM_SETUP,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the DevStatusAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_DevStatusAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_DevStatusAns_{
		DevStatusAns: pld,
	}
}

// MACCommand returns the DevStatusAns MAC command as a *MACCommand.
func (pld *MACCommand_DevStatusAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_DEV_STATUS,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the NewChannelReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_NewChannelReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_NewChannelReq_{
		NewChannelReq: pld,
	}
}

// MACCommand returns the NewChannelReq MAC command as a *MACCommand.
func (pld *MACCommand_NewChannelReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_NEW_CHANNEL,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the NewChannelAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_NewChannelAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_NewChannelAns_{
		NewChannelAns: pld,
	}
}

// MACCommand returns the NewChannelAns MAC command as a *MACCommand.
func (pld *MACCommand_NewChannelAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_NEW_CHANNEL,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the DLChannelReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_DLChannelReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_DlChannelReq{
		DlChannelReq: pld,
	}
}

// MACCommand returns the DLChannelReq MAC command as a *MACCommand.
func (pld *MACCommand_DLChannelReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_DL_CHANNEL,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the DLChannelAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_DLChannelAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_DlChannelAns{
		DlChannelAns: pld,
	}
}

// MACCommand returns the DLChannelAns MAC command as a *MACCommand.
func (pld *MACCommand_DLChannelAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_DL_CHANNEL,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the RxTimingSetupReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_RxTimingSetupReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_RxTimingSetupReq_{
		RxTimingSetupReq: pld,
	}
}

// MACCommand returns the RxTimingSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_RxTimingSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_RX_TIMING_SETUP,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the TxParamSetupReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_TxParamSetupReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_TxParamSetupReq_{
		TxParamSetupReq: pld,
	}
}

// MACCommand returns the TxParamSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_TxParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_TX_PARAM_SETUP,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the RekeyInd MAC command as a isMACCommand_Payload.
func (pld *MACCommand_RekeyInd) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_RekeyInd_{
		RekeyInd: pld,
	}
}

// MACCommand returns the RekeyInd MAC command as a *MACCommand.
func (pld *MACCommand_RekeyInd) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_REKEY,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the RekeyConf MAC command as a isMACCommand_Payload.
func (pld *MACCommand_RekeyConf) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_RekeyConf_{
		RekeyConf: pld,
	}
}

// MACCommand returns the RekeyConf MAC command as a *MACCommand.
func (pld *MACCommand_RekeyConf) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_REKEY,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the ADRParamSetupReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_ADRParamSetupReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_ADRParamSetupReq_{
		ADRParamSetupReq: pld,
	}
}

// MACCommand returns the ADRParamSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_ADRParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_ADR_PARAM_SETUP,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the DeviceTimeAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_DeviceTimeAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_DeviceTimeAns_{
		DeviceTimeAns: pld,
	}
}

// MACCommand returns the DeviceTimeAns MAC command as a *MACCommand.
func (pld *MACCommand_DeviceTimeAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_DEVICE_TIME,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the ForceRejoinReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_ForceRejoinReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_ForceRejoinReq_{
		ForceRejoinReq: pld,
	}
}

// MACCommand returns the ForceRejoinReq MAC command as a *MACCommand.
func (pld *MACCommand_ForceRejoinReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_FORCE_REJOIN,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the RejoinParamSetupReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_RejoinParamSetupReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_RejoinParamSetupReq_{
		RejoinParamSetupReq: pld,
	}
}

// MACCommand returns the RejoinParamSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_RejoinParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_REJOIN_PARAM_SETUP,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the RejoinParamSetupAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_RejoinParamSetupAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_RejoinParamSetupAns_{
		RejoinParamSetupAns: pld,
	}
}

// MACCommand returns the RejoinParamSetupAns MAC command as a *MACCommand.
func (pld *MACCommand_RejoinParamSetupAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_REJOIN_PARAM_SETUP,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the PingSlotInfoReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_PingSlotInfoReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_PingSlotInfoReq_{
		PingSlotInfoReq: pld,
	}
}

// MACCommand returns the PingSlotInfoReq MAC command as a *MACCommand.
func (pld *MACCommand_PingSlotInfoReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_PING_SLOT_INFO,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the PingSlotChannelReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_PingSlotChannelReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_PingSlotChannelReq_{
		PingSlotChannelReq: pld,
	}
}

// MACCommand returns the PingSlotChannelReq MAC command as a *MACCommand.
func (pld *MACCommand_PingSlotChannelReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_PING_SLOT_CHANNEL,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the PingSlotChannelAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_PingSlotChannelAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_PingSlotChannelAns_{
		PingSlotChannelAns: pld,
	}
}

// MACCommand returns the PingSlotChannelAns MAC command as a *MACCommand.
func (pld *MACCommand_PingSlotChannelAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_PING_SLOT_CHANNEL,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the BeaconTimingAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_BeaconTimingAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_BeaconTimingAns_{
		BeaconTimingAns: pld,
	}
}

// MACCommand returns the BeaconTimingAns MAC command as a *MACCommand.
func (pld *MACCommand_BeaconTimingAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_BEACON_TIMING,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the BeaconFreqReq MAC command as a isMACCommand_Payload.
func (pld *MACCommand_BeaconFreqReq) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_BeaconFreqReq_{
		BeaconFreqReq: pld,
	}
}

// MACCommand returns the BeaconFreqReq MAC command as a *MACCommand.
func (pld *MACCommand_BeaconFreqReq) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_BEACON_FREQ,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the BeaconFreqAns MAC command as a isMACCommand_Payload.
func (pld *MACCommand_BeaconFreqAns) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_BeaconFreqAns_{
		BeaconFreqAns: pld,
	}
}

// MACCommand returns the BeaconFreqAns MAC command as a *MACCommand.
func (pld *MACCommand_BeaconFreqAns) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_BEACON_FREQ,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the DeviceModeInd MAC command as a isMACCommand_Payload.
func (pld *MACCommand_DeviceModeInd) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_DeviceModeInd_{
		DeviceModeInd: pld,
	}
}

// MACCommand returns the DeviceModeInd MAC command as a *MACCommand.
func (pld *MACCommand_DeviceModeInd) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_DEVICE_MODE,
		Payload: pld.MACCommand_Payload(),
	}
}

// MACCommand_Payload returns the DeviceModeConf MAC command as a isMACCommand_Payload.
func (pld *MACCommand_DeviceModeConf) MACCommand_Payload() isMACCommand_Payload {
	return &MACCommand_DeviceModeConf_{
		DeviceModeConf: pld,
	}
}

// MACCommand returns the DeviceModeConf MAC command as a *MACCommand.
func (pld *MACCommand_DeviceModeConf) MACCommand() *MACCommand {
	return &MACCommand{
		CID:     CID_DEVICE_MODE,
		Payload: pld.MACCommand_Payload(),
	}
}

// Validate reports whether cmd represents a valid *MACCommand.
func (cmd *MACCommand) Validate() error {
	return cmd.CID.Validate()
}
