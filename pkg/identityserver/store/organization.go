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

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Organization model.
type Organization struct {
	Model
	SoftDelete

	Account Account `gorm:"polymorphic:Account;polymorphic_value:organization"`

	// BEGIN common fields
	Name        string       `gorm:"type:VARCHAR"`
	Description string       `gorm:"type:TEXT"`
	Attributes  []Attribute  `gorm:"polymorphic:Entity;polymorphic_value:organization"`
	APIKeys     []APIKey     `gorm:"polymorphic:Entity;polymorphic_value:organization"`
	Memberships []Membership `gorm:"polymorphic:Entity;polymorphic_value:organization"`
	// END common fields
}

func init() {
	registerModel(&Organization{})
}

// SetContext sets the context on the organization model and the embedded account model.
func (org *Organization) SetContext(ctx context.Context) {
	org.Model.SetContext(ctx)
	org.Account.Model.SetContext(ctx)
}

// functions to set fields from the organization model into the organization proto.
var organizationPBSetters = map[string]func(*ttnpb.Organization, *Organization){
	nameField:        func(pb *ttnpb.Organization, org *Organization) { pb.Name = org.Name },
	descriptionField: func(pb *ttnpb.Organization, org *Organization) { pb.Description = org.Description },
	attributesField:  func(pb *ttnpb.Organization, org *Organization) { pb.Attributes = attributes(org.Attributes).toMap() },
}

// functions to set fields from the organization proto into the organization model.
var organizationModelSetters = map[string]func(*Organization, *ttnpb.Organization){
	nameField:        func(org *Organization, pb *ttnpb.Organization) { org.Name = pb.Name },
	descriptionField: func(org *Organization, pb *ttnpb.Organization) { org.Description = pb.Description },
	attributesField: func(org *Organization, pb *ttnpb.Organization) {
		org.Attributes = attributes(org.Attributes).updateFromMap(pb.Attributes)
	},
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultOrganizationFieldMask = &types.FieldMask{}

func init() {
	paths := make([]string, 0, len(organizationPBSetters))
	for path := range organizationPBSetters {
		paths = append(paths, path)
	}
	defaultOrganizationFieldMask.Paths = paths
}

// fieldmask path to column name in organizations table.
var organizationColumnNames = map[string][]string{
	attributesField:  {},
	contactInfoField: {},
	nameField:        {nameField},
	descriptionField: {descriptionField},
}

func (org Organization) toPB(pb *ttnpb.Organization, fieldMask *types.FieldMask) {
	pb.OrganizationIdentifiers.OrganizationID = org.Account.UID
	pb.CreatedAt = cleanTime(org.CreatedAt)
	pb.UpdatedAt = cleanTime(org.UpdatedAt)
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultOrganizationFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := organizationPBSetters[path]; ok {
			setter(pb, &org)
		}
	}
}

func (org *Organization) fromPB(pb *ttnpb.Organization, fieldMask *types.FieldMask) (columns []string) {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultOrganizationFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := organizationModelSetters[path]; ok {
			setter(org, pb)
			if columnNames, ok := organizationColumnNames[path]; ok {
				columns = append(columns, columnNames...)
			}
			continue
		}
	}
	return
}
