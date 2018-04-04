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

package gatewayserver

import (
	"github.com/TheThingsNetwork/ttn/pkg/fetch"
	"github.com/TheThingsNetwork/ttn/pkg/frequencyplans"
)

// Config represents the GatewayServer configuration.
type Config struct {
	FileFrequencyPlansStore string `name:"frequency-plans-dir" description:"Directory where the frequency plans are stored"`
	HTTPFrequencyPlansStore string `name:"frequency-plans-uri" description:"URI from where the frequency plans will be fetched, if no directory is specified"`

	NSTags []string `name:"network-servers.tags" description:"Network server tags to accept to connect to"`
}

func (conf Config) store() frequencyplans.Store {
	store := frequencyplans.NewStore(fetch.FromGitHubRepository("TheThingsNetwork/gateway-conf", "yaml-master", "", true))
	if conf.FileFrequencyPlansStore != "" {
		store.Fetcher = fetch.FromFilesystem(conf.FileFrequencyPlansStore)
	} else if conf.HTTPFrequencyPlansStore != "" {
		store.Fetcher = fetch.FromHTTP(conf.HTTPFrequencyPlansStore, true)
	}

	return store
}
