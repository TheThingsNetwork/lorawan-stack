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

package mac

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtReceiveLinkCheckRequest = defineReceiveMACRequestEvent(
		"link_check", "link check",
	)()
	EvtEnqueueLinkCheckAnswer = defineEnqueueMACAnswerEvent(
		"link_check", "link check",
		events.WithDataType(&ttnpb.MACCommand_LinkCheckAns{}),
	)()
)

func HandleLinkCheckReq(ctx context.Context, dev *ttnpb.EndDevice, msg *ttnpb.UplinkMessage) (events.Builders, error) {
	evs := events.Builders{
		EvtReceiveLinkCheckRequest,
	}

	var floor float32
	if dr, ok := msg.Settings.DataRate.Modulation.(*ttnpb.DataRate_LoRa); ok {
		floor, ok = demodulationFloor[dr.LoRa.SpreadingFactor][dr.LoRa.Bandwidth]
		if !ok {
			return evs, ErrInvalidDataRate.New()
		}
	}
	if len(msg.RxMetadata) == 0 {
		return evs, nil
	}

	gtwCount, maxSNR := RXMetadataStats(ctx, msg.RxMetadata)
	ans := &ttnpb.MACCommand_LinkCheckAns{
		Margin:       uint32(uint8(maxSNR - floor)),
		GatewayCount: uint32(gtwCount),
	}
	dev.MACState.QueuedResponses = append(dev.MACState.QueuedResponses, ans.MACCommand())
	return append(evs,
		EvtEnqueueLinkCheckAnswer.With(events.WithData(ans)),
	), nil
}
