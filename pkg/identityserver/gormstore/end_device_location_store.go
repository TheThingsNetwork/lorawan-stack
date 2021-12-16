// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"runtime/trace"

	"github.com/jinzhu/gorm"
)

func (s *store) replaceEndDeviceLocations(ctx context.Context, endDeviceUUID string, old []EndDeviceLocation, new []EndDeviceLocation) error {
	return replaceEndDeviceLocations(ctx, s.DB, endDeviceUUID, old, new)
}

func replaceEndDeviceLocations(ctx context.Context, db *gorm.DB, endDeviceUUID string, old []EndDeviceLocation, new []EndDeviceLocation) (err error) {
	defer trace.StartRegion(ctx, "update end device locations").End()
	oldByUUID := make(map[string]EndDeviceLocation, len(old))
	for _, loc := range old {
		oldByUUID[loc.ID] = loc
	}
	newByUUID := make(map[string]EndDeviceLocation, len(new))
	for _, loc := range new {
		if loc.ID != "" {
			newByUUID[loc.ID] = loc
		}
	}
	var toCreate, toUpdate []EndDeviceLocation
	for _, loc := range new {
		if loc.ID == "" {
			toCreate = append(toCreate, loc)
			continue
		}
		if _, ok := oldByUUID[loc.ID]; ok {
			toUpdate = append(toUpdate, loc)
			continue
		}
	}
	var toDelete []string
	for _, loc := range old {
		if _, ok := newByUUID[loc.ID]; !ok {
			toDelete = append(toDelete, loc.ID)
			continue
		}
	}
	for _, loc := range toCreate {
		loc.EndDeviceID = endDeviceUUID
		if err = db.Save(&loc).Error; err != nil {
			return err
		}
	}
	for _, loc := range toUpdate {
		if err = db.Save(&loc).Error; err != nil {
			return err
		}
	}
	if len(toDelete) > 0 {
		if err = db.Where("id in (?)", toDelete).Delete(&EndDeviceLocation{}).Error; err != nil {
			return err
		}
	}
	return nil
}
