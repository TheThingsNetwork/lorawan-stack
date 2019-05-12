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

package formatters

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

type protobuf struct{}

func (protobuf) FromUp(msg *ttnpb.ApplicationUp) ([]byte, error) {
	return msg.Marshal()
}

func (protobuf) ToDownlinks(buf []byte) (*ttnpb.ApplicationDownlinks, error) {
	res := &ttnpb.ApplicationDownlinks{}
	if err := res.Unmarshal(buf); err != nil {
		return nil, err
	}
	return res, nil
}

func (protobuf) ToDownlinkQueueOperation(buf []byte) (*ttnpb.DownlinkQueueOperation, error) {
	res := &ttnpb.DownlinkQueueOperation{}
	if err := res.Unmarshal(buf); err != nil {
		return nil, err
	}
	return res, nil
}

// Protobuf is a formatter that uses proto marshaling.
var Protobuf Formatter = &protobuf{}
