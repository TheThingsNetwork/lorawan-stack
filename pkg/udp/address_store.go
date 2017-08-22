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
type staticStore struct {
	net.Addr
}

// StaticStore returns an AddressStore that always returns the same net.Addr.
func StaticStore(addr net.Addr) AddressStore { return staticStore{addr} }

// Get implements AddressStore
func (s staticStore) Get(types.EUI64) (net.Addr, bool) { return net.Addr(s), true }

// Set implements AddressStore
func (s staticStore) Set(types.EUI64, net.Addr) { return }
