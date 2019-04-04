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
	"context"
	"net/http"

	"github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	bs_cups "go.thethings.network/lorawan-stack/pkg/basicstation/cups"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/semtechudp"
	ttg_cups "go.thethings.network/lorawan-stack/pkg/thethingsgateway/cups"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc/metadata"
)

// Config contains the Gateway Configuration Server configuration.
type Config struct {
	// BasicStation defines the configuration for the BasicStation CUPS.
	BasicStation bs_cups.ServerConfig `name:"basic-station" description:"BasicStation CUPS configuration."`
	// TheThingsGateway defines the configuration for The Things Gateway CUPS.
	TheThingsGateway ttg_cups.Config `name:"the-things-gateway" description:"The Things Gateway CUPS configuration."`
	// RequreAuth defines if the HTTP endpoints should require authentication or not.
	RequireAuth bool `name:"require-auth" description:"Require authentication for the HTTP endpoints."`
}

// GatewayConfigurationServer implements the Gateway Configuration Server component.
type GatewayConfigurationServer struct {
	*component.Component
	config *Config

	registry ttnpb.GatewayRegistryClient

	ctx context.Context
}

// RegisterRoutes registers the web frontend routes.
func (gcs *GatewayConfigurationServer) RegisterRoutes(server *web.Server) {
	middleware := []echo.MiddlewareFunc{
		gcs.validateAndFillIDs(),
	}
	if gcs.config.RequireAuth {
		middleware = append(middleware, gcs.requireGatewayRights(ttnpb.RIGHT_GATEWAY_INFO))
	}
	group := server.Group(ttnpb.HTTPAPIPrefix+"/gcs/gateways/:gateway_id", middleware...)
	group.GET("/global_conf.json", gcs.handleGetGlobalConfig)
}

// New returns new *GatewayConfigurationServer.
func New(c *component.Component, conf *Config, opts ...Option) (*GatewayConfigurationServer, error) {
	gcs := &GatewayConfigurationServer{
		Component: c,
		config:    conf,
		ctx:       c.Context(),
	}
	for _, opt := range opts {
		opt(gcs)
	}

	bsCUPS := conf.BasicStation.NewServer(c)
	_ = bsCUPS

	ttgCUPS, err := conf.TheThingsGateway.NewServer(c)
	if err != nil {
		return nil, err
	}
	_ = ttgCUPS

	c.RegisterWeb(gcs)
	return gcs, nil
}

// Option represents a Gateway Configuration Server option.
type Option func(gcs *GatewayConfigurationServer)

// WithRegistry sets the gateway registry for the given server.
func WithRegistry(registry ttnpb.GatewayRegistryClient) Option {
	return func(gcs *GatewayConfigurationServer) {
		gcs.registry = registry
	}
}

// WithContext sets the context for the given server.
func WithContext(ctx context.Context) Option {
	return func(gcs *GatewayConfigurationServer) {
		gcs.ctx = ctx
	}
}

func (gcs *GatewayConfigurationServer) handleGetGlobalConfig(c echo.Context) error {
	ctx := gcs.getContext(c)
	gtwID := c.Get(gatewayIDKey).(ttnpb.GatewayIdentifiers)
	gtw, err := gcs.getRegistry(ctx).Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: gtwID,
		FieldMask: types.FieldMask{
			Paths: []string{
				"gateway_server_address",
				"frequency_plan_id",
			},
		},
	}, gcs.WithClusterAuth())
	if err != nil {
		return err
	}
	config, err := semtechudp.Build(gtw, gcs.FrequencyPlans)
	if err != nil {
		return err
	}
	return c.JSONPretty(http.StatusOK, config, "\t")
}

func (gcs *GatewayConfigurationServer) getContext(c echo.Context) context.Context {
	ctx := c.Request().Context()
	ctx = gcs.FillContext(ctx)
	md := metadata.New(map[string]string{
		"authorization": c.Request().Header.Get(echo.HeaderAuthorization),
	})
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	return metadata.NewIncomingContext(ctx, md)
}

func (gcs *GatewayConfigurationServer) getRegistry(ctx context.Context) ttnpb.GatewayRegistryClient {
	if gcs.registry != nil {
		return gcs.registry
	}
	return ttnpb.NewGatewayRegistryClient(gcs.GetPeer(ctx, ttnpb.PeerInfo_ENTITY_REGISTRY, nil).Conn())
}
