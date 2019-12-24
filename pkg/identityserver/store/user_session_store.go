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

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetUserSessionStore returns an UserSessionStore on the given db (or transaction).
func GetUserSessionStore(db *gorm.DB) UserSessionStore {
	return &userSessionStore{store: newStore(db)}
}

type userSessionStore struct {
	*store
}

func (s *userSessionStore) CreateSession(ctx context.Context, sess *ttnpb.UserSession) (*ttnpb.UserSession, error) {
	defer trace.StartRegion(ctx, "create user session").End()
	user, err := s.findEntity(ctx, sess.UserIdentifiers, "id")
	if err != nil {
		return nil, err
	}
	sessionModel := UserSession{
		UserID:    user.PrimaryKey(),
		ExpiresAt: cleanTimePtr(sess.ExpiresAt),
	}
	if err = s.createEntity(ctx, &sessionModel); err != nil {
		return nil, err
	}
	sessionProto := *sess
	sessionModel.toPB(&sessionProto)
	return &sessionProto, nil
}

func (s *userSessionStore) FindSessions(ctx context.Context, userIDs *ttnpb.UserIdentifiers) ([]*ttnpb.UserSession, error) {
	defer trace.StartRegion(ctx, "find user sessions").End()
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	query := s.query(ctx, UserSession{}).Where(UserSession{UserID: user.PrimaryKey()})
	query = query.Order(orderFromContext(ctx, "user_sessions", "id", "ASC"))
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(UserSession{}))
		query = query.Limit(limit).Offset(offset)
	}
	var sessionModels []UserSession
	query = query.Find(&sessionModels)
	setTotal(ctx, uint64(len(sessionModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	sessionProtos := make([]*ttnpb.UserSession, len(sessionModels))
	for i, sessionModel := range sessionModels {
		sessionProto := &ttnpb.UserSession{}
		sessionProto.UserID = userIDs.UserID
		sessionModel.toPB(sessionProto)
		sessionProtos[i] = sessionProto
	}
	return sessionProtos, nil
}

func (s *userSessionStore) GetSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) (*ttnpb.UserSession, error) {
	defer trace.StartRegion(ctx, "get user session").End()
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	query := s.query(ctx, UserSession{}).Where(UserSession{Model: Model{ID: sessionID}, UserID: user.PrimaryKey()})
	var sessionModel UserSession
	if err = query.Find(&sessionModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errSessionNotFound.WithAttributes("user_id", userIDs.UserID, "session_id", sessionID)
		}
		return nil, err
	}
	sessionProto := &ttnpb.UserSession{}
	sessionProto.UserID = userIDs.UserID
	sessionModel.toPB(sessionProto)
	return sessionProto, nil
}

func (s *userSessionStore) UpdateSession(ctx context.Context, sess *ttnpb.UserSession) (*ttnpb.UserSession, error) {
	defer trace.StartRegion(ctx, "update user session").End()
	user, err := s.findEntity(ctx, sess.UserIdentifiers, "id")
	if err != nil {
		return nil, err
	}
	query := s.query(ctx, UserSession{}).Where(UserSession{Model: Model{ID: sess.SessionID}, UserID: user.PrimaryKey()})
	var sessionModel UserSession
	if err = query.Find(&sessionModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errSessionNotFound.WithAttributes("user_id", sess.UserIdentifiers.UserID, "session_id", sess.SessionID)
		}
		return nil, err
	}
	columns := sessionModel.fromPB(sess)
	if err = s.updateEntity(ctx, &sessionModel, columns...); err != nil {
		return nil, err
	}
	updated := &ttnpb.UserSession{}
	sessionModel.toPB(updated)
	return updated, nil
}

func (s *userSessionStore) DeleteSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) error {
	defer trace.StartRegion(ctx, "delete user session").End()
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return err
	}
	query := s.query(ctx, UserSession{}).Where(UserSession{Model: Model{ID: sessionID}, UserID: user.PrimaryKey()})
	return query.Delete(&UserSession{}).Error
}
