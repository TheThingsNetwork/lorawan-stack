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

package cryptoutil_test

import (
	"encoding/hex"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestMemKeyVault(t *testing.T) {
	a := assertions.New(t)

	plaintext, _ := hex.DecodeString("00112233445566778899AABBCCDDEEFF")
	kek, _ := hex.DecodeString("000102030405060708090A0B0C0D0E0F")
	ciphertext, _ := hex.DecodeString("1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5")

	v := cryptoutil.NewMemKeyVault(map[string][]byte{
		"foo": kek,
	})

	// Non-existing KEK.
	{
		_, err := v.Wrap(plaintext, "bar")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}
	{
		_, err := v.Unwrap(ciphertext, "bar")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Existing KEK.
	{
		actual, err := v.Wrap(plaintext, "foo")
		a.So(err, should.BeNil)
		a.So(actual, should.Resemble, ciphertext)
	}
	{
		actual, err := v.Unwrap(ciphertext, "foo")
		a.So(err, should.BeNil)
		a.So(actual, should.Resemble, plaintext)
	}
}
