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

func TestHandleLinkCheckReq(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Message          *ttnpb.UplinkMessage
		Error            error
	}{
		{
			Name:     "SF13BW250",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Message: &ttnpb.UplinkMessage{
				Settings: ttnpb.TxSettings{
					SpreadingFactor: 13,
					Bandwidth:       250,
				},
			},
			Error: errInvalidDataRate,
		},
		{
			Name:   "SF12BW250/1 gateway/empty queue",
			Device: &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{
				QueuedMACResponses: []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_LinkCheckAns{
						Margin:       42, // 25-(-17)
						GatewayCount: 1,
					}).MACCommand(),
				},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: ttnpb.TxSettings{
					SpreadingFactor: 12,
					Bandwidth:       250,
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIdentifiers: ttnpb.GatewayIdentifiers{
							GatewayID: "test",
						},
						SNR: 25,
					},
				},
			},
			Error: nil,
		},
		{
			Name: "SF12BW250/1 gateway/non-empty queue",
			Device: &ttnpb.EndDevice{
				QueuedMACResponses: []*ttnpb.MACCommand{
					{},
					{},
					{},
				},
			},
			Expected: &ttnpb.EndDevice{
				QueuedMACResponses: []*ttnpb.MACCommand{
					{},
					{},
					{},
					(&ttnpb.MACCommand_LinkCheckAns{
						Margin:       42, // 25-(-17)
						GatewayCount: 1,
					}).MACCommand(),
				},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: ttnpb.TxSettings{
					SpreadingFactor: 12,
					Bandwidth:       250,
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIdentifiers: ttnpb.GatewayIdentifiers{
							GatewayID: "test",
						},
						SNR: 25,
					},
				},
			},
			Error: nil,
		},
		{
			Name: "SF12BW250/3 gateways/non-empty queue",
			Device: &ttnpb.EndDevice{
				QueuedMACResponses: []*ttnpb.MACCommand{
					{},
					{},
					{},
				},
			},
			Expected: &ttnpb.EndDevice{
				QueuedMACResponses: []*ttnpb.MACCommand{
					{},
					{},
					{},
					(&ttnpb.MACCommand_LinkCheckAns{
						Margin:       42, // 25-(-17)
						GatewayCount: 3,
					}).MACCommand(),
				},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: ttnpb.TxSettings{
					SpreadingFactor: 12,
					Bandwidth:       250,
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIdentifiers: ttnpb.GatewayIdentifiers{
							GatewayID: "test",
						},
						SNR: 24,
					},
					{
						GatewayIdentifiers: ttnpb.GatewayIdentifiers{
							GatewayID: "test2",
						},
						SNR: 25,
					},
					{
						GatewayIdentifiers: ttnpb.GatewayIdentifiers{
							GatewayID: "test3",
						},
						SNR: 2,
					},
				},
			},
			Error: nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleLinkCheckReq(test.Context(), dev, tc.Message)
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
