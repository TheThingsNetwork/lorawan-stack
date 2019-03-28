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
	"strings"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc/metadata"
)

func (gcs *GatewayConfigurationServer) handleError() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := gcs.ctx
			err := next(c)
			if err == nil || c.Response().Committed {
				return err
			}
			log.FromContext(ctx).WithError(err).Debug("HTTP request failed")
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

func (gcs *GatewayConfigurationServer) validateAndFillIDs() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := gcs.ctx
			gtwID := ttnpb.GatewayIdentifiers{
				GatewayID: c.Param(gatewayIDKey),
			}
			if err := gtwID.ValidateContext(ctx); err != nil {
				return err
			}
			c.Set(gatewayIDKey, gtwID)

			return next(c)
		}
	}
}

func (gcs *GatewayConfigurationServer) requireGatewayRights(required ...ttnpb.Right) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := gcs.ctx
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
