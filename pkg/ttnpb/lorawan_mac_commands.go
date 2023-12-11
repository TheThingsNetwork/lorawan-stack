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

// MACCommand returns a payload-less MAC command with this CID as a *MACCommand.
func (cid MACCommandIdentifier) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: cid,
	}
}

// MACCommand returns a MAC command with specified CID as a *MACCommand.
func (pld *MACCommand_RawPayload) MACCommand(cid MACCommandIdentifier) *MACCommand {
	return &MACCommand{
		Cid:     cid,
		Payload: pld,
	}
}

// MACCommand returns the ResetInd MAC command as a *MACCommand.
func (pld *MACCommand_ResetInd) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RESET,
		Payload: &MACCommand_ResetInd_{
			ResetInd: pld,
		},
	}
}

// MACCommand returns the ResetConf MAC command as a *MACCommand.
func (pld *MACCommand_ResetConf) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RESET,
		Payload: &MACCommand_ResetConf_{
			ResetConf: pld,
		},
	}
}

// MACCommand returns the LinkCheckAns MAC command as a *MACCommand.
func (pld *MACCommand_LinkCheckAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_LINK_CHECK,
		Payload: &MACCommand_LinkCheckAns_{
			LinkCheckAns: pld,
		},
	}
}

// MACCommand returns the LinkADRReq MAC command as a *MACCommand.
func (pld *MACCommand_LinkADRReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_LINK_ADR,
		Payload: &MACCommand_LinkAdrReq{
			LinkAdrReq: pld,
		},
	}
}

// MACCommand returns the LinkADRAns MAC command as a *MACCommand.
func (pld *MACCommand_LinkADRAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_LINK_ADR,
		Payload: &MACCommand_LinkAdrAns{
			LinkAdrAns: pld,
		},
	}
}

// MACCommand returns the DutyCycleReq MAC command as a *MACCommand.
func (pld *MACCommand_DutyCycleReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_DUTY_CYCLE,
		Payload: &MACCommand_DutyCycleReq_{
			DutyCycleReq: pld,
		},
	}
}

// MACCommand returns the RxParamSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_RxParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RX_PARAM_SETUP,
		Payload: &MACCommand_RxParamSetupReq_{
			RxParamSetupReq: pld,
		},
	}
}

// MACCommand returns the RxParamSetupAns MAC command as a *MACCommand.
func (pld *MACCommand_RxParamSetupAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RX_PARAM_SETUP,
		Payload: &MACCommand_RxParamSetupAns_{
			RxParamSetupAns: pld,
		},
	}
}

// MACCommand returns the DevStatusAns MAC command as a *MACCommand.
func (pld *MACCommand_DevStatusAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_DEV_STATUS,
		Payload: &MACCommand_DevStatusAns_{
			DevStatusAns: pld,
		},
	}
}

// MACCommand returns the NewChannelReq MAC command as a *MACCommand.
func (pld *MACCommand_NewChannelReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_NEW_CHANNEL,
		Payload: &MACCommand_NewChannelReq_{
			NewChannelReq: pld,
		},
	}
}

// MACCommand returns the NewChannelAns MAC command as a *MACCommand.
func (pld *MACCommand_NewChannelAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_NEW_CHANNEL,
		Payload: &MACCommand_NewChannelAns_{
			NewChannelAns: pld,
		},
	}
}

// MACCommand returns the DLChannelReq MAC command as a *MACCommand.
func (pld *MACCommand_DLChannelReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_DL_CHANNEL,
		Payload: &MACCommand_DlChannelReq{
			DlChannelReq: pld,
		},
	}
}

// MACCommand returns the DLChannelAns MAC command as a *MACCommand.
func (pld *MACCommand_DLChannelAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_DL_CHANNEL,
		Payload: &MACCommand_DlChannelAns{
			DlChannelAns: pld,
		},
	}
}

// MACCommand returns the RxTimingSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_RxTimingSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RX_TIMING_SETUP,
		Payload: &MACCommand_RxTimingSetupReq_{
			RxTimingSetupReq: pld,
		},
	}
}

// MACCommand returns the TxParamSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_TxParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_TX_PARAM_SETUP,
		Payload: &MACCommand_TxParamSetupReq_{
			TxParamSetupReq: pld,
		},
	}
}

