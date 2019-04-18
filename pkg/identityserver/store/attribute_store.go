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

func (s *store) replaceAttributes(ctx context.Context, entityType, entityUUID string, old []Attribute, new []Attribute) (err error) {
	return replaceAttributes(ctx, s.DB, entityType, entityUUID, old, new)
}

func replaceAttributes(ctx context.Context, db *gorm.DB, entityType, entityUUID string, old []Attribute, new []Attribute) (err error) {
	defer trace.StartRegion(ctx, "update attributes").End()
	oldByUUID := make(map[string]Attribute, len(old))
	for _, attr := range old {
		oldByUUID[attr.ID] = attr
	}
	newByUUID := make(map[string]Attribute, len(new))
	for _, attr := range new {
		if attr.ID != "" {
			newByUUID[attr.ID] = attr
		}
	}
	var toCreate, toUpdate []Attribute
	for _, attr := range new {
		if attr.ID == "" {
			toCreate = append(toCreate, attr)
			continue
		}
		if _, ok := oldByUUID[attr.ID]; ok {
			toUpdate = append(toUpdate, attr)
			continue
		}
	}
	var toDelete []string
	for _, attr := range old {
		if _, ok := newByUUID[attr.ID]; !ok {
			toDelete = append(toDelete, attr.ID)
			continue
		}
	}
	for _, attr := range toCreate {
		attr.EntityType, attr.EntityID = entityType, entityUUID
		if err = db.Save(&attr).Error; err != nil {
			return err
		}
	}
	for _, attr := range toUpdate {
		if err = db.Save(&attr).Error; err != nil {
			return err
		}
	}
	if len(toDelete) > 0 {
		if err = db.Where("id in (?)", toDelete).Delete(&Attribute{}).Error; err != nil {
			return err
		}
	}
	return nil
}
