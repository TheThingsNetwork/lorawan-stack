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

// ClientAuthorization model. Is also embedded by other OAuth models.
type ClientAuthorization struct {
	Model

	Client   *Client
	ClientID string `gorm:"type:UUID;index;not null"`

	User   *User
	UserID string `gorm:"type:UUID;index;not null"`

	Rights Rights `gorm:"type:INT ARRAY"`
}

func (a ClientAuthorization) toPB() *ttnpb.OAuthClientAuthorization {
	pb := &ttnpb.OAuthClientAuthorization{
		Rights:    a.Rights,
		CreatedAt: ttnpb.ProtoTimePtr(cleanTime(a.CreatedAt)),
		UpdatedAt: ttnpb.ProtoTimePtr(cleanTime(a.UpdatedAt)),
	}
	if a.Client != nil {
		pb.ClientIds = &ttnpb.ClientIdentifiers{ClientId: a.Client.ClientID}
	}
	if a.User != nil {
		pb.UserIds = &ttnpb.UserIdentifiers{UserId: a.User.Account.UID}
	}
	return pb
}

// AuthorizationCode model.
type AuthorizationCode struct {
	Model

	Client   *Client
	ClientID string `gorm:"type:UUID;index;not null"`

	User   *User
	UserID string `gorm:"type:UUID;index;not null"`

	UserSessionID *string `gorm:"type:UUID;index"`

	Rights Rights `gorm:"type:INT ARRAY"`

	Code        string `gorm:"type:VARCHAR;unique_index:authorization_code_code_index;not null"`
	RedirectURI string `gorm:"type:VARCHAR;column:redirect_uri"`
	State       string `gorm:"type:VARCHAR"`
	ExpiresAt   *time.Time
}

func (a AuthorizationCode) toPB() *ttnpb.OAuthAuthorizationCode {
	pb := &ttnpb.OAuthAuthorizationCode{
		Rights:      a.Rights,
		Code:        a.Code,
		RedirectUri: a.RedirectURI,
		State:       a.State,
		CreatedAt:   ttnpb.ProtoTimePtr(cleanTime(a.CreatedAt)),
		ExpiresAt:   ttnpb.ProtoTime(cleanTimePtr(a.ExpiresAt)),
	}
	if a.Client != nil {
		pb.ClientIds = &ttnpb.ClientIdentifiers{ClientId: a.Client.ClientID}
	}
	if a.User != nil {
		pb.UserIds = &ttnpb.UserIdentifiers{UserId: a.User.Account.UID}
	}
	if a.UserSessionID != nil {
		pb.UserSessionId = *a.UserSessionID
	}
	return pb
}

// AccessToken model.
type AccessToken struct {
	Model

	Client   *Client
	ClientID string `gorm:"type:UUID;index;not null"`

	User   *User
	UserID string `gorm:"type:UUID;index;not null"`

	UserSessionID *string `gorm:"type:UUID;index"`

	Rights Rights `gorm:"type:INT ARRAY"`

	TokenID string `gorm:"type:VARCHAR;unique_index:access_token_id_index;not null"`

	Previous   *AccessToken `gorm:"foreignkey:PreviousID;association_foreignkey:TokenID"`
	PreviousID string       `gorm:"type:VARCHAR;index:access_token_previous_index"`

	AccessToken  string `gorm:"type:VARCHAR;not null"`
	RefreshToken string `gorm:"type:VARCHAR;not null"`

	ExpiresAt *time.Time
}

func (a AccessToken) toPB() *ttnpb.OAuthAccessToken {
	pb := &ttnpb.OAuthAccessToken{
		Rights:       a.Rights,
		Id:           a.TokenID,
		AccessToken:  a.AccessToken,
		RefreshToken: a.RefreshToken,
		CreatedAt:    ttnpb.ProtoTimePtr(cleanTime(a.CreatedAt)),
		ExpiresAt:    ttnpb.ProtoTime(cleanTimePtr(a.ExpiresAt)),
	}
	if a.Client != nil {
		pb.ClientIds.ClientId = a.Client.ClientID
	}
	if a.User != nil {
		pb.UserIds = &ttnpb.UserIdentifiers{UserId: a.User.Account.UID}
	}
	if a.UserSessionID != nil {
		pb.UserSessionId = *a.UserSessionID
	}
	return pb
}

func init() {
	registerModel(
		&ClientAuthorization{},
		&AuthorizationCode{},
		&AccessToken{},
	)
}
