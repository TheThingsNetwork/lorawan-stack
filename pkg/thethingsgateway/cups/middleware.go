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
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const (
	gatewayIDKey       = "gateway_id"
	frequencyPlanIDKey = "frequency_plan_id"
)

var errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "call was not authenticated")

// validateAndFillGatewayIDs checks if the request contains a valid gateway ID.
func (s *Server) validateAndFillGatewayIDs() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			gatewayIDs := ttnpb.GatewayIdentifiers{
				GatewayID: c.Param(gatewayIDKey),
			}
			if err := gatewayIDs.ValidateContext(c.Request().Context()); err != nil {
				return err
			}
			c.Set(gatewayIDKey, gatewayIDs)

			return next(c)
		}
	}
}

// requireAuth checks if the request contains the Authorization header.
func (s *Server) requireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return errUnauthenticated
			}
			return next(c)
		}
	}
}
