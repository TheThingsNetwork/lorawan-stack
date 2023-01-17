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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// Attribute is the attribute model in the database.
type Attribute struct {
	bun.BaseModel `bun:"table:attributes,alias:attr"`

	UUID

	// EntityType is "application", "client", "end_device", "gateway", "organization" or "user".
	EntityType string `bun:"entity_type,notnull"`
	// EntityID is Application.ID, Client.ID, EndDevice.ID, Gateway.ID, Organization.ID or User.ID.
	EntityID string `bun:"entity_id,notnull"`

	Key   string `bun:"key,notnull"`
	Value string `bun:"value,notnull"`
}

func (Attribute) _isModel() {} // It doesn't embed Model, but it's still a model.

// AttributeSlice is a slice of Attributes.
type AttributeSlice []*Attribute

func (a AttributeSlice) Len() int           { return len(a) }
func (a AttributeSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AttributeSlice) Less(i, j int) bool { return a[i].Key < a[j].Key }

func attributeSlice(attributes map[string]string, entityType, entityID string) []*Attribute {
	result := make([]*Attribute, 0, len(attributes))
	for k, v := range attributes {
		result = append(result, &Attribute{
			EntityType: entityType,
			EntityID:   entityID,
			Key:        k,
			Value:      v,
		})
	}
	return result
}

func attributeMap(attributes []*Attribute) map[string]*Attribute {
	m := make(map[string]*Attribute, len(attributes))
	for _, a := range attributes {
		m[a.Key] = a
	}
	return m
}

func (s *baseStore) replaceAttributes(
	ctx context.Context, current []*Attribute, desired map[string]string, entityType, entityID string,
) ([]*Attribute, error) {
	var (
		oldMap   = attributeMap(current)
		newMap   = attributeMap(attributeSlice(desired, entityType, entityID))
		toCreate = make([]*Attribute, 0, len(newMap))
		toUpdate = make([]*Attribute, 0, len(newMap))
		toDelete = make([]*Attribute, 0, len(oldMap))
		result   = make(AttributeSlice, 0, len(newMap))
	)

	for k, v := range newMap {
		// Ignore attributes that are not updated
		if current, ok := oldMap[k]; ok {
			delete(oldMap, k) // Don't need to delete this one.
			delete(newMap, k) // Don't need to create this one.
			if current.Value == v.Value {
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
			return nil, errors.WrapDriverError(err)
		}
	}

	if len(toUpdate) > 0 {
		_, err := s.DB.NewUpdate().
			Model(&toUpdate).
			Column("value").
			Bulk().
			Exec(ctx)
		if err != nil {
			return nil, errors.WrapDriverError(err)
		}
	}

	if len(toCreate) > 0 {
		_, err := s.DB.NewInsert().
			Model(&toCreate).
			Exec(ctx)
		if err != nil {
			return nil, errors.WrapDriverError(err)
		}
	}

	sort.Sort(result)

	return result, nil
}
