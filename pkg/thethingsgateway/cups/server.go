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

package cups

import (
	"context"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
)

// Config is the configuration of the The Things Gateay CUPS server.
type Config struct {
	Default struct {
		UpdateChannel string `name:"update-channel" description:"The default update channel that the gateways should use"`
		MQTTServer    string `name:"mqtt-server" description:"The default MQTT server that the gateways should use"`
		FirmwareURL   string `name:"firmware-url" description:"The default URL to the firmware storage"`
	} `name:"default" description:"Default gateway settings"`
}

// NewServer returns a new CUPS server from this config on top of the component.
func (conf Config) NewServer(c *component.Component) *Server {
	s := NewServer(c, conf)
	c.RegisterWeb(s)
	return s
}

// Server implements the CUPS endpoints used by The Things Gateway.
type Server struct {
	component *component.Component

	config Config
}

const compatAPIPrefix = "/api/v2"

func (s *Server) getRegistry(ctx context.Context, ids *ttnpb.GatewayIdentifiers) ttnpb.GatewayRegistryClient {
	return ttnpb.NewGatewayRegistryClient(s.component.GetPeer(ctx, ttnpb.PeerInfo_ENTITY_REGISTRY, ids).Conn())
}

// RegisterRoutes implements the web.Registerer interface.
func (s *Server) RegisterRoutes(srv *web.Server) {
	group := srv.Group(compatAPIPrefix)
	group.GET("/gateways/:gateway_id", func(c echo.Context) error {
		return s.handleGatewayInfo(c)
	}, s.validateAndFillGatewayIDs())
	group.GET("/frequency-plans/:frequency_plan_id", func(c echo.Context) error {
		return s.handleFreqPlanInfo(c)
	})
}

// NewServer returns a new CUPS server on top of the given gateway registry.
func NewServer(c *component.Component, conf Config) *Server {
	return &Server{
		component: c,
		config:    conf,
	}
}
