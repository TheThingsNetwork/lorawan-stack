// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestAES128(t *testing.T) {
	a := assertions.New(t)

	aes := AES128Key{
		0x12, 0x34, 0xAE, 0x00, 0x3A, 0xB7, 0x38, 0x01,
		0x52, 0x31, 0x0B, 0x53, 0x3A, 0xB7, 0x38, 0x01,
	}

	// JSON
	{
		jsonBytes, err := json.Marshal(aes)
		if !a.So(err, should.BeNil) {
			panic(err)
		}
		jsonContent := string(jsonBytes)
		a.So(jsonContent, should.ContainSubstring, "1234AE003AB7380152310B533AB73801")
	}
}
