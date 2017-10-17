// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package oauth

import (
	"time"

	"github.com/RangelReale/osin"
	"github.com/TheThingsNetwork/ttn/pkg/auth"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/random"
)

// GenerateAccessToken generates the JWT access token based on the access token request.
func (s *Server) GenerateAccessToken(data *osin.AccessData, generateRefresh bool) (string, string, error) {
	_, key, err := s.keys.GetCurrentPrivateKey()
	if err != nil {
		return "", "", err
	}

	rights, err := ParseScope(data.Scope)
	if err != nil {
		return "", "", err
	}

	userID := ""
	udata, ok := data.UserData.(*UserData)
	if ok && udata != nil {
		userID = udata.UserID
	}

	claims := &auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    s.iss,
			ExpiresAt: data.ExpireAt().Unix(),
			Subject:   auth.UserSubject(userID),
			IssuedAt:  time.Now().Add(-5 * time.Second).Unix(),
		},
		User:   userID,
		Rights: rights,
		Client: data.Client.GetId(),
	}

	accessToken, err := claims.Sign(key)
	if err != nil {
		return "", "", err
	}

	if !generateRefresh {
		return accessToken, "", nil
	}

	refreshToken := random.String(64)

	return accessToken, refreshToken, err
}
