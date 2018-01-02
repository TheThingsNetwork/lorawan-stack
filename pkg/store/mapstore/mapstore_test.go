// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store/storetest"
)

func TestMapStore(t *testing.T) {
	s := New()
	storetest.TestTypedStore(t, s)
}
