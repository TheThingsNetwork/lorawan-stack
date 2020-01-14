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

package networkserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	evtReceiveLinkCheckRequest = defineReceiveMACRequestEvent("link_check", "link check")()
	evtEnqueueLinkCheckAnswer  = defineEnqueueMACAnswerEvent("link_check", "link check")()
)

func handleLinkCheckReq(ctx context.Context, dev *ttnpb.EndDevice, msg *ttnpb.UplinkMessage) ([]events.DefinitionDataClosure, error) {
	evs := []events.DefinitionDataClosure{
		evtReceiveLinkCheckRequest.BindData(nil),
	}

	var floor float32
	if dr, ok := msg.Settings.DataRate.Modulation.(*ttnpb.DataRate_LoRa); ok {
		floor, ok = demodulationFloor[dr.LoRa.SpreadingFactor][dr.LoRa.Bandwidth]
		if !ok {
			return evs, errInvalidDataRate
		}
	}
	if len(msg.RxMetadata) == 0 {
		return evs, nil
	}

	gtws := make(map[string]struct{}, len(msg.RxMetadata))
	maxSNR := msg.RxMetadata[0].SNR
	for _, md := range msg.RxMetadata {
		gtws[unique.ID(ctx, md.GatewayIdentifiers)] = struct{}{}
		if md.SNR > maxSNR {
			maxSNR = md.SNR
		}
	}

	ans := &ttnpb.MACCommand_LinkCheckAns{
		Margin:       uint32(uint8(maxSNR - floor)),
		GatewayCount: uint32(len(gtws)),
	}
	dev.MACState.QueuedResponses = append(dev.MACState.QueuedResponses, ans.MACCommand())
	return append(evs,
		evtEnqueueLinkCheckAnswer.BindData(ans),
	), nil
}
