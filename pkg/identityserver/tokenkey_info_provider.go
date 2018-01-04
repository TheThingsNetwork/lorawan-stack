// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/claims"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var _ claims.TokenKeyInfoProvider = new(tokenKeyProvider)

// tokenKeyProvider implements claims.TokenKeyInfoProvider interface.
type tokenKeyProvider struct {
	store *sql.Store
}

// TokenInfo fetches the access token from the OAuth store and returns the access data.
// It returns error if token is expired.
func (t *tokenKeyProvider) TokenInfo(accessToken string) (*types.AccessData, error) {
	data, err := t.store.OAuth.GetAccessToken(accessToken)
	if err != nil {
		return nil, err
	}

	// ensure the expiration
	if err := data.IsExpired(); err != nil {
		return nil, err
	}

	return data, nil
}

func (t *tokenKeyProvider) KeyInfo(key string, typ auth.APIKeyType) (string, *ttnpb.APIKey, error) {
	switch typ {
	case auth.UserKey:
		return t.store.Users.GetAPIKey(key)
	case auth.ApplicationKey:
		return t.store.Applications.GetAPIKey(key)
	case auth.GatewayKey:
		return t.store.Gateways.GetAPIKey(key)
	}

	return "", nil, sql.ErrAPIKeyNotFound.New(nil)
}
