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

package messages

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/basicstation"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func eui64Ptr(eui types.EUI64) *types.EUI64 { return &eui }

func TestDownlinkMessage(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name               string
		NSDownlinkMessage  ttnpb.DownlinkMessage
		GatewayIDs         ttnpb.GatewayIdentifiers
		LNSDownlinkMessage DownlinkMessage
		ExpectedError      error
	}{
		{
			"ValidExample",
			ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIDs: &ttnpb.EndDeviceIdentifiers{
					DeviceID: "testdevice",
					DevEUI:   eui64Ptr(types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}),
				},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_A,
						Rx1Delay:         ttnpb.RxDelay(5),
						Rx1DataRateIndex: 2,
						Rx2DataRateIndex: 0,
						Rx1Frequency:     868300000,
						Rx2Frequency:     869525000,
						Priority:         10,
					},
				},
			},
			ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
			DownlinkMessage{
				DevEUI:      basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				DeviceClass: 0,
				Pdu:         "Ymxhamthc25kJ3M==",
				RxDelay:     5,
				Rx1DR:       2,
				Rx2DR:       0,
				Rx1Freq:     868300000,
				Rx2Freq:     869525000,
				Priority:    10,
			},
			nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			dnmsg := DownlinkMessage{}
			err := dnmsg.GetFromNSDownlinkMessage(tc.GatewayIDs, tc.NSDownlinkMessage)
			if !(a.So(err, should.BeNil)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(dnmsg, should.Resemble, tc.LNSDownlinkMessage)) {
				t.Fatalf("Invalid DownlinkMessage: %v", dnmsg)
			}
		})
	}
}
