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

package cups_test

import (
	"bytes"
	"crypto/x509"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/basicstation/cups"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestUpdateInfoResponse(t *testing.T) {
	for _, tt := range []struct {
		Name string
		cups.UpdateInfoResponse
	}{
		{Name: "Empty"},
		{Name: "Full", UpdateInfoResponse: cups.UpdateInfoResponse{
			CUPSURI:         "https://cups.example.com",
			LNSURI:          "https://lns.example.com",
			CUPSCredentials: bytes.Repeat([]byte("CUPS CREDENTIALS"), 1000),
			LNSCredentials:  bytes.Repeat([]byte("LNS CREDENTIALS"), 1000),
			SignatureKeyCRC: 12345678,
			Signature:       bytes.Repeat([]byte("THIS IS THE SIGNATURE"), 100),
			UpdateData:      bytes.Repeat([]byte("THIS IS THE UPDATE DATA"), 1000),
		}},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)

			data, err := tt.UpdateInfoResponse.MarshalBinary()
			a.So(err, should.BeNil)

			var dec cups.UpdateInfoResponse
			err = dec.UnmarshalBinary(data)
			a.So(err, should.BeNil)
			a.So(dec, should.Resemble, tt.UpdateInfoResponse)
		})
	}
}

func TestTokenCredentials(t *testing.T) {
	expected := []byte{0x4c, 0x45, 0x54, 0x53, 0x20, 0x4e, 0x4f, 0x54, 0x20, 0x45, 0x4e, 0x43, 0x52, 0x59, 0x50, 0x54, 0x00, 0x00, 0x00, 0x00, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x3a, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x0d, 0x0a}
	a := assertions.New(t)
	creds, err := cups.TokenCredentials(&x509.Certificate{
		Raw: []byte("LETS NOT ENCRYPT"),
	}, "token")
	a.So(err, should.BeNil)
	a.So(creds, should.Resemble, expected)
}
