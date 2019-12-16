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

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
)

// Config is the configuration of the The Things Gateay CUPS.
type Config struct {
	Default struct {
		UpdateChannel string `name:"update-channel" description:"The default update channel that the gateways should use"`
		MQTTServer    string `name:"mqtt-server" description:"The default MQTT server that the gateways should use"`
		FirmwareURL   string `name:"firmware-url" description:"The default URL to the firmware storage"`
	} `name:"default" description:"Default gateway settings"`
}

var (
	errNoDefaultFirmwarePath  = errors.Define("no_default_firmware_path", "no default firmware path specified")
	errNoDefaultUpdateChannel = errors.Define("no_default_update_channel", "no default update channel specified")
)

// NewServer returns a new CUPS from this config on top of the component.
func (conf Config) NewServer(c *component.Component, customOpts ...Option) (*Server, error) {
	opts := []Option{
		WithConfig(conf),
	}
	if conf.Default.FirmwareURL == "" {
		return nil, errNoDefaultFirmwarePath
	}
	if conf.Default.UpdateChannel == "" {
		return nil, errNoDefaultUpdateChannel
	}
	s := NewServer(c, append(opts, customOpts...)...)
	c.RegisterWeb(s)
	return s, nil
}

// Server implements the CUPS endpoints used by The Things Gateway.
type Server struct {
	component *component.Component

	registry ttnpb.GatewayRegistryClient

	config Config
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

// Option configures the CUPS.
type Option func(s *Server)

// WithRegistry overrides the CUPS gateway registry.
func WithRegistry(registry ttnpb.GatewayRegistryClient) Option {
	return func(s *Server) {
		s.registry = registry
	}
}

// WithConfig overrides the CUPS configuration.
func WithConfig(conf Config) Option {
	return func(s *Server) {
		s.config = conf
	}
}

// WithDefaultUpdateChannel overrides the default CUPS gateway update channel.
func WithDefaultUpdateChannel(channel string) Option {
	return func(s *Server) {
		s.config.Default.UpdateChannel = channel
	}
}

// WithDefaultMQTTServer overrides the default CUPS gateway MQTT server.
func WithDefaultMQTTServer(server string) Option {
	return func(s *Server) {
		s.config.Default.MQTTServer = server
	}
}

// WithDefaultFirmwareURL overrides the default CUPS firmware base URL.
func WithDefaultFirmwareURL(url string) Option {
	return func(s *Server) {
		s.config.Default.FirmwareURL = url
	}
}

const compatAPIPrefix = "/api/v2"

// RegisterRoutes implements the web.Registerer interface.
func (s *Server) RegisterRoutes(srv *web.Server) {
	group := srv.Group(compatAPIPrefix, s.normalizeAuthorization)
	group.GET("/gateways/:gateway_id", s.handleGetGateway)
	group.GET("/frequency-plans/:frequency_plan_id", s.handleGetFrequencyPlan)
}

// NewServer returns a new CUPS on top of the given gateway registry.
func NewServer(c *component.Component, options ...Option) *Server {
	s := &Server{
		component: c,
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}
