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

package crypto_test

import (
	"bytes"
	"crypto/subtle"
	"encoding/gob"
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestHMACHash(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	test := struct {
		ID    string
		Value []byte
	}{
		ID:    "test-id",
		Value: []byte("test value"),
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(test)
	a.So(err, should.BeNil)

	for _, tc := range []struct {
		Name    string
		Payload []byte
	}{
		{
			Name:    "SmallMessage",
			Payload: []byte("hello"),
		},
		{
			Name:    "EncodedStruct",
			Payload: buf.Bytes(),
		},
		{
			Name:    "PrivateKey",
			Payload: []byte("-----BEGIN EC PRIVATE KEY-----MHcCAQEEIBVXljefOUPY++0sovcF0dboOLEJz4eZ9DoUE8o9Y7GHoAoGCCqGSM49AwEHoUQDQgAEjf3zZPXlc/sseTt7YzF0o61feXvk98JFyy+s/j0gzMzUjEka7+WzTPERi9uMQjERns1qXG/9DJLe/Qxi0r84hA==-----END EC PRIVATE KEY-----"), //nolint:lll
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			key := types.AES128Key{
				0x12, 0x34, 0xAE, 0x00, 0x3A, 0xB7, 0x38, 0x01,
				0x52, 0x31, 0x0B, 0x53, 0x3A, 0xB7, 0x38, 0x01,
			}
			hash1, err := HMACHash(key, tc.Payload)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to hash message: %v", err)
			}
			hash2, err := HMACHash(key, tc.Payload)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to hash message: %v", err)
			}
			a.So(subtle.ConstantTimeCompare(hash1, hash2), should.Equal, 1)
		})
	}
}
