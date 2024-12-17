// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package gatewayconfigurationserver_test

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttgc"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNoopManagedGCSServer(t *testing.T) {
	t.Parallel()

	a, ctx := test.New(t)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			TTGC: ttgc.Config{Enabled: false},
		},
	})

	_, err := gatewayconfigurationserver.New(c, &gatewayconfigurationserver.Config{})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_GATEWAY_CONFIGURATION_SERVER)

	client := ttnpb.NewManagedGatewayConfigurationServiceClient(c.LoopbackConn())
	a.So(client, should.NotBeNil)

	managedGateway, err := client.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw-id"},
	})
	a.So(managedGateway, should.BeNil)
	a.So(errors.IsNotFound(err), should.BeTrue)

	managedGateway, err = client.Update(ctx, &ttnpb.UpdateManagedGatewayRequest{
		Gateway: &ttnpb.ManagedGateway{
			Ids: &ttnpb.GatewayIdentifiers{
				GatewayId: "gtw-id",
			},
		},
	})
	a.So(managedGateway, should.BeNil)
	a.So(errors.IsNotFound(err), should.BeTrue)

	managedGatewayWiFiAccessPoints, err := client.ScanWiFiAccessPoints(ctx, &ttnpb.GatewayIdentifiers{
		GatewayId: "gtw-id",
	})
	a.So(managedGatewayWiFiAccessPoints, should.BeNil)
	a.So(errors.IsNotFound(err), should.BeTrue)

	streamEventsClient, err := client.StreamEvents(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "gtw-id"})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = streamEventsClient.RecvMsg(nil)
	a.So(errors.IsNotFound(err), should.BeTrue)
}
