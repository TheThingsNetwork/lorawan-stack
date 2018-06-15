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
	"testing"

	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleDLChannelAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_DLChannelAns
		Error            error
	}{
		{
			Name:     "nil payload",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Payload:  nil,
			Error:    common.ErrMissingPayload.New(nil),
		},
		{
			Name:     "no request",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Payload:  ttnpb.NewPopulatedMACCommand_DLChannelAns(test.Randy, false),
			Error:    ErrMACRequestNotFound.New(nil),
		},
		{
			Name: "both ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
				PendingMACCommands: []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_DLChannelReq{
						ChannelIndex: 4,
						Frequency:    42,
					}).MACCommand(),
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					// TODO: Handle (https://github.com/TheThingsIndustries/ttn/issues/834)
				},
				PendingMACCommands: []*ttnpb.MACCommand{},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				FrequencyAck:    true,
				ChannelIndexAck: true,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleDLChannelAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil {
				a.So(err, should.DescribeError, errors.Descriptor(tc.Error))
			} else {
				a.So(err, should.BeNil)
			}

			if !a.So(dev, should.Resemble, tc.Expected) {
				pretty.Ldiff(t, dev, tc.Expected)
			}
		})
	}
}
