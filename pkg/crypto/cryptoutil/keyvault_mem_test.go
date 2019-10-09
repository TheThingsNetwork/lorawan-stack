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
	"crypto/ecdsa"
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
		"kek1": kek,
		"cert1": []byte(`-----BEGIN CERTIFICATE-----
MIIBfDCCASKgAwIBAgIBBDAKBggqhkjOPQQDAjAkMRAwDgYDVQQKEwdBY21lIENv
MRAwDgYDVQQDEwdSb290IENBMCAXDTE5MDgwNzExMzYxOFoYDzIxMTkwNzE0MTEz
NjE4WjAyMRAwDgYDVQQKEwdBY21lIENvMR4wHAYDVQQDDBVjbGllbnRfYXV0aF90
ZXN0X2NlcnQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASN/fNk9eVz+yx5O3tj
MXSjrV95e+T3wkXLL6z+PSDMzNSMSRrv5bNM8RGL24xCMRGezWpcb/0Mkt79DGLS
vziEozUwMzAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0lBAwwCgYIKwYBBQUHAwIwDAYD
VR0TAQH/BAIwADAKBggqhkjOPQQDAgNIADBFAiEAylB8RCRTv3FJYonJkfKTVOMN
cr7idt4xexCs+l8ALzMCIGBu4+S8YWGq9yQ4BL86Rcf7j7veXm57o6kjxU4F6V7x
-----END CERTIFICATE-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBVXljefOUPY++0sovcF0dboOLEJz4eZ9DoUE8o9Y7GHoAoGCCqGSM49
AwEHoUQDQgAEjf3zZPXlc/sseTt7YzF0o61feXvk98JFyy+s/j0gzMzUjEka7+Wz
TPERi9uMQjERns1qXG/9DJLe/Qxi0r84hA==
-----END EC PRIVATE KEY-----
`),
	})

	// Existing KEK.
	{
		actual, err := v.Wrap(plaintext, "kek1")
		a.So(err, should.BeNil)
		a.So(actual, should.Resemble, ciphertext)
	}
	{
		actual, err := v.Unwrap(ciphertext, "kek1")
		a.So(err, should.BeNil)
		a.So(actual, should.Resemble, plaintext)
	}

	// Non-existing KEK.
	{
		_, err := v.Wrap(plaintext, "kek2")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}
	{
		_, err := v.Unwrap(ciphertext, "kek2")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Existing certificate.
	{
		cert, err := v.LoadCertificate("cert1")
		a.So(err, should.BeNil)
		if a.So(len(cert.Certificate), should.Equal, 1) {
			a.So(len(cert.Certificate[0]), should.Equal, 384)
		}
		a.So(cert.PrivateKey, should.HaveSameTypeAs, &ecdsa.PrivateKey{})
	}

	// Non-existing certificate.
	{
		_, err := v.LoadCertificate("cert2")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}
}
