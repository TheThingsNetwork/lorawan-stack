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
	"encoding/gob"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

func init() {
	gob.Register(&Message_MACPayload{})
	gob.Register(&Message_JoinRequestPayload{})
	gob.Register(&Message_JoinAcceptPayload{})
	gob.Register(&MACCommand_CID{})
	gob.Register(&MACCommand_Proprietary{})
	gob.Register(&MACCommand_Proprietary_{})
	gob.Register(&MACCommand_ResetInd{})
	gob.Register(&MACCommand_ResetInd_{})
	gob.Register(&MACCommand_ResetConf{})
	gob.Register(&MACCommand_ResetConf_{})
	gob.Register(&MACCommand_LinkCheckAns{})
	gob.Register(&MACCommand_LinkCheckAns_{})
	gob.Register(&MACCommand_LinkADRReq{})
	gob.Register(&MACCommand_LinkADRReq_{})
	gob.Register(&MACCommand_LinkADRAns{})
	gob.Register(&MACCommand_LinkADRAns_{})
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
	gob.Register(&MACCommand_ADRParamSetupReq{})
	gob.Register(&MACCommand_ADRParamSetupReq_{})
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

// IsValid reports whether v represents a valid MACVersion.
func (v MACVersion) Validate() error {
	if v < 0 || v >= MACVersion(len(MACVersion_name)) {
		return errors.Errorf("expected MACVersion to be between %d and %d, got %d", 0, len(MACVersion_name)-1, v)
	}
	return nil
}

// EncryptFOpts reports whether v requires MAC commands in FOpts to be encrypted.
// EncryptFOpts panics, if v.Validate() returns non-nil error.
func (v MACVersion) EncryptFOpts() bool {
	switch v {
	case MAC_V1_0, MAC_V1_0_1, MAC_V1_0_2:
		return false
	case MAC_V1_1:
		return true
	}
	panic(errors.Errorf("Unknown MACVersion: %v", v))
}
