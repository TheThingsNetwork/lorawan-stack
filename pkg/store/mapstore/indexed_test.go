// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store/storetest"
)

func TestIndexedMapStore(t *testing.T) {
	s := NewIndexed("foo", "bar")
	storetest.TestTypedStore(t, s)
}
