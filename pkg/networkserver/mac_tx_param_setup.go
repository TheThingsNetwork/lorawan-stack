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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueTxParamSetupRequest = defineEnqueueMACRequestEvent("tx_param_setup", "Tx parameter setup")()
	evtReceiveTxParamSetupAnswer  = defineReceiveMACAnswerEvent("tx_param_setup", "Tx parameter setup")()
)

func deviceNeedsTxParamSetupReq(dev *ttnpb.EndDevice, phy band.Band) bool {
	if !phy.TxParamSetupReqSupport || dev.MACState == nil || dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) < 0 {
		return false
	}
	if dev.MACState.DesiredParameters.MaxEIRP != dev.MACState.CurrentParameters.MaxEIRP {
		return true
	}
	if dev.MACState.DesiredParameters.UplinkDwellTime != nil &&
		(dev.MACState.CurrentParameters.UplinkDwellTime == nil ||
			dev.MACState.DesiredParameters.UplinkDwellTime.Value != dev.MACState.CurrentParameters.UplinkDwellTime.Value) {
		return true
	}
	if dev.MACState.DesiredParameters.DownlinkDwellTime != nil &&
		(dev.MACState.CurrentParameters.DownlinkDwellTime == nil ||
			dev.MACState.DesiredParameters.DownlinkDwellTime.Value != dev.MACState.CurrentParameters.DownlinkDwellTime.Value) {
		return true
	}
	return false
}

func enqueueTxParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, phy band.Band) macCommandEnqueueState {
	if !deviceNeedsTxParamSetupReq(dev, phy) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st macCommandEnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_TX_PARAM_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, []events.DefinitionDataClosure, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		req := &ttnpb.MACCommand_TxParamSetupReq{
			MaxEIRPIndex:      lorawan.Float32ToDeviceEIRP(dev.MACState.DesiredParameters.MaxEIRP),
			DownlinkDwellTime: dev.MACState.DesiredParameters.DownlinkDwellTime.GetValue(),
			UplinkDwellTime:   dev.MACState.DesiredParameters.UplinkDwellTime.GetValue(),
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"max_eirp_index", req.MaxEIRPIndex,
			"downlink_dwell_time", req.DownlinkDwellTime,
			"uplink_dwell_time", req.UplinkDwellTime,
		)).Debug("Enqueued TxParamSetupReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			[]events.DefinitionDataClosure{
				evtEnqueueTxParamSetupRequest.BindData(req),
			},
			true
	}, dev.MACState.PendingRequests...)
	return st
}

func handleTxParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) ([]events.DefinitionDataClosure, error) {
	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_TX_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetTxParamSetupReq()

		dev.MACState.CurrentParameters.MaxEIRP = lorawan.DeviceEIRPToFloat32(req.MaxEIRPIndex)
		dev.MACState.CurrentParameters.DownlinkDwellTime = &pbtypes.BoolValue{Value: req.DownlinkDwellTime}
		dev.MACState.CurrentParameters.UplinkDwellTime = &pbtypes.BoolValue{Value: req.UplinkDwellTime}

		if lorawan.Float32ToDeviceEIRP(dev.MACState.DesiredParameters.MaxEIRP) == req.MaxEIRPIndex {
			dev.MACState.DesiredParameters.MaxEIRP = dev.MACState.CurrentParameters.MaxEIRP
		}
		return nil
	}, dev.MACState.PendingRequests...)
	return []events.DefinitionDataClosure{
		evtReceiveTxParamSetupAnswer.BindData(nil),
	}, err
}
