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

package web

import (
	"fmt"
	"net/http"

	"github.com/golang/gddo/httputil"
	"github.com/labstack/echo"
)

// HTTPErrorBody is the type of an error body.
type HTTPErrorBody struct {
	Status      int    `json:"status"`
	Description string `json:"error_description"`
}

// Statusser is the interface of things that can have a specific http status.
type Statusser interface {
	// Status returns the http status
	Status() int
}

var (
	offers      = []string{"text/html", "application/json", "text/event-stream", "text/plain"}
	defaultType = "text/html"
)

// ErrorHandler is an echo.HTTPErrorHandler and renders them based on the accept header of the request.
func ErrorHandler(template string) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		httpe := HTTPFromError(err)

		body := &HTTPErrorBody{
			Status:      httpe.Code,
			Description: fmt.Sprintf("%s", httpe.Message),
		}

		if c.Response().Committed {
			return
		}

		var renderError error
		switch httputil.NegotiateContentType(c.Request(), offers, defaultType) {
		case "application/json", "text/event-stream":
			renderError = c.JSON(httpe.Code, body)
		case "text/html":
			renderError = c.Render(httpe.Code, template, map[string]interface{}{
				"error": body,
			})
		default:
			renderError = c.String(httpe.Code, fmt.Sprintf("%v %s\n", body.Status, body.Description))
		}

		if renderError != nil {
			_ = c.String(httpe.Code, fmt.Sprintf("%v %s\n", body.Status, body.Description))
		}
	}
}

// StatusCodeFromError returns the status code that represent the error.
func StatusCodeFromError(err error) int {
	switch v := err.(type) {
	case *echo.HTTPError:
		return v.Code
	case Statusser:
		return v.Status()
	default:
		return http.StatusInternalServerError
	}
}

// HTTPFromError creates an echo.HTTPError from an error.
func HTTPFromError(err error) *echo.HTTPError {
	switch e := err.(type) {
	case *echo.HTTPError:
		return e
	default:
		return echo.NewHTTPError(StatusCodeFromError(err), err.Error())
	}
}