// MACCommand returns the RekeyInd MAC command as a *MACCommand.
func (pld *MACCommand_RekeyInd) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_REKEY,
		Payload: &MACCommand_RekeyInd_{
			RekeyInd: pld,
		},
	}
}

// MACCommand returns the RekeyConf MAC command as a *MACCommand.
func (pld *MACCommand_RekeyConf) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_REKEY,
		Payload: &MACCommand_RekeyConf_{
			RekeyConf: pld,
		},
	}
}

// MACCommand returns the ADRParamSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_ADRParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_ADR_PARAM_SETUP,
		Payload: &MACCommand_AdrParamSetupReq{
			AdrParamSetupReq: pld,
		},
	}
}

// MACCommand returns the DeviceTimeAns MAC command as a *MACCommand.
func (pld *MACCommand_DeviceTimeAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_DEVICE_TIME,
		Payload: &MACCommand_DeviceTimeAns_{
			DeviceTimeAns: pld,
		},
	}
}

// MACCommand returns the ForceRejoinReq MAC command as a *MACCommand.
func (pld *MACCommand_ForceRejoinReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_FORCE_REJOIN,
		Payload: &MACCommand_ForceRejoinReq_{
			ForceRejoinReq: pld,
		},
	}
}

// MACCommand returns the RejoinParamSetupReq MAC command as a *MACCommand.
func (pld *MACCommand_RejoinParamSetupReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_REJOIN_PARAM_SETUP,
		Payload: &MACCommand_RejoinParamSetupReq_{
			RejoinParamSetupReq: pld,
		},
	}
}

// MACCommand returns the RejoinParamSetupAns MAC command as a *MACCommand.
func (pld *MACCommand_RejoinParamSetupAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_REJOIN_PARAM_SETUP,
		Payload: &MACCommand_RejoinParamSetupAns_{
			RejoinParamSetupAns: pld,
		},
	}
}

// MACCommand returns the PingSlotInfoReq MAC command as a *MACCommand.
func (pld *MACCommand_PingSlotInfoReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_PING_SLOT_INFO,
		Payload: &MACCommand_PingSlotInfoReq_{
			PingSlotInfoReq: pld,
		},
	}
}

// MACCommand returns the PingSlotChannelReq MAC command as a *MACCommand.
func (pld *MACCommand_PingSlotChannelReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_PING_SLOT_CHANNEL,
		Payload: &MACCommand_PingSlotChannelReq_{
			PingSlotChannelReq: pld,
		},
	}
}

// MACCommand returns the PingSlotChannelAns MAC command as a *MACCommand.
func (pld *MACCommand_PingSlotChannelAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_PING_SLOT_CHANNEL,
		Payload: &MACCommand_PingSlotChannelAns_{
			PingSlotChannelAns: pld,
		},
	}
}

// MACCommand returns the BeaconTimingAns MAC command as a *MACCommand.
func (pld *MACCommand_BeaconTimingAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_BEACON_TIMING,
		Payload: &MACCommand_BeaconTimingAns_{
			BeaconTimingAns: pld,
		},
	}
}

// MACCommand returns the BeaconFreqReq MAC command as a *MACCommand.
func (pld *MACCommand_BeaconFreqReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_BEACON_FREQ,
		Payload: &MACCommand_BeaconFreqReq_{
			BeaconFreqReq: pld,
		},
	}
}

// MACCommand returns the BeaconFreqAns MAC command as a *MACCommand.
func (pld *MACCommand_BeaconFreqAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_BEACON_FREQ,
		Payload: &MACCommand_BeaconFreqAns_{
			BeaconFreqAns: pld,
		},
	}
}

// MACCommand returns the DeviceModeInd MAC command as a *MACCommand.
func (pld *MACCommand_DeviceModeInd) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_DEVICE_MODE,
		Payload: &MACCommand_DeviceModeInd_{
			DeviceModeInd: pld,
		},
	}
}

// MACCommand returns the DeviceModeConf MAC command as a *MACCommand.
func (pld *MACCommand_DeviceModeConf) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_DEVICE_MODE,
		Payload: &MACCommand_DeviceModeConf_{
			DeviceModeConf: pld,
		},
	}
}

// MACCommand returns the RelayConfReq MAC command as a *MACCommand.
func (pld *MACCommand_RelayConfReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_CONF,
		Payload: &MACCommand_RelayConfReq_{
			RelayConfReq: pld,
		},
	}
}

