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
	"runtime/trace"

	"github.com/gogo/protobuf/proto"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetUserSessionStore returns an UserSessionStore on the given db (or transaction).
func GetUserSessionStore(db *gorm.DB) store.UserSessionStore {
	return &userSessionStore{baseStore: newStore(db)}
}

type userSessionStore struct {
	*baseStore
}

func (s *userSessionStore) CreateSession(ctx context.Context, sess *ttnpb.UserSession) (*ttnpb.UserSession, error) {
	defer trace.StartRegion(ctx, "create user session").End()
	user, err := s.findEntity(ctx, sess.GetUserIds(), "id")
	if err != nil {
		return nil, err
	}
	sessionModel := UserSession{
		UserID:        user.PrimaryKey(),
		SessionSecret: sess.SessionSecret,
		ExpiresAt:     cleanTimePtr(ttnpb.StdTime(sess.ExpiresAt)),
	}
	if err = s.createEntity(ctx, &sessionModel); err != nil {
		return nil, err
	}
	sessionProto := proto.Clone(sess).(*ttnpb.UserSession)
	sessionModel.toPB(sessionProto)
	return sessionProto, nil
}

func (s *userSessionStore) FindSessions(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers,
) ([]*ttnpb.UserSession, error) {
	defer trace.StartRegion(ctx, "find user sessions").End()
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	query := s.query(ctx, UserSession{}).Where(UserSession{UserID: user.PrimaryKey()})
	query = query.Order(store.OrderFromContext(ctx, "user_sessions", createdAt, "DESC"))
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
		query = query.Limit(limit).Offset(offset)
	}
	var sessionModels []UserSession
	query = query.Find(&sessionModels)
	store.SetTotal(ctx, uint64(len(sessionModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	sessionProtos := make([]*ttnpb.UserSession, len(sessionModels))
	for i, sessionModel := range sessionModels {
		sessionProto := &ttnpb.UserSession{}
		sessionProto.UserIds = userIDs
		sessionModel.toPB(sessionProto)
		sessionProtos[i] = sessionProto
	}
	return sessionProtos, nil
}

func (s *userSessionStore) findSession(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string,
) (*UserSession, error) {
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	query := s.query(ctx, UserSession{}).Where(UserSession{Model: Model{ID: sessionID}, UserID: user.PrimaryKey()})
	var sessionModel UserSession
	if err = query.Find(&sessionModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errSessionNotFound.WithAttributes("user_id", userIDs.GetUserId(), "session_id", sessionID)
		}
		return nil, err
	}
	return &sessionModel, nil
}

func (s *userSessionStore) GetSession(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string,
) (*ttnpb.UserSession, error) {
	defer trace.StartRegion(ctx, "get user session").End()
	sessionModel, err := s.findSession(ctx, userIDs, sessionID)
	if err != nil {
		return nil, err
	}
	sessionProto := &ttnpb.UserSession{}
	sessionProto.UserIds = userIDs
	sessionModel.toPB(sessionProto)
	return sessionProto, nil
}

func (s *userSessionStore) GetSessionByID(ctx context.Context, sessionID string) (*ttnpb.UserSession, error) {
	defer trace.StartRegion(ctx, "get user session by session ID").End()
	query := s.query(ctx, UserSession{}).Where(UserSession{Model: Model{ID: sessionID}})
	var sessionModel UserSession
	if err := query.Find(&sessionModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errSessionNotFound.WithAttributes("session_id", sessionID)
		}
		return nil, err
	}
	query = s.query(ctx, Account{}).Where(Account{
		AccountID:   sessionModel.UserID,
		AccountType: user,
	})
	var accountModel Account
	if err := query.Find(&accountModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errSessionNotFound.WithAttributes("user_id", sessionModel.UserID)
		}
		return nil, err
	}
	sessionProto := &ttnpb.UserSession{}
	sessionProto.UserIds = &ttnpb.UserIdentifiers{UserId: accountModel.UID}
	sessionModel.toPB(sessionProto)
	return sessionProto, nil
}

func (s *userSessionStore) DeleteSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) error {
	defer trace.StartRegion(ctx, "delete user session").End()
	sessionModel, err := s.findSession(ctx, userIDs, sessionID)
	if err != nil {
		return err
	}
	return s.query(ctx, UserSession{}).Delete(sessionModel).Error
}

func (s *userSessionStore) DeleteAllUserSessions(ctx context.Context, userIDs *ttnpb.UserIdentifiers) error {
	defer trace.StartRegion(ctx, "delete all user sessions").End()
	user, err := s.findEntity(store.WithSoftDeleted(ctx, false), userIDs, "id")
	if err != nil {
		return err
	}
	query := s.query(ctx, UserSession{}).Where(UserSession{UserID: user.PrimaryKey()})
	return query.Delete(&UserSession{}).Error
}
