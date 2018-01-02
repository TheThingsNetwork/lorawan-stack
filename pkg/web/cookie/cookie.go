// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package cookie

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
)

const (
	rootKey    = "root"
	encoderKey = "cookie.encoder"
	tombstone  = "deleted"
)

// Cookie is the type of cookies with arbitrary values.
type Cookie struct {
	Value    interface{}
	Path     string
	Expires  time.Time
	MaxAge   int
	HttpOnly bool
}

// Cookies is a middleware function that makes the handlers capable of handling cookies via
// methods of this package.
func Cookies(root string, block, hash []byte) echo.MiddlewareFunc {
	s := securecookie.New(hash, block)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(rootKey, root)
			c.Set(encoderKey, s)
			return next(c)
		}
	}
}

func getConfig(c echo.Context) (string, *securecookie.SecureCookie, error) {
	root, ok := c.Get(rootKey).(string)
	if !ok || root == "" {
		root = "/"
	}

	encoder, _ := c.Get(encoderKey).(*securecookie.SecureCookie)
	if encoder == nil {
		return "", nil, fmt.Errorf("No cookie.encoder set")
	}

	return root, encoder, nil
}

// Get the cookie with the specified name.
func Get(c echo.Context, name string) (*Cookie, error) {
	root, s, err := getConfig(c)

	cookie, err := c.Request().Cookie(name)
	if err != nil || cookie.Value == tombstone {
		return nil, nil
	}

	res := &Cookie{
		Path:     strings.TrimPrefix(cookie.Path, root),
		Expires:  cookie.Expires,
		MaxAge:   cookie.MaxAge,
		HttpOnly: cookie.HttpOnly,
	}

	err = s.Decode(name, cookie.Value, &res.Value)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Set the cookie with the specified name.
func Set(c echo.Context, name string, cookie *Cookie) error {
	root, s, err := getConfig(c)

	str, err := s.Encode(name, &cookie.Value)
	if err != nil {
		return err
	}

	http.SetCookie(c.Response().Writer, &http.Cookie{
		Name:     name,
		Value:    str,
		Path:     path.Join(root, cookie.Path),
		Expires:  cookie.Expires,
		MaxAge:   cookie.MaxAge,
		HttpOnly: cookie.HttpOnly,
	})

	return nil
}

// Remove the cookie with the specified name (if it exists).
func Remove(c echo.Context, name string) error {
	root, _, err := getConfig(c)

	cookie, err := Get(c, name)
	if err != nil {
		return err
	}

	if cookie == nil {
		return nil
	}

	http.SetCookie(c.Response().Writer, &http.Cookie{
		Name:     name,
		Value:    tombstone,
		Path:     path.Join(root, cookie.Path),
		Expires:  time.Unix(1, 0),
		MaxAge:   0,
		HttpOnly: cookie.HttpOnly,
	})

	return nil
}
