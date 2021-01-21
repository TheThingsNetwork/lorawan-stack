// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package devicerepository

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store/bleve"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// Config represents the DeviceRepository configuration.
type Config struct {
	Store StoreConfig `name:"store"`

	Source    string `name:"source" description:"Device Repository Source (directory)"`
	Directory string `name:"directory" description:"OS filesystem directory, which contains the Device Repository"`

	AssetsBaseURL string `name:"assets-base-url" description:"The base URL for Device Repository assets"`
}

// StoreConfig represents configuration for the Device Repository store.
type StoreConfig struct {
	Store store.Store `name:"-"`

	Bleve bleve.Config `name:"bleve"`
}

// NewStore creates a new Store for end devices.
func (c Config) NewStore(ctx context.Context, blobConf config.BlobConfig) (store.Store, error) {
	if c.Store.Store != nil {
		return c.Store.Store, nil
	}

	return c.Store.Bleve.NewStore(ctx)
}

var errUnknownSource = errors.DefineInvalidArgument("unknown_source", "unknown source `{source}`")

// Initialize sets up the Device Repository.
func (c Config) Initialize(ctx context.Context, blobConf config.BlobConfig, overwrite bool) error {
	switch c.Source {
	case "directory":
	default:
		return errUnknownSource.WithAttributes("source", c.Source)
	}

	return c.Store.Bleve.Initialize(ctx, c.Directory, overwrite)
}
