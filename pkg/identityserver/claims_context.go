// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/auth/oauth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

func (is *IdentityServer) claimsFromContext(ctx context.Context) (*claims, error) {
	md := rpcmetadata.FromIncomingContext(ctx)

	if md.AuthType == "" && md.AuthValue == "" {
		return new(claims), nil
	}

	if md.AuthType != "Bearer" {
		return nil, errors.Errorf("Expected authentication type to be Bearer but got `%s", md.AuthType)
	}

	header, payload, err := auth.DecodeTokenOrKey(md.AuthValue)
	if err != nil {
		return nil, err
	}

	var res *claims
	switch header.Type {
	case auth.Token:
		data, err := is.store.OAuth.GetAccessToken(md.AuthValue)
		if err != nil {
			return nil, err
		}

		err = data.IsExpired()
		if err != nil {
			return nil, err
		}

		rights, err := oauth.ParseScope(data.Scope)
		if err != nil {
			return nil, err
		}

		res = &claims{
			EntityID:   data.UserID,
			EntityType: entityUser,
			Source:     auth.Token,
			Rights:     rights,
		}
	case auth.Key:
		var entityID string
		var key *ttnpb.APIKey
		var err error

		res := &claims{
			Source: auth.Key,
		}

		switch payload.Type {
		case auth.ApplicationKey:
			entityID, key, err = is.store.Applications.GetAPIKey(md.AuthValue)

			res.EntityType = entityApplication
		case auth.GatewayKey:
			entityID, key, err = is.store.Gateways.GetAPIKey(md.AuthValue)

			res.EntityType = entityApplication
		case auth.UserKey:
			entityID, key, err = is.store.Users.GetAPIKey(md.AuthValue)

			res.EntityType = entityApplication
		default:
			return nil, errors.Errorf("Invalid API key type `%s`", payload.Type)
		}

		if err != nil {
			return nil, err
		}

		res.EntityID = entityID
		res.Rights = key.Rights
	default:
		return nil, errors.New("Invalid authentication value")
	}

	return res, nil
}
