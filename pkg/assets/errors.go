// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package assets

import (
	"net/http"

	"github.com/golang/gddo/httputil"
	"github.com/labstack/echo"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

var (
	offers      = []string{"text/html", "application/json", "text/event-stream", "text/plain"}
	defaultType = "text/html"
)

var httpError = errors.Define("http_error", "HTTP Error: {message}")

// Errors renders the errors with the specified template.
func (a *Assets) Errors(name string, env interface{}) echo.MiddlewareFunc {
	template, templateErr := a.template(name)
	if templateErr != nil { // can't offer HTML if we don't have a template where we can render the error.
		offers = []string{"application/json", "text/event-stream", "text/plain"}
		defaultType = "application/json"
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil || c.Response().Committed {
				return err
			}
			status := http.StatusInternalServerError
			if echoErr, ok := err.(*echo.HTTPError); ok {
				status = echoErr.Code
				if ttnErr, ok := errors.From(echoErr.Internal); ok {
					if status == http.StatusInternalServerError {
						status = errors.HTTPStatusCode(ttnErr)
					}
					err = ttnErr
				}
			} else if ttnErr, ok := errors.From(err); ok {
				status = errors.HTTPStatusCode(ttnErr)
				err = ttnErr
			} else {
				err = httpError.WithCause(err).WithAttributes("message", err.Error())
			}
			switch httputil.NegotiateContentType(c.Request(), offers, defaultType) {
			case "text/html":
				t, templateError := a.fresh(name, template)
				if templateError != nil {
					return templateError
				}
				c.Response().WriteHeader(status)
				return t.Execute(c.Response().Writer, data{
					Root:  a.config.CDN,
					Env:   env,
					Error: err,
				})
			case "application/json", "text/event-stream":
				return c.JSON(status, err)
			default:
				return c.String(status, err.Error())
			}
		}
	}
}
