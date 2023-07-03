// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package cluster_test

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestPeer(t *testing.T) {
	a := assertions.New(t)

	conn := new(grpc.ClientConn)

	p := cluster.NewTestPeer(conn)

	a.So(p.HasRole(ttnpb.ClusterRole_APPLICATION_SERVER), should.BeFalse)
	a.So(p.HasRole(ttnpb.ClusterRole_ACCESS), should.BeTrue)

	cc, err := p.Conn()
	a.So(err, should.BeNil)
	a.So(cc, should.Equal, conn)
}
