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

package pool_test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/pool"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPoolDownlinks(t *testing.T) {
	a := assertions.New(t)

	p := pool.NewPool(test.GetLogger(t), time.Millisecond)

	gatewayID := "gateway"
	link := &dummyLink{
		AcceptDownlink: true,

		NextUplink: make(chan *ttnpb.GatewayUp),
	}
	_, err := p.Subscribe(gatewayID, link, ttnpb.FrequencyPlan{BandID: band.EU_863_870})
	a.So(err, should.BeNil)

	obs, err := p.GetGatewayObservations(gatewayID)
	a.So(err, should.BeNil)
	a.So(obs.DownlinkCount, should.Equal, 0)
	a.So(obs.LastDownlinkReceivedAt, should.BeNil)

	err = p.Send("gateway-nonexistant", &ttnpb.GatewayDown{DownlinkMessage: &ttnpb.DownlinkMessage{}})
	a.So(err, should.NotBeNil)
	obs, err = p.GetGatewayObservations(gatewayID)
	a.So(err, should.BeNil)
	a.So(obs.DownlinkCount, should.Equal, 0)
	a.So(obs.LastDownlinkReceivedAt, should.BeNil)

	err = p.Send(gatewayID, &ttnpb.GatewayDown{})
	a.So(err, should.NotBeNil)
	err = p.Send(gatewayID, &ttnpb.GatewayDown{
		DownlinkMessage: &ttnpb.DownlinkMessage{
			Settings: ttnpb.TxSettings{
				Bandwidth:       125000,
				CodingRate:      "4/5",
				Frequency:       866000000,
				Modulation:      ttnpb.Modulation_LORA,
				SpreadingFactor: 7,
			},
			RawPayload: []byte{},
		},
	})
	a.So(err, should.BeNil)
	obs, err = p.GetGatewayObservations(gatewayID)
	a.So(err, should.BeNil)
	a.So(obs.DownlinkCount, should.Equal, 1)
	a.So(obs.LastDownlinkReceivedAt.Unix(), should.AlmostEqual, time.Now().Unix(), 3)
}
