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

package mockis

import (
	"context"
	"net"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var errNotFound = errors.DefineNotFound("not_found", "not found")

type MockDefinition struct {
	applicationRegistry *mockISApplicationRegistry
	gatewayRegistry     *mockISGatewayRegistry
	endDeviceRegistry   *mockISEndDeviceRegistry
}

type closeMock func()

func New(ctx context.Context) (*MockDefinition, string, closeMock) {
	is := &MockDefinition{
		applicationRegistry: &mockISApplicationRegistry{
			applications:      make(map[string]*ttnpb.Application),
			applicationAuths:  make(map[string][]string),
			applicationRights: make(map[string]authKeyToRights),
		},
		gatewayRegistry: &mockISGatewayRegistry{
			gateways:      make(map[string]*ttnpb.Gateway),
			gatewayAuths:  make(map[string][]string),
			gatewayRights: make(map[string][]ttnpb.Right),
		},
		endDeviceRegistry: &mockISEndDeviceRegistry{
			endDevices: make(map[string]*ttnpb.EndDevice),
		},
	}
	srv := rpcserver.New(ctx)

	ttnpb.RegisterApplicationRegistryServer(srv.Server, is.applicationRegistry)
	ttnpb.RegisterApplicationAccessServer(srv.Server, is.applicationRegistry)

	ttnpb.RegisterGatewayRegistryServer(srv.Server, is.gatewayRegistry)
	ttnpb.RegisterGatewayAccessServer(srv.Server, is.gatewayRegistry)

	ttnpb.RegisterEndDeviceRegistryServer(srv.Server, is.endDeviceRegistry)

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	return is, lis.Addr().String(), func() {
		lis.Close()
		srv.GracefulStop()
	}
}

func (m *MockDefinition) EndDeviceRegistry() *mockISEndDeviceRegistry {
	return m.endDeviceRegistry
}

func (m *MockDefinition) ApplicationRegistry() *mockISApplicationRegistry {
	return m.applicationRegistry
}

func (m *MockDefinition) GatewayRegistry() *mockISGatewayRegistry {
	return m.gatewayRegistry
}
