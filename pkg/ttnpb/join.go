// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// FieldIsZero returns whether path p is zero.
func (v *JoinRequest) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "cf_list":
		return v.CFList == nil
	case "cf_list.ch_masks":
		return v.CFList.FieldIsZero("ch_masks")
	case "cf_list.freq":
		return v.CFList.FieldIsZero("freq")
	case "cf_list.type":
		return v.CFList.FieldIsZero("type")
	case "consumed_airtime":
		return v.ConsumedAirtime == nil
	case "correlation_ids":
		return v.CorrelationIDs == nil
	case "dev_addr":
		return v.DevAddr == types.DevAddr{}
	case "downlink_settings":
		return v.DownlinkSettings == DLSettings{}
	case "downlink_settings.opt_neg":
		return v.DownlinkSettings.FieldIsZero("opt_neg")
	case "downlink_settings.rx1_dr_offset":
		return v.DownlinkSettings.FieldIsZero("rx1_dr_offset")
	case "downlink_settings.rx2_dr":
		return v.DownlinkSettings.FieldIsZero("rx2_dr")
	case "net_id":
		return v.NetId == types.NetID{}
	case "payload":
		return v.Payload == nil
	case "payload.Payload":
		return v.Payload.FieldIsZero("Payload")
	case "payload.Payload.join_accept_payload":
		return v.Payload.FieldIsZero("Payload.join_accept_payload")
	case "payload.Payload.join_accept_payload.cf_list":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.cf_list")
	case "payload.Payload.join_accept_payload.cf_list.ch_masks":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.cf_list.ch_masks")
	case "payload.Payload.join_accept_payload.cf_list.freq":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.cf_list.freq")
	case "payload.Payload.join_accept_payload.cf_list.type":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.cf_list.type")
	case "payload.Payload.join_accept_payload.dev_addr":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.dev_addr")
	case "payload.Payload.join_accept_payload.dl_settings":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.dl_settings")
	case "payload.Payload.join_accept_payload.dl_settings.opt_neg":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.dl_settings.opt_neg")
	case "payload.Payload.join_accept_payload.dl_settings.rx1_dr_offset":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.dl_settings.rx1_dr_offset")
	case "payload.Payload.join_accept_payload.dl_settings.rx2_dr":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.dl_settings.rx2_dr")
	case "payload.Payload.join_accept_payload.encrypted":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.encrypted")
	case "payload.Payload.join_accept_payload.join_nonce":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.join_nonce")
	case "payload.Payload.join_accept_payload.net_id":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.net_id")
	case "payload.Payload.join_accept_payload.rx_delay":
		return v.Payload.FieldIsZero("Payload.join_accept_payload.rx_delay")
	case "payload.Payload.join_request_payload":
		return v.Payload.FieldIsZero("Payload.join_request_payload")
	case "payload.Payload.join_request_payload.dev_eui":
		return v.Payload.FieldIsZero("Payload.join_request_payload.dev_eui")
	case "payload.Payload.join_request_payload.dev_nonce":
		return v.Payload.FieldIsZero("Payload.join_request_payload.dev_nonce")
	case "payload.Payload.join_request_payload.join_eui":
		return v.Payload.FieldIsZero("Payload.join_request_payload.join_eui")
	case "payload.Payload.mac_payload":
		return v.Payload.FieldIsZero("Payload.mac_payload")
	case "payload.Payload.mac_payload.decoded_payload":
		return v.Payload.FieldIsZero("Payload.mac_payload.decoded_payload")
	case "payload.Payload.mac_payload.f_hdr":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr")
	case "payload.Payload.mac_payload.f_hdr.dev_addr":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.dev_addr")
	case "payload.Payload.mac_payload.f_hdr.f_cnt":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.f_cnt")
	case "payload.Payload.mac_payload.f_hdr.f_ctrl":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.f_ctrl")
	case "payload.Payload.mac_payload.f_hdr.f_ctrl.ack":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.f_ctrl.ack")
	case "payload.Payload.mac_payload.f_hdr.f_ctrl.adr":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.f_ctrl.adr")
	case "payload.Payload.mac_payload.f_hdr.f_ctrl.adr_ack_req":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.f_ctrl.adr_ack_req")
	case "payload.Payload.mac_payload.f_hdr.f_ctrl.class_b":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.f_ctrl.class_b")
	case "payload.Payload.mac_payload.f_hdr.f_ctrl.f_pending":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.f_ctrl.f_pending")
	case "payload.Payload.mac_payload.f_hdr.f_opts":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_hdr.f_opts")
	case "payload.Payload.mac_payload.f_port":
		return v.Payload.FieldIsZero("Payload.mac_payload.f_port")
	case "payload.Payload.mac_payload.frm_payload":
		return v.Payload.FieldIsZero("Payload.mac_payload.frm_payload")
	case "payload.Payload.mac_payload.full_f_cnt":
		return v.Payload.FieldIsZero("Payload.mac_payload.full_f_cnt")
	case "payload.Payload.rejoin_request_payload":
		return v.Payload.FieldIsZero("Payload.rejoin_request_payload")
	case "payload.Payload.rejoin_request_payload.dev_eui":
		return v.Payload.FieldIsZero("Payload.rejoin_request_payload.dev_eui")
	case "payload.Payload.rejoin_request_payload.join_eui":
		return v.Payload.FieldIsZero("Payload.rejoin_request_payload.join_eui")
	case "payload.Payload.rejoin_request_payload.net_id":
		return v.Payload.FieldIsZero("Payload.rejoin_request_payload.net_id")
	case "payload.Payload.rejoin_request_payload.rejoin_cnt":
		return v.Payload.FieldIsZero("Payload.rejoin_request_payload.rejoin_cnt")
	case "payload.Payload.rejoin_request_payload.rejoin_type":
		return v.Payload.FieldIsZero("Payload.rejoin_request_payload.rejoin_type")
	case "payload.m_hdr":
		return v.Payload.FieldIsZero("m_hdr")
	case "payload.m_hdr.m_type":
		return v.Payload.FieldIsZero("m_hdr.m_type")
	case "payload.m_hdr.major":
		return v.Payload.FieldIsZero("m_hdr.major")
	case "payload.mic":
		return v.Payload.FieldIsZero("mic")
	case "raw_payload":
		return v.RawPayload == nil
	case "rx_delay":
		return v.RxDelay == 0
	case "selected_mac_version":
		return v.SelectedMACVersion == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}
