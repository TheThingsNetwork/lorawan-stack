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

package networkserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtMACRxTimingRequest = events.Define("ns.mac.rx_timing.request", "request receive window timing setup") // TODO(#988): publish when requesting
	evtMACRxTimingAccept  = events.Define("ns.mac.rx_timing.accept", "device accepted receive window timing setup request")
)

func handleRxTimingSetupAns(ctx context.Context, dev *ttnpb.EndDevice) (err error) {
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_RX_TIMING_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetRxTimingSetupReq()

		dev.MACState.Rx1Delay = req.Delay

		events.Publish(evtMACRxTimingAccept(ctx, dev.EndDeviceIdentifiers, req))
		return nil

	}, dev.MACState.PendingRequests...)
	return
}
