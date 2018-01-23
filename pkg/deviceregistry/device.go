// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry

import (
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Device represents the device stored in the registry.
type Device struct {
	*ttnpb.EndDevice
	key   store.PrimaryKey
	store store.Client
}

func newDevice(ed *ttnpb.EndDevice, s store.Client, k store.PrimaryKey) *Device {
	return &Device{
		EndDevice: ed,
		store:     s,
		key:       k,
	}
}

// Update updates devices data in the underlying store.Interface.
func (d *Device) Update(fields ...string) error {
	return d.store.Update(d.key, d.EndDevice, fields...)
}

// Delete removes device from the underlying store.Interface.
func (d *Device) Delete() error {
	return d.store.Delete(d.key)
}
