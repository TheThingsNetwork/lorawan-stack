// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

// LoginToken model.
type LoginToken struct {
	Model

	User   *User
	UserID string `gorm:"type:UUID"`

	Token     string `gorm:"type:VARCHAR;unique_index:login_token_index;not null"`
	ExpiresAt time.Time
	Used      bool
}

func init() {
	registerModel(&LoginToken{})
}

func (t LoginToken) toPB() *ttnpb.LoginToken {
	pb := &ttnpb.LoginToken{
		Token:     t.Token,
		ExpiresAt: cleanTime(t.ExpiresAt),
		CreatedAt: cleanTime(t.CreatedAt),
		UpdatedAt: cleanTime(t.UpdatedAt),
	}
	if t.User != nil {
		pb.UserIdentifiers = ttnpb.UserIdentifiers{UserId: t.User.Account.UID}
	}
	return pb
}
