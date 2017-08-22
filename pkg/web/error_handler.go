// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package web

import (
	"fmt"
	"net/http"

	"github.com/golang/gddo/httputil"
	"github.com/labstack/echo"
)

// HTTPErrorBody is the type of an error body
type HTTPErrorBody struct {
	Status      int    `json:"status"`
	Description string `json:"error_description"`
}

// Statusser is the interface of things that can have a specific http status
type Statusser interface {
	// Status returns the http status
	Status() int
}

var (
	offers      = []string{"text/html", "application/json", "text/event-stream", "text/plain"}
	defaultType = "text/html"
)

// ErrorHandler is an echo.HTTPErrorHandler and renders them based
// on the accept header of the request.
func ErrorHandler(err error, c echo.Context) {
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
		renderError = c.Render(httpe.Code, "index", body)
	default:
		renderError = c.String(httpe.Code, fmt.Sprintf("%v %s\n", body.Status, body.Description))
	}

	if renderError != nil {
		c.String(httpe.Code, fmt.Sprintf("%v %s\n", body.Status, body.Description))
	}
}

// statusCodeFromError returns the statuscode that represent the error
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

// HTTPFromError creates an echo.HTTPError from an error
func HTTPFromError(err error) *echo.HTTPError {
	switch e := err.(type) {
	case *echo.HTTPError:
		return e
	default:
		return echo.NewHTTPError(StatusCodeFromError(err), err.Error())
	}
}
