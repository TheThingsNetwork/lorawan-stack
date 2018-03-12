// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

// AddressStore exposes methods to set and get network addresses.
type AddressStore interface {
	GetDownlinkAddress(types.EUI64) (net.Addr, bool)
	SetUplinkAddress(types.EUI64, net.Addr)
	SetDownlinkAddress(types.EUI64, net.Addr)
}

// StaticStore is a store that always returns the address in the struct
type staticStore struct {
	net.Addr
}

// StaticStore returns an AddressStore that always returns the same net.Addr.
func StaticStore(addr net.Addr) AddressStore { return staticStore{addr} }

func (s staticStore) GetDownlinkAddress(types.EUI64) (net.Addr, bool) { return net.Addr(s), true }
func (s staticStore) SetDownlinkAddress(types.EUI64, net.Addr)        { return }
func (s staticStore) SetUplinkAddress(types.EUI64, net.Addr)          { return }
