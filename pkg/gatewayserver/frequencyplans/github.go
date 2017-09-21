// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	yaml "gopkg.in/yaml.v2"
)

func fetchHTTPContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.NewWithCause("HTTP request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Error with the HTTP exchange: expected successful status code, received %s status", resp.Status)
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewWithCause("Copying the HTTP response to a local buffer failed", err)
	}

	return buffer, nil
}

func fetchList(baseURI storeFetchingConfiguration) ([]frequencyPlanDescription, error) {
	list := make([]frequencyPlanDescription, 0)

	url := fmt.Sprintf("%s/%s", string(baseURI), "frequency-plans.yml")

	buffer, err := fetchHTTPContent(url)
	if err != nil {
		return nil, errors.NewWithCause("Fetching content failed", err)
	}

	err = yaml.Unmarshal(buffer, &list)
	if err != nil {
		return nil, errors.NewWithCause("Failed to parse the HTTP content as a list of frequency plans", err)
	}

	return list, nil
}

func fetchFrequencyPlan(baseURI storeFetchingConfiguration, filename string) (ttnpb.FrequencyPlan, error) {
	frequencyPlan := ttnpb.FrequencyPlan{}

	url := fmt.Sprintf("%s/%s", string(baseURI), filename)

	buffer, err := fetchHTTPContent(url)
	if err != nil {
		return frequencyPlan, errors.NewWithCause("Fetching content failed", err)
	}

	err = yaml.Unmarshal(buffer, &frequencyPlan)
	if err != nil {
		return frequencyPlan, errors.NewWithCause("Failed to parse the HTTP content as a frequency plan", err)
	}

	return frequencyPlan, nil
}

func fetchFrequencyPlans(config storeFetchingConfiguration) (store, error) {
	frequencyPlansInfo, err := fetchList(config)
	if err != nil {
		return nil, errors.NewWithCause("Fetching list of frequency plans failed", err)
	}

	frequencyPlansStorage := make(store, 0)
	for _, description := range frequencyPlansInfo {
		frequencyPlanContent, err := fetchFrequencyPlan(config, description.Filename)
		if err != nil {
			return nil, errors.NewWithCause(fmt.Sprintf("Failed to retrieve %s frequency plan content", description.ID), err)
		}

		frequencyPlansStorage[description.ID] = frequencyPlanContent
	}

	return frequencyPlansStorage, nil
}

// RetrieveHTTPStore returns a new Store of frequency plans, based on the options given, fetched from a HTTP server.
//
// By default, RetrieveHTTPStore fetches its frequency plans on GitHub in the TheThingsNetwork/gateway-conf repository, in the yaml-master branch. It is possible to specify another HTTP root through options.
func RetrieveHTTPStore(options ...RetrieveHTTPStoreOption) (Store, error) {
	baseURI := DefaultBaseURL

	for _, option := range options {
		option(&baseURI)
	}

	store, err := fetchFrequencyPlans(baseURI)
	return &store, err
}
