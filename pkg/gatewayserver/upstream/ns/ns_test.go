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

package ns

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	. "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/upstream/mock"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var timeout = (1 << 5) * test.Delay

func TestNSHandler(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	gtwIds := ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"}
	ns, nsAddr := mock.StartNS(ctx)
	c := NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				NetworkServer: nsAddr,
			},
		},
	})
	StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_NETWORK_SERVER)
	h := NewHandler(ctx, "cluster", c, nil)

	for _, tc := range []struct {
		Name                 string
		UplinkMessage        *ttnpb.UplinkMessage
		EndDeviceIdentifiers ttnpb.EndDeviceIdentifiers
	}{
		{
			Name: "OneUplink",
			UplinkMessage: &ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
						DevEUI:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
						DevNonce: [2]byte{0x46, 0x50},
					}},
				},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: gtwIds,
					RSSI:               89,
					ChannelRSSI:        89,
					SNR:                9.25,
				}},
				Settings: ttnpb.TxSettings{
					Frequency:  868300000,
					CodingRate: "4/5",
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			gtwUp := &ttnpb.GatewayUp{UplinkMessages: []*ttnpb.UplinkMessage{tc.UplinkMessage}}
			err := h.HandleUp(ctx, gtwIds, tc.EndDeviceIdentifiers, gtwUp)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Error sending upstream message: %v", err)
			}
			select {
			case msg := <-ns.Up():
				if !a.So(msg, should.Resemble, tc.UplinkMessage) {
					t.Fatalf("Unexpected upstream message: %v", msg)
				}
			case <-time.After(timeout):
				t.Fatal("Expected uplink event timeout")
			}
		})
	}
}
