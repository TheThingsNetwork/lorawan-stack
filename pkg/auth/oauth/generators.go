// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package oauth

import (
	"github.com/RangelReale/osin"
	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/random"
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
