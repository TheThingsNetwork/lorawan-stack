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
	"github.com/golang/gddo/httputil"
	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/httperrors"
)

var (
	offers      = []string{"text/html", "application/json", "text/event-stream", "text/plain"}
	defaultType = "text/html"
)

// Errors renders the errors with the specified template.
func (a *Assets) Errors(name string, env interface{}) echo.MiddlewareFunc {
	template := a.template(name)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			if err == nil {
				return nil
			}

			if c.Response().Committed {
				return nil
			}

			e := from(err)
			status := httperrors.HTTPStatusCode(e)
			httperrors.SetErrorHeaders(e, c.Response().Header())

			switch httputil.NegotiateContentType(c.Request(), offers, defaultType) {
			case "text/html":
				t := a.fresh(name, template)
				data := data{
					Root:  a.config.CDN,
					Env:   env,
					Error: &e,
				}

				c.Response().WriteHeader(status)
				return t.Execute(c.Response().Writer, data)
			case "application/json", "text/event-stream":
				return c.JSON(status, e)
			default:
				return c.String(status, e.Error())
			}
		}
	}
}

type httpError struct {
	id      string      `json:"error_id"`
	message string      `json:"error_message"`
	typ     errors.Type `json:"error_type"`
}

// Error implements error.
func (e httpError) Error() string {
	return e.message
}

// Message implements errors.Error.
func (e httpError) Message() string {
	return e.message
}

// Code implements errors.Error.
func (e httpError) Code() errors.Code {
	return errors.NoCode
}

// Type implements errors.Error.
func (e httpError) Type() errors.Type {
	return e.typ
}

// Namespace implements errors.Error.
func (e httpError) Namespace() string {
	return ""
}

// Attributes implements errors.Error.
func (e httpError) Attributes() errors.Attributes {
	return nil
}

// ID implements errors.Error.
func (e httpError) ID() string {
	return e.id
}

func from(e error) errors.Error {
	if httpe, ok := e.(*echo.HTTPError); ok {
		msg, ok := httpe.Message.(string)
		if !ok {
			msg = httpe.Error()
		}

		return errors.ToImpl(httpError{
			id:      errors.NewID(),
			message: msg,
			typ:     httperrors.HTTPStatusToType(httpe.Code),
		})
	}

	return errors.From(e)
}
