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

package applicationserver

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// ApplicationServer implements the Application Server component.
//
// The Application Server exposes the As, AppAs and AsEndDeviceRegistry services.
type ApplicationServer struct {
	*component.Component

	config *Config

	linkRegistry   LinkRegistry
	deviceRegistry DeviceRegistry
}

// Config represents the ApplicationServer configuration.
type Config struct {
	Devices DeviceRegistry
	Links   LinkRegistry
}

// New returns new *ApplicationServer.
func New(c *component.Component, conf *Config) (*ApplicationServer, error) {
	as := &ApplicationServer{
		Component:      c,
		config:         conf,
		linkRegistry:   conf.Links,
		deviceRegistry: conf.Devices,
	}

	c.RegisterGRPC(as)
	return as, nil
}

// RegisterServices registers services provided by as at s.
func (as *ApplicationServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsServer(s, as)
	ttnpb.RegisterAppAsServer(s, as)
	// TODO: Register AsEndDeviceRegistryServer (https://github.com/TheThingsIndustries/lorawan-stack/issues/1117)
}

// RegisterHandlers registers gRPC handlers.
func (as *ApplicationServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterAsHandler(as.Context(), s, conn)
	// TODO: Register AsEndDeviceRegistryHandler (https://github.com/TheThingsIndustries/lorawan-stack/issues/1117)
}

// Roles returns the roles that the Application Server fulfills.
func (as *ApplicationServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_APPLICATION_SERVER}
}
