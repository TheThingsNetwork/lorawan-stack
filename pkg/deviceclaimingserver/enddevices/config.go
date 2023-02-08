// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package enddevices

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// JSClientConfigurationName is the filename of Join Server client configuration.
const JSClientConfigurationName = "config.yml"

// NetworkServer contains information related to the Network Server.
// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/6048)
type NetworkServer struct {
	HomeNSID *types.EUI64 `name:"home-ns-id" description:"DEPRECATED"`
	Hostname string       `name:"hostname" description:"DEPRECATED"`
}

// Config contains options for end device claiming clients.
//
//nolint:lll
type Config struct {
	NetID types.NetID  `name:"net-id" description:"NetID of the Network Server to configure when claiming"`
	NSID  *types.EUI64 `name:"ns-id" description:"NSID of the Network Server to configure when claiming"`
	ASID  string       `name:"as-id" description:"AS-ID of the Application Server to configure when claiming"`

	// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/6048)
	NetworkServer NetworkServer `name:"network-server" description:"DEPRECATED"`

	Source    string                `name:"source" description:"Source of the file containing Join Server settings (directory, url, blob)"`
	Directory string                `name:"directory" description:"OS filesystem directory, which contains the config.yml and the client-specific files"`
	URL       string                `name:"url" description:"URL, which contains Join Server client configuration"`
	Blob      config.BlobPathConfig `name:"blob"`
}

// Fetcher returns a fetch.Interface based on the configuration.
// If no configuration source is set, this method returns nil, nil.
func (c Config) Fetcher(
	ctx context.Context, blobConf config.BlobConfig, httpClientProvider httpclient.Provider,
) (fetch.Interface, error) {
	switch c.Source {
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		httpClient, err := httpClientProvider.HTTPClient(ctx, httpclient.WithCache(true))
		if err != nil {
			return nil, err
		}
		return fetch.FromHTTP(httpClient, c.URL)
	case "blob":
		b, err := blobConf.Bucket(ctx, c.Blob.Bucket, httpClientProvider)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

type baseConfig struct {
	JoinServers []struct {
		File     string              `yaml:"file"`
		JoinEUIs []types.EUI64Prefix `yaml:"join-euis"`
		Type     string              `yaml:"type"`
	} `yaml:"join-servers"`
}
