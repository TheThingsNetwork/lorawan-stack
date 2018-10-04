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
	evtMACTxParamRequest = events.Define("ns.mac.tx_param.request", "request transmit parameter setup") // TODO(#988): publish when requesting
	evtMACTxParamAccept  = events.Define("ns.mac.tx_param.accept", "device accepted transmit parameter setup request")
)

func enqueueTxParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice) {
	if dev.MACState.DesiredParameters.MaxEIRP == dev.MACState.CurrentParameters.MaxEIRP &&
		dev.MACState.DesiredParameters.DownlinkDwellTime == dev.MACState.CurrentParameters.DownlinkDwellTime &&
		dev.MACState.DesiredParameters.UplinkDwellTime == dev.MACState.CurrentParameters.UplinkDwellTime {
		return
	}

	dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_TxParamSetupReq{
		MaxEIRPIndex:      ttnpb.Float32ToDeviceEIRP(dev.MACState.DesiredParameters.MaxEIRP),
		DownlinkDwellTime: dev.MACState.DesiredParameters.DownlinkDwellTime,
		UplinkDwellTime:   dev.MACState.DesiredParameters.UplinkDwellTime,
	}).MACCommand())
}

func handleTxParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) (err error) {
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_TX_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetTxParamSetupReq()

		dev.MACState.CurrentParameters.DownlinkDwellTime = req.DownlinkDwellTime
		dev.MACState.CurrentParameters.UplinkDwellTime = req.UplinkDwellTime
		dev.MACState.CurrentParameters.MaxEIRP = ttnpb.DeviceEIRPToFloat32(req.MaxEIRPIndex)

		if ttnpb.Float32ToDeviceEIRP(dev.MACState.DesiredParameters.MaxEIRP) == req.MaxEIRPIndex {
			dev.MACState.DesiredParameters.MaxEIRP = dev.MACState.CurrentParameters.MaxEIRP
		}

		events.Publish(evtMACTxParamAccept(ctx, dev.EndDeviceIdentifiers, req))
		return nil

	}, dev.MACState.PendingRequests...)
	return
}
