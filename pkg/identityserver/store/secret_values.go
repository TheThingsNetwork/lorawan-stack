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

// SecretValue model.
type SecretValue struct {
	ID    string `gorm:"type:UUID;primary_key;default:gen_random_uuid()"`
	Key   string `gorm:"type:VARCHAR"`
	Value []byte `gorm:"type:BYTEA"`
}

func init() {
	registerModel(&SecretValue{})
}

type secretValues []SecretValue

func (v secretValues) toMap() map[string][]byte {
	secretValues := make(map[string][]byte, len(v))
	for _, secretValue := range v {
		secretValues[secretValue.Key] = secretValue.Value
	}
	return secretValues
}

func (v secretValues) updateFromMap(m map[string][]byte) secretValues {
	type secretValues struct {
		SecretValue
		deleted bool
	}
	sv := make(map[string]*secretValues)
	for _, existing := range v {
		sv[existing.Key] = &secretValues{
			SecretValue: existing,
			deleted:     true,
		}
	}
	var updated []SecretValue
	for k, v := range m {
		if existing, ok := sv[k]; ok {
			existing.deleted = false
			existing.Value = v
		} else {
			updated = append(updated, SecretValue{Key: k, Value: v})
		}
	}
	for _, existing := range sv {
		if existing.deleted {
			continue
		}
		updated = append(updated, existing.SecretValue)
	}
	return updated
}
