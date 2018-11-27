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

// Attribute model.
type Attribute struct {
	ID string `gorm:"type:UUID;primary_key;default:gen_random_uuid()"`

	EntityID   string `gorm:"type:UUID;index:entity"`
	EntityType string `gorm:"index:entity"`

	Key   string `gorm:"type:VARCHAR"`
	Value string `gorm:"type:VARCHAR"`
}

func init() {
	registerModel(&Attribute{})
}

type attributes []Attribute

func (a attributes) toMap() map[string]string {
	attributes := make(map[string]string, len(a))
	for _, attr := range a {
		attributes[attr.Key] = attr.Value
	}
	return attributes
}

func (a attributes) updateFromMap(m map[string]string) attributes {
	type attribute struct {
		Attribute
		deleted bool
	}
	attributes := make(map[string]*attribute)
	for _, existing := range a {
		attributes[existing.Key] = &attribute{
			Attribute: existing,
			deleted:   true,
		}
	}
	var updated []Attribute
	for k, v := range m {
		if existing, ok := attributes[k]; ok {
			existing.deleted = false
			existing.Value = v
		} else {
			updated = append(updated, Attribute{Key: k, Value: v})
		}
	}
	for _, existing := range attributes {
		if existing.deleted {
			continue
		}
		updated = append(updated, existing.Attribute)
	}
	return updated
}
