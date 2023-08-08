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

// Package mockis provides a mock structure to the Identity Server.
package mockis

import (
	"context"
	"net"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var errNotFound = errors.DefineNotFound("not_found", "not found")

type authKeyToRights map[string][]ttnpb.Right

// MockDefinition contains the structure that is returned by the New(ctx) method of the package, might be used to
// specify IS mock in test cases definitons.
type MockDefinition struct {
	applicationRegistry    *mockISApplicationRegistry
	gatewayRegistry        *mockISGatewayRegistry
	endDeviceRegistry      *mockISEndDeviceRegistry
	endDeviceBatchRegistry *isEndDeviceBatchRegistry
	entityAccess           *mockEntityAccess
}

type closeMock func()

// New returns a identityserver mock along side its address and closing function.
func New(ctx context.Context) (*MockDefinition, string, closeMock) {
	endDeviceRegistry := &mockISEndDeviceRegistry{}
	is := &MockDefinition{
		applicationRegistry: &mockISApplicationRegistry{
			applications:      make(map[string]*ttnpb.Application),
			applicationAuths:  make(map[string][]string),
			applicationRights: make(map[string]authKeyToRights),
		},
		gatewayRegistry: &mockISGatewayRegistry{
			gateways:      make(map[string]*ttnpb.Gateway),
			gatewayAuths:  make(map[string][]string),
			gatewayRights: make(map[string]authKeyToRights),
		},
		endDeviceRegistry: endDeviceRegistry,
		endDeviceBatchRegistry: &isEndDeviceBatchRegistry{
			reg: endDeviceRegistry,
		},
		entityAccess: &mockEntityAccess{},
	}

	srv := rpcserver.New(ctx)

	ttnpb.RegisterApplicationRegistryServer(srv.Server, is.applicationRegistry)
	ttnpb.RegisterApplicationAccessServer(srv.Server, is.applicationRegistry)

	ttnpb.RegisterGatewayRegistryServer(srv.Server, is.gatewayRegistry)
	ttnpb.RegisterGatewayAccessServer(srv.Server, is.gatewayRegistry)

	ttnpb.RegisterEndDeviceRegistryServer(srv.Server, is.endDeviceRegistry)

	ttnpb.RegisterEndDeviceBatchRegistryServer(srv.Server, is.endDeviceBatchRegistry)

	ttnpb.RegisterEntityAccessServer(srv.Server, is.entityAccess)

	lis, err := net.Listen("tcp", "")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis) //nolint:errcheck
	return is, lis.Addr().String(), func() {
		lis.Close()
		srv.GracefulStop()
	}
}

// EndDeviceRegistry returns the methods related to the device registry.
func (m *MockDefinition) EndDeviceRegistry() *mockISEndDeviceRegistry {
	return m.endDeviceRegistry
}

// ApplicationRegistry returns the methods related to the application registry.
func (m *MockDefinition) ApplicationRegistry() *mockISApplicationRegistry {
	return m.applicationRegistry
}

// GatewayRegistry returns the methods related to the gateway registry.
func (m *MockDefinition) GatewayRegistry() *mockISGatewayRegistry {
	return m.gatewayRegistry
}

// EntityAccess returns the methods related to the access entity.
func (m *MockDefinition) EntityAccess() *mockEntityAccess {
	return m.entityAccess
}

func (m *MockDefinition) EndDeviceBatchRegistry() *isEndDeviceBatchRegistry { //nolint:revive
	return m.endDeviceBatchRegistry
}
