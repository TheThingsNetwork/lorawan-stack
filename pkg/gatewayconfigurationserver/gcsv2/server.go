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

package gcsv2

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc"
)

// TheThingsGatewayConfig is the configuration for The Things Gateway.
type TheThingsGatewayConfig struct {
	Default struct {
		UpdateChannel string `name:"update-channel" description:"The default update channel that the gateways should use"`
		MQTTServer    string `name:"mqtt-server" description:"The default MQTT server that the gateways should use"`
		FirmwareURL   string `name:"firmware-url" description:"The default URL to the firmware storage"`
	} `name:"default" description:"Default gateway settings"`
}

// Server implements the CUPS endpoints used by The Things Gateway.
type Server struct {
	component *component.Component

	ttgConfig TheThingsGatewayConfig

	registry ttnpb.GatewayRegistryClient
	auth     func(context.Context) grpc.CallOption
}

func (s *Server) getRegistry(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayRegistryClient, error) {
	if s.registry != nil {
		return s.registry, nil
	}
	cc, err := s.component.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, ids)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayRegistryClient(cc), nil
}

func (s *Server) getAuth(ctx context.Context) grpc.CallOption {
	if s.auth != nil {
		return s.auth(ctx)
	}
	return s.component.WithClusterAuth()
}

// Option configures the Server.
type Option func(s *Server)

// WithRegistry overrides the Server's gateway registry.
func WithRegistry(registry ttnpb.GatewayRegistryClient) Option {
	return func(s *Server) {
		s.registry = registry
	}
}

// WithAuth overrides the Server's auth func.
func WithAuth(auth func(ctx context.Context) grpc.CallOption) Option {
	return func(s *Server) {
		s.auth = auth
	}
}

// WithTheThingsGatewayConfig overrides the Server's configuration for The Things Gateway.
func WithTheThingsGatewayConfig(config TheThingsGatewayConfig) Option {
	return func(s *Server) {
		s.ttgConfig = config
	}
}

const compatAPIPrefix = "/api/v2"

// RegisterRoutes implements the web.Registerer interface.
func (s *Server) RegisterRoutes(srv *web.Server) {
	group := srv.Group(compatAPIPrefix, s.normalizeAuthorization)
	group.GET("/gateways/:gateway_id", s.handleGetGateway)
	group.GET("/frequency-plans/:frequency_plan_id", s.handleGetFrequencyPlan)
}

// New returns a new v2 GCS on top of the given gateway registry.
func New(c *component.Component, options ...Option) *Server {
	s := &Server{
		component: c,
	}
	for _, opt := range options {
		opt(s)
	}
	c.RegisterWeb(s)
	return s
}
