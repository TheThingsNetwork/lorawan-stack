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

import (
	"net/http"
)

type redirector struct {
	Path     string
	Code     int
	Location string
}

func (red redirector) RegisterRoutes(s *Server) {
	s.RootRouter().Path(red.Path).Handler(http.RedirectHandler(red.Location, red.Code))
}

// Redirect returns a Registerer that redirects requests to the given path to
// the given location with the given code.
func Redirect(path string, code int, location string) Registerer {
	return redirector{
		Path:     path,
		Code:     code,
		Location: location,
	}
}
