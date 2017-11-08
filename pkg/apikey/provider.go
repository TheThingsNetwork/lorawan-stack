// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Provider is anything that can provide an API key information.
type Provider interface {
	// GetAPIKey online validates the provided API key by returning the
	// entity ID it is associated to, plus its rights and name.
	GetAPIKey(key string) (string, *ttnpb.APIKey, error)
}
