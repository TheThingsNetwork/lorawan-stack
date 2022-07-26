// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import (
	"context"
	"sort"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// EndDeviceLocation is the end device location model in the database.
type EndDeviceLocation struct {
	bun.BaseModel `bun:"table:end_device_locations,alias:edloc"`

	Model

	EndDeviceID string     `bun:"end_device_id,notnull"`
	EndDevice   *EndDevice `bun:"rel:belongs-to,join:end_device_id=id"`

	Service string `bun:"service,nullzero"`

	Location

	Source int `bun:"source"`
}

// EndDeviceLocationSlice is a slice of EndDeviceLocation.
type EndDeviceLocationSlice []*EndDeviceLocation

func (a EndDeviceLocationSlice) Len() int      { return len(a) }
func (a EndDeviceLocationSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EndDeviceLocationSlice) Less(i, j int) bool {
	return a[i].Service < a[j].Service
}

func endDeviceLocationSliceFromPB(pbs map[string]*ttnpb.Location, endDeviceID string) []*EndDeviceLocation {
	out := make([]*EndDeviceLocation, 0, len(pbs))
	for service, pb := range pbs {
		out = append(out, &EndDeviceLocation{
			EndDeviceID: endDeviceID,
			Service:     service,
			Location:    locationFromPB(pb),
			Source:      int(pb.Source),
		})
	}
	return out
}

func endDeviceLocationMap(models []*EndDeviceLocation) map[string]*EndDeviceLocation {
	m := make(map[string]*EndDeviceLocation, len(models))
	for _, model := range models {
		m[model.Service] = model
	}
	return m
}

func (s *endDeviceStore) replaceEndDeviceLocations(
	ctx context.Context, current []*EndDeviceLocation, desired map[string]*ttnpb.Location, gatewayID string,
) ([]*EndDeviceLocation, error) {
	var (
		oldMap   = endDeviceLocationMap(current)
		newMap   = endDeviceLocationMap(endDeviceLocationSliceFromPB(desired, gatewayID))
		toCreate = make([]*EndDeviceLocation, 0, len(newMap))
		toUpdate = make([]*EndDeviceLocation, 0, len(newMap))
		toDelete = make([]*EndDeviceLocation, 0, len(oldMap))
		result   = make(EndDeviceLocationSlice, 0, len(newMap))
	)

	for k, v := range newMap {
		// Ignore end device location that has not been updated.
		if current, ok := oldMap[k]; ok {
			delete(oldMap, k) // Don't need to delete this one.
			delete(newMap, k) // Don't need to create this one.
			if current.Location == v.Location && current.Source == v.Source {
				result = append(result, v)
				continue // Don't need to update this one.
			}
			v.ID = current.ID
			toUpdate = append(toUpdate, v)
			result = append(result, v)
			continue
		}
		toCreate = append(toCreate, v)
		result = append(result, v)
	}
	for _, v := range oldMap {
		toDelete = append(toDelete, v)
	}

	if len(toDelete) > 0 {
		_, err := s.DB.NewDelete().
			Model(&toDelete).
			WherePK().
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
	}

	if len(toUpdate) > 0 {
		_, err := s.DB.NewUpdate().
			Model(&toUpdate).
			Column("latitude", "longitude", "altitude", "accuracy", "source").
			Bulk().
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
	}

	if len(toCreate) > 0 {
		_, err := s.DB.NewInsert().
			Model(&toCreate).
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
	}

	sort.Sort(result)

	return result, nil
}
