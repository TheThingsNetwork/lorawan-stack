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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var timeout = (1 << 5) * test.Delay

func TestNSHandler(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	gtwIDs := ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
	ns, nsAddr := mock.StartNS(ctx)
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				NetworkServer: nsAddr,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_NETWORK_SERVER)
	h := NewHandler(ctx, c, c, nil)

	for _, tc := range []struct {
		Name                 string
		Message              *ttnpb.GatewayUplinkMessage
		EndDeviceIdentifiers *ttnpb.EndDeviceIdentifiers
	}{
		{
			Name: "OneUplink",
			Message: &ttnpb.GatewayUplinkMessage{
				BandId: band.EU_863_870,
				Message: &ttnpb.UplinkMessage{
					Payload: &ttnpb.Message{
						MHdr: &ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
						Mic:  []byte{0x4E, 0x61, 0xBC, 0x00},
						Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
							JoinEui:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}.Bytes(),
							DevEui:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(),
							DevNonce: []byte{0x46, 0x50},
						}},
					},
					RxMetadata: []*ttnpb.RxMetadata{{
						GatewayIds:  &gtwIDs,
						Rssi:        89,
						ChannelRssi: 89,
						Snr:         9.25,
					}},
					Settings: &ttnpb.TxSettings{
						Frequency: 868300000,
						DataRate: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
							SpreadingFactor: 11,
							Bandwidth:       125000,
							CodingRate:      band.Cr4_5,
						}}},
					},
				},
			},
			EndDeviceIdentifiers: &ttnpb.EndDeviceIdentifiers{
				DeviceId: "test-device",
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			err := h.HandleUplink(ctx, nil, tc.EndDeviceIdentifiers, tc.Message)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Error sending upstream message: %v", err)
			}
			select {
			case msg := <-ns.Up():
				if !a.So(msg, should.Resemble, tc.Message.Message) {
					t.Fatalf("Unexpected upstream message: %v", msg)
				}
			case <-time.After(timeout):
				t.Fatal("Expected uplink event timeout")
			}
		})
	}
}
