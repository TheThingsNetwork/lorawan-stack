// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package crypto

import (
	"crypto/hmac"
	"crypto/sha256"

	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// HMACHash calculates the  Keyed-Hash Message Authentication Code (HMAC, RFC 2104) hash of the data.
func HMACHash(key types.AES128Key, payload []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key[:])
	_, err := h.Write(payload)
	if err != nil {
		return nil, err
	}
	return h.Sum([]byte{}), nil
}
