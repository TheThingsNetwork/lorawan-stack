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
	"fmt"
	"sort"
	"time"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// ContactInfo is the contact info model in the database.
type ContactInfo struct {
	bun.BaseModel `bun:"table:contact_infos,alias:ci"`

	UUID

	// EntityType is "application", "client", "end_device", "gateway", "organization" or "user".
	EntityType string `bun:"entity_type,notnull"`
	// EntityID is Application.ID, Client.ID, EndDevice.ID, Gateway.ID, Organization.ID or User.ID.
	EntityID string `bun:"entity_id,notnull"`

	ContactType   int    `gorm:"contact_type,notnull"`
	ContactMethod int    `gorm:"contact_method,notnull"`
	Value         string `gorm:"value,notnull"`

	Public bool `bun:"public"`

	ValidatedAt *time.Time `bun:"validated_at"`
}

func contactInfoFromPB(pb *ttnpb.ContactInfo, entityType, entityID string) *ContactInfo {
	return &ContactInfo{
		EntityType:    entityType,
		EntityID:      entityID,
		ContactType:   int(pb.ContactType),
		ContactMethod: int(pb.ContactMethod),
		Value:         pb.Value,
		Public:        pb.Public,
		ValidatedAt:   ttnpb.StdTime(pb.ValidatedAt),
	}
}

func contactInfoToPB(m *ContactInfo) *ttnpb.ContactInfo {
	return &ttnpb.ContactInfo{
		ContactType:   ttnpb.ContactType(m.ContactType),
		ContactMethod: ttnpb.ContactMethod(m.ContactMethod),
		Value:         m.Value,
		Public:        m.Public,
		ValidatedAt:   ttnpb.ProtoTime(m.ValidatedAt),
	}
}

// ContactInfoSlice is a slice of ContactInfo.
type ContactInfoSlice []*ContactInfo

func (a ContactInfoSlice) Len() int      { return len(a) }
func (a ContactInfoSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ContactInfoSlice) Less(i, j int) bool {
	if a[i].Value < a[j].Value {
		return true
	}
	return a[i].ContactType < a[j].ContactType
}

func contactInfoSliceFromPB(pbs []*ttnpb.ContactInfo, entityType, entityID string) []*ContactInfo {
	out := make([]*ContactInfo, len(pbs))
	for i, pb := range pbs {
		out[i] = contactInfoFromPB(pb, entityType, entityID)
	}
	return out
}

func contactInfoMap(models []*ContactInfo) map[string]*ContactInfo {
	m := make(map[string]*ContactInfo, len(models))
	for _, model := range models {
		key := fmt.Sprintf("%d-%d-%s", model.ContactMethod, model.ContactType, model.Value)
		m[key] = model
	}
	return m
}

func (s *baseStore) replaceContactInfo(
	ctx context.Context, current []*ContactInfo, desired []*ttnpb.ContactInfo, entityType, entityID string,
) ([]*ContactInfo, error) {
	var (
		oldMap   = contactInfoMap(current)
		newMap   = contactInfoMap(contactInfoSliceFromPB(desired, entityType, entityID))
		toCreate = make([]*ContactInfo, 0, len(newMap))
		toUpdate = make([]*ContactInfo, 0, len(newMap))
		toDelete = make([]*ContactInfo, 0, len(oldMap))
		result   = make(ContactInfoSlice, 0, len(newMap))
	)

	for k, v := range newMap {
		// Ignore contact info that has not been updated.
		if current, ok := oldMap[k]; ok {
			delete(oldMap, k) // Don't need to delete this one.
			delete(newMap, k) // Don't need to create this one.
			if current.Public == v.Public && equalTime(current.ValidatedAt, v.ValidatedAt) {
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
			Column("public", "validated_at").
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
