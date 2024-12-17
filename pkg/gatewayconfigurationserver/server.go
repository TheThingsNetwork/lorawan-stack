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

package gatewayconfigurationserver

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver/managed"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver/ttkg"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Server implements the Gateway Configuration Server component.
type Server struct {
	ttnpb.UnimplementedGatewayConfigurationServiceServer

	*component.Component
	config *Config

	managedServer *managed.Server
}

// Roles returns the roles that the Gateway Configuration Server fulfills.
func (*Server) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_GATEWAY_CONFIGURATION_SERVER}
}

// RegisterServices registers services provided by gcs at s.
func (s *Server) RegisterServices(grpcServer *grpc.Server) {
	ttnpb.RegisterGatewayConfigurationServiceServer(grpcServer, s)
	if s.managedServer != nil {
		s.managedServer.RegisterServices(grpcServer)
	}
}

// RegisterHandlers registers gRPC handlers.
//
//nolint:errcheck
func (s *Server) RegisterHandlers(mux *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterGatewayConfigurationServiceHandler(s.Context(), mux, conn)
	if s.managedServer != nil {
		s.managedServer.RegisterHandlers(mux, conn)
	}
}

// New returns new *Server.
func New(c *component.Component, conf *Config) (*Server, error) {
	gcs := &Server{
		Component: c,
		config:    conf,
	}

	bsCUPS := conf.BasicStation.NewServer(c)
	_ = bsCUPS

	ttkgServer := ttkg.New(c, ttkg.WithConfig(conf.TheThingsKickstarterGateway))
	_ = ttkgServer

	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.GatewayConfigurationService", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("gatewayconfigurationserver")) //nolint:lll
	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.GatewayConfigurationService", cluster.HookName, c.ClusterAuthUnaryHook())

	ttgcConf := c.GetBaseConfig(c.Context()).TTGC
	managedServer, err := managed.New(c.Context(), c, ttgcConf)
	if err != nil {
		return nil, err
	}
	gcs.managedServer = managedServer

	c.RegisterGRPC(gcs)
	c.RegisterWeb(gcs)
	return gcs, nil
}
