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
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const gatewayIDKey = "gateway_id"

func (gcs *GatewayConfigurationServer) validateAndFillIDs() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := gcs.getContext(c)
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
			ctx := gcs.getContext(c)
			gtwID := c.Get(gatewayIDKey).(ttnpb.GatewayIdentifiers)
			if err := rights.RequireGateway(ctx, gtwID, required...); err != nil {
				return err
			}
			return next(c)
		}
	}
}
