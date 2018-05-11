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

package oauth

import (
	"github.com/RangelReale/osin"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/random"
)

// GenerateAuthorizeToken generates a 64-length authorization code based on the ttn random generator.
func (s *Server) GenerateAuthorizeToken(data *osin.AuthorizeData) (string, error) {
	return random.String(64), nil
}

// GenerateAccessToken generates 64-length access and refresh tokens based on the ttn random generator.
func (s *Server) GenerateAccessToken(data *osin.AccessData, generateRefresh bool) (accessToken string, refreshToken string, err error) {
	accessToken, err = auth.GenerateAccessToken(s.iss)
	if err != nil {
		return
	}

	if generateRefresh {
		refreshToken = random.String(64)
	}

	return
}
