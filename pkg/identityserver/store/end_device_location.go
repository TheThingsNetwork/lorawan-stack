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

package store

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

// EndDeviceLocation model.
type EndDeviceLocation struct {
	Model

	EndDeviceID string `gorm:"type:UUID;unique_index:id;index;not null"`
	Service     string `gorm:"unique_index:id"`

	Location

	Source int `gorm:"not null"`
}

func init() {
	registerModel(&EndDeviceLocation{})
}

func (l EndDeviceLocation) toPB() *ttnpb.Location {
	return &ttnpb.Location{
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		Altitude:  l.Altitude,
		Accuracy:  l.Accuracy,
		Source:    ttnpb.LocationSource(l.Source),
	}
}

func (l *EndDeviceLocation) fromPB(pb *ttnpb.Location) {
	l.Latitude = pb.Latitude
	l.Longitude = pb.Longitude
	l.Altitude = pb.Altitude
	l.Accuracy = pb.Accuracy
	l.Source = int(pb.Source)
}

type deviceLocations []EndDeviceLocation

func (a deviceLocations) toMap() map[string]*ttnpb.Location {
	deviceLocations := make(map[string]*ttnpb.Location, len(a))
	for _, loc := range a {
		deviceLocations[loc.Service] = loc.toPB()
	}
	return deviceLocations
}

func (a deviceLocations) updateFromMap(m map[string]*ttnpb.Location) deviceLocations {
	type deviceLocation struct {
		EndDeviceLocation
		deleted bool
	}
	deviceLocations := make(map[string]*deviceLocation)
	for _, existing := range a {
		deviceLocations[existing.Service] = &deviceLocation{
			EndDeviceLocation: existing,
			deleted:           true,
		}
	}
	var updated []EndDeviceLocation
	for k, v := range m {
		if existing, ok := deviceLocations[k]; ok {
			existing.deleted = false
			existing.fromPB(v)
		} else {
			loc := EndDeviceLocation{Service: k}
			loc.fromPB(v)
			updated = append(updated, loc)
		}
	}
	for _, existing := range deviceLocations {
		if existing.deleted {
			continue
		}
		updated = append(updated, existing.EndDeviceLocation)
	}
	return updated
}
