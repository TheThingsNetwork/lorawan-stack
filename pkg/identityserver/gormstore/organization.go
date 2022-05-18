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
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Organization model.
type Organization struct {
	Model
	SoftDelete

	Account Account `gorm:"polymorphic:Account;polymorphic_value:organization"`

	// BEGIN common fields
	Name        string `gorm:"type:VARCHAR"`
	Description string `gorm:"type:TEXT"`

	Attributes  []Attribute  `gorm:"polymorphic:Entity;polymorphic_value:organization"`
	APIKeys     []APIKey     `gorm:"polymorphic:Entity;polymorphic_value:organization"`
	Memberships []Membership `gorm:"polymorphic:Entity;polymorphic_value:organization"`

	AdministrativeContactID *string  `gorm:"type:UUID;index"`
	AdministrativeContact   *Account `gorm:"save_associations:false"`

	TechnicalContactID *string  `gorm:"type:UUID;index"`
	TechnicalContact   *Account `gorm:"save_associations:false"`
	// END common fields
}

func init() {
	registerModel(&Organization{})
}

// functions to set fields from the organization model into the organization proto.
var organizationPBSetters = map[string]func(*ttnpb.Organization, *Organization){
	nameField: func(pb *ttnpb.Organization, org *Organization) {
		pb.Name = org.Name
	},
	descriptionField: func(pb *ttnpb.Organization, org *Organization) {
		pb.Description = org.Description
	},
	attributesField: func(pb *ttnpb.Organization, org *Organization) {
		pb.Attributes = attributes(org.Attributes).toMap()
	},
	administrativeContactField: func(pb *ttnpb.Organization, org *Organization) {
		if org.AdministrativeContact != nil {
			pb.AdministrativeContact = org.AdministrativeContact.OrganizationOrUserIdentifiers()
		}
	},
	technicalContactField: func(pb *ttnpb.Organization, org *Organization) {
		if org.TechnicalContact != nil {
			pb.TechnicalContact = org.TechnicalContact.OrganizationOrUserIdentifiers()
		}
	},
}

// functions to set fields from the organization proto into the organization model.
var organizationModelSetters = map[string]func(*Organization, *ttnpb.Organization){
	nameField: func(org *Organization, pb *ttnpb.Organization) {
		org.Name = pb.Name
	},
	descriptionField: func(org *Organization, pb *ttnpb.Organization) {
		org.Description = pb.Description
	},
	attributesField: func(org *Organization, pb *ttnpb.Organization) {
		org.Attributes = attributes(org.Attributes).updateFromMap(pb.Attributes)
	},
	administrativeContactField: func(org *Organization, pb *ttnpb.Organization) {
		if pb.AdministrativeContact == nil {
			org.AdministrativeContact = nil
			return
		}
		org.AdministrativeContact = &Account{
			AccountType: pb.AdministrativeContact.EntityType(),
			UID:         pb.AdministrativeContact.IDString(),
		}
	},
	technicalContactField: func(org *Organization, pb *ttnpb.Organization) {
		if pb.TechnicalContact == nil {
			org.TechnicalContact = nil
			return
		}
		org.TechnicalContact = &Account{
			AccountType: pb.TechnicalContact.EntityType(),
			UID:         pb.TechnicalContact.IDString(),
		}
	},
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultOrganizationFieldMask store.FieldMask

func init() {
	paths := make([]string, 0, len(organizationPBSetters))
	for _, path := range ttnpb.OrganizationFieldPathsNested {
		if _, ok := organizationPBSetters[path]; ok {
			paths = append(paths, path)
		}
	}
	defaultOrganizationFieldMask = paths
}

// fieldmask path to column name in organizations table.
var organizationColumnNames = map[string][]string{
	attributesField:            {},
	contactInfoField:           {},
	nameField:                  {nameField},
	descriptionField:           {descriptionField},
	administrativeContactField: {administrativeContactField + "_id"},
	technicalContactField:      {technicalContactField + "_id"},
}

func (org Organization) toPB(pb *ttnpb.Organization, fieldMask store.FieldMask) {
	pb.Ids = &ttnpb.OrganizationIdentifiers{OrganizationId: org.Account.UID}
	pb.CreatedAt = ttnpb.ProtoTimePtr(cleanTime(org.CreatedAt))
	pb.UpdatedAt = ttnpb.ProtoTimePtr(cleanTime(org.UpdatedAt))
	pb.DeletedAt = ttnpb.ProtoTime(cleanTimePtr(org.DeletedAt))
	if len(fieldMask) == 0 {
		fieldMask = defaultOrganizationFieldMask
	}
	for _, path := range fieldMask {
		if setter, ok := organizationPBSetters[path]; ok {
			setter(pb, &org)
		}
	}
}

func (org *Organization) fromPB(pb *ttnpb.Organization, fieldMask store.FieldMask) (columns []string) {
	if len(fieldMask) == 0 {
		fieldMask = defaultOrganizationFieldMask
	}
	for _, path := range fieldMask {
		if setter, ok := organizationModelSetters[path]; ok {
			setter(org, pb)
			if columnNames, ok := organizationColumnNames[path]; ok {
				columns = append(columns, columnNames...)
			}
			continue
		}
	}
	return columns
}

type organizationWithUID struct {
	UID          string
	Organization `gorm:"embedded"`
}

func (organizationWithUID) TableName() string { return "organizations" }

func (u organizationWithUID) toPB(pb *ttnpb.Organization, fieldMask store.FieldMask) {
	u.Organization.Account.UID = u.UID
	u.Organization.toPB(pb, fieldMask)
}
