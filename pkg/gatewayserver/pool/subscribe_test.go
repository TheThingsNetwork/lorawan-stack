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
	"context"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/pool"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPoolUplinks(t *testing.T) {
	a := assertions.New(t)

	p := pool.NewPool(test.GetLogger(t), time.Millisecond)

	gatewayID := "gateway"
	gatewayIdentifiers := ttnpb.GatewayIdentifiers{GatewayID: gatewayID}
	link := &dummyLink{
		AcceptSendingUplinks: true,

		NextUplink: make(chan *ttnpb.GatewayUp),
	}
	emptyUplink := &ttnpb.GatewayUp{}
	upstream, err := p.Subscribe(gatewayIdentifiers, link, ttnpb.FrequencyPlan{BandID: band.EU_863_870})
	a.So(err, should.BeNil)

	obs, err := p.GetGatewayObservations(&gatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(obs.UplinkCount, should.Equal, 0)
	a.So(obs.LastUplinkReceivedAt, should.BeNil)

	go func() { link.NextUplink <- emptyUplink }()
	newUplink := <-upstream
	a.So(newUplink, should.Equal, emptyUplink)

	obs, err = p.GetGatewayObservations(&gatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(obs.UplinkCount, should.Equal, 0)
	a.So(obs.LastUplinkReceivedAt, should.BeNil)

	go func() {
		link.NextUplink <- &ttnpb.GatewayUp{
			UplinkMessages: []*ttnpb.UplinkMessage{
				{},
			},
		}
	}()
	select {
	case <-upstream:
	case <-time.After(time.Second * 10):
		t.Log("Timeout: uplink was not received")
		t.Fail()
	}

	obs, err = p.GetGatewayObservations(&gatewayIdentifiers)
	a.So(err, should.BeNil)
	a.So(obs.UplinkCount, should.Equal, 1)
	a.So(obs.LastUplinkReceivedAt.Unix(), should.AlmostEqual, time.Now().Unix(), 1)

	link.AcceptSendingUplinks = false
	go func() { link.NextUplink <- emptyUplink }()
	newUplink = <-upstream
	a.So(newUplink, should.BeNil)
}

func TestDoneContextUplinks(t *testing.T) {
	a := assertions.New(t)

	p := pool.NewPool(test.GetLogger(t), time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	gatewayID := "gateway"
	link := &dummyLink{
		AcceptSendingUplinks: true,

		context:       ctx,
		cancelContext: cancel,

		NextUplink: make(chan *ttnpb.GatewayUp),
	}
	cancel()

	emptyUplink := &ttnpb.GatewayUp{}
	upstream, err := p.Subscribe(ttnpb.GatewayIdentifiers{GatewayID: gatewayID}, link, ttnpb.FrequencyPlan{BandID: band.EU_863_870})
	a.So(err, should.BeNil)

	go func() { link.NextUplink <- emptyUplink }()
	time.Sleep(time.Millisecond)
	select {
	case _, ok := <-upstream:
		if ok {
			t.Error("Stream not closed, message received")
		}
	default:
		t.Error("Stream not closed, no message")
	}
}

func TestSubscribeTwice(t *testing.T) {
	a := assertions.New(t)

	p := pool.NewPool(test.GetLogger(t), time.Millisecond)

	gateway := ttnpb.GatewayIdentifiers{GatewayID: "gateway"}

	link := &dummyLink{}
	newLink := &dummyLink{}

	_, err := p.Subscribe(gateway, link, ttnpb.FrequencyPlan{BandID: band.EU_863_870})
	a.So(err, should.BeNil)
	_, err = p.Subscribe(gateway, newLink, ttnpb.FrequencyPlan{BandID: band.EU_863_870})
	a.So(err, should.BeNil)
}
