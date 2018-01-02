// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/mohae/deepcopy"
)

// Interface represents the interface exposed by the *Registry.
type Interface interface {
	Create(ed *ttnpb.EndDevice) (*Device, error)
	FindBy(ed ...*ttnpb.EndDevice) ([]*Device, error)
}

// Registry is responsible for mapping devices to their identities.
type Registry struct {
	store store.Client
}

// New returns a new Registry with s as an internal Store.
func New(s store.Client) *Registry {
	return &Registry{
		store: s,
	}
}

// Create stores devices data in underlying store.Interface and returns a new *Device.
func (r *Registry) Create(ed *ttnpb.EndDevice) (*Device, error) {
	id, err := r.store.Create(ed)
	if err != nil {
		return nil, err
	}
	return newDevice(ed, r.store, id, ed), nil
}

// FindBy searches for devices matching specified device fields in underlying store.Interface.
func (r *Registry) FindBy(eds ...*ttnpb.EndDevice) ([]*Device, error) {
	found := make(map[store.PrimaryKey]*ttnpb.EndDevice)
	for i, ed := range eds {
		if ed == nil {
			return nil, errors.Errorf("Device %d is nil", i)
		}
		m, err := r.store.FindBy(ed, func() interface{} { return &ttnpb.EndDevice{} })
		if err != nil {
			return nil, err
		}
		for id, dev := range m {
			found[id] = dev.(*ttnpb.EndDevice)
		}
	}

	devices := make([]*Device, 0, len(found))
	for id, ed := range found {
		devices = append(devices, newDevice(ed, r.store, id, deepcopy.Copy(ed).(*ttnpb.EndDevice)))
	}
	return devices, nil
}

// FindDeviceByIdentifiers searches for devices matching specified device identifiers in r.
func FindDeviceByIdentifiers(r Interface, ids ...*ttnpb.EndDeviceIdentifiers) ([]*Device, error) {
	devs := make([]*ttnpb.EndDevice, len(ids))
	for i, id := range ids {
		if id == nil {
			return nil, errors.Errorf("Identifier %d is nil", i)
		}
		devs[i] = &ttnpb.EndDevice{EndDeviceIdentifiers: *id}
	}
	return r.FindBy(devs...)
}

// FindOneDeviceByIdentifiers searches for exactly one device matching specified device identifiers in r.
func FindOneDeviceByIdentifiers(r Interface, ids ...*ttnpb.EndDeviceIdentifiers) (*Device, error) {
	devs, err := FindDeviceByIdentifiers(r, ids...)
	if err != nil {
		return nil, err
	}
	switch len(devs) {
	case 0:
		return nil, ErrDeviceNotFound.New(errors.Attributes{
			"identifiers": ids,
		})
	case 1:
		return devs[0], nil
	default:
		return nil, ErrTooManyDevices.New(errors.Attributes{
			"identifiers": ids,
		})
	}
}
