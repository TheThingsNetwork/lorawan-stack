// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver

import "github.com/TheThingsNetwork/ttn/pkg/gatewayserver/frequencyplans"

// Config represents the GatewayServer configuration.
type Config struct {
	LocalFrequencyPlansStore    string `name:"frequency-plans-dir" description:"Directory where the frequency plans are stored"`
	HTTPFrequencyPlansStoreRoot string `name:"frequency-plans-uri" description:"URI from where the frequency plans will be fetched, if no directory is specified"`

	NSTags []string `name:"network-servers.tags" description:"Network server tags to accept to connect to"`
}

func (conf Config) store() (fpStore frequencyplans.Store, err error) {
	defer func() {
		if err == nil {
			fpStore = frequencyplans.Cache(fpStore, frequencyPlansCacheDuration)
		}
	}()

	if conf.LocalFrequencyPlansStore != "" {
		fpStore, err = frequencyplans.ReadFileSystemStore(frequencyplans.FileSystemRootPathOption(conf.LocalFrequencyPlansStore))
		return
	}

	fpStore, err = frequencyplans.RetrieveHTTPStore(frequencyplans.BaseURIOption(conf.HTTPFrequencyPlansStoreRoot))
	return
}
