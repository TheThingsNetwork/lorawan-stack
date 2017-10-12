// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gwpool_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/gwpool"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPoolDownlinks(t *testing.T) {
	a := assertions.New(t)

	p := gwpool.NewPool(log.Noop)

	gatewayID := "gateway"
	link := &dummyLink{
		AcceptDownlink: true,

		NextUplink: make(chan *ttnpb.GatewayUp),
	}
	p.Subscribe(ttnpb.GatewayIdentifier{GatewayID: gatewayID}, link)

	err := p.Send(ttnpb.GatewayIdentifier{GatewayID: "gateway-nonexistant"}, &ttnpb.GatewayDown{})
	a.So(err, should.NotBeNil)

	err = p.Send(ttnpb.GatewayIdentifier{GatewayID: gatewayID}, &ttnpb.GatewayDown{})
	a.So(err, should.BeNil)
}
