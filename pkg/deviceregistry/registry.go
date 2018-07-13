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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Interface represents the interface exposed by the *Registry.
type Interface interface {
	Create(ed *ttnpb.EndDevice, fields ...string) (*Device, error)
	Range(ed *ttnpb.EndDevice, batchSize uint64, f func(*Device) bool, fields ...string) error
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
	start := time.Now()

	now := start.UTC()
	ed.CreatedAt = now
	ed.UpdatedAt = now

	if len(fields) != 0 {
		fields = append(fields, "CreatedAt", "UpdatedAt")
	}

	id, err := r.store.Create(ed, fields...)
	if err != nil {
		return nil, err
	}

	dev := newDevice(ed, r.store, id)

	latency.WithLabelValues("create").Observe(time.Since(start).Seconds())

	return dev, nil
}

// Range calls f sequentially for each device stored, matching specified device fields.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Registry's
// contents: no device will be visited more than once, but if the device is
// created or deleted concurrently, Range may or may not call f on that device.
//
// If batchSize argument is non-zero, Range will retrieve devices
// from the underlying store in chunks of (approximately) batchSize devices.
//
// If len(fields) == 0, then Range uses all fields in ed to match devices.
func (r *Registry) Range(ed *ttnpb.EndDevice, batchSize uint64, f func(*Device) bool, fields ...string) error {
	if ed == nil {
		return errors.New("Device specified is nil")
	}
	start := time.Now()
	err := r.store.Range(
		ed,
		func() interface{} { return &ttnpb.EndDevice{} },
		batchSize,
		func(k store.PrimaryKey, v interface{}) bool {
			return f(newDevice(v.(*ttnpb.EndDevice), r.store, k))
		},
		fields...,
	)
	duration := time.Since(start).Seconds()
	latency.WithLabelValues("range").Observe(duration)
	rangeLatency.WithLabelValues(strings.Join(fields, ",")).Observe(duration)
	return err
}

// Identifiers supported in RangeByIdentifiers.
var Identifiers = []string{
	"EndDeviceIdentifiers.DeviceID",
	"EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID",
	"EndDeviceIdentifiers.DevEUI",
	"EndDeviceIdentifiers.JoinEUI",
	"EndDeviceIdentifiers.DevAddr",
}

// RangeByIdentifiers calls f sequentially for each device stored in r, matching specified device identifiers.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Intefaces's
// contents: no device will be visited more than once, but if the device is
// created or deleted concurrently, Range may or may not call f on that device.
//
// If batchSize argument is non-zero, Range will retrieve devices
// from the underlying store in chunks of (approximately) batchSize devices.
func RangeByIdentifiers(r Interface, id *ttnpb.EndDeviceIdentifiers, batchSize uint64, f func(*Device) bool) error {
	if id == nil {
		return errors.New("Identifiers specified are nil")
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
	return r.Range(&ttnpb.EndDevice{EndDeviceIdentifiers: *id}, batchSize, f, fields...)
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
		return nil, errDeviceNotFound
	case 1:
		return dev, nil
	default:
		return nil, errTooManyDevices
	}
}
