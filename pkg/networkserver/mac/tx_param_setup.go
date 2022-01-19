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

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtEnqueueTxParamSetupRequest = defineEnqueueMACRequestEvent(
		"tx_param_setup", "Tx parameter setup",
		events.WithDataType(&ttnpb.MACCommand_TxParamSetupReq{}),
	)()
	EvtReceiveTxParamSetupAnswer = defineReceiveMACAnswerEvent(
		"tx_param_setup", "Tx parameter setup",
	)()
)

func DeviceNeedsTxParamSetupReq(dev *ttnpb.EndDevice, phy *band.Band) bool {
	if !phy.TxParamSetupReqSupport ||
		dev.GetMulticast() ||
		dev.GetMacState() == nil ||
		dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_0_2) < 0 {
		return false
	}
	if dev.MacState.DesiredParameters.MaxEirp != dev.MacState.CurrentParameters.MaxEirp {
		return true
	}
	if dev.MacState.DesiredParameters.UplinkDwellTime != nil &&
		(dev.MacState.CurrentParameters.UplinkDwellTime == nil ||
			dev.MacState.DesiredParameters.UplinkDwellTime.Value != dev.MacState.CurrentParameters.UplinkDwellTime.Value) {
		return true
	}
	if dev.MacState.DesiredParameters.DownlinkDwellTime != nil &&
		(dev.MacState.CurrentParameters.DownlinkDwellTime == nil ||
			dev.MacState.DesiredParameters.DownlinkDwellTime.Value != dev.MacState.CurrentParameters.DownlinkDwellTime.Value) {
		return true
	}
	return false
}

func EnqueueTxParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, phy *band.Band) EnqueueState {
	if !DeviceNeedsTxParamSetupReq(dev, phy) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		req := &ttnpb.MACCommand_TxParamSetupReq{
			MaxEirpIndex:      lorawan.Float32ToDeviceEIRP(dev.MacState.DesiredParameters.MaxEirp),
			DownlinkDwellTime: dev.MacState.DesiredParameters.DownlinkDwellTime.GetValue(),
			UplinkDwellTime:   dev.MacState.DesiredParameters.UplinkDwellTime.GetValue(),
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"max_eirp_index", req.MaxEirpIndex,
			"downlink_dwell_time", req.DownlinkDwellTime,
			"uplink_dwell_time", req.UplinkDwellTime,
		)).Debug("Enqueued TxParamSetupReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			events.Builders{
				EvtEnqueueTxParamSetupRequest.With(events.WithData(req)),
			},
			true
	}, dev.MacState.PendingRequests...)
	return st
}

func HandleTxParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) (events.Builders, error) {
	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetTxParamSetupReq()

		dev.MacState.CurrentParameters.MaxEirp = lorawan.DeviceEIRPToFloat32(req.MaxEirpIndex)
		dev.MacState.CurrentParameters.DownlinkDwellTime = &ttnpb.BoolValue{Value: req.DownlinkDwellTime}
		dev.MacState.CurrentParameters.UplinkDwellTime = &ttnpb.BoolValue{Value: req.UplinkDwellTime}

		if lorawan.Float32ToDeviceEIRP(dev.MacState.DesiredParameters.MaxEirp) == req.MaxEirpIndex {
			dev.MacState.DesiredParameters.MaxEirp = dev.MacState.CurrentParameters.MaxEirp
		}
		return nil
	}, dev.MacState.PendingRequests...)
	return events.Builders{
		EvtReceiveTxParamSetupAnswer,
	}, err
}
