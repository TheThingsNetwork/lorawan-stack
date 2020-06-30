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

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// UserSession is the session of a logged in user.
type UserSession struct {
	Model

	User          *User
	UserID        string `gorm:"type:UUID;index:user_session_user_index;not null"`
	SessionSecret string `gorm:"type:VARCHAR;not null"`

	ExpiresAt *time.Time
}

func init() {
	registerModel(&UserSession{})
}

func (sess UserSession) toPB(pb *ttnpb.UserSession) {
	pb.SessionID = sess.ID
	pb.SessionSecret = sess.SessionSecret
	pb.CreatedAt = cleanTime(sess.CreatedAt)
	pb.UpdatedAt = cleanTime(sess.UpdatedAt)
	pb.ExpiresAt = cleanTimePtr(sess.ExpiresAt)
	if sess.User != nil {
		pb.UserID = sess.User.Account.UID
	}
}

func (sess *UserSession) fromPB(pb *ttnpb.UserSession) []string {
	sess.ExpiresAt = cleanTimePtr(pb.ExpiresAt)
	return []string{"expires_at"}
}
