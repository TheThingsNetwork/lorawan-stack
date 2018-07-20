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
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Interface represents the interface exposed by the *Registry.
//
// Create stores device data and returns a new *Device.
// It may modify CreatedAt and UpdatedAt fields of ed and may return error if either of them is set to non-zero value on ed.
//
// Range calls f sequentially for each device stored, matching specified device fields.
// If f returns false, Range stops the iteration.
// If orderBy is set to non-empty string, it represents the fieldpath of the field, which the devices, that Range will iterate over will be sorted by.
// If count > 0, then Range will do it's best effort to iterate over at most count devices.
// If count == 0, then Range will iterate over all matching devices.
// Note, that Range provides no guarantees on the count of devices iterated over if count > 0 and
// it's caller's responsibility to handle cases where such are required.
// Range starts iteration at the index specified by the offset. Offset it 0-indexed.
// If len(fields) == 0, then Range uses all fields in ed to match devices.
type Interface interface {
	Create(ed *ttnpb.EndDevice, fields ...string) (*Device, error)
	Range(ed *ttnpb.EndDevice, orderBy string, count, offset uint64, f func(*Device) bool, fields ...string) (total uint64, err error)
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

// Create implements Interface.
// Create modifies CreatedAt and UpdatedAt fields of ed and returns error if either of them is set to non-zero value on ed.
func (r *Registry) Create(ed *ttnpb.EndDevice, fields ...string) (dev *Device, err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		latency.WithLabelValues("create").Observe(time.Since(start).Seconds())
	}(time.Now())

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

// Range implements Interface.
func (r *Registry) Range(ed *ttnpb.EndDevice, orderBy string, count, offset uint64, f func(*Device) bool, fields ...string) (total uint64, err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		duration := time.Since(start).Seconds()
		latency.WithLabelValues("range").Observe(duration)
		rangeLatency.WithLabelValues(strings.Join(fields, ",")).Observe(duration)
	}(time.Now())

	if ed == nil {
		return 0, errNilDevice
	}
	return r.store.Range(
		ed,
		func() interface{} { return &ttnpb.EndDevice{} },
		orderBy, count, offset,
		func(k store.PrimaryKey, v interface{}) bool {
			return f(newDevice(v.(*ttnpb.EndDevice), r.store, k))
		},
		fields...,
	)
}

// Identifiers supported in RangeByIdentifiers.
var Identifiers = []string{
	"EndDeviceIdentifiers.DeviceID",
	"EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID",
	"EndDeviceIdentifiers.DevEUI",
	"EndDeviceIdentifiers.JoinEUI",
	"EndDeviceIdentifiers.DevAddr",
}

// RangeByIdentifiers is a helper function, which allows ranging over r by matching identifiers instead of *ttnpb.EndDevice.
// See Interface documentation for more details.
func RangeByIdentifiers(r Interface, id *ttnpb.EndDeviceIdentifiers, orderBy string, count, offset uint64, f func(*Device) bool) (uint64, error) {
	if id == nil {
		return 0, errNilIdentifiers
	}
	fields := make([]string, 0, 5)
	if id.DeviceID != "" {
		fields = append(fields, "EndDeviceIdentifiers.DeviceID")
	}
	if id.ApplicationID != "" {
		fields = append(fields, "EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID")
	}
	if id.DevEUI != nil && !id.DevEUI.IsZero() {
		fields = append(fields, "EndDeviceIdentifiers.DevEUI")
	}
	if id.JoinEUI != nil && !id.JoinEUI.IsZero() {
		fields = append(fields, "EndDeviceIdentifiers.JoinEUI")
	}
	if id.DevAddr != nil && !id.DevAddr.IsZero() {
		fields = append(fields, "EndDeviceIdentifiers.DevAddr")
	}
	return r.Range(&ttnpb.EndDevice{EndDeviceIdentifiers: *id}, orderBy, count, offset, f, fields...)
}

// FindByIdentifiers searches for exactly one device matching specified device identifiers in r.
// See Interface documentation for more details.
func FindByIdentifiers(r Interface, id *ttnpb.EndDeviceIdentifiers) (*Device, error) {
	var dev *Device
	total, err := RangeByIdentifiers(r, id, "", 1, 0, func(d *Device) bool {
		dev = d
		return false
	})
	if err != nil {
		return nil, err
	}

	switch {
	case total == 0:
		return nil, errDeviceNotFound
	case total > 1:
		return nil, errTooManyDevices
	}
	return dev, nil
}
