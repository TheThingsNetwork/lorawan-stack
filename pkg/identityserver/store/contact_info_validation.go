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
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// ContactInfoValidation model.
type ContactInfoValidation struct {
	Model

	Reference string `gorm:"type:VARCHAR;index:contact_info_validation_id_index"`
	Token     string `gorm:"type:VARCHAR;index:contact_info_validation_id_index"`

	EntityID   string `gorm:"type:UUID;index:contact_info_validation_entity_index;not null"`
	EntityType string `gorm:"type:VARCHAR(32);index:contact_info_validation_entity_index;not null"`

	ContactMethod int    `gorm:"not null"`
	Value         string `gorm:"type:VARCHAR"`

	ExpiresAt time.Time
}

func init() {
	registerModel(&ContactInfoValidation{})
}

func (c ContactInfoValidation) toPB() *ttnpb.ContactInfoValidation {
	return &ttnpb.ContactInfoValidation{
		ID:        c.Reference,
		Token:     c.Token,
		CreatedAt: &c.CreatedAt,
		ExpiresAt: &c.ExpiresAt,
	}
}

func (c *ContactInfoValidation) fromPB(pb *ttnpb.ContactInfoValidation) {
	c.Reference = pb.ID
	c.Token = pb.Token
	if pb.CreatedAt != nil {
		c.CreatedAt = cleanTime(*pb.CreatedAt)
	}
	if pb.ExpiresAt != nil {
		c.ExpiresAt = cleanTime(*pb.ExpiresAt)
	}
}
