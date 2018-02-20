// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

const (
	// DefaultBaseURL is the default URL where files will be fetched
	DefaultBaseURL storeFetchingConfiguration = "https://raw.githubusercontent.com/TheThingsNetwork/gateway-conf/yaml-master"
)

type storeFetchingConfiguration string

// RetrieveHTTPStoreOption is an option applied when creating the store.
type RetrieveHTTPStoreOption func(*storeFetchingConfiguration)

// BaseURIOption returns an option allowing to change the base URI to retrieve the frequency plans from. When this option is not used, the URI `https://raw.githubusercontent.com/TheThingsNetwork/gateway-conf/yaml-master` is used.
//
// Frequency plans are then retrieved from the `<base>/frequency-plans.yml` file, then from the `<base>/<file path>` files.
func BaseURIOption(baseURI string) RetrieveHTTPStoreOption {
	return func(config *storeFetchingConfiguration) {
		*config = storeFetchingConfiguration(baseURI)
	}
}
