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
	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Application model.
type Application struct {
	Model
	SoftDelete

	// BEGIN common fields
	ApplicationID string       `gorm:"unique_index:application_id_index;type:VARCHAR(36);not null"`
	Name          string       `gorm:"type:VARCHAR"`
	Description   string       `gorm:"type:TEXT"`
	Attributes    []Attribute  `gorm:"polymorphic:Entity;polymorphic_value:application"`
	APIKeys       []APIKey     `gorm:"polymorphic:Entity;polymorphic_value:application"`
	Memberships   []Membership `gorm:"polymorphic:Entity;polymorphic_value:application"`
	// END common fields
}

func init() {
	registerModel(&Application{})
}

// functions to set fields from the application model into the application proto.
var applicationPBSetters = map[string]func(*ttnpb.Application, *Application){
	nameField:        func(pb *ttnpb.Application, app *Application) { pb.Name = app.Name },
	descriptionField: func(pb *ttnpb.Application, app *Application) { pb.Description = app.Description },
	attributesField:  func(pb *ttnpb.Application, app *Application) { pb.Attributes = attributes(app.Attributes).toMap() },
}

// functions to set fields from the application proto into the application model.
var applicationModelSetters = map[string]func(*Application, *ttnpb.Application){
	nameField:        func(app *Application, pb *ttnpb.Application) { app.Name = pb.Name },
	descriptionField: func(app *Application, pb *ttnpb.Application) { app.Description = pb.Description },
	attributesField: func(app *Application, pb *ttnpb.Application) {
		app.Attributes = attributes(app.Attributes).updateFromMap(pb.Attributes)
	},
}

// fieldMask to use if a nil or empty fieldmask is passed.
var defaultApplicationFieldMask = &types.FieldMask{}

func init() {
	paths := make([]string, 0, len(applicationPBSetters))
	for _, path := range ttnpb.ApplicationFieldPathsNested {
		if _, ok := applicationPBSetters[path]; ok {
			paths = append(paths, path)
		}
	}
	defaultApplicationFieldMask.Paths = paths
}

// fieldmask path to column name in applications table.
var applicationColumnNames = map[string][]string{
	attributesField:  {},
	contactInfoField: {},
	nameField:        {nameField},
	descriptionField: {descriptionField},
}

func (app Application) toPB(pb *ttnpb.Application, fieldMask *types.FieldMask) {
	pb.ApplicationIdentifiers.ApplicationID = app.ApplicationID
	pb.CreatedAt = cleanTime(app.CreatedAt)
	pb.UpdatedAt = cleanTime(app.UpdatedAt)
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultApplicationFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := applicationPBSetters[path]; ok {
			setter(pb, &app)
		}
	}
}

func (app *Application) fromPB(pb *ttnpb.Application, fieldMask *types.FieldMask) (columns []string) {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		fieldMask = defaultApplicationFieldMask
	}
	for _, path := range fieldMask.Paths {
		if setter, ok := applicationModelSetters[path]; ok {
			setter(app, pb)
			if columnNames, ok := applicationColumnNames[path]; ok {
				columns = append(columns, columnNames...)
			}
			continue
		}
	}
	return
}
