// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry

import (
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/mohae/deepcopy"
)

// Device represents the device stored in the registry.
type Device struct {
	*ttnpb.EndDevice
	stored *ttnpb.EndDevice

	key   store.PrimaryKey
	store store.Client
}

func newDevice(ed *ttnpb.EndDevice, s store.Client, k store.PrimaryKey, stored *ttnpb.EndDevice) *Device {
	if stored == nil {
		stored = deepcopy.Copy(ed).(*ttnpb.EndDevice)
	}
	return &Device{
		EndDevice: ed,
		store:     s,
		key:       k,
		stored:    stored,
	}
}

// Update updates devices data in the underlying store.Interface.
func (d *Device) Update() error {
	if err := d.store.Update(d.key, d.EndDevice, d.stored); err != nil {
		return err
	}
	d.stored = deepcopy.Copy(d.EndDevice).(*ttnpb.EndDevice)
	return nil
}

// Delete removes device from the underlying store.Interface.
func (d *Device) Delete() error {
	return d.store.Delete(d.key)
}
