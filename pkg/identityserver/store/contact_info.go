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

import (
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// ContactInfo model.
type ContactInfo struct {
	ID string `gorm:"type:UUID;primary_key;default:gen_random_uuid()"`

	ContactType   int    `gorm:"not_null"`
	ContactMethod int    `gorm:"not_null"`
	Value         string `gorm:"type:VARCHAR"`

	Public bool

	ValidatedAt *time.Time

	EntityID   string `gorm:"type:UUID;index:entity"`
	EntityType string `gorm:"index:entity"`
}

func init() {
	registerModel(&ContactInfo{})
}

type contactInfos []ContactInfo

func (c contactInfos) toPB() []*ttnpb.ContactInfo {
	pb := make([]*ttnpb.ContactInfo, len(c))
	for i, info := range c {
		pb[i] = &ttnpb.ContactInfo{
			ContactType:   ttnpb.ContactType(info.ContactType),
			ContactMethod: ttnpb.ContactMethod(info.ContactMethod),
			Value:         info.Value,
			Public:        info.Public,
			ValidatedAt:   cleanTimePtr(info.ValidatedAt),
		}
	}
	return pb
}

func (c contactInfos) updateFromPB(pb []*ttnpb.ContactInfo) contactInfos {
	type contactInfoID struct {
		contactType   int
		contactMethod int
		value         string
	}
	type contactInfo struct {
		ContactInfo
		deleted bool
	}
	contactInfos := make(map[contactInfoID]*contactInfo)
	validatedInfos := make(map[contactInfoID]time.Time)
	for _, existing := range c {
		contactInfos[contactInfoID{existing.ContactType, existing.ContactMethod, existing.Value}] = &contactInfo{
			ContactInfo: existing,
			deleted:     true,
		}
		if existing.ValidatedAt != nil { // Mark this contact as validated.
			id := contactInfoID{-1, existing.ContactMethod, existing.Value}
			if existing.ValidatedAt.After(validatedInfos[id]) {
				validatedInfos[id] = cleanTime(*existing.ValidatedAt)
			}
		}
	}
	var updated []ContactInfo
	for _, new := range pb {
		if existing, ok := contactInfos[contactInfoID{int(new.ContactType), int(new.ContactMethod), new.Value}]; ok {
			existing.deleted = false
			existing.Public = new.Public
			existing.ValidatedAt = cleanTimePtr(new.ValidatedAt)
		} else {
			info := ContactInfo{
				ContactType:   int(new.ContactType),
				ContactMethod: int(new.ContactMethod),
				Value:         new.Value,
				Public:        new.Public,
				ValidatedAt:   cleanTimePtr(new.ValidatedAt),
			}
			if new.ValidatedAt == nil { // See if this contact was previously validated.
				if validated := validatedInfos[contactInfoID{-1, int(new.ContactMethod), new.Value}]; !validated.IsZero() {
					info.ValidatedAt = &validated
				}
			}
			updated = append(updated, info)
		}
	}
	for _, existing := range contactInfos {
		if existing.deleted {
			continue
		}
		updated = append(updated, existing.ContactInfo)
	}
	return updated
}
