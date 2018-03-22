// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// SettingStore is the settings store.
type SettingStore interface {
	// Get returns the settings.
	Get() (*ttnpb.IdentityServerSettings, error)

	// Set sets the settings.
	Set(ttnpb.IdentityServerSettings) error
}
