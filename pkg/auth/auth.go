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

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

const (
	idLength  = 24
	keyLength = 32
)

var enc = base32.StdEncoding.WithPadding(base32.NoPadding)

var (
	// APIKey authenticates calls on behalf of an entity on itself, or of users or organizations.
	APIKey = TokenType(enc.EncodeToString([]byte("key")))
	// AccessToken authenticates calls on behalf of a user that authorized an OAuth client.
	AccessToken = TokenType(enc.EncodeToString([]byte("acc")))
	// RefreshToken is used by OAuth clients to refresh AccessTokens.
	RefreshToken = TokenType(enc.EncodeToString([]byte("ref")))
	// AuthorizationCode is used by OAuth clients to exchange AccessTokens.
	AuthorizationCode = TokenType(enc.EncodeToString([]byte("aut")))
)

// TokenType indicates the type of a token.
type TokenType string

// Generate a token of this type.
// The ID is only generated if not already given.
func (t TokenType) Generate(ctx context.Context, id string) (token string, err error) {
	if id == "" {
		id, err = GenerateID(ctx)
		if err != nil {
			return "", err
		}
	}
	key, err := GenerateKey(ctx)
	if err != nil {
		return "", err
	}
	return JoinToken(t, id, key), nil
}

var errInvalidToken = errors.DefineInvalidArgument("token", "invalid token")

// SplitToken splits the token from "<prefix>.<id>.<key>".
func SplitToken(token string) (tokenType TokenType, id, key string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", "", errInvalidToken
	}
	switch TokenType(parts[0]) {
	case APIKey, AccessToken, RefreshToken, AuthorizationCode:
		return TokenType(parts[0]), parts[1], parts[2], nil
	default:
		return "", "", "", errInvalidToken
	}
}

// JoinToken joins the token as "<prefix>.<id>.<key>".
func JoinToken(tokenType TokenType, id, key string) string {
	return fmt.Sprintf("%s.%s.%s", string(tokenType), id, key)
}

// GenerateID generates the "id" part of the token.
func GenerateID(_ context.Context) (string, error) {
	var b [idLength]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return "", err
	}
	return enc.EncodeToString(b[:]), nil
}

// GenerateKey generates the "key" part of the token.
func GenerateKey(_ context.Context) (string, error) {
	var b [keyLength]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return "", err
	}
	return enc.EncodeToString(b[:]), nil
}
