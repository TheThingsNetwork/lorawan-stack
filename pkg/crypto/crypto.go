// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package crypto

func reverse(in []byte) []byte {
	l := len(in)
	out := make([]byte, l)
	for i := 0; i < l; i++ {
		out[l-i-1] = in[i]
	}
	return out
}