// MACCommand returns the RelayConfAns MAC command as a *MACCommand.
func (pld *MACCommand_RelayConfAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_CONF,
		Payload: &MACCommand_RelayConfAns_{
			RelayConfAns: pld,
		},
	}
}

// MACCommand returns the EndDeviceConfReq MAC command as a *MACCommand.
func (pld *MACCommand_RelayEndDeviceConfReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF,
		Payload: &MACCommand_RelayEndDeviceConfReq_{
			RelayEndDeviceConfReq: pld,
		},
	}
}

// MACCommand returns the EndDeviceConfAns MAC command as a *MACCommand.
func (pld *MACCommand_RelayEndDeviceConfAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF,
		Payload: &MACCommand_RelayEndDeviceConfAns_{
			RelayEndDeviceConfAns: pld,
		},
	}
}

// MACCommand returns the UpdateUplinkListReq MAC command as a *MACCommand.
func (pld *MACCommand_RelayUpdateUplinkListReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST,
		Payload: &MACCommand_RelayUpdateUplinkListReq_{
			RelayUpdateUplinkListReq: pld,
		},
	}
}

// MACCommand returns the UpdateUplinkListAns MAC command as a *MACCommand.
func (pld *MACCommand_RelayUpdateUplinkListAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST,
		Payload: &MACCommand_RelayUpdateUplinkListAns_{
			RelayUpdateUplinkListAns: pld,
		},
	}
}

// MACCommand returns the RelayCtrlUplinkListReq MAC command as a *MACCommand.
func (pld *MACCommand_RelayCtrlUplinkListReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST,
		Payload: &MACCommand_RelayCtrlUplinkListReq_{
			RelayCtrlUplinkListReq: pld,
		},
	}
}

// MACCommand returns the RelayCtrlUplinkListAns MAC command as a *MACCommand.
func (pld *MACCommand_RelayCtrlUplinkListAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST,
		Payload: &MACCommand_RelayCtrlUplinkListAns_{
			RelayCtrlUplinkListAns: pld,
		},
	}
}

// MACCommand returns the ConfigureFwdLimitReq MAC command as a *MACCommand.
func (pld *MACCommand_RelayConfigureFwdLimitReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT,
		Payload: &MACCommand_RelayConfigureFwdLimitReq_{
			RelayConfigureFwdLimitReq: pld,
		},
	}
}

// MACCommand returns the ConfigureFwdLimitAns MAC command as a *MACCommand.
func (pld *MACCommand_RelayConfigureFwdLimitAns) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT,
		Payload: &MACCommand_RelayConfigureFwdLimitAns_{
			RelayConfigureFwdLimitAns: pld,
		},
	}
}

// MACCommand returns the NotifyNewEndDeviceReq MAC command as a *MACCommand.
func (pld *MACCommand_RelayNotifyNewEndDeviceReq) MACCommand() *MACCommand {
	return &MACCommand{
		Cid: MACCommandIdentifier_CID_RELAY_NOTIFY_NEW_END_DEVICE,
		Payload: &MACCommand_RelayNotifyNewEndDeviceReq_{
			RelayNotifyNewEndDeviceReq: pld,
		},
	}
}

// sanitizableMACCommandPayload is a MAC command payload that can be sanitized.
type sanitizableMACCommandPayload interface {
	// sanitizedMACCommand returns a sanitized copy of the MAC command.
	sanitizedMACCommand() *MACCommand
}

// Sanitized returns a sanitized copy of the MAC command.
func (m *MACCommand) Sanitized() *MACCommand {
	if v, ok := m.GetPayload().(sanitizableMACCommandPayload); ok {
		return v.sanitizedMACCommand()
	}
	return m
}

var _ sanitizableMACCommandPayload = (*MACCommand_RelayUpdateUplinkListReq_)(nil)

// Sanitized returns a sanitized copy of the payload.
func (pld *MACCommand_RelayUpdateUplinkListReq) Sanitized() *MACCommand_RelayUpdateUplinkListReq {
	pld = Clone(pld)
	pld.RootWorSKey = nil
	return pld
}

// sanitizedMACCommand returns a sanitized copy of the MAC command.
func (pld *MACCommand_RelayUpdateUplinkListReq_) sanitizedMACCommand() *MACCommand {
	return pld.RelayUpdateUplinkListReq.Sanitized().MACCommand()
}
