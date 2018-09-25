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

package applicationserver

import (
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/pkg/fetch"
)

// LinkMode defines how applications are linked to their Network Server.
type LinkMode int

const (
	// LinkAll links all applications in the link registry to their Network Server automatically.
	LinkAll LinkMode = iota
	// LinkExplicit links applications on request.
	LinkExplicit
)

// Config represents the ApplicationServer configuration.
type Config struct {
	LinkMode         LinkMode
	Devices          DeviceRegistry
	Links            LinkRegistry
	KeyVault         crypto.KeyVault
	DeviceRepository DeviceRepositoryConfig
}

// DeviceRepositoryConfig defines the source of the device repository.
type DeviceRepositoryConfig struct {
	Static    map[string][]byte `name:"-"`
	Directory string            `name:"directory" description:"Retrieve the device repository from the filesystem"`
	URL       string            `name:"url" description:"Retrieve the device repository from a web server"`
}

// Client instantiates a new devicerepository.Client with a fetcher based on the configuration.
// The order of precedence is Static, Directory and URL.
// If neither Static, Directory nor a URL is set, this function returns nil.
func (c DeviceRepositoryConfig) Client() *devicerepository.Client {
	var fetcher fetch.Interface
	switch {
	case c.Static != nil:
		fetcher = fetch.NewMemFetcher(c.Static)
	case c.Directory != "":
		fetcher = fetch.FromFilesystem(c.Directory)
	case c.URL != "":
		fetcher = fetch.FromHTTP(c.URL, true)
	default:
		return nil
	}
	return &devicerepository.Client{
		Fetcher: fetcher,
	}
}
