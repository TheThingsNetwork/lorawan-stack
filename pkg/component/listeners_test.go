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

package component_test

import (
	"net"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

const udpListenAddr = "0.0.0.0:8056"

func TestListenUDP(t *testing.T) {
	a := assertions.New(t)

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	conn, err := c.ListenUDP(udpListenAddr)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	senderConn, err := net.Dial("udp", udpListenAddr)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	content := []byte{0xaa, 0xbb, 0xcc, 0x03}
	go func() {
		_, err := senderConn.Write(content)
		a.So(err, should.BeNil)
	}()
	receptionBuf := make([]byte, 256)
	_, err = conn.Read(receptionBuf)
	a.So(err, should.BeNil)
	for i := range content {
		a.So(receptionBuf[i], should.Equal, content[i])
	}
}
