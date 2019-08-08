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

package middleware

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	ulid "github.com/oklog/ulid/v2"
)

// used to mock time
var now = time.Now

// ID adds a request id to the request.
func ID(prefix string) echo.MiddlewareFunc {
	gen := newID(prefix)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get(echo.HeaderXRequestID)

			if id == "" {
				var err error
				id, err = gen.generate()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to generate request ID: %s", err))
				}
			}
			c.Request().Header.Set(echo.HeaderXRequestID, id)
			c.Request().Header.Set("Grpc-Metadata-Request-ID", id)
			c.Response().Header().Set(echo.HeaderXRequestID, id)
			return next(c)
		}
	}
}

// id a generator of new ids, which uses ULID under the hood.
type id struct {
	prefixer func(ulid.ULID) string
}

// newID creates a new id.
func newID(prefix string) *id {
	return &id{
		prefixer: prefixer(prefix),
	}
}

// generate generates a new ULID.
func (i *id) generate() (string, error) {
	id, err := ulid.New(ulid.Timestamp(now()), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("Failed to generate a new ULID")
	}

	return i.prefixer(id), nil
}

// prefixer returns a function that can smartly prefix the request id.
func prefixer(prefix string) func(ulid.ULID) string {
	if prefix == "" {
		return func(id ulid.ULID) string {
			return id.String()
		}
	}

	return func(id ulid.ULID) string {
		return prefix + "." + id.String()
	}
}
