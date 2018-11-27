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

// GetUserStore returns an UserStore on the given db (or transaction).
func GetUserStore(db *gorm.DB) UserStore {
	return &userStore{db: db}
}

type userStore struct {
	db *gorm.DB
}

// selectUserFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectUserFields(query *gorm.DB, fieldMask *types.FieldMask) *gorm.DB {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query.Select([]string{"users.*", "accounts.uid"})
	}
	var userColumns []string
	for _, column := range modelColumns {
		userColumns = append(userColumns, "users."+column)
	}
	for _, path := range fieldMask.Paths {
		switch path {
		case "ids.user_id":
			// accounts.uid is always selected
		default:
			if column, ok := userColumnNames[path]; ok {
				userColumns = append(userColumns, column)
			} else {
				userColumns = append(userColumns, path)
			}
		}
	}
	return query.Select(append(userColumns, "accounts.uid"))
}

func (s *userStore) CreateUser(ctx context.Context, usr *ttnpb.User) (*ttnpb.User, error) {
	userModel := User{
		Account: Account{UID: usr.UserID}, // The ID is not mutated by fromPB.
	}
	fieldMask := &types.FieldMask{Paths: append(defaultUserFieldMask.Paths, passwordField)}
	userModel.fromPB(usr, fieldMask)
	userModel.SetContext(ctx)
	query := s.db.Create(&userModel)
	if query.Error != nil {
		return nil, query.Error
	}
	var userProto ttnpb.User
	userModel.toPB(&userProto, nil)
	return &userProto, nil
}

func (s *userStore) FindUsers(ctx context.Context, ids []*ttnpb.UserIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.User, error) {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetUserID()
	}
	query := s.db.Scopes(withContext(ctx), withUserID(idStrings...))
	query = selectUserFields(query, fieldMask)
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(&User{}))
		query = query.Limit(limit).Offset(offset)
	}
	var userModels []User
	query = query.Preload("Account").Find(&userModels)
	setTotal(ctx, uint64(len(userModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	userProtos := make([]*ttnpb.User, len(userModels))
	for i, userModel := range userModels {
		userProto := new(ttnpb.User)
		userModel.toPB(userProto, nil)
		userProtos[i] = userProto
	}
	return userProtos, nil
}

func (s *userStore) GetUser(ctx context.Context, id *ttnpb.UserIdentifiers, fieldMask *types.FieldMask) (*ttnpb.User, error) {
	query := s.db.Scopes(withContext(ctx), withUserID(id.GetUserID()))
	query = selectUserFields(query, fieldMask)
	var userModel User
	err := query.Preload("Account").First(&userModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id.EntityIdentifiers())
		}
		return nil, err
	}
	userProto := new(ttnpb.User)
	userModel.toPB(userProto, nil)
	return userProto, nil
}

func (s *userStore) UpdateUser(ctx context.Context, usr *ttnpb.User, fieldMask *types.FieldMask) (updated *ttnpb.User, err error) {
	query := s.db.Scopes(withContext(ctx), withUserID(usr.GetUserID()))
	query = selectUserFields(query, fieldMask)
	var userModel User
	err = query.First(&userModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(usr.UserIdentifiers.EntityIdentifiers())
		}
		return nil, err
	}
	if !usr.UpdatedAt.IsZero() && usr.UpdatedAt != userModel.UpdatedAt {
		return nil, errConcurrentWrite
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	// NOTE: fromPB replaces these, it should not mutate
	columns := userModel.fromPB(usr, fieldMask)
	if len(columns) > 0 {
		query = s.db.Model(&userModel).Select(columns).Updates(&userModel)
		if query.Error != nil {
			return nil, query.Error
		}
	}
	updated = new(ttnpb.User)
	userModel.toPB(updated, nil)
	return updated, nil
}

func (s *userStore) DeleteUser(ctx context.Context, id *ttnpb.UserIdentifiers) (err error) {
	defer func() {
		if err != nil && gorm.IsRecordNotFoundError(err) {
			err = errNotFoundForID(id.EntityIdentifiers())
		}
	}()
	query := s.db.Scopes(withContext(ctx), withUserID(id.GetUserID()))
	query = query.Select("users.id")
	var userModel User
	err = query.First(&userModel).Error
	if err != nil {
		return err
	}
	return s.db.Delete(&userModel).Error
}
