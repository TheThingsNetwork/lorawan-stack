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
	"fmt"
	"strings"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"google.golang.org/grpc/metadata"
)

func (s *Server) normalizeAuthorization(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		authorization := c.Request().Header.Get(echo.HeaderAuthorization)
		if authorization == "" {
			return next(c)
		}
		authorizationParts := strings.SplitN(authorization, " ", 2)
		if len(authorizationParts) != 2 {
			return errUnauthenticated
		}
		authType, authValue := strings.ToLower(authorizationParts[0]), authorizationParts[1]
		switch authType {
		case "bearer", "key":
			tokenType, _, _, err := auth.SplitToken(authValue)
			if err == nil && (tokenType == auth.APIKey || tokenType == auth.AccessToken) {
				authType = "bearer"
			} else {
				authType = "key"
			}
		default:
			return errUnauthenticated
		}
		md := metadata.New(map[string]string{
			"authorization": fmt.Sprintf("%s %s", authType, authValue),
		})
		if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
			md = metadata.Join(ctxMd, md)
		}
		ctx = metadata.NewIncomingContext(ctx, md)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
