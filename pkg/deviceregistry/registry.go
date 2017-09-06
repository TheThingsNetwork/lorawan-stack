// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry

import (
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Device represents the device stored in the registry.
type Device struct {
	*ttnpb.EndDevice
	stored map[string]interface{}

	key   store.PrimaryKey
	store store.Interface
}

func newDevice(ed *ttnpb.EndDevice, s store.Interface, k store.PrimaryKey, stored map[string]interface{}) *Device {
	if stored == nil {
		stored = store.Marshal(ed)
	}
	return &Device{
		EndDevice: ed,
		store:     s,
		key:       k,
		stored:    stored,
	}
}

// Registry is reponsible for mapping devices to their identities.
type Registry struct {
	store store.Interface
}

// New returns a new Registry with s as an internal Store.
func New(s store.Interface) *Registry {
	return &Registry{
		store: s,
	}
}

// Create stores devices data in underlying store.Interface and returns a new *Device.
func (r *Registry) Create(ed *ttnpb.EndDevice) (*Device, error) {
	m := store.Marshal(ed)

	id, err := r.store.Create(m)
	if err != nil {
		return nil, err
	}
	return newDevice(ed, r.store, id, m), nil
}

// FindDeviceByIdentifiers searches for devices matching specified device identifiers in underlying store.Interface.
func (r *Registry) FindDeviceByIdentifiers(ids ...*ttnpb.EndDeviceIdentifiers) ([]*Device, error) {
	if len(ids) == 0 {
		return []*Device{}, nil
	}

	intersection, err := r.store.FindBy(store.Marshal(&ttnpb.EndDevice{EndDeviceIdentifiers: *ids[0]}))
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(ids); i++ {
		m, err := r.store.FindBy(store.Marshal(&ttnpb.EndDevice{EndDeviceIdentifiers: *ids[i]}))
		if err != nil {
			return nil, err
		}
		for k := range m {
			if _, ok := intersection[k]; !ok {
				delete(intersection, k)
			}
		}
	}

	devices := make([]*Device, 0, len(intersection))
	for id, fields := range intersection {
		ed := &ttnpb.EndDevice{}
		if err := store.Unmarshal(fields, ed); err != nil {
			return nil, err
		}
		devices = append(devices, newDevice(ed, r.store, id, fields))
	}
	return devices, nil
}

// Update updates devices data in the underlying store.Interface.
func (d *Device) Update() error {
	diff := store.Diff(store.Marshal(d.EndDevice), d.stored)
	if len(diff) == 0 {
		return nil
	}
	if err := d.store.Update(d.key, diff); err != nil {
		return err
	}
	for k, v := range diff {
		d.stored[k] = v
	}
	return nil
}

// Delete removes device from the underlying store.Interface.
func (d *Device) Delete() error {
	return d.store.Delete(d.key)
}
