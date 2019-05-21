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
		Rights:    a.Rights.Rights,
		CreatedAt: cleanTime(a.CreatedAt),
		UpdatedAt: cleanTime(a.UpdatedAt),
	}
	if a.Client != nil {
		pb.ClientIDs.ClientID = a.Client.ClientID
	}
	if a.User != nil {
		pb.UserIDs.UserID = a.User.Account.UID
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

	Rights Rights `gorm:"type:INT ARRAY"`

	Code        string `gorm:"type:VARCHAR;unique_index:authorization_code_code_index;not null"`
	RedirectURI string `gorm:"type:VARCHAR;column:redirect_uri"`
	State       string `gorm:"type:VARCHAR"`
	ExpiresAt   time.Time
}

func (a AuthorizationCode) toPB() *ttnpb.OAuthAuthorizationCode {
	pb := &ttnpb.OAuthAuthorizationCode{
		Rights:      a.Rights.Rights,
		Code:        a.Code,
		RedirectURI: a.RedirectURI,
		State:       a.State,
		CreatedAt:   cleanTime(a.CreatedAt),
		ExpiresAt:   cleanTime(a.ExpiresAt),
	}
	if a.Client != nil {
		pb.ClientIDs.ClientID = a.Client.ClientID
	}
	if a.User != nil {
		pb.UserIDs.UserID = a.User.Account.UID
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

	Rights Rights `gorm:"type:INT ARRAY"`

	TokenID string `gorm:"type:VARCHAR;unique_index:access_token_id_index;not null"`

	Previous   *AccessToken `gorm:"foreignkey:PreviousID;association_foreignkey:TokenID"`
	PreviousID string       `gorm:"type:VARCHAR;index:access_token_previous_index"`

	AccessToken  string `gorm:"type:VARCHAR;not null"`
	RefreshToken string `gorm:"type:VARCHAR;not null"`

	ExpiresAt time.Time
}

func (a AccessToken) toPB() *ttnpb.OAuthAccessToken {
	pb := &ttnpb.OAuthAccessToken{
		Rights:       a.Rights.Rights,
		ID:           a.TokenID,
		AccessToken:  a.AccessToken,
		RefreshToken: a.RefreshToken,
		CreatedAt:    cleanTime(a.CreatedAt),
		ExpiresAt:    cleanTime(a.ExpiresAt),
	}
	if a.Client != nil {
		pb.ClientIDs.ClientID = a.Client.ClientID
	}
	if a.User != nil {
		pb.UserIDs.UserID = a.User.Account.UID
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
