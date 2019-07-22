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

// Package web implements a middleware to handle HTTP errors.
package web

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/golang/gddo/httputil"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/errors"
	_ "go.thethings.network/lorawan-stack/pkg/ttnpb" // imported for side-effect of correct TTN error rendering.
)

var globalRenderers = map[string]ErrorRenderer{}

// ErrorRenderer is an interface for rendering errors to HTTP responses.
type ErrorRenderer interface {
	RenderError(c echo.Context, statusCode int, err error) error
}

// ErrorRendererFunc is a function signature that implements ErrorRenderer.
type ErrorRendererFunc func(c echo.Context, statusCode int, err error) error

// RenderError implements the ErrorRenderer interface.
func (f ErrorRendererFunc) RenderError(c echo.Context, statusCode int, err error) error {
	return f(c, statusCode, err)
}

// RegisterRenderer registers a global error renderer.
func RegisterRenderer(contentType string, renderer ErrorRenderer) {
	globalRenderers[contentType] = renderer
}

// ProcessError processes an HTTP error by converting it if appropriate, and
// determining the HTTP status code to return.
func ProcessError(in error) (statusCode int, err error) {
	statusCode, err = http.StatusInternalServerError, in
	if echoErr, ok := err.(*echo.HTTPError); ok {
		if echoErr.Code != 0 {
			statusCode = echoErr.Code
		}
		if echoErr.Internal == nil {
			ttnErr := errors.FromHTTPStatusCode(statusCode, "message")
			return statusCode, ttnErr.WithAttributes("message", fmt.Sprint(echoErr.Message))
		}
		err = echoErr.Internal
	}
	if ttnErr, ok := errors.From(err); ok {
		statusCode = errors.ToHTTPStatusCode(ttnErr)
		return statusCode, ttnErr
	}
	ttnErr := errors.FromHTTPStatusCode(statusCode, "message")
	return statusCode, ttnErr.WithCause(err).WithAttributes("message", err.Error())
}

// ErrorMiddleware returns an Echo middleware that catches errors in the chain,
// and renders them using the negotiated renderer. Global renderers can be registered
// with RegisterRenderer, and extra renderers can be passed to this function.
func ErrorMiddleware(extraRenderers map[string]ErrorRenderer) echo.MiddlewareFunc {
	renderers := make(map[string]ErrorRenderer)
	for contentType, renderer := range globalRenderers {
		renderers[contentType] = renderer
	}
	for contentType, renderer := range extraRenderers {
		renderers[contentType] = renderer
	}
	offers := make([]string, 0, len(renderers))
	for k := range renderers {
		offers = append(offers, k)
	}
	sort.Strings(offers) // Send offers in alphabetical order

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil || c.Response().Committed {
				return err
			}
			statusCode, err := ProcessError(err)
			renderer := httputil.NegotiateContentType(c.Request(), offers, "application/json")
			if renderer != "" {
				return renderers[renderer].RenderError(c, statusCode, err)
			}
			return err
		}
	}
}

func init() {
	RegisterRenderer("application/json", ErrorRendererFunc(func(c echo.Context, statusCode int, err error) error {
		return c.JSON(statusCode, err)
	}))
}
