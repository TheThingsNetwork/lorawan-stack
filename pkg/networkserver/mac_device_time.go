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
)

var (
	evtReceiveDeviceTimeRequest = defineReceiveMACRequestEvent("device_time", "device time")()
	evtEnqueueDeviceTimeAnswer  = defineEnqueueMACAnswerEvent("device_time", "device time")()
)

func handleDeviceTimeReq(ctx context.Context, dev *ttnpb.EndDevice, msg *ttnpb.UplinkMessage) ([]events.DefinitionDataClosure, error) {
	ans := &ttnpb.MACCommand_DeviceTimeAns{
		Time: msg.ReceivedAt,
	}
	dev.MACState.QueuedResponses = append(dev.MACState.QueuedResponses, ans.MACCommand())
	return []events.DefinitionDataClosure{
		evtReceiveDeviceTimeRequest.BindData(nil),
		evtEnqueueDeviceTimeAnswer.BindData(ans),
	}, nil
}
