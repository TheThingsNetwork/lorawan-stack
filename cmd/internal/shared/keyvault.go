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

package shared

import (
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
)

// KeyVault returns an initialized crypto.KeyVault from the given config.
func KeyVault(config config.ServiceBase) crypto.KeyVault {
	switch config.KeyVault.Backend {
	case "static":
		return cryptoutil.NewMemKeyVault(config.KeyVault.Static)
	default:
		return cryptoutil.NewMemKeyVault(map[string][]byte{})
	}
}
