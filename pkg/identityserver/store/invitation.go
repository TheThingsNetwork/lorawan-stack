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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Invitation model.
type Invitation struct {
	Model

	Email     string `gorm:"type:VARCHAR;unique_index:invitation_email_index;not null"`
	Token     string `gorm:"type:VARCHAR;unique_index:invitation_token_index;not null"`
	ExpiresAt time.Time

	AcceptedBy   *User
	AcceptedByID *string `gorm:"type:UUID"`
	AcceptedAt   *time.Time
}

func init() {
	registerModel(&Invitation{})
}

func (i Invitation) toPB() *ttnpb.Invitation {
	pb := &ttnpb.Invitation{
		Email:     i.Email,
		Token:     i.Token,
		ExpiresAt: cleanTime(i.ExpiresAt),
		CreatedAt: cleanTime(i.CreatedAt),
		UpdatedAt: cleanTime(i.UpdatedAt),
	}
	return pb
}
