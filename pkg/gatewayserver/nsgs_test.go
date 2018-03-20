// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver_test

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/pool"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestScheduleDownlinkUnregisteredGateway(t *testing.T) {
	a := assertions.New(t)

	dir := createFPStore(a)
	defer removeFPStore(a, dir)

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	gs := gatewayserver.New(c, &gatewayserver.Config{
		FileFrequencyPlansStore: dir,
	})

	_, err := gs.ScheduleDownlink(context.Background(), &ttnpb.DownlinkMessage{
		TxMetadata: ttnpb.TxMetadata{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{
				GatewayID: "unknown-downlink",
			},
		},
	})
	a.So(err, should.NotBeNil)
	a.So(pool.ErrGatewayNotConnected.Caused(err), should.BeTrue)

	defer gs.Close()
}
