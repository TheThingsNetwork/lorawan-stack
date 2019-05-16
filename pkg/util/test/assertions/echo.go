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

package assertions

import (
	"fmt"

	echo "github.com/labstack/echo/v4"
)

const (
	needEcho           = "This assertion requires an *echo.Echo but got %T"
	needMethodString   = "This assertions requires the method to be a string"
	needPathString     = "This assertions requires the path to be a string"
	shouldHaveRoute    = "Should have a route %s %s"
	shouldNotHaveRoute = "Should not have a route %s %s"
)

// ShouldHaveRoute takes as argument an *echo.Echo, the method string and path string.
// If a route is found with the given method and path, this function returns an empty string.
// Otherwise, it returns a string describing the error.
func ShouldHaveRoute(actual interface{}, expected ...interface{}) (message string) {
	e, ok := actual.(*echo.Echo)
	if !ok {
		return fmt.Sprintf(needEcho, actual)
	}
	if message = need(2, expected); message != success {
		return
	}
	method, ok := expected[0].(string)
	if !ok {
		return needMethodString
	}
	path, ok := expected[1].(string)
	if !ok {
		return needPathString
	}
	for _, r := range e.Routes() {
		if r.Method == method && r.Path == path {
			return success
		}
	}
	return fmt.Sprintf(shouldHaveRoute, method, path)
}

// ShouldNotHaveRoute takes as argument an *echo.Echo, the method string and path string.
// If no route is found with the given method and path, this function returns an empty string.
// Otherwise, it returns a string describing the error.
func ShouldNotHaveRoute(actual interface{}, expected ...interface{}) (message string) {
	e, ok := actual.(*echo.Echo)
	if !ok {
		return fmt.Sprintf(needEcho, actual)
	}
	if message = need(2, expected); message != success {
		return
	}
	method, ok := expected[0].(string)
	if !ok {
		return needMethodString
	}
	path, ok := expected[1].(string)
	if !ok {
		return needPathString
	}
	for _, r := range e.Routes() {
		if r.Method == method && r.Path == path {
			return fmt.Sprintf(shouldNotHaveRoute, method, path)
		}
	}
	return success
}
