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

import "github.com/jinzhu/gorm"

func replaceContactInfos(db *gorm.DB, entityType, entityUUID string, old []ContactInfo, new []ContactInfo) (err error) {
	oldByUUID := make(map[string]ContactInfo, len(old))
	for _, info := range old {
		oldByUUID[info.ID] = info
	}
	newByUUID := make(map[string]ContactInfo, len(new))
	for _, info := range new {
		if info.ID != "" {
			newByUUID[info.ID] = info
		}
	}
	var toCreate, toUpdate []ContactInfo
	for _, info := range new {
		if info.ID == "" {
			toCreate = append(toCreate, info)
			continue
		}
		if _, ok := oldByUUID[info.ID]; ok {
			toUpdate = append(toUpdate, info)
			continue
		}
	}
	var toDelete []string
	for _, info := range old {
		if _, ok := newByUUID[info.ID]; !ok {
			toDelete = append(toDelete, info.ID)
			continue
		}
	}
	for _, info := range toCreate {
		info.EntityType, info.EntityID = entityType, entityUUID
		err = db.Save(&info).Error
		if err != nil {
			return err
		}
	}
	for _, info := range toUpdate {
		err = db.Save(&info).Error
		if err != nil {
			return err
		}
	}
	if len(toDelete) > 0 {
		err = db.Where("id in (?)", toDelete).Delete(&ContactInfo{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
