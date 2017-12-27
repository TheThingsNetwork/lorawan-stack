// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMACLinkCheck(t *testing.T) {
	a := assertions.New(t)

	dev := newDev()

	uplink := newUplink()
	uplink.Payload.GetMACPayload().FOpts = []byte{0x02}
	uplink.RxMetadata = []ttnpb.RxMetadata{
		ttnpb.RxMetadata{GatewayIdentifier: ttnpb.GatewayIdentifier{GatewayID: "testgw"}, AntennaIndex: 0, SNR: 1.5},
		ttnpb.RxMetadata{GatewayIdentifier: ttnpb.GatewayIdentifier{GatewayID: "testgw"}, AntennaIndex: 1, SNR: 2.0},
		ttnpb.RxMetadata{GatewayIdentifier: ttnpb.GatewayIdentifier{GatewayID: "othergw"}, SNR: -2.0},
	}

	ctx := context.Background()

	err := HandleUplink(ctx, dev, uplink)
	a.So(err, should.BeNil)

	if a.So(dev.QueuedMACCommands, should.NotBeEmpty) {
		mac := dev.QueuedMACCommands[0].GetActualPayload()
		if a.So(mac, should.HaveSameTypeAs, &ttnpb.MACCommand_LinkCheckAns{}) {
			ans := mac.(*ttnpb.MACCommand_LinkCheckAns)
			a.So(ans.Margin, should.Equal, 22)
			a.So(ans.GatewayCount, should.Equal, 2)
		}
	}

	// TODO: Test that MAC and Radio state was reset
}
