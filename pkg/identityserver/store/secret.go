// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

// Secret model.
type Secret struct {
	ID string `gorm:"type:UUID;primary_key;default:gen_random_uuid()"`

	EntityID   string `gorm:"type:UUID;index:secret_entity_index;not null"`
	EntityType string `gorm:"type:VARCHAR(32);index:secret_entity_index;not null"`

	Key string `gorm:"type:VARCHAR"`

	// This value is marshaled to/from ttnpb.Secret.
	CipherData []byte `gorm:"type:BYTEA"`
}

func init() {
	registerModel(&Attribute{})
}

type secrets []Secret

func (s secrets) toMap() map[string]*ttnpb.Secret {
	pbSecrets := make(map[string]*ttnpb.Secret, len(s))
	for _, secret := range s {
		pbSecret := ttnpb.Secret{}
		_ = pbSecret.Unmarshal(secret.CipherData)
		pbSecrets[secret.Key] = &pbSecret
	}
	return pbSecrets
}

func (s secrets) updateFromMap(m map[string]*ttnpb.Secret) secrets {
	type secret struct {
		Secret
		deleted bool
	}
	secrets := make(map[string]*secret)
	for _, existing := range s {
		secrets[existing.Key] = &secret{
			Secret:  existing,
			deleted: true,
		}
	}
	var updated []Secret
	for k, v := range m {
		marshaled, _ := v.Marshal()
		if existing, ok := secrets[k]; ok {
			existing.deleted = false
			existing.CipherData = marshaled
		} else {
			updated = append(updated, Secret{Key: k, CipherData: marshaled})
		}
	}
	for _, existing := range secrets {
		if existing.deleted {
			continue
		}
		updated = append(updated, existing.Secret)
	}
	return updated
}
