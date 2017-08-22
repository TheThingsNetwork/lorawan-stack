// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

// AddressStore exposes methods to set and get network addresses.
type AddressStore interface {
	Get(types.EUI64) (net.Addr, bool)
	Set(types.EUI64, net.Addr)
}

// StaticStore is a store that always returns the address in the struct
type StaticStore struct {
	net.Addr
}

// Get implements AddressStore
func (s StaticStore) Get(types.EUI64) (net.Addr, bool) { return net.Addr(s), true }

// Set implements AddressStore
func (s StaticStore) Set(types.EUI64, net.Addr) { return }
