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

// Package deviceregistry contains the implementation of a device registry service.
package deviceregistry

import (
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Interface represents the interface exposed by the *Registry.
type Interface interface {
	Create(ed *ttnpb.EndDevice, fields ...string) (*Device, error)
	Range(ed *ttnpb.EndDevice, count uint64, f func(*Device) bool, fields ...string) error
}

var _ Interface = &Registry{}

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
// It modifies CreatedAt and UpdatedAt fields of ed and returns error if either of them is non-zero on ed.
func (r *Registry) Create(ed *ttnpb.EndDevice, fields ...string) (*Device, error) {
	now := time.Now().UTC()
	ed.CreatedAt = now
	ed.UpdatedAt = now

	if len(fields) != 0 {
		fields = append(fields, "CreatedAt", "UpdatedAt")
	}

	id, err := r.store.Create(ed, fields...)
	if err != nil {
		return nil, err
	}
	return newDevice(ed, r.store, id), nil
}

// Range ranges over devices matching specified device in underlying store.Interface.
func (r *Registry) Range(ed *ttnpb.EndDevice, count uint64, f func(*Device) bool, fields ...string) error {
	if ed == nil {
		return errors.New("Device specified is nil")
	}
	return r.store.Range(
		ed,
		func() interface{} { return &ttnpb.EndDevice{} },
		count,
		func(k store.PrimaryKey, v interface{}) bool {
			return f(newDevice(v.(*ttnpb.EndDevice), r.store, k))
		},
		fields...,
	)
}

// RangeByIdentifiers ranges over devices matching specified device identifiers in r.
func RangeByIdentifiers(r Interface, id *ttnpb.EndDeviceIdentifiers, count uint64, f func(*Device) bool) error {
	if id == nil {
		return errors.New("Identifiers specified are nil")
	}

	fields := make([]string, 0, 5)
	switch {
	case id.DeviceID != "":
		fields = append(fields, "EndDeviceIdentifiers.DeviceID")
	case id.ApplicationID != "":
		fields = append(fields, "EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID")
	case id.DevEUI != nil && !id.DevEUI.IsZero():
		fields = append(fields, "EndDeviceIdentifiers.DevEUI")
	case id.JoinEUI != nil && !id.JoinEUI.IsZero():
		fields = append(fields, "EndDeviceIdentifiers.JoinEUI")
	case id.DevAddr != nil && !id.DevAddr.IsZero():
		fields = append(fields, "EndDeviceIdentifiers.DevAddr")
	}
	return r.Range(&ttnpb.EndDevice{EndDeviceIdentifiers: *id}, count, f, fields...)
}

// FindByIdentifiers searches for exactly one device matching specified device identifiers in r.
func FindByIdentifiers(r Interface, id *ttnpb.EndDeviceIdentifiers) (*Device, error) {
	var dev *Device
	var i uint64
	err := RangeByIdentifiers(r, id, 1, func(d *Device) bool {
		i++
		if i > 1 {
			return false
		}
		dev = d
		return true
	})
	if err != nil {
		return nil, err
	}
	switch i {
	case 0:
		return nil, ErrDeviceNotFound.New(nil)
	case 1:
		return dev, nil
	default:
		return nil, ErrTooManyDevices.New(nil)
	}
}
