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
	"sort"
	"time"

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtReceiveDeviceTimeRequest = defineReceiveMACRequestEvent("device_time", "device time")()
	evtEnqueueDeviceTimeAnswer  = defineEnqueueMACAnswerEvent("device_time", "device time")()
)

func handleDeviceTimeReq(ctx context.Context, dev *ttnpb.EndDevice, msg *ttnpb.UplinkMessage) ([]events.DefinitionDataClosure, error) {
	evs := []events.DefinitionDataClosure{
		evtReceiveDeviceTimeRequest.BindData(nil),
	}

	ts := make([]time.Time, 0, len(msg.RxMetadata))
	for _, md := range msg.RxMetadata {
		if md.Time == nil {
			continue
		}
		ts = append(ts, *md.Time)
	}
	if len(ts) == 0 {
		return evs, nil
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Before(ts[j])
	})

	var t time.Time
	if n := len(ts); n%2 == 1 {
		t = ts[n/2]
	} else {
		i := (n - 1) / 2
		t = time.Unix(0, (ts[i].UnixNano()+ts[i+1].UnixNano())/2)
	}

	ans := &ttnpb.MACCommand_DeviceTimeAns{
		Time: t,
	}
	dev.MACState.QueuedResponses = append(dev.MACState.QueuedResponses, ans.MACCommand())
	return append(evs,
		evtEnqueueDeviceTimeAnswer.BindData(ans),
	), nil
}
