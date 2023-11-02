// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package crypto_test

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDeriveRootWorSKey(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	nwkSEncKey := types.AES128Key{
		0xCE, 0x07, 0xA0, 0x09, 0xA3, 0x97, 0x0A, 0xC0, 0x51, 0x9A, 0x09, 0x9E, 0xD5, 0x3E, 0x55, 0x0B,
	}
	a.So(crypto.DeriveRootWorSKey(nwkSEncKey), should.Equal, types.AES128Key{
		0xEE, 0x91, 0xDC, 0x1A, 0x66, 0x66, 0xC0, 0x6E, 0x82, 0x77, 0xDE, 0x6D, 0xB4, 0xDB, 0x94, 0x5F,
	})
}
