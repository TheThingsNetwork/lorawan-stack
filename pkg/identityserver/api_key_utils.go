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

package identityserver

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var apiKeyHashSettings auth.HashValidator = pbkdf2.PBKDF2{
	Iterations: 1000,
	KeyLength:  32,
	Algorithm:  pbkdf2.Sha256,
	SaltLength: 16,
}

// GenerateAPIKey generates a new API key with the given name for the set of rights
func GenerateAPIKey(ctx context.Context, name string, expiresAt *time.Time, rights ...ttnpb.Right) (key *ttnpb.APIKey, token string, err error) {
	token, err = auth.APIKey.Generate(ctx, "")
	if err != nil {
		return nil, "", err
	}
	_, generatedID, generatedKey, err := auth.SplitToken(token)
	if err != nil {
		panic(err) // Bug in either Generate or SplitToken.
	}
	hashedKey, err := auth.Hash(auth.NewContextWithHashValidator(ctx, apiKeyHashSettings), generatedKey)
	if err != nil {
		return nil, "", err
	}
	key = &ttnpb.APIKey{
		Id:        generatedID,
		Key:       hashedKey,
		Name:      name,
		Rights:    rights,
		ExpiresAt: ttnpb.ProtoTime(expiresAt),
	}
	return key, token, nil
}
