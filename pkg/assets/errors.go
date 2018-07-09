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
	"html/template"
	"net/http"

	"github.com/golang/gddo/httputil"
	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/assets/templates"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

var (
	errInvalidConfiguration = errors.DefineInvalidArgument("invalid_configuration", "invalid configuration")
	errTemplateNotFound     = errors.DefineNotFound("not_found", "template `{name}` not found", "name")
	errHTTP                 = errors.Define("http_error", "HTTP error: {message}")
)

// Errors renders an error according to a negotiated content type.
func (a *Assets) Errors(name string, env interface{}) echo.MiddlewareFunc {
	var (
		offers       = []string{"application/json", "text/event-stream", "text/plain"}
		defaultType  = "application/json"
		template     *template.Template
		templateErr  error
		templateData = templates.Data{
			Env: env,
		}
	)
	if a.fs != nil {
		template, templateErr = a.loadTemplate(name)
		templateData.Root = a.config.Mount
	} else {
		template = templates.Error
		templateData.Root = a.config.CDN
	}
	if templateErr == nil {
		offers = append(offers, "text/html")
		defaultType = "text/html"
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
				err = errHTTP.WithCause(err).WithAttributes("message", err.Error())
			}

			switch httputil.NegotiateContentType(c.Request(), offers, defaultType) {
			case "text/html":
				c.Response().WriteHeader(status)
				templateData.Data = templates.ErrorData{
					Title: http.StatusText(status),
					Error: err,
				}
				return template.Execute(c.Response().Writer, templateData)
			case "application/json", "text/event-stream":
				return c.JSON(status, err)
			default:
				return c.String(status, err.Error())
			}
		}
	}
}
