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

package deviceregistry

import (
	"time"

	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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

// Store updates devices data in the underlying store.Interface.
// It modifies the UpdatedAt field of d.EndDevice.
func (d *Device) Store(fields ...string) (err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		latency.WithLabelValues("update").Observe(time.Since(start).Seconds())
	}(time.Now())

	d.EndDevice.UpdatedAt = time.Now().UTC()
	if len(fields) != 0 {
		fields = append(fields, "UpdatedAt")
	}
	return d.store.Update(d.key, d.EndDevice, fields...)
}

// Load returns a snapshot of current device data in underlying store.Interface.
func (d *Device) Load() (dev *Device, err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		latency.WithLabelValues("load").Observe(time.Since(start).Seconds())
	}(time.Now())

	pb := &ttnpb.EndDevice{}
	if err := d.store.Find(d.key, pb); err != nil {
		return nil, err
	}
	return newDevice(pb, d.store, d.key), nil
}

// Delete removes device from the underlying store.Interface.
func (d *Device) Delete() (err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		latency.WithLabelValues("delete").Observe(time.Since(start).Seconds())
	}(time.Now())

	return d.store.Delete(d.key)
}
