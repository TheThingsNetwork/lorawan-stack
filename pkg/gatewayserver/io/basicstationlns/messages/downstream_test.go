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
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRateIndex: 2,
						Frequency:     868500000,
						RequestInfo: &ttnpb.RequestInfo{
							Class:        ttnpb.CLASS_A,
							RxWindow:     1,
							AntennaIndex: 2,
						},
					},
				},
			},
			ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
			DownlinkMessage{
				DevEUI:      basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				DeviceClass: 0,
				Pdu:         "Ymxhamthc25kJ3M==",
				RxDelay:     1,
				Rx2DR:       2,
				Rx2Freq:     868500000,
				RCtx:        2,
				Priority:    0,
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
			dnmsg.XTime = tc.LNSDownlinkMessage.XTime
			if !(a.So(dnmsg, should.Resemble, tc.LNSDownlinkMessage)) {
				t.Fatalf("Invalid DownlinkMessage: %v", dnmsg)
			}
		})
	}
}
