// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Interface represents the interface exposed by the *Registry.
type Interface interface {
	Create(ed *ttnpb.EndDevice, fields ...string) (*Device, error)
	FindBy(ed *ttnpb.EndDevice, fields ...string) ([]*Device, error)
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
func (r *Registry) Create(ed *ttnpb.EndDevice, fields ...string) (*Device, error) {
	id, err := r.store.Create(ed, fields...)
	if err != nil {
		return nil, err
	}
	return newDevice(ed, r.store, id), nil
}

// FindBy searches for devices matching specified device fields in underlying store.Interface. The returned slice contains unique devices, matching at least one of values in eds.
func (r *Registry) FindBy(ed *ttnpb.EndDevice, fields ...string) ([]*Device, error) {
	if ed == nil {
		return nil, errors.New("Device specified is nil")
	}

	found, err := r.store.FindBy(ed, func() interface{} { return &ttnpb.EndDevice{} }, fields...)
	if err != nil {
		return nil, err
	}

	devices := make([]*Device, 0, len(found))
	for id, ed := range found {
		devices = append(devices, newDevice(ed.(*ttnpb.EndDevice), r.store, id))
	}
	return devices, nil
}

// FindDeviceByIdentifiers searches for devices matching specified device identifiers in r.
func FindDeviceByIdentifiers(r Interface, id *ttnpb.EndDeviceIdentifiers) ([]*Device, error) {
	if id == nil {
		return nil, errors.New("Identifiers specified are nil")
	}
	return r.FindBy(&ttnpb.EndDevice{EndDeviceIdentifiers: *id}, "EndDeviceIdentifiers")
}

// FindOneDeviceByIdentifiers searches for exactly one device matching specified device identifiers in r.
func FindOneDeviceByIdentifiers(r Interface, id *ttnpb.EndDeviceIdentifiers) (*Device, error) {
	devs, err := FindDeviceByIdentifiers(r, id)
	if err != nil {
		return nil, err
	}
	switch len(devs) {
	case 0:
		return nil, ErrDeviceNotFound.New(nil)
	case 1:
		return devs[0], nil
	default:
		return nil, ErrTooManyDevices.New(nil)
	}
}
