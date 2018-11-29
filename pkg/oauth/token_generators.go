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
)

func (s *server) GenerateAuthorizeToken(_ *osin.AuthorizeData) (string, error) {
	return auth.AuthorizationCode.Generate(s.ctx, "")
}

func (s *server) GenerateAccessToken(_ *osin.AccessData, generateRefresh bool) (accessToken string, refreshToken string, err error) {
	var id string
	if generateRefresh {
		id, err = auth.GenerateID(s.ctx)
		if err != nil {
			return
		}
	}
	accessToken, err = auth.AccessToken.Generate(s.ctx, id)
	if err != nil {
		return
	}
	if generateRefresh {
		refreshToken, err = auth.RefreshToken.Generate(s.ctx, id)
		if err != nil {
			return
		}
	}
	return
}
