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
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetApplicationStore returns an ApplicationStore on the given db (or transaction).
func GetApplicationStore(db *gorm.DB) ApplicationStore {
	return &applicationStore{db: db}
}

type applicationStore struct {
	db *gorm.DB
}

// selectApplicationFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectApplicationFields(query *gorm.DB, fieldMask *types.FieldMask) *gorm.DB {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query
	}
	var applicationColumns []string
	for _, path := range fieldMask.Paths {
		if column, ok := applicationColumnNames[path]; ok {
			applicationColumns = append(applicationColumns, column)
		} else {
			applicationColumns = append(applicationColumns, path)
		}
	}
	return query.Select(append(append(modelColumns, "application_id"), applicationColumns...)) // TODO: remove possible duplicate application_id
}

func (s *applicationStore) CreateApplication(ctx context.Context, app *ttnpb.Application) (*ttnpb.Application, error) {
	appModel := Application{
		ApplicationID: app.ApplicationID, // The ID is not mutated by fromPB.
	}
	appModel.fromPB(app, nil)
	appModel.SetContext(ctx)
	query := s.db.Create(&appModel)
	if query.Error != nil {
		return nil, query.Error
	}
	var appProto ttnpb.Application
	appModel.toPB(&appProto, nil)
	return &appProto, nil
}

func (s *applicationStore) FindApplications(ctx context.Context, ids []*ttnpb.ApplicationIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Application, error) {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetApplicationID()
	}
	query := s.db.Scopes(withContext(ctx), withApplicationID(idStrings...))
	query = selectApplicationFields(query, fieldMask)
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(&Application{}))
		query = query.Limit(limit).Offset(offset)
	}
	var appModels []Application
	query = query.Find(&appModels)
	setTotal(ctx, uint64(len(appModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	appProtos := make([]*ttnpb.Application, len(appModels))
	for i, appModel := range appModels {
		appProto := new(ttnpb.Application)
		appModel.toPB(appProto, nil)
		appProtos[i] = appProto
	}
	return appProtos, nil
}

func (s *applicationStore) GetApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Application, error) {
	query := s.db.Scopes(withContext(ctx), withApplicationID(id.GetApplicationID()))
	query = selectApplicationFields(query, fieldMask)
	var appModel Application
	err := query.First(&appModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id.EntityIdentifiers())
		}
		return nil, err
	}
	appProto := new(ttnpb.Application)
	appModel.toPB(appProto, nil)
	return appProto, nil
}

func (s *applicationStore) UpdateApplication(ctx context.Context, app *ttnpb.Application, fieldMask *types.FieldMask) (updated *ttnpb.Application, err error) {
	query := s.db.Scopes(withContext(ctx), withApplicationID(app.GetApplicationID()))
	query = selectApplicationFields(query, fieldMask)
	var appModel Application
	err = query.First(&appModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(app.ApplicationIdentifiers.EntityIdentifiers())
		}
		return nil, err
	}
	if !app.UpdatedAt.IsZero() && app.UpdatedAt != appModel.UpdatedAt {
		return nil, errConcurrentWrite
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	columns := appModel.fromPB(app, fieldMask)
	if len(columns) > 0 {
		query = s.db.Model(&appModel).Select(columns).Updates(&appModel)
		if query.Error != nil {
			return nil, query.Error
		}
	}
	updated = new(ttnpb.Application)
	appModel.toPB(updated, nil)
	return updated, nil
}

func (s *applicationStore) DeleteApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error {
	return deleteEntity(ctx, s.db, id.EntityIdentifiers())
}
