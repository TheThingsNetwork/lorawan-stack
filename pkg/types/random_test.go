// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"math/rand"
	"testing"
)

func TestRandom(t *testing.T) {
	source := rand.NewSource(1)
	NewPopulatedDevNonce(source)
	NewPopulatedJoinNonce(source)
	NewPopulatedNetID(source)
	NewPopulatedDevAddr(source)
	NewPopulatedDevAddrPrefix(source)
	NewPopulatedEUI64(source)
	NewPopulatedAES128Key(source)
}
