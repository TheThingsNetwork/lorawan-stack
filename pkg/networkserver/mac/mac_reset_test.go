// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMACReset(t *testing.T) {
	a := assertions.New(t)

	dev := newDev()
	msg := newUplink()
	msg.Payload.GetMACPayload().FOpts = []byte{0x01, 0x01}

	err := HandleUplink(context.Background(), dev, msg)
	a.So(err, should.BeNil)

	if a.So(dev.QueuedMACCommands, should.NotBeEmpty) {
		mac := dev.QueuedMACCommands[0].GetActualPayload()
		if a.So(mac, should.HaveSameTypeAs, &ttnpb.MACCommand_ResetConf{}) {
			a.So(mac.(*ttnpb.MACCommand_ResetConf).MinorVersion, should.Equal, 1)
		}
	}

	// TODO: Test that MAC and Radio state was reset
}
