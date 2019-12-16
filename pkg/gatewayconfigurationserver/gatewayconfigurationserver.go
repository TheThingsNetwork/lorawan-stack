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
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	echo "github.com/labstack/echo/v4"
	bscups "go.thethings.network/lorawan-stack/pkg/basicstation/cups"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/gatewayconfigurationserver/gcsv2"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/semtechudp"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Config contains the Gateway Configuration Server configuration.
type Config struct {
	// BasicStation defines the configuration for the BasicStation CUPS.
	BasicStation bscups.ServerConfig `name:"basic-station" description:"BasicStation CUPS configuration."`
	// TheThingsGateway defines the configuration for The Things Gateway CUPS.
	TheThingsGateway gcsv2.Config `name:"the-things-gateway" description:"The Things Gateway CUPS configuration."`
	// RequreAuth defines if the HTTP endpoints should require authentication or not.
	RequireAuth bool `name:"require-auth" description:"Require authentication for the HTTP endpoints."`
}

// GatewayConfigurationServer implements the Gateway Configuration Server component.
type GatewayConfigurationServer struct {
	*component.Component
	config *Config
}

// Roles returns the roles that the Gateway Configuration Server fulfills.
func (gcs *GatewayConfigurationServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_GATEWAY_CONFIGURATION_SERVER}
}

// RegisterServices registers services provided by gcs at s.
func (gcs *GatewayConfigurationServer) RegisterServices(_ *grpc.Server) {}

// RegisterHandlers registers gRPC handlers.
func (gcs *GatewayConfigurationServer) RegisterHandlers(_ *runtime.ServeMux, _ *grpc.ClientConn) {}

// RegisterRoutes registers the web frontend routes.
func (gcs *GatewayConfigurationServer) RegisterRoutes(server *web.Server) {
	middleware := []echo.MiddlewareFunc{
		gcs.validateAndFillIDs(),
	}
	if gcs.config.RequireAuth {
		middleware = append(middleware, gcs.requireGatewayRights(ttnpb.RIGHT_GATEWAY_INFO))
	}
	group := server.Group(ttnpb.HTTPAPIPrefix+"/gcs/gateways/:gateway_id", middleware...)
	group.GET("/semtechudp/global_conf.json", gcs.handleGetGlobalConfig)
}

// New returns new *GatewayConfigurationServer.
func New(c *component.Component, conf *Config) (*GatewayConfigurationServer, error) {
	gcs := &GatewayConfigurationServer{
		Component: c,
		config:    conf,
	}

	bsCUPS := conf.BasicStation.NewServer(c)
	_ = bsCUPS

	v2CUPS, err := conf.TheThingsGateway.NewServer(c)
	if err != nil {
		return nil, err
	}
	_ = v2CUPS

	c.RegisterGRPC(gcs)
	c.RegisterWeb(gcs)
	return gcs, nil
}

func (gcs *GatewayConfigurationServer) handleGetGlobalConfig(c echo.Context) error {
	ctx := gcs.getContext(c)
	gtwID := c.Get(gatewayIDKey).(ttnpb.GatewayIdentifiers)
	cc, err := gcs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return err
	}
	client := ttnpb.NewGatewayRegistryClient(cc)
	gtw, err := client.Get(ctx, &ttnpb.GetGatewayRequest{
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
	ctx := gcs.FillContext(c.Request().Context())
	md := metadata.New(map[string]string{
		"authorization": c.Request().Header.Get(echo.HeaderAuthorization),
	})
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	return metadata.NewIncomingContext(ctx, md)
}
