// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
