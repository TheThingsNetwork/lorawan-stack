// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package encoding_test

import (
	"encoding"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/encoding"
)

func TestStdlibCompatibility(t *testing.T) {
	var (
		textMarshaler     TextMarshaler
		textUnmarshaler   TextUnmarshaler
		binaryMarshaler   BinaryMarshaler
		binaryUnmarshaler BinaryUnmarshaler

		_ encoding.TextMarshaler     = textMarshaler
		_ encoding.TextUnmarshaler   = textUnmarshaler
		_ encoding.BinaryMarshaler   = binaryMarshaler
		_ encoding.BinaryUnmarshaler = binaryUnmarshaler
	)
}
