// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "encoding/gob"

func init() {
	gob.Register(&Message_MACPayload{})
	gob.Register(&Message_JoinRequestPayload{})
	gob.Register(&Message_JoinAcceptPayload{})
	gob.Register(&MACCommand_CID{})
	gob.Register(&MACCommand_Proprietary_{})
	gob.Register(&MACCommand_ResetInd_{})
	gob.Register(&MACCommand_ResetConf_{})
	gob.Register(&MACCommand_LinkCheckAns_{})
	gob.Register(&MACCommand_LinkAdrReq{})
	gob.Register(&MACCommand_LinkAdrAns{})
	gob.Register(&MACCommand_DutyCycleReq_{})
	gob.Register(&MACCommand_RxParamSetupReq_{})
	gob.Register(&MACCommand_RxParamSetupAns_{})
	gob.Register(&MACCommand_DevStatusAns_{})
	gob.Register(&MACCommand_NewChannelReq_{})
	gob.Register(&MACCommand_NewChannelAns_{})
	gob.Register(&MACCommand_DlChannelReq{})
	gob.Register(&MACCommand_DlChannelAns{})
	gob.Register(&MACCommand_RxTimingSetupReq_{})
	gob.Register(&MACCommand_TxParamSetupReq_{})
	gob.Register(&MACCommand_RekeyInd_{})
	gob.Register(&MACCommand_RekeyConf_{})
	gob.Register(&MACCommand_AdrParamSetupReq{})
	gob.Register(&MACCommand_DeviceTimeAns_{})
	gob.Register(&MACCommand_ForceRejoinReq_{})
	gob.Register(&MACCommand_RejoinParamSetupReq_{})
	gob.Register(&MACCommand_RejoinParamSetupAns_{})
	gob.Register(&MACCommand_PingSlotInfoReq_{})
	gob.Register(&MACCommand_PingSlotChannelReq_{})
	gob.Register(&MACCommand_PingSlotChannelAns_{})
	gob.Register(&MACCommand_BeaconTimingAns_{})
	gob.Register(&MACCommand_BeaconFreqReq_{})
	gob.Register(&MACCommand_BeaconFreqAns_{})
	gob.Register(&MACCommand_DeviceModeInd_{})
	gob.Register(&MACCommand_DeviceModeConf_{})
}

func (v MACVersion) Compare(other MACVersion) int {
	vStr := v.String()
	oStr := other.String()
	switch {
	case MACVersion_value[vStr] > MACVersion_value[oStr]:
		return 1
	case MACVersion_value[vStr] == MACVersion_value[oStr]:
		return 0
	default:
		return -1
	}
}
