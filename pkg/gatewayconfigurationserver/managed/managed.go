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

// Package managed implements the API gateway for The Things Gateway Controller.
package managed

import (
	"context"
	"crypto/tls"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttgc"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Component is the component interface required for this package.
type Component interface {
	Context() context.Context
	GetBaseConfig(context.Context) config.ServiceBase
	GRPCServer() *rpcserver.Server
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
	GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error)
}

// Server implements provides configuration services for managed gateways via The Things Gateway Controller.
type Server struct {
	Component
	grpc struct {
		server           ttnpb.ManagedGatewayConfigurationServiceServer
		wifiProfiles     ttnpb.ManagedGatewayWiFiProfileConfigurationServiceServer
		ethernetProfiles ttnpb.ManagedGatewayEthernetProfileConfigurationServiceServer
	}
}

// New returns a new Server.
func New(ctx context.Context, c Component, conf ttgc.Config) (*Server, error) {
	srv := &Server{Component: c}
	srv.grpc.server = &noopManagedGCSServer{}
	srv.grpc.wifiProfiles = &noopManagedGatewayWiFiProfileServer{}
	srv.grpc.ethernetProfiles = &noopManagedGatewayEthernetProfileServer{}

	if conf.Enabled {
		client, err := ttgc.NewClient(ctx, c, conf)
		if err != nil {
			return nil, err
		}
		srv.grpc.server = &managedGCSServer{
			Component:   c,
			client:      client,
			gatewayEUIs: conf.GatewayEUIs,
		}
		srv.grpc.wifiProfiles = &managedGatewayWiFiProfileServer{
			client: client,
		}
		srv.grpc.ethernetProfiles = &managedGatewayEthernetProfileServer{
			client: client,
		}
	}

	c.GRPCServer().RegisterUnaryHook(
		"/ttn.lorawan.v3.ManagedGatewayConfigurationService",
		rpclog.NamespaceHook,
		rpclog.UnaryNamespaceHook("gatewayconfigurationserver/managed"),
	)
	c.GRPCServer().RegisterUnaryHook(
		"/ttn.lorawan.v3.ManagedGatewayWiFiProfileConfigurationService",
		rpclog.NamespaceHook,
		rpclog.UnaryNamespaceHook("gatewayconfigurationserver/managed"),
	)
	c.GRPCServer().RegisterUnaryHook(
		"/ttn.lorawan.v3.ManagedGatewayEthernetProfileConfigurationService",
		rpclog.NamespaceHook,
		rpclog.UnaryNamespaceHook("gatewayconfigurationserver/managed"),
	)

	return srv, nil
}

// RegisterServices registers services provided by gcs at s.
func (s *Server) RegisterServices(grpcServer *grpc.Server) {
	ttnpb.RegisterManagedGatewayConfigurationServiceServer(grpcServer, s.grpc.server)
	ttnpb.RegisterManagedGatewayWiFiProfileConfigurationServiceServer(grpcServer, s.grpc.wifiProfiles)
	ttnpb.RegisterManagedGatewayEthernetProfileConfigurationServiceServer(grpcServer, s.grpc.ethernetProfiles)
}

// RegisterHandlers registers gRPC handlers.
//
//nolint:errcheck
func (s *Server) RegisterHandlers(mux *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterManagedGatewayConfigurationServiceHandler(s.Context(), mux, conn)
	ttnpb.RegisterManagedGatewayWiFiProfileConfigurationServiceHandler(s.Context(), mux, conn)
	ttnpb.RegisterManagedGatewayEthernetProfileConfigurationServiceHandler(s.Context(), mux, conn)
}
