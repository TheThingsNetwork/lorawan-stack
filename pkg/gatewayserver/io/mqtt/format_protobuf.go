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

package mqtt

import (
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type protobuf struct {
	topics.Layout
}

func (protobuf) FromDownlink(down *ttnpb.DownlinkMessage) ([]byte, error) {
	gwDown := &ttnpb.GatewayDown{
		DownlinkMessage: down,
	}
	return gwDown.Marshal()
}

func (protobuf) ToUplink(message []byte) (*ttnpb.UplinkMessage, error) {
	uplink := &ttnpb.UplinkMessage{}
	if err := uplink.Unmarshal(message); err != nil {
		return nil, err
	}
	return uplink, nil
}

func (protobuf) ToStatus(message []byte) (*ttnpb.GatewayStatus, error) {
	status := &ttnpb.GatewayStatus{}
	if err := status.Unmarshal(message); err != nil {
		return nil, err
	}
	return status, nil
}

func (protobuf) ToTxAck(message []byte) (*ttnpb.TxAcknowledgment, error) {
	ack := &ttnpb.TxAcknowledgment{}
	if err := ack.Unmarshal(message); err != nil {
		return nil, err
	}
	return ack, nil
}

// Protobuf is a format that uses Protocol Buffers marshaling and unmarshaling.
var Protobuf Format = &protobuf{
	Layout: topics.Default,
}
