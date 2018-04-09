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

package gatewayserver_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func Example() {
	logger, err := log.NewLogger()
	if err != nil {
		panic(err)
	}

	c := component.MustNew(logger, &component.Config{ServiceBase: config.ServiceBase{}})

	gs, err := gatewayserver.New(c, gatewayserver.Config{})
	if err != nil {
		panic(err)
	}
	gs.Run()
}

func TestGatewayServer(t *testing.T) {
	a := assertions.New(t)

	dir := createFPStore(a)
	defer removeFPStore(a, dir)

	logger := test.GetLogger(t)
	c := component.MustNew(logger, &component.Config{ServiceBase: config.ServiceBase{FrequencyPlans: config.FrequencyPlans{
		StoreDirectory: dir,
	}}})
	gs, err := gatewayserver.New(c, gatewayserver.Config{})
	if !a.So(err, should.BeNil) {
		logger.Fatal("Gateway server could not start")
	}

	roles := gs.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER)

	defer gs.Close()
}
