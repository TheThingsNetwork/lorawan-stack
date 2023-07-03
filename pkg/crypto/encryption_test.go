// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestEncryptDecrypt(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Message []byte
	}{
		{
			Name:    "SmallMessage",
			Message: []byte("hello"),
		},
		{
			Name:    "APIKey",
			Message: []byte("NNSXS.W5A2226537J2PWZ7YOCJZPUUSFKM6RJ4RGCL2YY.AICDXTLG7CIEDTP3JD4JVYTMGZN75DU5WSKU7RNRZELAJFRWQCHA"),
		},
		{
			Name:    "PrivateKey",
			Message: []byte("-----BEGIN EC PRIVATE KEY-----MHcCAQEEIBVXljefOUPY++0sovcF0dboOLEJz4eZ9DoUE8o9Y7GHoAoGCCqGSM49AwEHoUQDQgAEjf3zZPXlc/sseTt7YzF0o61feXvk98JFyy+s/j0gzMzUjEka7+WzTPERi9uMQjERns1qXG/9DJLe/Qxi0r84hA==-----END EC PRIVATE KEY-----"),
		},
	} {
		t.Run(fmt.Sprintf("%s", tc.Name), func(t *testing.T) {
			a := assertions.New(t)
			var key types.AES128Key
			_, err := rand.Read(key[:])
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to generate key: %v", err)
			}
			encrypted, err := Encrypt(key, tc.Message)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to encrypt message: %v", err)
			}
			expectedCipherLen := 12 + 16 + len(tc.Message) // Nonce + Tag + Message
			if !a.So(len(encrypted), should.Equal, expectedCipherLen) {
				t.Fatalf("Invalid cipher length: %v", len(encrypted))
			}
			decrypted, err := Decrypt(key, encrypted)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to decrypt message: %v", err)
			}
			if !a.So(tc.Message, should.Resemble, decrypted) {
				t.Fatalf("Failed to decrypt message: %v", err)
			}
		})
	}
}
