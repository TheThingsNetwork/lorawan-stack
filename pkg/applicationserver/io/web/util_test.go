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

package web_test

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

var (
	registeredApplicationID  = &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"}
	registeredApplicationUID = unique.ID(test.Context(), registeredApplicationID)
	registeredApplicationKey = "secret"
	registeredDeviceID       = &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: registeredApplicationID,
		DeviceId:       "foo-device",
		DevAddr:        types.DevAddr{0x42, 0xff, 0xff, 0xff}.Bytes(),
		DevEui:         test.DefaultDevEUI.Bytes(),
		JoinEui:        test.DefaultJoinEUI.Bytes(),
	}
	unregisteredDeviceID = ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "bar-app",
		},
		DeviceId: "bar-device",
		DevAddr:  types.DevAddr{0x42, 0x42, 0x42, 0x42}.Bytes(),
	}
	registeredWebhookID = "foo-hook"

	Timeout = (1 << 5) * test.Delay
)

type mockRegisterer struct {
	context.Context
	ttnpb.ApplicationWebhookRegistryServer
}

func (m *mockRegisterer) Roles() []ttnpb.ClusterRole {
	return nil
}

func (m *mockRegisterer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterApplicationWebhookRegistryServer(s, m.ApplicationWebhookRegistryServer)
}

func (m *mockRegisterer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterApplicationWebhookRegistryHandler(m.Context, s, conn)
}

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}
