// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/storetest"
)

func TestIndexedMapStore(t *testing.T) {
	storetest.TestTypedStore(t, func() store.TypedStore {
		return NewIndexed(storetest.IndexedFields...)
	})
}
