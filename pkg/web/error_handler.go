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

	"github.com/labstack/echo"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

// ErrorHandler is an echo.HTTPErrorHandler.
func ErrorHandler(err error, c echo.Context) {
	status := errors.HTTPStatusCode(err)
	msg := err.Error()
	if h, ok := err.(*echo.HTTPError); ok {
		status = h.Code
		msg = fmt.Sprintf("%s", h.Message)
	}
	c.String(status, msg)
}
