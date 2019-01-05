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

package cookie

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
)

const (
	// encoderKey is the key where the encoder will be stored on the request.
	encoderKey = "cookie.encoder"

	// tombstone is the cookie tombstone value.
	tombstone = "<deleted>"
)

// Cookie is a description of a cookie used for consistent cookie setting and deleting.
type Cookie struct {
	// Name is the name of the cookie.
	Name string

	// Path is path of the cookie.
	Path string

	// MaxAge is the max age of the cookie.
	MaxAge time.Duration

	// HTTPOnly restricts the cookie to HTTP (no javascript access).
	HTTPOnly bool
}

// Cookies is a middleware function that makes the handlers capable of handling cookies via
// methods of this package.
func Cookies(block, hash []byte) echo.MiddlewareFunc {
	s := securecookie.New(hash, block)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(encoderKey, s)
			return next(c)
		}
	}
}

func getConfig(c echo.Context) (*securecookie.SecureCookie, error) {
	encoder, _ := c.Get(encoderKey).(*securecookie.SecureCookie)
	if encoder == nil {
		return nil, fmt.Errorf("No cookie.encoder set")
	}

	return encoder, nil
}

// Get decodes the cookie into the value. Returns false if the cookie is not there.
func (d *Cookie) Get(c echo.Context, v interface{}) (bool, error) {
	s, err := getConfig(c)
	if err != nil {
		return false, err
	}

	cookie, err := c.Request().Cookie(d.Name)
	if err != nil || cookie.Value == tombstone {
		return false, nil
	}

	err = s.Decode(d.Name, cookie.Value, v)
	if err != nil {
		d.Remove(c)
		return false, nil
	}

	return true, nil
}

// Set the value of the cookie.
func (d *Cookie) Set(c echo.Context, v interface{}) error {
	s, err := getConfig(c)
	if err != nil {
		return err
	}

	str, err := s.Encode(d.Name, v)
	if err != nil {
		return err
	}

	http.SetCookie(c.Response().Writer, &http.Cookie{
		Name:     d.Name,
		Path:     d.Path,
		MaxAge:   int(d.MaxAge.Nanoseconds() / 1000),
		HttpOnly: d.HTTPOnly,
		Value:    str,
	})

	return nil
}

// Exists checks if the cookies exists.
func (d *Cookie) Exists(c echo.Context) bool {
	cookie, err := c.Request().Cookie(d.Name)
	return err == nil && cookie.Value != tombstone
}

// Remove the cookie with the specified name (if it exists).
func (d *Cookie) Remove(c echo.Context) {
	if !d.Exists(c) {
		return
	}

	http.SetCookie(c.Response().Writer, &http.Cookie{
		Name:     d.Name,
		Path:     d.Path,
		HttpOnly: d.HTTPOnly,
		Value:    tombstone,
		Expires:  time.Unix(1, 0),
		MaxAge:   0,
	})
}
