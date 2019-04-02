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
	"fmt"
	"reflect"
	"runtime/trace"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/warning"
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
func selectUserFields(ctx context.Context, query *gorm.DB, fieldMask *types.FieldMask) *gorm.DB {
	query = query.Preload("Account")
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query.Preload("Attributes").Preload("ProfilePicture")
	}
	var userColumns []string
	var notFoundPaths []string
	for _, column := range modelColumns {
		userColumns = append(userColumns, "users."+column)
	}
	for _, path := range ttnpb.TopLevelFields(fieldMask.Paths) {
		switch path {
		case "ids", "created_at", "updated_at":
			// always selected
		case attributesField:
			query = query.Preload("Attributes")
		case profilePictureField:
			userColumns = append(userColumns, "profile_picture_id")
			query = query.Preload("ProfilePicture")
		default:
			if columns, ok := userColumnNames[path]; ok {
				userColumns = append(userColumns, columns...)
			} else {
				notFoundPaths = append(notFoundPaths, path)
			}
		}
	}
	if len(notFoundPaths) > 0 {
		warning.Add(ctx, fmt.Sprintf("unsupported field mask paths: %s", strings.Join(notFoundPaths, ", ")))
	}
	return query.Select(userColumns)
}

func (s *userStore) CreateUser(ctx context.Context, usr *ttnpb.User) (*ttnpb.User, error) {
	defer trace.StartRegion(ctx, "create user").End()
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
	defer trace.StartRegion(ctx, "find users").End()
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetUserID()
	}
	query := s.db.Scopes(withContext(ctx), withUserID(idStrings...))
	query = selectUserFields(ctx, query, fieldMask)
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(User{}))
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
		userProto := &ttnpb.User{}
		userModel.toPB(userProto, fieldMask)
		userProtos[i] = userProto
	}
	return userProtos, nil
}

func (s *userStore) GetUser(ctx context.Context, id *ttnpb.UserIdentifiers, fieldMask *types.FieldMask) (*ttnpb.User, error) {
	defer trace.StartRegion(ctx, "get user").End()
	query := s.db.Scopes(withContext(ctx), withUserID(id.GetUserID()))
	query = selectUserFields(ctx, query, fieldMask)
	var userModel User
	if err := query.Preload("Account").First(&userModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id.EntityIdentifiers())
		}
		return nil, err
	}
	userProto := &ttnpb.User{}
	userModel.toPB(userProto, fieldMask)
	return userProto, nil
}

func (s *userStore) UpdateUser(ctx context.Context, usr *ttnpb.User, fieldMask *types.FieldMask) (updated *ttnpb.User, err error) {
	defer trace.StartRegion(ctx, "update user").End()
	query := s.db.Scopes(withContext(ctx), withUserID(usr.GetUserID()))
	query = selectUserFields(ctx, query, fieldMask)
	var userModel User
	if err = query.First(&userModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(usr.UserIdentifiers.EntityIdentifiers())
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	oldAttributes, oldProfilePicture := userModel.Attributes, userModel.ProfilePicture
	columns := userModel.fromPB(usr, fieldMask)
	newProfilePicture := userModel.ProfilePicture
	if newProfilePicture != oldProfilePicture {
		if oldProfilePicture != nil {
			if err = s.db.Delete(oldProfilePicture).Error; err != nil {
				return nil, err
			}
		}
		if newProfilePicture != nil {
			if err = s.db.Create(newProfilePicture).Error; err != nil {
				return nil, err
			}
			userModel.ProfilePictureID, userModel.ProfilePicture = &newProfilePicture.ID, nil
			columns = append(columns, "profile_picture_id")
		}
	}
	if len(columns) > 0 {
		query = s.db.Select(append(columns, "updated_at")).Save(&userModel)
		if query.Error != nil {
			return nil, query.Error
		}
	}
	if !reflect.DeepEqual(oldAttributes, userModel.Attributes) {
		if err = replaceAttributes(ctx, s.db, "user", userModel.ID, oldAttributes, userModel.Attributes); err != nil {
			return nil, err
		}
	}
	userModel.ProfilePicture = newProfilePicture
	updated = &ttnpb.User{}
	userModel.toPB(updated, fieldMask)
	return updated, nil
}

func (s *userStore) DeleteUser(ctx context.Context, id *ttnpb.UserIdentifiers) (err error) {
	defer trace.StartRegion(ctx, "delete user").End()
	defer func() {
		if err != nil && gorm.IsRecordNotFoundError(err) {
			err = errNotFoundForID(id.EntityIdentifiers())
		}
	}()
	query := s.db.Scopes(withContext(ctx), withUserID(id.GetUserID()))
	query = query.Select("users.id")
	var userModel User
	if err = query.First(&userModel).Error; err != nil {
		return err
	}
	return s.db.Delete(&userModel).Error
}
