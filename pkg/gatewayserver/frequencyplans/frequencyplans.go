// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package frequencyplans contains abstractions to fetch and manipulate frequency plans
package frequencyplans

// RetrieveHTTPStore returns a new Store of frequency plans, based on the options given, fetched from a HTTP server.
//
// By default, RetrieveHTTPStore fetches its frequency plans on GitHub in the TheThingsNetwork/gateway-conf repository, in the yaml-master branch. It is possible to specify another HTTP root through options.
func RetrieveHTTPStore(options ...RetrieveHTTPStoreOption) (Store, error) {
	baseURI := DefaultBaseURL

	for _, option := range options {
		option(&baseURI)
	}

	store, err := fetchStore(baseURI)
	return &store, err
}
