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

package web

import "fmt"

func ShouldHaveRoute(actual interface{}, expected ...interface{}) string {
	e, ok := actual.(*Server)
	if !ok {
		return fmt.Sprintf("expected *Server but got %T", actual)
	}

	if len(expected) != 2 {
		return "expected 2 arguments to ShouldHaveRoute (method and path)"
	}

	method, ok := expected[0].(string)
	if !ok {
		return "expected method to be string"
	}

	path, ok := expected[1].(string)
	if !ok {
		return "expected path to be string"
	}

	for _, r := range e.Routes() {
		if r.Method == method && r.Path == path {
			return ""
		}
	}

	return fmt.Sprintf("should have a route %s %s", method, path)
}

func ShouldNotHaveRoute(actual interface{}, expected ...interface{}) string {
	e, ok := actual.(*Server)
	if !ok {
		return fmt.Sprintf("expected *Server but got %T", actual)
	}

	if len(expected) != 2 {
		return "expected 2 arguments to ShouldHaveRoute (method and path)"
	}

	method, ok := expected[0].(string)
	if !ok {
		return "expected method to be string"
	}

	path, ok := expected[1].(string)
	if !ok {
		return "expected path to be string"
	}

	for _, r := range e.Routes() {
		if r.Method == method && r.Path == path {
			return fmt.Sprintf("should not have a route %s %s", method, path)
		}
	}

	return ""
}
