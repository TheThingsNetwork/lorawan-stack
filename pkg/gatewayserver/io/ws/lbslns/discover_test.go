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

package lbslns

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws/id6"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDiscover(t *testing.T) {
	a := assertions.New(t)
	ctx := context.Background()
	var lbsLNS lbsLNS
	eui := types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}
	mockServer := mockServer{
		ids: &ttnpb.GatewayIdentifiers{
			GatewayId: "eui-1111111111111111",
			Eui:       eui.Bytes(),
		},
	}
	info := ws.ServerInfo{
		Scheme:  "wss",
		Address: "thethings.example.com:8887",
	}

	for _, tc := range []struct {
		Name             string
		Query            DiscoverQuery
		ExpectedResponse DiscoverResponse
	}{
		{
			Name: "Valid",
			Query: DiscoverQuery{
				EUI: id6.EUI{
					Prefix: "router",
					EUI64:  eui,
				},
			},
			ExpectedResponse: DiscoverResponse{
				EUI: id6.EUI{Prefix: "router", EUI64: eui},
				Muxs: id6.EUI{
					Prefix: "muxs",
				},
				URI: "wss://thethings.example.com:8887/traffic/eui-1111111111111111",
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			msg, err := json.Marshal(tc.Query)
			a.So(err, should.BeNil)
			resp := lbsLNS.HandleConnectionInfo(ctx, msg, mockServer, info, noopAssertRights)
			expected, _ := json.Marshal(tc.ExpectedResponse)
			a.So(string(resp), should.Equal, string(expected))
		})
	}
}

func noopAssertRights(context.Context, *ttnpb.GatewayIdentifiers) error { return nil }
