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

	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func generateAPIKey(ctx context.Context, name string, rights ...ttnpb.Right) (key *ttnpb.APIKey, token string, err error) {
	token, err = auth.APIKey.Generate(ctx, "")
	if err != nil {
		return nil, "", err
	}
	_, generatedID, generatedKey, err := auth.SplitToken(token)
	if err != nil {
		panic(err) // Bug in either Generate or SplitToken.
	}
	hashedKey, err := auth.Hash(ctx, generatedKey)
	if err != nil {
		return nil, "", err
	}
	key = &ttnpb.APIKey{
		ID:     generatedID,
		Key:    hashedKey,
		Name:   name,
		Rights: rights,
	}
	return key, token, nil
}
