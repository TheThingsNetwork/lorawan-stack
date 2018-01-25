// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gwpool_test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/gwpool"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPoolDownlinks(t *testing.T) {
	a := assertions.New(t)

	p := gwpool.NewPool(test.GetLogger(t), time.Millisecond)

	gatewayID := "gateway"
	link := &dummyLink{
		AcceptDownlink: true,

		NextUplink: make(chan *ttnpb.GatewayUp),
	}
	_, err := p.Subscribe(ttnpb.GatewayIdentifier{GatewayID: gatewayID}, link, ttnpb.FrequencyPlan{BandID: band.EU_863_870})
	a.So(err, should.BeNil)

	err = p.Send(ttnpb.GatewayIdentifier{GatewayID: "gateway-nonexistant"}, &ttnpb.GatewayDown{})
	a.So(err, should.NotBeNil)

	err = p.Send(ttnpb.GatewayIdentifier{GatewayID: gatewayID}, &ttnpb.GatewayDown{
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
}
