// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package cluster

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
)

func TestPeer(t *testing.T) {
	a := assertions.New(t)

	conn := new(grpc.ClientConn)

	p := &peer{
		name:   "name",
		roles:  []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_IDENTITY_SERVER},
		tags:   []string{"tag"},
		target: "target",
		conn:   conn,
	}

	a.So(p.HasRole(ttnpb.PeerInfo_APPLICATION_SERVER), should.BeFalse)
	a.So(p.HasRole(ttnpb.PeerInfo_IDENTITY_SERVER), should.BeTrue)

	a.So(p.HasTag("no-tag"), should.BeFalse)
	a.So(p.HasTag("tag"), should.BeTrue)

	a.So(p.Conn(), should.Equal, conn)
}
