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

package oauth

import (
	"time"

	"github.com/RangelReale/osin"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var clientSpecializer = func(base ttnpb.Client) store.Client {
	return &base
}

// storage implements osin.Storage.
type storage struct {
	*store.Store
}

// UserData is the userdata that gets carried around with authorization requests.
type UserData struct {
	UserID string
}

// getUserID returns the UserID of the input if it a ptr to an UserData, otherwise empty string.
func getUserID(data interface{}) string {
	userID := ""
	udata, ok := data.(*UserData)
	if ok && udata != nil {
		userID = udata.UserID
	}
	return userID
}

// Clone the store if needed.
func (s *storage) Clone() osin.Storage {
	return s
}

// Close the store, releasing resources it might be holding.
func (s *storage) Close() {}

// GetClient loads the OAuth Client by client_id.
func (s *storage) GetClient(clientID string) (osin.Client, error) {
	client, err := s.Clients.GetByID(ttnpb.ClientIdentifiers{ClientID: clientID}, clientSpecializer)
	if err != nil {
		return nil, err
	}

	if client.GetClient().State != ttnpb.STATE_APPROVED {
		return nil, nil
	}

	return client, nil
}

// SaveAuthorize saves authorization data.
func (s *storage) SaveAuthorize(data *osin.AuthorizeData) error {
	return s.OAuth.SaveAuthorizationCode(store.AuthorizationData{
		AuthorizationCode: data.Code,
		ClientID:          data.Client.GetId(),
		CreatedAt:         data.CreatedAt,
		ExpiresIn:         time.Duration(data.ExpiresIn) * time.Second,
		Scope:             data.Scope,
		RedirectURI:       data.RedirectUri,
		State:             data.State,
		UserID:            getUserID(data.UserData),
	})
}

// LoadAuthorize loads the client and authorization data for the authorization code.
func (s *storage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	data, err := s.OAuth.GetAuthorizationCode(code)
	if err != nil {
		return nil, err
	}

	// ensure the expiration
	err = data.IsExpired()
	if err != nil {
		return nil, err
	}

	client, err := s.Clients.GetByID(ttnpb.ClientIdentifiers{ClientID: data.ClientID}, clientSpecializer)
	if err != nil {
		return nil, err
	}

	return &osin.AuthorizeData{
		Code:        data.AuthorizationCode,
		Client:      client,
		ExpiresIn:   int32(data.ExpiresIn.Seconds()),
		Scope:       data.Scope,
		RedirectUri: data.RedirectURI,
		State:       data.State,
		CreatedAt:   data.CreatedAt,
		UserData: &UserData{
			UserID: data.UserID,
		},
	}, nil
}

// RemoveAuthorize deletes the authorization code.
func (s *storage) RemoveAuthorize(code string) error {
	return s.OAuth.DeleteAuthorizationCode(code)
}

// SaveAccess saves the access data for later use.
func (s *storage) SaveAccess(data *osin.AccessData) error {
	err := s.Transact(func(s *store.Store) error {
		err := s.OAuth.SaveAccessToken(store.AccessData{
			AccessToken: data.AccessToken,
			ClientID:    data.Client.GetId(),
			UserID:      getUserID(data.UserData),
			Scope:       data.Scope,
			CreatedAt:   data.CreatedAt.Add(time.Second),
			RedirectURI: data.RedirectUri,
			ExpiresIn:   time.Duration(data.ExpiresIn) * time.Second,
		})
		if err != nil {
			return err
		}

		if data.RefreshToken == "" {
			return nil
		}

		return s.OAuth.SaveRefreshToken(store.RefreshData{
			RefreshToken: data.RefreshToken,
			ClientID:     data.Client.GetId(),
			UserID:       getUserID(data.UserData),
			Scope:        data.Scope,
			CreatedAt:    data.CreatedAt,
			RedirectURI:  data.RedirectUri,
		})
	})

	return err
}

// LoadAccess loads the access data based on the access token.
func (s *storage) LoadAccess(accessToken string) (*osin.AccessData, error) {
	data, err := s.OAuth.GetAccessToken(accessToken)
	if err != nil {
		return nil, err
	}

	// ensure the expiration
	err = data.IsExpired()
	if err != nil {
		return nil, err
	}

	client, err := s.Clients.GetByID(ttnpb.ClientIdentifiers{ClientID: data.ClientID}, clientSpecializer)
	if err != nil {
		return nil, err
	}

	return &osin.AccessData{
		AccessToken: data.AccessToken,
		Client:      client,
		ExpiresIn:   int32(data.ExpiresIn.Seconds()),
		Scope:       data.Scope,
		RedirectUri: data.RedirectURI,
		CreatedAt:   data.CreatedAt,
		UserData: &UserData{
			UserID: data.UserID,
		},
	}, nil
}

// RemoveAccess revokes access data.
func (s *storage) RemoveAccess(accessToken string) error {
	return s.OAuth.DeleteAccessToken(accessToken)
}

// LoadRefresh loads the access data based on the refresh token.
func (s *storage) LoadRefresh(refreshToken string) (*osin.AccessData, error) {
	data, err := s.OAuth.GetRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	client, err := s.Clients.GetByID(ttnpb.ClientIdentifiers{ClientID: data.ClientID}, clientSpecializer)
	if err != nil {
		return nil, err
	}

	return &osin.AccessData{
		RefreshToken: data.RefreshToken,
		Client:       client,
		Scope:        data.Scope,
		CreatedAt:    data.CreatedAt,
		RedirectUri:  data.RedirectURI,
		UserData: &UserData{
			UserID: data.UserID,
		},
	}, nil
}

// RemoveRefresh deletes the refresh token.
func (s *storage) RemoveRefresh(refreshToken string) error {
	return s.OAuth.DeleteRefreshToken(refreshToken)
}
