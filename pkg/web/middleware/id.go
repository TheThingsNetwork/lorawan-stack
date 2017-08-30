// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/pseudorandom"
	"github.com/labstack/echo"
	"github.com/oklog/ulid"
)

// used to mock time
var now = time.Now

// ID adds a request id to the request.
func ID(prefix string) echo.MiddlewareFunc {
	id := newID(prefix)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id, err := id.generate()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to generate request ID: %s", err))
			}

			c.Response().Header().Set("X-Request-ID", id)
			return next(c)
		}
	}
}

// id a generator of new ids. It uses ULID and a pool of entropy sources under the hood.
type id struct {
	prefixer func(ulid.ULID) string
	pool     sync.Pool
}

// newID creates a new id.
func newID(prefix string) *id {
	return &id{
		prefixer: prefixer(prefix),
		pool: sync.Pool{
			New: func() interface{} {
				return rand.New(rand.NewSource(int64(pseudorandom.Intn(int(math.MaxInt32)))))
			},
		},
	}
}

// generate generates a new ULID.
func (i *id) generate() (string, error) {
	entropy, ok := i.pool.Get().(*rand.Rand)
	defer i.pool.Put(entropy)
	if !ok {
		return "", fmt.Errorf("Failed to get an entropy source")
	}

	id, err := ulid.New(ulid.Timestamp(now()), entropy)
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
