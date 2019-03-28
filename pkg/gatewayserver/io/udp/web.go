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

package udp

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/semtechudp"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	ttnweb "go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc/metadata"
)

// WebServer is an interface for registering the UDP web frontend.
type WebServer interface {
	ttnweb.Registerer
}

type webSrv struct {
	ctx    context.Context
	config Config

	server io.Server
}

// RegisterRoutes registers the UDP web frontend routes.
func (s *webSrv) RegisterRoutes(server *ttnweb.Server) {
	middleware := []echo.MiddlewareFunc{
		s.handleError(),
		s.validateAndFillIDs(),
	}
	if s.config.RequireAuth {
		middleware = append(middleware, s.requireGatewayRights(ttnpb.RIGHT_GATEWAY_INFO))
	}
	group := server.Group(ttnpb.HTTPAPIPrefix+"/gs/gateways/:gateway_id", middleware...)
	group.GET("/global_conf.json", func(c echo.Context) error {
		return s.handleGet(c)
	})
}

// StartWeb starts the UDP web frontend.
func StartWeb(ctx context.Context, server io.Server, config Config) WebServer {
	s := &webSrv{
		ctx:    ctx,
		server: server,
		config: config,
	}

	return s
}

func (s *webSrv) handleGet(c echo.Context) error {
	ctx := s.ctx
	gtwID := c.Get(gatewayIDKey).(ttnpb.GatewayIdentifiers)
	gtw, err := s.server.GetGateway(ctx, gtwID, types.FieldMask{
		Paths: []string{
			"gateway_server_address",
		},
	})
	if err != nil {
		return err
	}
	frequencyPlan, err := s.server.GetFrequencyPlan(ctx, gtwID)
	if err != nil {
		return err
	}
	config, err := semtechudp.BuildSimple(gtw, frequencyPlan)
	if err != nil {
		return err
	}
	return c.JSONPretty(http.StatusOK, config, "\t")
}

func (s *webSrv) handleError() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil || c.Response().Committed {
				return err
			}
			log.FromContext(s.ctx).WithError(err).Debug("HTTP request failed")
			statusCode, err := web_errors.ProcessError(err)
			if strings.Contains(c.Request().Header.Get(echo.HeaderAccept), "application/json") {
				return c.JSON(statusCode, err)
			}
			return c.String(statusCode, err.Error())
		}
	}
}

const (
	gatewayIDKey = "gateway_id"
)

func (s *webSrv) validateAndFillIDs() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			gtwID := ttnpb.GatewayIdentifiers{
				GatewayID: c.Param(gatewayIDKey),
			}
			if err := gtwID.ValidateContext(s.ctx); err != nil {
				return err
			}
			c.Set(gatewayIDKey, gtwID)

			return next(c)
		}
	}
}

func (s *webSrv) requireGatewayRights(required ...ttnpb.Right) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := s.ctx
			gtwID := c.Get(gatewayIDKey).(ttnpb.GatewayIdentifiers)
			md := metadata.New(map[string]string{
				"id":            gtwID.GatewayID,
				"authorization": c.Request().Header.Get(echo.HeaderAuthorization),
			})
			if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
				md = metadata.Join(ctxMd, md)
			}
			ctx = metadata.NewIncomingContext(ctx, md)
			if err := rights.RequireGateway(ctx, gtwID, required...); err != nil {
				return err
			}
			return next(c)
		}
	}
}
