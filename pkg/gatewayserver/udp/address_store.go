// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
