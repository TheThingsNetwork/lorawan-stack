// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
	gatewayIdentifiers := ttnpb.GatewayIdentifiers{GatewayID: gatewayID}
	link := &dummyLink{
		AcceptDownlink: true,

		NextUplink: make(chan *ttnpb.GatewayUp),
	}
	_, err := p.Subscribe(gatewayIdentifiers, link, ttnpb.FrequencyPlan{BandID: band.EU_863_870})
	a.So(err, should.BeNil)

	obs, err := p.GetGatewayObservations(&gatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(obs.DownlinkCount, should.Equal, 0)
	a.So(obs.LastDownlinkReceivedAt, should.BeNil)

	err = p.Send(ttnpb.GatewayIdentifiers{GatewayID: "gateway-nonexistant"}, &ttnpb.GatewayDown{DownlinkMessage: &ttnpb.DownlinkMessage{}})
	a.So(err, should.NotBeNil)
	obs, err = p.GetGatewayObservations(&gatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(obs.DownlinkCount, should.Equal, 0)
	a.So(obs.LastDownlinkReceivedAt, should.BeNil)

	err = p.Send(gatewayIdentifiers, &ttnpb.GatewayDown{})
	a.So(err, should.NotBeNil)
	err = p.Send(gatewayIdentifiers, &ttnpb.GatewayDown{
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
	obs, err = p.GetGatewayObservations(&gatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(obs.DownlinkCount, should.Equal, 1)
	a.So(obs.LastDownlinkReceivedAt.Unix(), should.AlmostEqual, time.Now().Unix(), 3)
}
