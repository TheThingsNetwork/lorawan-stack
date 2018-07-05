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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleLinkADRAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_LinkADRAns
		Error            error
	}{
		{
			Name:     "nil payload",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Payload:  nil,
			Error:    errMissingPayload,
		},
		{
			Name:     "no request",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Payload:  ttnpb.NewPopulatedMACCommand_LinkADRAns(test.Randy, false),
			Error:    errMACRequestNotFound,
		},
		{
			Name: "1 request/all ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
				PendingMACRequests: []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_LinkADRReq{
						DataRateIndex: 4,
						TxPowerIndex:  42,
					}).MACCommand(),
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					ADRDataRateIndex: 4,
					ADRTXPowerIndex:  42,
				},
				PendingMACRequests: []*ttnpb.MACCommand{},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
		},
		{
			Name: "2 requests/all ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
				PendingMACRequests: []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_LinkADRReq{
						DataRateIndex: 4,
						TxPowerIndex:  42,
					}).MACCommand(),
					(&ttnpb.MACCommand_LinkADRReq{
						DataRateIndex: 5,
						TxPowerIndex:  43,
					}).MACCommand(),
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					ADRDataRateIndex: 5,
					ADRTXPowerIndex:  43,
				},
				PendingMACRequests: []*ttnpb.MACCommand{},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleLinkADRAns(test.Context(), dev, tc.Payload)
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
