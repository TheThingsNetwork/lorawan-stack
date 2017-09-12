// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package helpers

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Location constructs a ttnpb.Location type with the provided arguments.
// It returns nil if all provided arguments are nil.
func Location(longitude, latitude *float32, altitude *int32) *ttnpb.Location {
	if longitude == nil && latitude == nil && altitude == nil {
		return nil
	}

	location := &ttnpb.Location{}
	if longitude != nil {
		location.Longitude = *longitude
	}
	if latitude != nil {
		location.Latitude = *latitude
	}
	if altitude != nil {
		location.Altitude = *altitude
	}

	return location
}
