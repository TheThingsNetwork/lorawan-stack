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

package mqtt

import (
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type pb struct {
	topics.Layout
}

func (pb) FromDownlink(down *ttnpb.DownlinkMessage) ([]byte, error) {
	gwDown := &ttnpb.GatewayDown{
		DownlinkMessage: down,
	}
	return gwDown.Marshal()
}

func (pb) ToUplink(message []byte) (*ttnpb.UplinkMessage, error) {
	uplink := &ttnpb.UplinkMessage{}
	if err := uplink.Unmarshal(message); err != nil {
		return nil, err
	}
	return uplink, nil
}

func (pb) ToStatus(message []byte) (*ttnpb.GatewayStatus, error) {
	status := &ttnpb.GatewayStatus{}
	if err := status.Unmarshal(message); err != nil {
		return nil, err
	}
	return status, nil
}

func (pb) ToTxAck(message []byte) (*ttnpb.TxAcknowledgment, error) {
	ack := &ttnpb.TxAcknowledgment{}
	if err := ack.Unmarshal(message); err != nil {
		return nil, err
	}
	return ack, nil
}

// Protobuf is a formatter that uses proto marshaling and unmarshaling.
var Protobuf Formatter = &pb{
	Layout: topics.Default,
}
